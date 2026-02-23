local M = {}

M.setup = function(opts)
    M.config = vim.tbl_deep_extend("force", {
        bin = "zk",
    }, opts or {})
end

M.create_note = function(opts_or_type, project)
    local opts = {}

    -- Handle legacy call signature: create_note(type, project)
    if type(opts_or_type) == "string" then
        opts = {
            note_type = opts_or_type,
            project = project,
        }
    else
        opts = opts_or_type or {}
    end

    local title = opts.title
    if not title or title == "" then
        title = vim.fn.input("Note Title: ")
        if title == "" then return end
    end

    local cwd = vim.fn.getcwd()
    local args = { "create", title, "--category", opts.note_type or "untethered" }

    if opts.project and opts.project ~= "" then
        table.insert(args, "--project")
        table.insert(args, opts.project)
    end

    local created_file = ""
    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = args,
        cwd = cwd,
        on_stdout = function(_, data)
            local match = data:match("Created: (.+)")
            if match then
                created_file = match
            end
        end,
        on_exit = function(_job, return_val)
            vim.schedule(function()
                if return_val == 0 then
                    print("Note created successfully")
                    if created_file ~= "" and vim.fn.filereadable(created_file) == 1 then
                        vim.cmd("edit " .. vim.fn.fnameescape(created_file))
                    end
                else
                    print("Failed to create note")
                end
            end)
        end,
    })
    job:start()
end

-- Available templates (cached)
M._templates_cache = nil

-- Daily notes functions

-- Create or open a daily note
M.daily = function(opts)
    opts = opts or {}

    local args = { "daily" }

    -- Handle date options
    if opts.date then
        if opts.date == "yesterday" then
            table.insert(args, "--yesterday")
        else
            table.insert(args, "--date")
            table.insert(args, opts.date)
        end
    end

    local daily_file = ""
    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = args,
        on_stdout = function(_, data)
            -- Capture the file path from output
            local match = data:match("Daily note: (.+)")
            if match then
                daily_file = match
            end
            match = data:match("Created daily note: (.+)")
            if match then
                daily_file = match
            end
        end,
        on_exit = function(_job, return_val)
            vim.schedule(function()
                if return_val == 0 and daily_file ~= "" then
                    vim.cmd("edit " .. vim.fn.fnameescape(daily_file))
                elseif return_val ~= 0 then
                    print("Failed to open daily note")
                end
            end)
        end,
    })
    job:start()
end

-- List daily notes (async)
M.list_daily = function(opts, callback)
    opts = opts or {}

    local args = { "daily", "--list", "--json" }

    if opts.week then
        table.insert(args, "--week")
    elseif opts.month then
        table.insert(args, "--month")
    end

    local results_json = ""
    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = args,
        on_stdout = function(_, data)
            results_json = results_json .. data
        end,
        on_exit = function(_job, return_val)
            vim.schedule(function()
                if return_val ~= 0 then
                    callback({})
                    return
                end

                local ok, results = pcall(vim.json.decode, results_json)
                if not ok or type(results) ~= "table" then
                    callback({})
                    return
                end

                callback(results)
            end)
        end,
    })
    job:start()
end

-- List daily notes synchronously
M.list_daily_sync = function(opts)
    opts = opts or {}

    local args = { "daily", "--list", "--json" }

    if opts.week then
        table.insert(args, "--week")
    elseif opts.month then
        table.insert(args, "--month")
    end

    local Job = require("plenary.job")
    local results_json = ""

    local job = Job:new({
        command = M.config.bin,
        args = args,
        on_stdout = function(_, data)
            results_json = results_json .. data
        end,
    })

    job:sync(5000)

    local ok, results = pcall(vim.json.decode, results_json)
    if not ok or type(results) ~= "table" then
        return {}
    end

    return results
end

-- Telescope picker for daily notes
M.daily_picker = function(opts)
    opts = opts or {}

    local has_telescope = pcall(require, "telescope")
    if not has_telescope then
        print("Telescope required for daily picker")
        return
    end

    local pickers = require("telescope.pickers")
    local finders = require("telescope.finders")
    local conf = require("telescope.config").values
    local actions = require("telescope.actions")
    local action_state = require("telescope.actions.state")
    local previewers = require("telescope.previewers")

    -- Get daily notes
    local daily_notes = M.list_daily_sync(opts)

    if #daily_notes == 0 then
        vim.notify("No daily notes found", vim.log.levels.INFO)
        return
    end

    pickers.new(opts, {
        prompt_title = "Daily Notes",
        finder = finders.new_table({
            results = daily_notes,
            entry_maker = function(entry)
                return {
                    value = entry,
                    display = entry.title or entry.date,
                    ordinal = entry.date .. " " .. (entry.title or ""),
                    path = entry.file_path,
                }
            end,
        }),
        sorter = conf.generic_sorter(opts),
        previewer = previewers.new_buffer_previewer({
            title = "Daily Note Preview",
            define_preview = function(self, entry)
                if entry.path and entry.path ~= "" then
                    conf.buffer_previewer_maker(entry.path, self.state.bufnr, {
                        bufname = self.state.bufname,
                    })
                end
            end,
        }),
        attach_mappings = function(prompt_bufnr, _)
            actions.select_default:replace(function()
                actions.close(prompt_bufnr)
                local selection = action_state.get_selected_entry()
                if selection and selection.path then
                    vim.cmd("edit " .. vim.fn.fnameescape(selection.path))
                end
            end)
            return true
        end,
    }):find()
end

-- Todo management functions

-- Create a new todo
M.todo = function(opts)
    opts = opts or {}

    local title = opts.title
    if not title or title == "" then
        title = vim.fn.input("Todo: ")
        if title == "" then return end
    end

    local args = { "todo", title }

    if opts.project and opts.project ~= "" then
        table.insert(args, "--project")
        table.insert(args, opts.project)
    end

    if opts.due and opts.due ~= "" then
        table.insert(args, "--due")
        table.insert(args, opts.due)
    end

    if opts.priority and opts.priority ~= "" then
        table.insert(args, "--priority")
        table.insert(args, opts.priority)
    end

    local created_file = ""
    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = args,
        on_stdout = function(_, data)
            local match = data:match("File: (.+)")
            if match then
                created_file = match
            end
        end,
        on_exit = function(_job, return_val)
            vim.schedule(function()
                if return_val == 0 then
                    print("Todo created: " .. title)
                    if created_file ~= "" and vim.fn.filereadable(created_file) == 1 then
                        vim.cmd("edit " .. vim.fn.fnameescape(created_file))
                    end
                else
                    print("Failed to create todo")
                end
            end)
        end,
    })
    job:start()
