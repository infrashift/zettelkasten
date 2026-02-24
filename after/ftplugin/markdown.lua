-- zk.nvim - Markdown-specific settings
-- This file is auto-loaded for markdown files

-- Only set up if zk plugin is loaded
if not vim.g.loaded_zk then
    return
end

-- Check if this looks like a zettel (has frontmatter with id field)
local function is_zettel()
    local lines = vim.api.nvim_buf_get_lines(0, 0, 10, false)
    if #lines < 2 then return false end
    if lines[1] ~= "---" then return false end

    for i = 2, math.min(10, #lines) do
        if lines[i]:match("^id:") then
            return true
        end
        if lines[i] == "---" then
            break
        end
    end
    return false
end

-- Get the type field from frontmatter (e.g. "todo", "note", "daily-note")
local function get_type()
    local lines = vim.api.nvim_buf_get_lines(0, 0, 15, false)
    if #lines < 2 then return nil end
    if lines[1] ~= "---" then return nil end

    for i = 2, math.min(15, #lines) do
        if lines[i] == "---" then
            break
        end
        local val = lines[i]:match("^type:%s*(.+)$")
        if val then
            -- Strip surrounding quotes and whitespace
            val = vim.trim(val)
            val = val:gsub('^"(.*)"$', '%1'):gsub("^'(.*)'$", '%1')
            return val
        end
    end
    return nil
end

-- Set up buffer-local settings for zettel files
if is_zettel() then
    -- Set omnifunc for tag completion
    vim.bo.omnifunc = "v:lua.require'zk'.omnifunc"

    -- Auto-index on save
    vim.api.nvim_create_autocmd("BufWritePost", {
        buffer = 0,
        callback = function()
            local file_path = vim.fn.expand("%:p")
            require("zk").index_file(file_path)
        end,
        desc = "Auto-index zettel on save",
    })

    -- Buffer-local keymaps for zettel files
    local opts = { buffer = true, silent = true }

    -- Tag completion in insert mode
    vim.keymap.set("i", "<C-x><C-t>", function()
        require("zk").complete_tags()
    end, vim.tbl_extend("force", opts, { desc = "Complete zk tags" }))

    -- Quick link insertion
    vim.keymap.set("n", "<localleader>l", function()
        require("zk").link_picker()
    end, vim.tbl_extend("force", opts, { desc = "Insert zettel link" }))

    vim.keymap.set("n", "<localleader>L", function()
        require("zk").link_picker({ include_title = true })
    end, vim.tbl_extend("force", opts, { desc = "Insert zettel link with title" }))

    -- Toggle backlinks panel
    vim.keymap.set("n", "<localleader>b", function()
        require("zk").toggle_backlinks()
    end, vim.tbl_extend("force", opts, { desc = "Toggle backlinks" }))

    -- Add tags
    vim.keymap.set("n", "<localleader>a", function()
        vim.ui.input({ prompt = "Tags (space-separated): " }, function(input)
            if input and input ~= "" then
                require("zk").add_tags(input)
            end
        end)
    end, vim.tbl_extend("force", opts, { desc = "Add tags" }))

    -- Validate frontmatter
    vim.keymap.set("n", "<localleader>v", function()
        require("zk").validate()
    end, vim.tbl_extend("force", opts, { desc = "Validate frontmatter" }))

    local ztype = get_type()

    -- Set project (note and todo types)
    if ztype == "note" or ztype == "todo" then
        vim.keymap.set("n", "<localleader>p", function()
            require("zk").set_project()
        end, vim.tbl_extend("force", opts, { desc = "Set project" }))

        vim.keymap.set("n", "<localleader>t", function()
            vim.ui.select(
                { "tether", "untether" },
                { prompt = "Tether / Untether:" },
                function(choice)
                    if choice == "tether" then
                        require("zk").tether_note()
                    elseif choice == "untether" then
                        require("zk").untether_note()
                    end
                end
            )
        end, vim.tbl_extend("force", opts, { desc = "Tether / Untether" }))
    end

    -- Status picker for todo-type zettels
    if ztype == "todo" then
        vim.keymap.set("n", "<localleader>s", function()
            vim.ui.select(
                { "open", "in_progress", "closed" },
                { prompt = "Set todo status:" },
                function(choice)
                    if choice then
                        require("zk").set_status(choice)
                    end
                end
            )
        end, vim.tbl_extend("force", opts, { desc = "Set todo status" }))
    end
end
