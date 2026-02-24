-- Telescope integration for zk
-- Requires telescope.nvim to be installed

local M = {}

local has_telescope, telescope = pcall(require, "telescope")
if not has_telescope then
    return M
end

local pickers = require("telescope.pickers")
local finders = require("telescope.finders")
local conf = require("telescope.config").values
local actions = require("telescope.actions")
local action_state = require("telescope.actions.state")
local previewers = require("telescope.previewers")

-- Get the zk config from the main module
local function get_config()
    local ok, zk = pcall(require, "zk")
    if ok and zk.config then
        return zk.config
    end
    return { bin = "zk" }
end

-- Parse JSON search results
local function parse_results(json_str)
    local ok, results = pcall(vim.json.decode, json_str)
    if not ok or type(results) ~= "table" then
        return {}
    end
    return results
end

-- Search zettels with optional filters
M.search = function(opts)
    opts = opts or {}

    local config = get_config()
    local args = { "search", "--json" }

    -- Add query if provided
    if opts.query and opts.query ~= "" then
        table.insert(args, opts.query)
    end

    -- Add filters
    if opts.project and opts.project ~= "" then
        table.insert(args, "--project")
        table.insert(args, opts.project)
    end

    if opts.category and opts.category ~= "" then
        table.insert(args, "--category")
        table.insert(args, opts.category)
    end

    if opts.type and opts.type ~= "" then
        table.insert(args, "--type")
        table.insert(args, opts.type)
    end

    if opts.tags and type(opts.tags) == "table" then
        for _, tag in ipairs(opts.tags) do
            table.insert(args, "--tag")
            table.insert(args, tag)
        end
    end

    if opts.status and opts.status ~= "" then
        table.insert(args, "--status")
        table.insert(args, opts.status)
    end

    if opts.priority and opts.priority ~= "" then
        table.insert(args, "--priority")
        table.insert(args, opts.priority)
    end

    if opts.due_before and opts.due_before ~= "" then
        table.insert(args, "--due-before")
        table.insert(args, opts.due_before)
    end

    if opts.due_after and opts.due_after ~= "" then
        table.insert(args, "--due-after")
        table.insert(args, opts.due_after)
    end

    if opts.limit then
        table.insert(args, "--limit")
        table.insert(args, tostring(opts.limit))
    else
        table.insert(args, "--limit")
        table.insert(args, "50")
    end

    -- Execute search
    local Job = require("plenary.job")
    local results_json = ""

    local job = Job:new({
        command = config.bin,
        args = args,
        on_stdout = function(_, data)
            results_json = results_json .. data
        end,
    })

    job:sync(10000)

    local results = parse_results(results_json)

    if #results == 0 then
        vim.notify("No zettels found", vim.log.levels.INFO)
        return
    end

    pickers.new(opts, {
        prompt_title = "Zettels",
        finder = finders.new_table({
            results = results,
            entry_maker = function(entry)
                local title = entry.title or entry.id or ""
                local project = (entry.project and entry.project ~= "") and entry.project or ""
                local category = entry.category or ""
                local tags = (type(entry.tags) == "table") and entry.tags or {}

                local display = title
                if project ~= "" then
                    display = display .. " [" .. project .. "]"
                end
                if category ~= "" then
                    display = display .. " (" .. category .. ")"
                end

                return {
                    value = entry,
                    display = display,
                    ordinal = title .. " " .. project .. " " .. table.concat(tags, " "),
                    path = entry.file_path,
                }
            end,
        }),
        sorter = conf.generic_sorter(opts),
        previewer = previewers.new_buffer_previewer({
            title = "Zettel Preview",
            define_preview = function(self, entry)
                if entry.path and entry.path ~= "" then
                    conf.buffer_previewer_maker(entry.path, self.state.bufnr, {
                        bufname = self.state.bufname,
                    })
                else
                    vim.api.nvim_buf_set_lines(self.state.bufnr, 0, -1, false, {
                        "No file path available",
                        "",
                        "ID: " .. (entry.value.id or ""),
                        "Title: " .. (entry.value.title or ""),
                        "Project: " .. (entry.value.project or ""),
                        "Category: " .. (entry.value.category or ""),
                        "Tags: " .. table.concat(entry.value.tags or {}, ", "),
                    })
                end
            end,
        }),
        attach_mappings = function(prompt_bufnr, map)
            actions.select_default:replace(function()
                actions.close(prompt_bufnr)
                local selection = action_state.get_selected_entry()
                if selection and selection.path then
                    vim.cmd("edit " .. vim.fn.fnameescape(selection.path))
                end
            end)

            -- Open in floating preview window with Ctrl-p
            map("i", "<C-p>", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.path then
                    actions.close(prompt_bufnr)
                    local zk = require("zk")
                    zk.preview_note(selection.path)
                end
            end)
            map("n", "<C-p>", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.path then
                    actions.close(prompt_bufnr)
                    local zk = require("zk")
                    zk.preview_note(selection.path)
                end
            end)

            -- Insert link [[id]] with Ctrl-l
            map("i", "<C-l>", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    actions.close(prompt_bufnr)
                    local zk = require("zk")
                    zk.insert_link(selection.value.id, selection.value.title, false)
                end
            end)
            map("n", "<C-l>", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    actions.close(prompt_bufnr)
                    local zk = require("zk")
                    zk.insert_link(selection.value.id, selection.value.title, false)
                end
            end)

            -- Insert link [[id|title]] with Ctrl-Shift-l (Ctrl-L)
            map("i", "<C-S-l>", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    actions.close(prompt_bufnr)
                    local zk = require("zk")
                    zk.insert_link(selection.value.id, selection.value.title, true)
                end
            end)
            map("n", "<C-S-l>", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    actions.close(prompt_bufnr)
                    local zk = require("zk")
                    zk.insert_link(selection.value.id, selection.value.title, true)
                end
            end)
            return true
        end,
    }):find()