end

-- Mark a todo as done
M.done = function(id_or_file)
    id_or_file = id_or_file or vim.fn.expand("%:p")
    if id_or_file == "" then
        print("No todo specified")
        return
    end

    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = { "done", id_or_file },
        on_exit = function(j, return_val)
            vim.schedule(function()
                if return_val == 0 then
                    local result = table.concat(j:result(), "\n")
                    print(result)
                    -- Reload current buffer if it's the todo
                    if vim.fn.expand("%:p") == id_or_file or vim.fn.expand("%:p"):match(id_or_file) then
                        vim.cmd("edit!")
                    end
                else
                    local err = table.concat(j:stderr_result(), "\n")
                    print("Failed to mark done: " .. err)
                end
            end)
        end,
    })
    job:start()
end

-- Reopen a closed todo
M.reopen = function(id_or_file)
    id_or_file = id_or_file or vim.fn.expand("%:p")
    if id_or_file == "" then
        print("No todo specified")
        return
    end

    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = { "reopen", id_or_file },
        on_exit = function(j, return_val)
            vim.schedule(function()
                if return_val == 0 then
                    local result = table.concat(j:result(), "\n")
                    print(result)
                    -- Reload current buffer if it's the todo
                    if vim.fn.expand("%:p") == id_or_file or vim.fn.expand("%:p"):match(id_or_file) then
                        vim.cmd("edit!")
                    end
                else
                    local err = table.concat(j:stderr_result(), "\n")
                    print("Failed to reopen: " .. err)
                end
            end)
        end,
    })
    job:start()
end

-- Search todos (async)
M.todos = function(opts, callback)
    opts = opts or {}

    local args = { "todos", "--json" }

    if opts.closed then
        table.insert(args, "--closed")
    end

    if opts.project and opts.project ~= "" then
        table.insert(args, "--project")
        table.insert(args, opts.project)
    end

    if opts.priority and opts.priority ~= "" then
        table.insert(args, "--priority")
        table.insert(args, opts.priority)
    end

    if opts.overdue then
        table.insert(args, "--overdue")
    end

    if opts.today then
        table.insert(args, "--today")
    end

    if opts.this_week then
        table.insert(args, "--this-week")
    end

    if opts.query and opts.query ~= "" then
        table.insert(args, opts.query)
    end

    local results_json = ""
    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = args,
        on_stdout = function(_, data)
            results_json = results_json .. data
        end,
        on_exit = function(_job, return_val)
            vim.schedule(function()
                if return_val ~= 0 then
                    if callback then callback({}) end
                    return
                end

                local ok, results = pcall(vim.json.decode, results_json)
                if not ok or type(results) ~= "table" then
                    if callback then callback({}) end
                    return
                end

                if callback then
                    callback(results)
                else
                    -- Default: print results
                    if #results == 0 then
                        print("No todos found")
                    else
                        for _, r in ipairs(results) do
                            local status_icon = "[ ]"
                            if r.status == "closed" then
                                status_icon = "[x]"
                            elseif r.status == "in_progress" then
                                status_icon = "[~]"
                            end
                            print(string.format("%s %s: %s", status_icon, r.id, r.title))
                        end
                    end
                end
            end)
        end,
    })
    job:start()
end

-- Search todos synchronously
M.todos_sync = function(opts)
    opts = opts or {}

    local args = { "todos", "--json" }

    if opts.closed then
        table.insert(args, "--closed")
    end

    if opts.project and opts.project ~= "" then
        table.insert(args, "--project")
        table.insert(args, opts.project)
    end

    if opts.priority and opts.priority ~= "" then
        table.insert(args, "--priority")
        table.insert(args, opts.priority)
    end

    if opts.overdue then
        table.insert(args, "--overdue")
    end

    if opts.today then
        table.insert(args, "--today")
    end

    if opts.this_week then
        table.insert(args, "--this-week")
    end

    local Job = require("plenary.job")
    local results_json = ""

    local job = Job:new({
        command = M.config.bin,
        args = args,
        on_stdout = function(_, data)
            results_json = results_json .. data
        end,
    })

    job:sync(5000)

    local ok, results = pcall(vim.json.decode, results_json)
    if not ok or type(results) ~= "table" then
        return {}
    end

    return results
end

-- Telescope picker for todos
M.todo_picker = function(opts)
    opts = opts or {}

    local has_telescope = pcall(require, "telescope")
    if not has_telescope then
        print("Telescope required for todo picker")
        return
    end

    local pickers = require("telescope.pickers")
    local finders = require("telescope.finders")
    local conf = require("telescope.config").values
    local actions = require("telescope.actions")
    local action_state = require("telescope.actions.state")
    local previewers = require("telescope.previewers")

    -- Get todos
    local todos = M.todos_sync(opts)

    if #todos == 0 then
        vim.notify("No todos found", vim.log.levels.INFO)
        return
    end

    pickers.new(opts, {
        prompt_title = "Todos",
        finder = finders.new_table({
            results = todos,
            entry_maker = function(entry)
                local status_icon = "[ ]"
                if entry.status == "closed" then
                    status_icon = "[x]"
                elseif entry.status == "in_progress" then
                    status_icon = "[~]"
                end

                local display = string.format("%s %s", status_icon, entry.title or entry.id)
                if entry.due and entry.due ~= "" then
                    display = display .. " (due: " .. entry.due .. ")"
                end
                if entry.priority and entry.priority ~= "" then
                    display = display .. " [" .. entry.priority .. "]"
                end

                return {
                    value = entry,
                    display = display,
                    ordinal = entry.title .. " " .. (entry.id or ""),
                    path = entry.file_path,
                }
            end,
        }),
        sorter = conf.generic_sorter(opts),
        previewer = previewers.new_buffer_previewer({
            title = "Todo Preview",
            define_preview = function(self, entry)
                if entry.path and entry.path ~= "" then
                    conf.buffer_previewer_maker(entry.path, self.state.bufnr, {
                        bufname = self.state.bufname,
                    })
                end
            end,
        }),
        attach_mappings = function(prompt_bufnr, map)
            -- Default: open todo
            actions.select_default:replace(function()
                actions.close(prompt_bufnr)
                local selection = action_state.get_selected_entry()
                if selection and selection.path then
                    vim.cmd("edit " .. vim.fn.fnameescape(selection.path))
                end
            end)

            -- d: mark as done
            map("n", "d", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    actions.close(prompt_bufnr)
                    M.done(selection.path)
                end
            end)

            -- r: reopen
            map("n", "r", function()
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    actions.close(prompt_bufnr)
                    M.reopen(selection.path)
                end
            end)

            return true
        end,
    }):find()
