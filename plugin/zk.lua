-- zk.nvim - Zettelkasten CLI integration for NeoVim
-- This file is auto-loaded by NeoVim when the plugin is installed

-- Prevent loading twice
if vim.g.loaded_zk then
    return
end
vim.g.loaded_zk = true

-- Minimum NeoVim version check
if vim.fn.has("nvim-0.9") ~= 1 then
    vim.notify("zk.nvim requires NeoVim 0.9+", vim.log.levels.ERROR)
    return
end

-- User commands
local function create_commands()
    -- Daily notes
    vim.api.nvim_create_user_command("ZkDaily", function(opts)
        local zk = require("zk")
        if opts.args == "yesterday" then
            zk.daily({ date = "yesterday" })
        elseif opts.args ~= "" then
            zk.daily({ date = opts.args })
        else
            zk.daily()
        end
    end, {
        nargs = "?",
        complete = function()
            return { "yesterday" }
        end,
        desc = "Open or create daily note",
    })

    vim.api.nvim_create_user_command("ZkDailyList", function(opts)
        local zk = require("zk")
        if opts.bang then
            zk.daily_picker({ week = true })
        else
            zk.daily_picker()
        end
    end, {
        bang = true,
        desc = "Browse daily notes (! for this week only)",
    })

    -- Note creation
    vim.api.nvim_create_user_command("ZkNew", function(opts)
        local zk = require("zk")
        local args_parsed = {}

        -- Parse args
        local args_list = vim.split(opts.args, "%s+")
        local i = 1
        local links = {}
        while i <= #args_list do
            local arg = args_list[i]
            if arg == "--type" and args_list[i + 1] then
                args_parsed.note_type = args_list[i + 1]
                i = i + 2
            elseif arg == "--project" and args_list[i + 1] then
                args_parsed.project = args_list[i + 1]
                i = i + 2
            elseif arg == "--link" and args_list[i + 1] then
                table.insert(links, args_list[i + 1])
                i = i + 2
            elseif arg == "--link-daily" then
                args_parsed.link_daily = true
                i = i + 1
            elseif arg == "fleeting" or arg == "permanent" then
                args_parsed.note_type = arg
                i = i + 1
            else
                i = i + 1
            end
        end

        if #links > 0 then
            args_parsed.links = links
        end

        -- Default to fleeting
        if not args_parsed.note_type then
            args_parsed.note_type = "fleeting"
        end

        zk.create_note(args_parsed)
    end, {
        nargs = "*",
        complete = function(_, line)
            local args = vim.split(line, "%s+")
            if #args == 2 then
                return { "fleeting", "permanent", "--type", "--project", "--link", "--link-daily" }
            end
            return {}
        end,
        desc = "Create new zettel (--type fleeting/permanent --link ID --link-daily)",
    })

    vim.api.nvim_create_user_command("ZkTemplate", function(opts)
        local zk = require("zk")
        local args_parsed = {}

        -- Parse args
        local args_list = vim.split(opts.args, "%s+")
        local i = 1
        local links = {}
        while i <= #args_list do
            local arg = args_list[i]
            if arg == "--link" and args_list[i + 1] then
                table.insert(links, args_list[i + 1])
                i = i + 2
            elseif arg == "--link-daily" then
                args_parsed.link_daily = true
                i = i + 1
            elseif arg == "--project" and args_list[i + 1] then
                args_parsed.project = args_list[i + 1]
                i = i + 2
            elseif not arg:match("^%-%-") then
                -- Assume it's the template name
                args_parsed.template = arg
                i = i + 1
            else
                i = i + 1
            end
        end

        if #links > 0 then
            args_parsed.links = links
        end

        if args_parsed.template then
            zk.create_from_template(args_parsed)
        else
            zk.template_picker()
        end
    end, {
        nargs = "*",
        complete = function()
            local zk = require("zk")
            local names = {}
            for _, tmpl in ipairs(zk.templates) do
                table.insert(names, tmpl.name)
            end
            table.insert(names, "--link")
            table.insert(names, "--link-daily")
            table.insert(names, "--project")
            return names
        end,
        desc = "Create note from template (--link ID --link-daily)",
    })

    -- Note management
    vim.api.nvim_create_user_command("ZkPromote", function()
        require("zk").promote_note()
    end, {
        desc = "Promote current note to permanent",
    })

    vim.api.nvim_create_user_command("ZkSetProject", function(opts)
        local zk = require("zk")
        if opts.args ~= "" then
            zk.set_project(nil, opts.args)
        else
            zk.set_project()
        end
    end, {
        nargs = "?",
        desc = "Set project for current note",
    })

    -- Search and navigation
    vim.api.nvim_create_user_command("ZkSearch", function(opts)
        local has_telescope = pcall(require, "telescope")
        if has_telescope then
            if opts.bang then
                require("zk.telescope").live_search()
            else
                require("zk.telescope").search({ default_text = opts.args })
            end
        else
            require("zk").search(opts.args)
        end
    end, {
        nargs = "?",
        bang = true,
        desc = "Search zettels (! for live search)",
    })

    vim.api.nvim_create_user_command("ZkFleeting", function()
        local has_telescope = pcall(require, "telescope")
        if has_telescope then
            require("zk.telescope").fleeting()
        else
            require("zk").search("", { category = "fleeting" })
        end
    end, {
        desc = "Browse fleeting notes",
    })

    vim.api.nvim_create_user_command("ZkPermanent", function()
        local has_telescope = pcall(require, "telescope")
        if has_telescope then
            require("zk.telescope").permanent()
        else
            require("zk").search("", { category = "permanent" })
        end
    end, {
        desc = "Browse permanent notes",
    })

    -- Backlinks
    vim.api.nvim_create_user_command("ZkBacklinks", function(opts)
        local zk = require("zk")
        if opts.bang then
            zk.backlinks_split()
        else
            zk.backlinks_panel()
        end
    end, {
        bang = true,
        desc = "Show backlinks (! for split)",
    })

    -- Graph
    vim.api.nvim_create_user_command("ZkGraph", function(opts)
        local zk = require("zk")
        local limit = tonumber(opts.args) or 10
        zk.graph({ limit = limit })
    end, {
        nargs = "?",
        desc = "Generate graph visualization",
    })

    -- Index
    vim.api.nvim_create_user_command("ZkIndex", function(opts)
        local zk = require("zk")
        local path = opts.args ~= "" and opts.args or nil
        zk.index(path)
    end, {
        nargs = "?",
        complete = "dir",
        desc = "Index zettels",
    })

    -- Linking
    vim.api.nvim_create_user_command("ZkInsertLink", function(opts)
        local zk = require("zk")
        if opts.bang then
            zk.link_picker({ include_title = true })
        else
            zk.link_picker()
        end
    end, {
        bang = true,
        desc = "Insert link to zettel (! includes title)",
    })

    -- Preview
    vim.api.nvim_create_user_command("ZkPreview", function(opts)
        local zk = require("zk")
        if opts.args ~= "" then
            zk.preview_by_id(opts.args)
        else
            zk.preview_note()
        end
    end, {
        nargs = "?",
        desc = "Preview note (optionally by ID)",
    })

    -- Tags
    vim.api.nvim_create_user_command("ZkRefreshTags", function()
        require("zk").refresh_tags()
    end, {
        desc = "Refresh tag cache",
    })

    -- Templates list
    vim.api.nvim_create_user_command("ZkTemplates", function()
        local zk = require("zk")
        local lines = { "Available templates:", "" }
        for _, tmpl in ipairs(zk.templates) do
            table.insert(lines, string.format("  %-15s %s [%s]", tmpl.name, tmpl.description, tmpl.category))
        end
        print(table.concat(lines, "\n"))
    end, {
        desc = "List available templates",
    })

    -- Todo commands
    vim.api.nvim_create_user_command("ZkTodo", function(opts)
        local zk = require("zk")
        local args_parsed = {}

        -- Parse args for --due, --priority, --link, --link-daily
        local args_list = vim.split(opts.args, "%s+")
        local i = 1
        local title_parts = {}
        local links = {}
        while i <= #args_list do
            local arg = args_list[i]
            if arg == "--due" and args_list[i + 1] then
                args_parsed.due = args_list[i + 1]
                i = i + 2
            elseif arg == "--priority" and args_list[i + 1] then
                args_parsed.priority = args_list[i + 1]
                i = i + 2
            elseif arg == "--project" and args_list[i + 1] then
                args_parsed.project = args_list[i + 1]
                i = i + 2
            elseif arg == "--link" and args_list[i + 1] then
                table.insert(links, args_list[i + 1])
                i = i + 2
            elseif arg == "--link-daily" then
                args_parsed.link_daily = true
                i = i + 1
            else
                table.insert(title_parts, arg)
                i = i + 1
            end
        end

        if #links > 0 then
            args_parsed.links = links
        end

        if #title_parts > 0 then
            args_parsed.title = table.concat(title_parts, " ")
        end

        zk.todo(args_parsed)
    end, {
        nargs = "*",
        desc = "Create new todo (--due DATE --priority high/medium/low --link ID --link-daily)",
    })

    -- Convenience command for creating todo linked to daily note
    vim.api.nvim_create_user_command("ZkTodoDaily", function(opts)
        local zk = require("zk")
        local args_parsed = { link_daily = true }

        if opts.args ~= "" then
            args_parsed.title = opts.args
        end

        zk.todo(args_parsed)
    end, {
        nargs = "?",
        desc = "Create todo linked to today's daily note",
    })

    vim.api.nvim_create_user_command("ZkTodos", function(opts)
        local zk = require("zk")
        local picker_opts = {}

        if opts.bang then
            picker_opts.closed = true
        end

        if opts.args == "overdue" then
            picker_opts.overdue = true
        elseif opts.args == "today" then
            picker_opts.today = true
        elseif opts.args == "week" then
            picker_opts.this_week = true
        elseif opts.args ~= "" then
            picker_opts.project = opts.args
        end

        local has_telescope = pcall(require, "telescope")
        if has_telescope then
            zk.todo_picker(picker_opts)
        else
            zk.todos(picker_opts)
        end
    end, {
        nargs = "?",
        bang = true,
        complete = function()
            return { "overdue", "today", "week" }
        end,
        desc = "Browse todos (! for closed)",
    })

    vim.api.nvim_create_user_command("ZkDone", function(opts)
        local zk = require("zk")
        if opts.args ~= "" then
            zk.done(opts.args)
        else
            zk.done()
        end
    end, {
        nargs = "?",
        desc = "Mark todo as done",
    })

    vim.api.nvim_create_user_command("ZkReopen", function(opts)
        local zk = require("zk")
        if opts.args ~= "" then
            zk.reopen(opts.args)
        else
            zk.reopen()
        end
    end, {
        nargs = "?",
        desc = "Reopen a closed todo",
    })

    vim.api.nvim_create_user_command("ZkTodoList", function(opts)
        local zk = require("zk")
        local list_opts = {}

        if opts.args == "today" then
            list_opts.today = true
        elseif opts.args == "week" then
            list_opts.this_week = true
        elseif opts.args ~= "" then
            list_opts.project = opts.args
        end

        zk.todo_list(list_opts)
    end, {
        nargs = "?",
        complete = function()
            return { "today", "week" }
        end,
        desc = "Generate todo list markdown",
    })
end

-- Create commands when plugin loads
create_commands()

-- Provide a convenient setup function that auto-configures everything
-- Users can call require("zk").setup() for more control
vim.api.nvim_create_autocmd("User", {
    pattern = "VeryLazy",
    callback = function()
        -- If user hasn't called setup() yet, do a minimal setup
        local zk = require("zk")
        if not zk.config then
            zk.setup()
        end
    end,
})
