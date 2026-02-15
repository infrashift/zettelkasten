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

    -- Preview current note
    vim.keymap.set("n", "<localleader>p", function()
        require("zk").preview_note()
    end, vim.tbl_extend("force", opts, { desc = "Preview note" }))

    -- Toggle backlinks panel
    vim.keymap.set("n", "<localleader>b", function()
        require("zk").toggle_backlinks()
    end, vim.tbl_extend("force", opts, { desc = "Toggle backlinks" }))

    -- Promote note
    vim.keymap.set("n", "<localleader>P", function()
        require("zk").promote_note()
    end, vim.tbl_extend("force", opts, { desc = "Promote to permanent" }))
end