end

-- Generate todo list markdown file
M.todo_list = function(opts)
    opts = opts or {}

    local args = { "todo-list" }

    if opts.project and opts.project ~= "" then
        table.insert(args, "--project")
        table.insert(args, opts.project)
    end

    if opts.today then
        table.insert(args, "--today")
    end

    if opts.this_week then
        table.insert(args, "--this-week")
    end

    if opts.output and opts.output ~= "" then
        table.insert(args, "--output")
        table.insert(args, opts.output)
    end

    local output_file = ""
    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = args,
        on_stdout = function(_, data)
            local match = data:match("Output: (.+)")
            if match then
                output_file = match
            end
        end,
        on_exit = function(j, return_val)
            vim.schedule(function()
                if return_val == 0 then
                    local result = table.concat(j:result(), "\n")
                    print(result)
                    if output_file ~= "" and vim.fn.filereadable(output_file) == 1 then
                        vim.cmd("vsplit " .. vim.fn.fnameescape(output_file))
                    end
                else
                    print("Failed to generate todo list")
                end
            end)
        end,
    })
    job:start()
end

-- Template definitions matching CLI
M.templates = {
    {
        name = "meeting",
        description = "Meeting notes with attendees and action items",
        category = "untethered",
    },
    {
        name = "book-review",
        description = "Book review with rating and key takeaways",
        category = "tethered",
    },
    {
        name = "snippet",
        description = "Code snippet with context and explanation",
        category = "untethered",
    },
    {
        name = "project-idea",
        description = "Project idea with goals and next steps",
        category = "untethered",
    },
    {
        name = "user-story",
        description = "User story with acceptance criteria",
        category = "untethered",
    },
    {
        name = "feature",
        description = "Feature specification with requirements",
        category = "untethered",
    },
    {
        name = "daily",
        description = "Daily note for thoughts, tasks, and reflections",
        category = "untethered",
    },
    {
        name = "todo",
        description = "Actionable task with status tracking",
        category = "untethered",
        type = "todo",
    },
    {
        name = "issue",
        description = "Issue tracking like GitHub (bug, enhancement, question)",
        category = "untethered",
        type = "issue",
    },
}

-- Get template by name
M.get_template = function(name)
    for _, tmpl in ipairs(M.templates) do
        if tmpl.name == name then
            return tmpl
        end
    end
    return nil
end

-- Create note from template
M.create_from_template = function(opts)
    opts = opts or {}

    local function do_create(template_name, project)
        local title = vim.fn.input("Note Title: ")
        if title == "" then return end

        local args = { "create", title, "--template", template_name }

        if project and project ~= "" then
            table.insert(args, "--project")
            table.insert(args, project)
        end

        local created_file = ""
        local job = require('plenary.job'):new({
            command = M.config.bin,
            args = args,
            on_stdout = function(_, data)
                local match = data:match("Created: (.+)")
                if match then
                    created_file = match
                end
            end,
            on_exit = function(_job, return_val)
                vim.schedule(function()
                    if return_val == 0 then
                        print("Note created from template: " .. template_name)
                        -- Open the created file
                        if created_file ~= "" and vim.fn.filereadable(created_file) == 1 then
                            vim.cmd("edit " .. vim.fn.fnameescape(created_file))
                        end
                    else
                        print("Failed to create note from template")
                    end
                end)
            end,
        })
        job:start()
    end

    -- If template specified, use it directly
    if opts.template then
        do_create(opts.template, opts.project)
        return
    end

    -- Otherwise show picker
    local has_telescope = pcall(require, "telescope")
    if has_telescope then
        M.template_picker(opts)
    else
        -- Fallback: prompt for template name
        local names = {}
        for _, tmpl in ipairs(M.templates) do
            table.insert(names, tmpl.name)
        end
        print("Available templates: " .. table.concat(names, ", "))
        local template_name = vim.fn.input("Template: ")
        if template_name ~= "" then
            do_create(template_name, opts.project)
        end
    end
end