end

-- Search with live query input
M.live_search = function(opts)
    opts = opts or {}
    local config = get_config()

    pickers.new(opts, {
        prompt_title = "Search Zettels",
        finder = finders.new_dynamic({
            fn = function(prompt)
                if not prompt or prompt == "" then
                    return {}
                end

                local args = { "search", "--json", "--limit", "30" }

                -- Add filters before the query
                if opts.project and opts.project ~= "" then
                    table.insert(args, "--project")
                    table.insert(args, opts.project)
                end

                if opts.category and opts.category ~= "" then
                    table.insert(args, "--category")
                    table.insert(args, opts.category)
                end

                if opts.type and opts.type ~= "" then
                    table.insert(args, "--type")
                    table.insert(args, opts.type)
                end

                if opts.status and opts.status ~= "" then
                    table.insert(args, "--status")
                    table.insert(args, opts.status)
                end

                if opts.priority and opts.priority ~= "" then
                    table.insert(args, "--priority")
                    table.insert(args, opts.priority)
                end

                if opts.due_before and opts.due_before ~= "" then
                    table.insert(args, "--due-before")
                    table.insert(args, opts.due_before)
                end

                if opts.due_after and opts.due_after ~= "" then
                    table.insert(args, "--due-after")
                    table.insert(args, opts.due_after)
                end

                -- Query goes last
                table.insert(args, prompt)

                local Job = require("plenary.job")
                local results_json = ""

                local job = Job:new({
                    command = config.bin,
                    args = args,
                    on_stdout = function(_, data)
                        results_json = results_json .. data
                    end,
                })

                job:sync(5000)

                local results = parse_results(results_json)
                local entries = {}

                for _, entry in ipairs(results) do
                    local display = entry.title or entry.id
                    if entry.project and entry.project ~= "" then
                        display = display .. " [" .. entry.project .. "]"
                    end

                    table.insert(entries, {
                        value = entry,
                        display = display,
                        ordinal = (entry.title or "") .. " " .. (entry.project or ""),
                        path = entry.file_path,
                    })
                end

                return entries
            end,
            entry_maker = function(entry)
                return entry
            end,
        }),
        previewer = previewers.new_buffer_previewer({
            title = "Zettel Preview",
            define_preview = function(self, entry)
                if entry.path and entry.path ~= "" then
                    conf.buffer_previewer_maker(entry.path, self.state.bufnr, {
                        bufname = self.state.bufname,
                    })
                end
            end,
        }),
        attach_mappings = function(prompt_bufnr, map)
            actions.select_default:replace(function()
                actions.close(prompt_bufnr)
                local selection = action_state.get_selected_entry()
                if selection and selection.path then
                    vim.cmd("edit " .. vim.fn.fnameescape(selection.path))
                end
            end)

            -- Open in floating preview window with Ctrl-p
            map("i", "<C-p>", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.path then
                    actions.close(prompt_bufnr)
                    local zk = require("zk")
                    zk.preview_note(selection.path)
                end
            end)
            map("n", "<C-p>", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.path then
                    actions.close(prompt_bufnr)
                    local zk = require("zk")
                    zk.preview_note(selection.path)
                end
            end)

            -- Insert link [[id]] with Ctrl-l
            map("i", "<C-l>", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    actions.close(prompt_bufnr)
                    local zk = require("zk")
                    zk.insert_link(selection.value.id, selection.value.title, false)
                end
            end)
            map("n", "<C-l>", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    actions.close(prompt_bufnr)
                    local zk = require("zk")
                    zk.insert_link(selection.value.id, selection.value.title, false)
                end
            end)

            -- Insert link [[id|title]] with Ctrl-Shift-l
            map("i", "<C-S-l>", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    actions.close(prompt_bufnr)
                    local zk = require("zk")
                    zk.insert_link(selection.value.id, selection.value.title, true)
                end
            end)
            map("n", "<C-S-l>", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    actions.close(prompt_bufnr)
                    local zk = require("zk")
                    zk.insert_link(selection.value.id, selection.value.title, true)
                end
            end)
            return true
        end,
    }):find()
end

-- Browse by project
M.projects = function(opts)
    M.search(vim.tbl_extend("force", opts or {}, {
        prompt_title = "Zettels by Project",
    }))
end

-- Browse untethered notes
M.untethered = function(opts)
    M.search(vim.tbl_extend("force", opts or {}, {
        category = "untethered",
        prompt_title = "Untethered Notes",
    }))
end

-- Browse tethered notes
M.tethered = function(opts)
    M.search(vim.tbl_extend("force", opts or {}, {
        category = "tethered",
        prompt_title = "Tethered Notes",
    }))
end

-- Insert link picker (dedicated function for linking)
M.insert_link = function(opts)
    opts = opts or {}
    local zk = require("zk")
    zk.link_picker(opts)
end

-- Register as telescope extension
telescope.register_extension({
    exports = {
        zk = M.search,
        search = M.search,
        live_search = M.live_search,
        untethered = M.untethered,
        tethered = M.tethered,
        projects = M.projects,
        insert_link = M.insert_link,
    },
})

return M
