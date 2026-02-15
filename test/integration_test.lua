-- Integration test for zk NeoVim plugin
-- Run with: nvim --headless -u NONE -c "luafile test/integration_test.lua"

local test_passed = true
local test_output = {}

local function log(msg)
    table.insert(test_output, msg)
    print(msg)
end

local function assert_eq(actual, expected, msg)
    if actual ~= expected then
        log(string.format("FAIL: %s - expected '%s', got '%s'", msg, tostring(expected), tostring(actual)))
        test_passed = false
    else
        log(string.format("PASS: %s", msg))
    end
end

local function assert_true(condition, msg)
    if not condition then
        log(string.format("FAIL: %s", msg))
        test_passed = false
    else
        log(string.format("PASS: %s", msg))
    end
end

-- Setup runtime path
local plugin_path = vim.fn.getcwd() .. "/lua"
local plenary_path = "/tmp/nvim-test-plugins/plenary.nvim/lua"
vim.opt.runtimepath:append(vim.fn.getcwd())
vim.opt.runtimepath:append("/tmp/nvim-test-plugins/plenary.nvim")

log("=== ZK Plugin Integration Test ===")
log("")

-- Test 1: Plugin loads successfully
log("Test 1: Plugin loads")
local ok, zk = pcall(require, "zk")
assert_true(ok, "Plugin loads without error")
if not ok then
    log("Error: " .. tostring(zk))
    vim.cmd("cq 1")
end

-- Test 2: Setup function works
log("")
log("Test 2: Setup function")
local setup_ok, setup_err = pcall(function()
    zk.setup({ bin = "./zk" })
end)
assert_true(setup_ok, "Setup completes without error")
if not setup_ok then
    log("Error: " .. tostring(setup_err))
end

-- Test 3: Config is set correctly
log("")
log("Test 3: Config values")
assert_eq(zk.config.bin, "./zk", "Binary path is set correctly")

-- Test 4: Plenary.job is available
log("")
log("Test 4: Plenary dependency")
local plenary_ok, plenary_job = pcall(function()
    return require("plenary.job")
end)
assert_true(plenary_ok, "Plenary.job loads successfully")

-- Test 5: Integration test - actually run zk via plenary.job
log("")
log("Test 5: CLI integration via plenary.job")