-- Telescope picker for templates
M.template_picker = function(opts)
    opts = opts or {}

    local has_telescope = pcall(require, "telescope")
    if not has_telescope then
        print("Telescope required for template picker")
        return
    end

    local pickers = require("telescope.pickers")
    local finders = require("telescope.finders")
    local conf = require("telescope.config").values
    local actions = require("telescope.actions")
    local action_state = require("telescope.actions.state")

    pickers.new(opts, {
        prompt_title = "Select Template",
        finder = finders.new_table({
            results = M.templates,
            entry_maker = function(entry)
                local display = string.format("%-15s %s [%s]",
                    entry.name, entry.description, entry.category)
                return {
                    value = entry,
                    display = display,
                    ordinal = entry.name .. " " .. entry.description,
                }
            end,
        }),
        sorter = conf.generic_sorter(opts),
        attach_mappings = function(prompt_bufnr, _)
            actions.select_default:replace(function()
                actions.close(prompt_bufnr)
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    local tmpl = selection.value

                    -- Prompt for title
                    local title = vim.fn.input("Note Title: ")
                    if title == "" then return end

                    -- For tethered templates, ensure project
                    local project = opts.project
                    if tmpl.category == "tethered" and (not project or project == "") then
                        project = vim.fn.input("Project (required for tethered): ")
                        if project == "" then
                            print("Tethered notes require a project")
                            return
                        end
                    end

                    local args = { "create", title, "--template", tmpl.name }
                    if project and project ~= "" then
                        table.insert(args, "--project")
                        table.insert(args, project)
                    end

                    local created_file = ""
                    local job = require('plenary.job'):new({
                        command = M.config.bin,
                        args = args,
                        on_stdout = function(_, data)
                            local match = data:match("Created: (.+)")
                            if match then
                                created_file = match
                            end
                        end,
                        on_exit = function(_job, return_val)
                            vim.schedule(function()
                                if return_val == 0 then
                                    print("Created from template: " .. tmpl.name)
                                    if created_file ~= "" then
                                        vim.cmd("edit " .. vim.fn.fnameescape(created_file))
                                    end
                                else
                                    print("Failed to create note")
                                end
                            end)
                        end,
                    })
                    job:start()
                end
            end)
            return true
        end,
    }):find()
end

M.tether_note = function(file_path, project)
    file_path = file_path or vim.fn.expand("%:p")
    if file_path == "" then
        print("No file specified")
        return
    end

    local cwd = vim.fn.getcwd()
    local args = { "tether", file_path }

    if project and project ~= "" then
        table.insert(args, "--project")
        table.insert(args, project)
    end

    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = args,
        cwd = cwd,
        on_exit = function(_job, return_val)
            if return_val == 0 then
                print("Note tethered")
                vim.schedule(function()
                    if vim.fn.expand("%:p") == file_path then
                        vim.cmd("edit!")
                    end
                end)
            else
                print("Failed to tether note (project may be required)")
            end
        end,
    })
    job:start()
end

M.untether_note = function(file_path)
    file_path = file_path or vim.fn.expand("%:p")
    if file_path == "" then
        print("No file specified")
        return
    end

    local cwd = vim.fn.getcwd()
    local args = { "untether", file_path }

    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = args,
        cwd = cwd,
        on_exit = function(_job, return_val)
            if return_val == 0 then
                print("Note untethered")
                vim.schedule(function()
                    if vim.fn.expand("%:p") == file_path then
                        vim.cmd("edit!")
                    end
                end)
            else
                print("Failed to untether note")
            end
        end,
    })
    job:start()
end

M.set_project = function(file_path, project)
    file_path = file_path or vim.fn.expand("%:p")
    if file_path == "" then
        print("No file specified")
        return
    end

    if not project or project == "" then
        project = vim.fn.input("Project: ")
        if project == "" then return end
    end

    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = { "set-project", file_path, project },
        on_exit = function(_job, return_val)
            if return_val == 0 then
                print("Project updated to: " .. project)
                vim.schedule(function()
                    if vim.fn.expand("%:p") == file_path then
                        vim.cmd("edit!")
                    end
                end)
            else
                print("Failed to set project")
            end
        end,
    })
    job:start()
end

-- Search zettels (requires results to be displayed somehow)
M.search = function(query, opts)
    opts = opts or {}
    local args = { "search", "--json" }

    if query and query ~= "" then
        table.insert(args, query)
    end

    if opts.project and opts.project ~= "" then
        table.insert(args, "--project")
        table.insert(args, opts.project)
    end

    if opts.category and opts.category ~= "" then
        table.insert(args, "--category")
        table.insert(args, opts.category)
    end

    if opts.tags and type(opts.tags) == "table" then
        for _, tag in ipairs(opts.tags) do
            table.insert(args, "--tag")
            table.insert(args, tag)
        end
    end

    if opts.limit then
        table.insert(args, "--limit")
        table.insert(args, tostring(opts.limit))
    end

    local results_json = ""
    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = args,
        on_stdout = function(_, data)
            results_json = results_json .. data
        end,
        on_exit = function(_job, return_val)
            vim.schedule(function()
                if return_val == 0 then
                    local ok, results = pcall(vim.json.decode, results_json)
                    if ok and opts.on_results then
                        opts.on_results(results)
                    elseif ok then
                        -- Default: print results
                        if #results == 0 then
                            print("No results found")
                        else
                            for _, r in ipairs(results) do
                                print(string.format("%s: %s [%s]", r.id, r.title, r.category))
                            end
                        end
                    end
                else
                    print("Search failed")
                end
            end)
        end,
    })
    job:start()
end

-- Index files
M.index = function(path)
    path = path or vim.fn.getcwd()

    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = { "index", path },
        on_exit = function(j, return_val)
            vim.schedule(function()
                if return_val == 0 then
                    local result = table.concat(j:result(), "\n")
                    print(result)
                else
                    print("Failed to index")
                end
            end)
        end,
    })
    job:start()
end

-- Index a single file silently (for auto-indexing on save)
M.index_file = function(path)
    if not path or path == "" then
        return
    end

    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = { "index", path },
        -- Silent - no output on success
        on_exit = function(_, return_val)
            if return_val ~= 0 then
                vim.schedule(function()
                    vim.notify("zk: failed to index " .. vim.fn.fnamemodify(path, ":t"), vim.log.levels.WARN)
                end)
            end
        end,
    })
    job:start()
end

