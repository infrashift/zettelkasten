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
        while i <= #args_list do
            local arg = args_list[i]
            if arg == "--category" and args_list[i + 1] then
                args_parsed.note_type = args_list[i + 1]
                i = i + 2
            elseif arg == "--project" and args_list[i + 1] then
                args_parsed.project = args_list[i + 1]
                i = i + 2
            else
                i = i + 1
            end
        end

        -- Default to untethered
        if not args_parsed.note_type then
            args_parsed.note_type = "untethered"
        end

        zk.create_note(args_parsed)
    end, {
        nargs = "*",
        complete = function(_, line)
            local args = vim.split(line, "%s+")
            local last = args[#args] or ""
            local prev = args[#args - 1] or ""

            -- Complete values after flags
            if prev == "--category" then
                return { "untethered", "tethered" }
            end
            if prev == "--project" then
                return {}
            end

            -- Complete flags (filter out already-used ones)
            local used = {}
            for _, a in ipairs(args) do
                used[a] = true
            end
            local completions = {}
            if not used["--category"] then
                table.insert(completions, "--category")
            end
            if not used["--project"] then
                table.insert(completions, "--project")
            end
            return completions
        end,
        desc = "Create new zettel (--category untethered/tethered --project NAME)",
    })

    vim.api.nvim_create_user_command("ZkTemplate", function(opts)
        local zk = require("zk")
        local args_parsed = {}

        -- Parse args
        local args_list = vim.split(opts.args, "%s+")
        local i = 1
        while i <= #args_list do
            local arg = args_list[i]
            if arg == "--project" and args_list[i + 1] then
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

        if args_parsed.template then
            zk.create_from_template(args_parsed)
        else
            zk.template_picker()
        end
    end, {
        nargs = "*",
        complete = function(_, line)
            local args = vim.split(line, "%s+")
            local prev = args[#args - 1] or ""

            -- Complete project name (user types freely)
            if prev == "--project" then
                return {}
            end

            -- Check what's already been provided
            local has_template = false
            local has_project = false
            for _, a in ipairs(args) do
                if a == "--project" then
                    has_project = true
                elseif not a:match("^%-%-") and a ~= "" and a ~= "ZkTemplate" then
                    has_template = true
                end
            end

            local completions = {}
            -- Offer template names if none selected yet
            if not has_template then
                local zk = require("zk")
                for _, tmpl in ipairs(zk.templates) do
                    table.insert(completions, tmpl.name)
                end
            end
            if not has_project then
                table.insert(completions, "--project")
            end
            return completions
        end,
        desc = "Create note from template (--project NAME)",
    })

    -- Note management
    vim.api.nvim_create_user_command("ZkTether", function()
        require("zk").tether_note()
    end, {
        desc = "Tether current note (set category to tethered)",
    })

    vim.api.nvim_create_user_command("ZkUntether", function()
        require("zk").untether_note()
    end, {
        desc = "Untether current note (set category to untethered)",
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

    vim.api.nvim_create_user_command("ZkUntethered", function()
        local has_telescope = pcall(require, "telescope")
        if has_telescope then
            require("zk.telescope").untethered()
        else
            require("zk").search("", { category = "untethered" })
        end
    end, {
        desc = "Browse untethered notes",
    })

    vim.api.nvim_create_user_command("ZkTethered", function()
        local has_telescope = pcall(require, "telescope")
        if has_telescope then
            require("zk.telescope").tethered()
        else
            require("zk").search("", { category = "tethered" })
        end
    end, {
        desc = "Browse tethered notes",
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
        local graph_opts = { html = not opts.bang }

        -- Parse args for --depth N or plain number
        local args_list = vim.split(opts.args, "%s+")
        local i = 1
        while i <= #args_list do
            local arg = args_list[i]
            if arg == "--depth" and args_list[i + 1] then
                graph_opts.limit = tonumber(args_list[i + 1])
                i = i + 2
            elseif tonumber(arg) then
                graph_opts.limit = tonumber(arg)
                i = i + 1
            else
                i = i + 1
            end
        end

        zk.graph(graph_opts)
    end, {
        nargs = "*",
        bang = true,
        desc = "Generate graph (--depth N, ! for markdown-only)",
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

        -- Parse args for --due, --priority, --project
        local args_list = vim.split(opts.args, "%s+")
        local i = 1
        local title_parts = {}
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
            else
                table.insert(title_parts, arg)
                i = i + 1
            end
        end

        if #title_parts > 0 then
            args_parsed.title = table.concat(title_parts, " ")
        end

        zk.todo(args_parsed)
    end, {
        nargs = "*",
        complete = function(_, line)
            local args = vim.split(line, "%s+")
            local prev = args[#args - 1] or ""

            -- Complete values after flags
            if prev == "--priority" then
                return { "high", "medium", "low" }
            end
            if prev == "--due" or prev == "--project" then
                return {}
            end

            -- Complete flags (filter out already-used ones)
            local used = {}
            for _, a in ipairs(args) do
                used[a] = true
            end
            local completions = {}
            if not used["--due"] then
                table.insert(completions, "--due")
            end
            if not used["--priority"] then
                table.insert(completions, "--priority")
            end
            if not used["--project"] then
                table.insert(completions, "--project")
            end
            return completions
        end,
        desc = "Create new todo (--due DATE --priority high/medium/low --project NAME)",
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