if plenary_ok then
    local Job = require("plenary.job")
    local job_output = {}
    local job_code = nil

    local job = Job:new({
        command = "./zk",
        args = { "create", "Integration Test Note", "--type", "fleeting" },
        cwd = vim.fn.getcwd(),
        on_stdout = function(_, data)
            table.insert(job_output, data)
        end,
        on_exit = function(_, return_val)
            job_code = return_val
        end,
    })

    job:sync(5000) -- Wait up to 5 seconds

    assert_eq(job_code, 0, "zk create exits with code 0")
    assert_true(#job_output > 0 or job_code == 0, "zk create produces output or succeeds")

    if #job_output > 0 then
        log("  CLI output: " .. job_output[1])
    end
end

-- Test 6: Verify create_note function exists and is callable
log("")
log("Test 6: create_note function")
assert_true(type(zk.create_note) == "function", "create_note is a function")

-- Test 7: Verify promote_note function exists
log("")
log("Test 7: promote_note function")
assert_true(type(zk.promote_note) == "function", "promote_note is a function")

-- Test 8: Verify set_project function exists
log("")
log("Test 8: set_project function")
assert_true(type(zk.set_project) == "function", "set_project is a function")

-- Test 9: Test promote command via CLI
log("")
log("Test 9: CLI promote command")
if plenary_ok then
    local Job = require("plenary.job")

    -- First copy the test file
    os.execute("cp testdata/valid_fleeting_no_project.md /tmp/test_promote_integration.md")

    local promote_output = {}
    local promote_code = nil

    local promote_job = Job:new({
        command = "./zk",
        args = { "promote", "/tmp/test_promote_integration.md", "--project", "integration-test" },
        cwd = vim.fn.getcwd(),
        on_stdout = function(_, data)
            table.insert(promote_output, data)
        end,
        on_exit = function(_, return_val)
            promote_code = return_val
        end,
    })

    promote_job:sync(5000)

    assert_eq(promote_code, 0, "zk promote exits with code 0")

    if #promote_output > 0 then
        log("  CLI output: " .. promote_output[1])
    end
end

-- Test 10: Test set-project command via CLI
log("")
log("Test 10: CLI set-project command")
if plenary_ok then
    local Job = require("plenary.job")

    local setproj_output = {}
    local setproj_code = nil

    local setproj_job = Job:new({
        command = "./zk",
        args = { "set-project", "/tmp/test_promote_integration.md", "new-project" },
        cwd = vim.fn.getcwd(),
        on_stdout = function(_, data)
            table.insert(setproj_output, data)
        end,
        on_exit = function(_, return_val)
            setproj_code = return_val
        end,
    })

    setproj_job:sync(5000)

    assert_eq(setproj_code, 0, "zk set-project exits with code 0")

    if #setproj_output > 0 then
        log("  CLI output: " .. setproj_output[1])
    end
end

-- Test 11: Test search command via CLI
log("")
log("Test 11: CLI search command")
if plenary_ok then
    local Job = require("plenary.job")

    local search_output = {}
    local search_code = nil

    local search_job = Job:new({
        command = "./zk",
        args = { "search", "--json" },
        cwd = vim.fn.getcwd(),
        on_stdout = function(_, data)
            table.insert(search_output, data)
        end,
        on_exit = function(_, return_val)
            search_code = return_val
        end,
    })

    search_job:sync(5000)

    assert_eq(search_code, 0, "zk search exits with code 0")

    -- Verify JSON output is parseable
    if #search_output > 0 then
        local json_str = table.concat(search_output, "")
        local ok, results = pcall(vim.json.decode, json_str)
        assert_true(ok, "zk search JSON output is valid")
        if ok then
            log("  Found " .. #results .. " results")
        end
    end
end

-- Test 12: Verify search and index functions exist
log("")
log("Test 12: search and index functions")
assert_true(type(zk.search) == "function", "search is a function")
assert_true(type(zk.index) == "function", "index is a function")

-- Test 13: Verify graph function exists
log("")
log("Test 13: graph function")
assert_true(type(zk.graph) == "function", "graph is a function")

-- Test 14: Verify preview functions exist
log("")
log("Test 14: preview functions")
assert_true(type(zk.preview_note) == "function", "preview_note is a function")
assert_true(type(zk.preview_by_id) == "function", "preview_by_id is a function")

-- Test 15: Verify link functions exist
log("")
log("Test 15: link functions")
assert_true(type(zk.insert_link) == "function", "insert_link is a function")
assert_true(type(zk.insert_link_prompt) == "function", "insert_link_prompt is a function")
assert_true(type(zk.link_picker) == "function", "link_picker is a function")

-- Test 16: Verify tag completion functions exist
log("")
log("Test 16: tag completion functions")
assert_true(type(zk.get_tags) == "function", "get_tags is a function")
assert_true(type(zk.get_tags_sync) == "function", "get_tags_sync is a function")
assert_true(type(zk.refresh_tags) == "function", "refresh_tags is a function")
assert_true(type(zk.complete_tags) == "function", "complete_tags is a function")
assert_true(type(zk.setup_tag_completion) == "function", "setup_tag_completion is a function")
assert_true(type(zk.setup_cmp) == "function", "setup_cmp is a function")
assert_true(type(zk.omnifunc) == "function", "omnifunc is a function")

-- Test 17: Verify backlinks functions exist
log("")
log("Test 17: backlinks functions")
assert_true(type(zk.get_backlinks) == "function", "get_backlinks is a function")
assert_true(type(zk.get_backlinks_sync) == "function", "get_backlinks_sync is a function")
assert_true(type(zk.backlinks_panel) == "function", "backlinks_panel is a function")
assert_true(type(zk.backlinks_split) == "function", "backlinks_split is a function")
assert_true(type(zk.toggle_backlinks) == "function", "toggle_backlinks is a function")

-- Test 18: CLI backlinks command
log("")
log("Test 18: CLI backlinks command")
if plenary_ok then
    local Job = require("plenary.job")
    local backlinks_output = {}
    local backlinks_code = nil

    local backlinks_job = Job:new({
        command = "./zk",
        args = { "backlinks", "202602131045", "--json" },
        cwd = vim.fn.getcwd(),
        on_stdout = function(_, data)
            table.insert(backlinks_output, data)
        end,
        on_exit = function(_, return_val)
            backlinks_code = return_val
        end,
    })

    backlinks_job:sync(5000)

    assert_eq(backlinks_code, 0, "zk backlinks exits with code 0")

    if #backlinks_output > 0 then
        local json_str = table.concat(backlinks_output, "")
        local ok, results = pcall(vim.json.decode, json_str)
        assert_true(ok, "zk backlinks JSON output is valid")
        if ok and type(results) == "table" then
            log("  Found " .. #results .. " backlink(s)")
        end
    end
end

-- Test 19: Verify template functions exist
log("")
log("Test 19: template functions")
assert_true(type(zk.templates) == "table", "templates is a table")
assert_true(type(zk.get_template) == "function", "get_template is a function")
assert_true(type(zk.create_from_template) == "function", "create_from_template is a function")
assert_true(type(zk.template_picker) == "function", "template_picker is a function")

-- Test 20: CLI templates command
log("")
log("Test 20: CLI templates command")
if plenary_ok then
    local Job = require("plenary.job")
    local templates_output = {}
    local templates_code = nil

    local templates_job = Job:new({
        command = "./zk",
        args = { "templates" },
        cwd = vim.fn.getcwd(),
        on_stdout = function(_, data)
            table.insert(templates_output, data)
        end,
        on_exit = function(_, return_val)
            templates_code = return_val
        end,
    })

    templates_job:sync(5000)

    assert_eq(templates_code, 0, "zk templates exits with code 0")
    assert_true(#templates_output > 0, "zk templates produces output")

    -- Check that expected templates are listed
    local output_str = table.concat(templates_output, "\n")
    assert_true(output_str:find("meeting") ~= nil, "templates output contains 'meeting'")
    assert_true(output_str:find("user%-story") ~= nil, "templates output contains 'user-story'")
end

-- Test 21: CLI create with template
log("")
log("Test 21: CLI create with template")
if plenary_ok then
    local Job = require("plenary.job")
    local create_tmpl_output = {}
    local create_tmpl_code = nil

    local create_tmpl_job = Job:new({
        command = "./zk",
        args = { "create", "Template Test Meeting", "--template", "meeting" },
        cwd = vim.fn.getcwd(),
        on_stdout = function(_, data)
            table.insert(create_tmpl_output, data)
        end,
        on_exit = function(_, return_val)
            create_tmpl_code = return_val
        end,
    })

    create_tmpl_job:sync(5000)

    assert_eq(create_tmpl_code, 0, "zk create --template meeting exits with code 0")

    if #create_tmpl_output > 0 then
        log("  CLI output: " .. create_tmpl_output[1])
    end
end

-- Test 22: Verify daily notes functions exist
log("")
log("Test 22: daily notes functions")
assert_true(type(zk.daily) == "function", "daily is a function")
assert_true(type(zk.list_daily) == "function", "list_daily is a function")
assert_true(type(zk.list_daily_sync) == "function", "list_daily_sync is a function")
assert_true(type(zk.daily_picker) == "function", "daily_picker is a function")

-- Test 23: CLI daily command
log("")
log("Test 23: CLI daily command")
if plenary_ok then
    local Job = require("plenary.job")
    local daily_output = {}
    local daily_code = nil

    local daily_job = Job:new({
        command = "./zk",
        args = { "daily" },
        cwd = vim.fn.getcwd(),
        on_stdout = function(_, data)
            table.insert(daily_output, data)
        end,
        on_exit = function(_, return_val)
            daily_code = return_val
        end,
    })

    daily_job:sync(5000)

    assert_eq(daily_code, 0, "zk daily exits with code 0")
    assert_true(#daily_output > 0, "zk daily produces output")

    if #daily_output > 0 then
        log("  CLI output: " .. daily_output[1])
    end
end

-- Test 24: CLI daily --list command
log("")
log("Test 24: CLI daily --list command")
if plenary_ok then
    local Job = require("plenary.job")
    local daily_list_output = {}
    local daily_list_code = nil

    local daily_list_job = Job:new({
        command = "./zk",
        args = { "daily", "--list", "--json" },
        cwd = vim.fn.getcwd(),
        on_stdout = function(_, data)
            table.insert(daily_list_output, data)
        end,
        on_exit = function(_, return_val)
            daily_list_code = return_val
        end,
    })

    daily_list_job:sync(5000)

    assert_eq(daily_list_code, 0, "zk daily --list --json exits with code 0")

    if #daily_list_output > 0 then
        local json_str = table.concat(daily_list_output, "")
        local ok, results = pcall(vim.json.decode, json_str)
        assert_true(ok, "zk daily --list JSON output is valid")
        if ok and type(results) == "table" then
            log("  Found " .. #results .. " daily note(s)")
        end
    end
end

-- Summary
log("")
log("=== Test Summary ===")
if test_passed then
    log("All tests PASSED")
    vim.cmd("qa!")
else
    log("Some tests FAILED")
    vim.cmd("cq 1")
end