-- Generate graph visualization
M.graph = function(opts)
    opts = opts or {}
    local path = opts.path or vim.fn.getcwd()
    local limit = opts.limit or 10
    local html = opts.html ~= false -- default true

    local args = { "graph", path, "--limit", tostring(limit) }

    if opts.output and opts.output ~= "" then
        table.insert(args, "--output")
        table.insert(args, opts.output)
    end

    local output_file = ""
    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = args,
        on_stdout = function(_, data)
            -- Capture output path from CLI
            local match = data:match("Output: (.+)")
            if match then
                output_file = match
            end
        end,
        on_exit = function(_job, return_val)
            vim.schedule(function()
                if return_val ~= 0 then
                    vim.notify("Failed to generate graph", vim.log.levels.ERROR)
                    return
                end

                if output_file == "" or vim.fn.filereadable(output_file) ~= 1 then
                    vim.notify("Graph generated but output file not found", vim.log.levels.WARN)
                    return
                end

                if html and vim.fn.exists(":MarkdownPreview") == 2 then
                    vim.cmd("edit " .. vim.fn.fnameescape(output_file))
                    vim.cmd("MarkdownPreview")
                    vim.notify("Graph preview started — open http://localhost:8080 in your browser")
                else
                    vim.cmd("vsplit " .. vim.fn.fnameescape(output_file))
                    if html then
                        vim.notify("Install markdown-preview.nvim for browser rendering", vim.log.levels.INFO)
                    end
                end
            end)
        end,
    })
    job:start()
end

