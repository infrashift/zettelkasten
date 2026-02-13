local M = {}

M.setup = function(opts)
    M.config = vim.tbl_deep_extend("force", {
        bin = "zk",
    }, opts or {})
end

M.create_note = function(note_type)
    local title = vim.fn.input("Note Title: ")
    if title == "" then return end

    -- Pass current working directory to detect git repo
    local cwd = vim.fn.getcwd()
    
    local job = require('plenary.job'):new({
        command = M.config.bin,
        args = { "create", title, "--type", note_type },
        cwd = cwd,
        on_exit = function(j, return_val)
            if return_val == 0 then
                print("Note created successfully")
            end
        end,
    })
    job:start()
end

return M