-- Preview note in floating window
M.preview_note = function(file_path)
    file_path = file_path or vim.fn.expand("%:p")
    if file_path == "" then
        print("No file specified")
        return
    end

    -- Check file exists
    if vim.fn.filereadable(file_path) ~= 1 then
        print("File not found: " .. file_path)
        return
    end

    -- Read file content
    local lines = vim.fn.readfile(file_path)
    if #lines == 0 then
        print("File is empty")
        return
    end

    -- Calculate window dimensions
    local width = math.min(100, vim.o.columns - 10)
    local height = math.min(#lines + 2, vim.o.lines - 10)
    local row = math.floor((vim.o.lines - height) / 2)
    local col = math.floor((vim.o.columns - width) / 2)

    -- Create buffer
    local buf = vim.api.nvim_create_buf(false, true)
    vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
    vim.api.nvim_buf_set_option(buf, "modifiable", false)
    vim.api.nvim_buf_set_option(buf, "bufhidden", "wipe")
    vim.api.nvim_buf_set_option(buf, "filetype", "markdown")

    -- Create floating window
    local win = vim.api.nvim_open_win(buf, true, {
        relative = "editor",
        width = width,
        height = height,
        row = row,
        col = col,
        style = "minimal",
        border = "rounded",
        title = " " .. vim.fn.fnamemodify(file_path, ":t") .. " ",
        title_pos = "center",
    })

    -- Set window options
    vim.api.nvim_win_set_option(win, "wrap", true)
    vim.api.nvim_win_set_option(win, "cursorline", true)

    -- Store file path for opening
    vim.b[buf].zk_preview_path = file_path

    -- Keymaps for the preview window
    local close_preview = function()
        if vim.api.nvim_win_is_valid(win) then
            vim.api.nvim_win_close(win, true)
        end
    end

    local open_note = function()
        local path = vim.b[buf].zk_preview_path
        close_preview()
        vim.cmd("edit " .. vim.fn.fnameescape(path))
    end

    vim.keymap.set("n", "q", close_preview, { buffer = buf, nowait = true })
    vim.keymap.set("n", "<Esc>", close_preview, { buffer = buf, nowait = true })
    vim.keymap.set("n", "<CR>", open_note, { buffer = buf, nowait = true })

    return { buf = buf, win = win }
end

-- Preview note by ID (searches for matching file)
M.preview_by_id = function(id)
    if not id or id == "" then
        id = vim.fn.input("Note ID: ")
        if id == "" then return end
    end

    -- Search for the note
    M.search("", {
        limit = 1,
        on_results = function(results)
            -- Filter by ID
            for _, r in ipairs(results) do
                if r.id == id then
                    M.preview_note(r.file_path)
                    return
                end
            end
            print("Note not found: " .. id)
        end,
    })
end

-- Insert a zettel link at cursor position
-- Format: [[id]] or [[id|title]]
M.insert_link = function(id, title, include_title)
    if not id or id == "" then
        print("No ID provided")
        return
    end

    -- Prevent self-linking
    local current_id = get_current_zettel_id()
    if current_id and current_id == id then
        vim.notify("Cannot link a note to itself", vim.log.levels.WARN)
        return
    end

    local link
    if include_title and title and title ~= "" then
        link = string.format("[[%s|%s]]", id, title)
    else
        link = string.format("[[%s]]", id)
    end

    -- Insert at cursor position
    local row, col = unpack(vim.api.nvim_win_get_cursor(0))
    local line = vim.api.nvim_get_current_line()
    local new_line = line:sub(1, col) .. link .. line:sub(col + 1)
    vim.api.nvim_set_current_line(new_line)

    -- Move cursor to end of inserted link
    vim.api.nvim_win_set_cursor(0, { row, col + #link })
end

-- Prompt for ID and insert link
M.insert_link_prompt = function(include_title)
    local id = vim.fn.input("Note ID: ")
    if id == "" then return end

    if include_title then
        -- Search for title
        M.search("", {
            limit = 100,
            on_results = function(results)
                for _, r in ipairs(results) do
                    if r.id == id then
                        M.insert_link(id, r.title, true)
                        return
                    end
                end
                -- Not found, insert without title
                M.insert_link(id, nil, false)
            end,
        })
    else
        M.insert_link(id, nil, false)
    end
end

-- Open picker to search notes and insert link (requires Telescope)
M.link_picker = function(opts)
    opts = opts or {}
    local include_title = opts.include_title

    local has_telescope = pcall(require, "telescope")
    if not has_telescope then
        print("Telescope required for link picker. Use insert_link_prompt() instead.")
        return
    end

    local pickers = require("telescope.pickers")
    local finders = require("telescope.finders")
    local conf = require("telescope.config").values
    local actions = require("telescope.actions")
    local action_state = require("telescope.actions.state")
    local previewers = require("telescope.previewers")

    -- Get search results
    local Job = require("plenary.job")
    local results_json = ""

    local args = { "search", "--json", "--limit", "50" }
    if opts.query and opts.query ~= "" then
        table.insert(args, opts.query)
    end

    local job = Job:new({
        command = M.config.bin,
        args = args,
        on_stdout = function(_, data)
            results_json = results_json .. data
        end,
    })

    job:sync(10000)

    local ok, results = pcall(vim.json.decode, results_json)
    if not ok or type(results) ~= "table" then
        vim.notify("No zettels found to link", vim.log.levels.INFO)
        return
    end

    -- Filter out the current note to prevent self-linking
    local current_id = get_current_zettel_id()
    if current_id then
        local filtered = {}
        for _, r in ipairs(results) do
            if r.id ~= current_id then
                table.insert(filtered, r)
            end
        end
        results = filtered
    end

    if #results == 0 then
        vim.notify("No zettels found to link", vim.log.levels.INFO)
        return
    end

    pickers.new(opts, {
        prompt_title = "Insert Link to Zettel",
        finder = finders.new_table({
            results = results,
            entry_maker = function(entry)
                local display = entry.title or entry.id
                if entry.project and entry.project ~= "" then
                    display = display .. " [" .. entry.project .. "]"
                end

                return {
                    value = entry,
                    display = display,
                    ordinal = (entry.title or "") .. " " .. (entry.id or ""),
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
                end
            end,
        }),
        attach_mappings = function(prompt_bufnr, map)
            -- Default: insert [[id]]
            actions.select_default:replace(function()
                actions.close(prompt_bufnr)
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    local entry = selection.value
                    M.insert_link(entry.id, entry.title, include_title)
                end
            end)

            -- Ctrl-t: insert [[id|title]]
            map("i", "<C-t>", function()
                actions.close(prompt_bufnr)
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    local entry = selection.value
                    M.insert_link(entry.id, entry.title, true)
                end
            end)
            map("n", "<C-t>", function()
                actions.close(prompt_bufnr)
                local selection = action_state.get_selected_entry()
                if selection and selection.value then
                    local entry = selection.value
                    M.insert_link(entry.id, entry.title, true)
                end
            end)

            return true
        end,
    }):find()
end

-- Tag completion support
-- Cache for tags to avoid repeated searches
M._tag_cache = nil
M._tag_cache_time = 0
M._tag_cache_ttl = 60 -- Cache TTL in seconds

-- Get all unique tags from indexed zettels (async)
M.get_tags = function(callback)
    local Job = require("plenary.job")
    local results_json = ""

    local job = Job:new({
        command = M.config.bin,
        args = { "search", "--json", "--limit", "500" },
        on_stdout = function(_, data)
            results_json = results_json .. data
        end,
        on_exit = function(_, return_val)
            vim.schedule(function()
                if return_val ~= 0 then
                    callback({})
                    return
                end

                local ok, results = pcall(vim.json.decode, results_json)
                if not ok then
                    callback({})
                    return
                end

                -- Collect unique tags
                local tag_set = {}
                for _, r in ipairs(results) do
                    if r.tags and type(r.tags) == "table" then
                        for _, tag in ipairs(r.tags) do
                            tag_set[tag] = true
                        end
                    end
                end

                -- Convert to sorted list
                local tags = {}
                for tag in pairs(tag_set) do
                    table.insert(tags, tag)
                end
                table.sort(tags)

                -- Update cache
                M._tag_cache = tags
                M._tag_cache_time = os.time()

                callback(tags)
            end)
        end,
    })
    job:start()
end

-- Get tags synchronously (for completion)
M.get_tags_sync = function()
    -- Check cache
    local now = os.time()
    if M._tag_cache and (now - M._tag_cache_time) < M._tag_cache_ttl then
        return M._tag_cache
    end

    local Job = require("plenary.job")
    local results_json = ""

    local job = Job:new({
        command = M.config.bin,
        args = { "search", "--json", "--limit", "500" },
        on_stdout = function(_, data)
            results_json = results_json .. data
        end,
    })

    job:sync(5000)

    local ok, results = pcall(vim.json.decode, results_json)
    if not ok then
        return M._tag_cache or {}
    end

    -- Collect unique tags
    local tag_set = {}
    for _, r in ipairs(results) do
        if r.tags and type(r.tags) == "table" then
            for _, tag in ipairs(r.tags) do
                tag_set[tag] = true
            end
        end
    end

    -- Convert to sorted list
    local tags = {}
    for tag in pairs(tag_set) do
        table.insert(tags, tag)
    end
    table.sort(tags)

    -- Update cache
    M._tag_cache = tags
    M._tag_cache_time = now

    return tags
end

-- Refresh tag cache
M.refresh_tags = function()
    M._tag_cache = nil
    M._tag_cache_time = 0
    M.get_tags(function(tags)
        print("Refreshed " .. #tags .. " tags")
    end)
end

-- Manual tag completion using vim's completion menu
M.complete_tags = function()
    local tags = M.get_tags_sync()
    if #tags == 0 then
        print("No tags found. Run :lua require('zk').index() first.")
        return
    end

    -- Get current line context
    local line = vim.api.nvim_get_current_line()
    local col = vim.api.nvim_win_get_cursor(0)[2]
    local before_cursor = line:sub(1, col)

    -- Find the start of the current word
    local start_col = col
    for i = col, 1, -1 do
        local char = before_cursor:sub(i, i)
        if char:match("[%s,\"']") then
            break
        end
        start_col = i - 1
    end

    -- Get partial input for filtering
    local partial = before_cursor:sub(start_col + 1, col)

    -- Filter tags by partial match
    local matches = {}
    for _, tag in ipairs(tags) do
        if partial == "" or tag:lower():find(partial:lower(), 1, true) then
            table.insert(matches, tag)
        end
    end

    if #matches == 0 then
        print("No matching tags")
        return
    end

    -- Show completion menu
    vim.fn.complete(start_col + 1, matches)
end

-- Check if cursor is in YAML frontmatter tags section
local function in_tags_section()
    local bufnr = vim.api.nvim_get_current_buf()
    local row = vim.api.nvim_win_get_cursor(0)[1]
    local lines = vim.api.nvim_buf_get_lines(bufnr, 0, row, false)

    local in_frontmatter = false
    local frontmatter_end = false
    local in_tags = false

    for i, line in ipairs(lines) do
        if i == 1 and line == "---" then
            in_frontmatter = true
        elseif in_frontmatter and line == "---" then
            frontmatter_end = true
            break
        elseif in_frontmatter then
            if line:match("^tags:") then
                in_tags = true
            elseif line:match("^%w+:") and not line:match("^%s+-") then
                in_tags = false
            end
        end
    end

    return in_frontmatter and not frontmatter_end and in_tags
end

-- Omni completion function for tags
M.omnifunc = function(findstart, base)
    if findstart == 1 then
        -- Find the start of the word
        local line = vim.api.nvim_get_current_line()
        local col = vim.api.nvim_win_get_cursor(0)[2]
        local start = col

        while start > 0 do
            local char = line:sub(start, start)
            if char:match("[%s,\"'-]") then
                break
            end
            start = start - 1
        end

        return start
    else
        -- Return matching tags
        local tags = M.get_tags_sync()
        local matches = {}

        for _, tag in ipairs(tags) do
            if base == "" or tag:lower():find(base:lower(), 1, true) then
                table.insert(matches, {
                    word = tag,
                    menu = "[zk-tag]",
                })
            end
        end

        return matches
    end
end

-- Set up tag completion for markdown files
M.setup_tag_completion = function()
    vim.api.nvim_create_autocmd("FileType", {
        pattern = "markdown",
        callback = function()
            -- Set omnifunc for manual completion
            vim.bo.omnifunc = "v:lua.require'zk'.omnifunc"

            -- Set up <C-x><C-o> for omni completion in insert mode
            vim.keymap.set("i", "<C-x><C-t>", function()
                if in_tags_section() then
                    require("zk").complete_tags()
                else
                    -- Fall back to default behavior
                    vim.api.nvim_feedkeys(
                        vim.api.nvim_replace_termcodes("<C-x><C-o>", true, false, true),
                        "n",
                        false
                    )
                end
            end, { buffer = true, desc = "Complete zk tags" })
        end,
    })
end

-- nvim-cmp source for tag completion (if nvim-cmp is available)
M.cmp_source = function()
    local source = {}

    source.new = function()
        return setmetatable({}, { __index = source })
    end

    source.get_trigger_characters = function()
        return { "-", " " }
    end

    source.is_available = function()
        -- Only available in markdown files in tags section
        if vim.bo.filetype ~= "markdown" then
            return false
        end
        return in_tags_section()
    end

    source.complete = function(self, request, callback)
        local tags = M.get_tags_sync()
        local items = {}

        for _, tag in ipairs(tags) do
            table.insert(items, {
                label = tag,
                kind = require("cmp").lsp.CompletionItemKind.Keyword,
                documentation = "Zettel tag",
            })
        end

        callback({ items = items })
    end

    return source
end

-- Register nvim-cmp source if available
M.setup_cmp = function()
    local has_cmp, cmp = pcall(require, "cmp")
    if not has_cmp then
        return false
    end

    cmp.register_source("zk_tags", M.cmp_source().new())
    return true
end

-- Backlinks panel
-- State for the backlinks panel
M._backlinks_state = {
    buf = nil,
    win = nil,
    backlinks = {},
}

-- Get backlinks for a zettel (async)
M.get_backlinks = function(id_or_file, callback)
    local Job = require("plenary.job")
    local results_json = ""

    local job = Job:new({
        command = M.config.bin,
        args = { "backlinks", id_or_file, "--json" },
        on_stdout = function(_, data)
            results_json = results_json .. data
        end,
        on_exit = function(_, return_val)
            vim.schedule(function()
                if return_val ~= 0 then
                    callback({})
                    return
                end

                local ok, results = pcall(vim.json.decode, results_json)
                if not ok or type(results) ~= "table" then
                    callback({})
                    return
                end

                callback(results)
            end)
        end,
    })
    job:start()
end

-- Get backlinks synchronously
M.get_backlinks_sync = function(id_or_file)
    local Job = require("plenary.job")
    local results_json = ""

    local job = Job:new({
        command = M.config.bin,
        args = { "backlinks", id_or_file, "--json" },
        on_stdout = function(_, data)
            results_json = results_json .. data
        end,
    })

    job:sync(5000)

    local ok, results = pcall(vim.json.decode, results_json)
    if not ok or type(results) ~= "table" then
        return {}
    end

    return results
end

-- Extract zettel ID from current buffer frontmatter
local function get_current_zettel_id()
    local lines = vim.api.nvim_buf_get_lines(0, 0, 20, false)
    local in_frontmatter = false

    for i, line in ipairs(lines) do
        if i == 1 and line == "---" then
            in_frontmatter = true
        elseif in_frontmatter and line == "---" then
            break
        elseif in_frontmatter then
            local id = line:match('^id:%s*"?([%w%-]+)"?')
            if id then
                return id
            end
        end
    end

    return nil
end

-- Open backlinks panel as a floating window
M.backlinks_panel = function(opts)
    opts = opts or {}

    -- Save all modified zettel buffers so backlinks scan finds latest links
    for _, bufnr in ipairs(vim.api.nvim_list_bufs()) do
        if vim.api.nvim_buf_is_loaded(bufnr) and vim.bo[bufnr].modified then
            local name = vim.api.nvim_buf_get_name(bufnr)
            if name:match("%.md$") then
                vim.api.nvim_buf_call(bufnr, function() vim.cmd("silent write") end)
            end
        end
    end

    -- Get the zettel ID
    local target = opts.id or opts.file or get_current_zettel_id()
    if not target then
        print("No zettel ID found. Open a zettel file first.")
        return
    end

    -- Get backlinks
    local backlinks = M.get_backlinks_sync(target)
    M._backlinks_state.backlinks = backlinks

    -- Build panel content
    local lines = {}
    table.insert(lines, " Backlinks (" .. #backlinks .. ")")
    table.insert(lines, string.rep("─", 40))

    if #backlinks == 0 then
        table.insert(lines, "")
        table.insert(lines, "  No backlinks found")
        table.insert(lines, "")
    else
        for i, bl in ipairs(backlinks) do
            table.insert(lines, "")
            local title = bl.title or bl.id
            if #title > 35 then
                title = title:sub(1, 32) .. "..."
            end
            table.insert(lines, string.format(" %d. %s", i, title))
            table.insert(lines, string.format("    [%s] %s", bl.category, bl.id))
        end
        table.insert(lines, "")
    end

    table.insert(lines, string.rep("─", 40))
    table.insert(lines, " <CR> open │ q close │ p preview")

    -- Calculate window dimensions
    local width = 44
    local height = math.min(#lines, vim.o.lines - 10)
    local row = 2
    local col = vim.o.columns - width - 2

    -- Close existing panel if open
    if M._backlinks_state.win and vim.api.nvim_win_is_valid(M._backlinks_state.win) then
        vim.api.nvim_win_close(M._backlinks_state.win, true)
    end

    -- Create buffer
    local buf = vim.api.nvim_create_buf(false, true)
    vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
    vim.api.nvim_buf_set_option(buf, "modifiable", false)
    vim.api.nvim_buf_set_option(buf, "bufhidden", "wipe")
    vim.api.nvim_buf_set_option(buf, "buftype", "nofile")
    vim.api.nvim_buf_set_name(buf, "zk-backlinks")

    -- Create floating window
    local win = vim.api.nvim_open_win(buf, true, {
        relative = "editor",
        width = width,
        height = height,
        row = row,
        col = col,
        style = "minimal",
        border = "rounded",
        title = " Backlinks ",
        title_pos = "center",
    })

    -- Set window options
    vim.api.nvim_win_set_option(win, "wrap", true)
    vim.api.nvim_win_set_option(win, "cursorline", true)

    -- Store state
    M._backlinks_state.buf = buf
    M._backlinks_state.win = win

    -- Helper to get selected backlink index from cursor position
    local function get_selected_index()
        local cursor_row = vim.api.nvim_win_get_cursor(win)[1]
        -- Find which backlink the cursor is on
        local current_index = 0
        local line_num = 0
        for _, line in ipairs(lines) do
            line_num = line_num + 1
            local idx = line:match("^ (%d+)%.")
            if idx then
                current_index = tonumber(idx)
            end
            if line_num == cursor_row and current_index > 0 then
                return current_index
            end
        end
        return current_index
    end

    -- Keymaps
    local close_panel = function()
        if vim.api.nvim_win_is_valid(win) then
            vim.api.nvim_win_close(win, true)
        end
        M._backlinks_state.win = nil
        M._backlinks_state.buf = nil
    end

    local open_selected = function()
        local idx = get_selected_index()
        if idx > 0 and idx <= #backlinks then
            local bl = backlinks[idx]
            close_panel()
            vim.cmd("edit " .. vim.fn.fnameescape(bl.file_path))
        end
    end

    local preview_selected = function()
        local idx = get_selected_index()
        if idx > 0 and idx <= #backlinks then
            local bl = backlinks[idx]
            M.preview_note(bl.file_path)
        end
    end

    vim.keymap.set("n", "q", close_panel, { buffer = buf, nowait = true })
    vim.keymap.set("n", "<Esc>", close_panel, { buffer = buf, nowait = true })
    vim.keymap.set("n", "<CR>", open_selected, { buffer = buf, nowait = true })
    vim.keymap.set("n", "p", preview_selected, { buffer = buf, nowait = true })
    vim.keymap.set("n", "o", open_selected, { buffer = buf, nowait = true })

    -- Position cursor on first backlink if any
    if #backlinks > 0 then
        vim.api.nvim_win_set_cursor(win, { 4, 0 })
    end

    return { buf = buf, win = win, backlinks = backlinks }
end

-- Open backlinks in split window
M.backlinks_split = function(opts)
    opts = opts or {}
    local position = opts.position or "right"

    -- Get the zettel ID
    local target = opts.id or opts.file or get_current_zettel_id()
    if not target then
        print("No zettel ID found. Open a zettel file first.")
        return
    end

    -- Get backlinks
    local backlinks = M.get_backlinks_sync(target)

    -- Build content
    local lines = {}
    table.insert(lines, "# Backlinks")
    table.insert(lines, "")
    table.insert(lines, "Notes linking to: " .. target)
    table.insert(lines, "")

    if #backlinks == 0 then
        table.insert(lines, "*No backlinks found*")
    else
        for _, bl in ipairs(backlinks) do
            local title = bl.title or bl.id
            table.insert(lines, string.format("- **%s** `%s`", title, bl.id))
            table.insert(lines, string.format("  [%s] %s", bl.category, bl.file_path))
            table.insert(lines, "")
        end
    end

    -- Create split
    if position == "right" then
        vim.cmd("vsplit")
    elseif position == "left" then
        vim.cmd("leftabove vsplit")
    elseif position == "bottom" then
        vim.cmd("split")
    else
        vim.cmd("aboveleft split")
    end

    -- Create buffer
    local buf = vim.api.nvim_create_buf(false, true)
    vim.api.nvim_win_set_buf(0, buf)
    vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
    vim.api.nvim_buf_set_option(buf, "modifiable", false)
    vim.api.nvim_buf_set_option(buf, "bufhidden", "wipe")
    vim.api.nvim_buf_set_option(buf, "buftype", "nofile")
    vim.api.nvim_buf_set_option(buf, "filetype", "markdown")
    vim.api.nvim_buf_set_name(buf, "zk-backlinks-" .. target)

    -- Resize to reasonable width
    if position == "right" or position == "left" then
        vim.cmd("vertical resize 50")
    else
        vim.cmd("resize 15")
    end

    return { buf = buf, backlinks = backlinks }
end

-- Toggle backlinks panel
M.toggle_backlinks = function(opts)
    if M._backlinks_state.win and vim.api.nvim_win_is_valid(M._backlinks_state.win) then
        vim.api.nvim_win_close(M._backlinks_state.win, true)
        M._backlinks_state.win = nil
        M._backlinks_state.buf = nil
    else
        M.backlinks_panel(opts)
    end
end

-- Telescope integration (lazy loaded)
M.telescope = setmetatable({}, {
    __index = function(_, key)
        local ok, telescope_ext = pcall(require, "zk.telescope")
        if ok then
            return telescope_ext[key]
        end
        return function()
            print("Telescope not available. Install telescope.nvim for this feature.")
        end
    end,
})

return M
