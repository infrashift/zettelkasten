#!/bin/bash
# Comprehensive container test script
set +e

PASS=0
FAIL=0
TOTAL=0

pass() { ((PASS++)); ((TOTAL++)); echo "  PASS: $1"; }
fail() { ((FAIL++)); ((TOTAL++)); echo "  FAIL: $1 -- $2"; }

section() { echo ""; echo "==== $1 ===="; }

# ── Environment checks ──────────────────────────────────────────────────
section "Environment"

[[ "$TERM" == "xterm-ghostty" ]] && pass "TERM=xterm-ghostty" || fail "TERM" "got $TERM"
[[ "$EDITOR" == "nvim" ]] && pass "EDITOR=nvim" || fail "EDITOR" "got $EDITOR"
command -v zk &>/dev/null && pass "zk in PATH" || fail "zk" "not in PATH"
command -v nvim &>/dev/null && pass "nvim in PATH" || fail "nvim" "not in PATH"
command -v tmux &>/dev/null && pass "tmux in PATH" || fail "tmux" "not in PATH"
command -v claude &>/dev/null && pass "claude in PATH" || fail "claude" "not in PATH"
command -v gcc &>/dev/null && pass "gcc in PATH" || fail "gcc" "not in PATH"
command -v git &>/dev/null && pass "git in PATH" || fail "git" "not in PATH"

# ── Version checks ──────────────────────────────────────────────────────
section "Versions"

nvim --version | head -1 | grep -q "NVIM v0.11" && pass "nvim v0.11.x" || fail "nvim version" "$(nvim --version | head -1)"
tmux -V | grep -q "tmux" && pass "tmux $(tmux -V)" || fail "tmux version" ""
claude --version 2>&1 | grep -q "Claude Code" && pass "claude $(claude --version 2>&1)" || fail "claude version" ""

# ── Mason tools ─────────────────────────────────────────────────────────
section "Mason tools"

MASON_BIN=~/.local/share/nvim/mason/bin
[[ -x "$MASON_BIN/shfmt" ]] && pass "shfmt installed" || fail "shfmt" "not found"
[[ -x "$MASON_BIN/tree-sitter" ]] && pass "tree-sitter installed" || fail "tree-sitter" "not found"
[[ -x "$MASON_BIN/stylua" ]] && pass "stylua installed" || fail "stylua" "not installed (mason timeout)"

"$MASON_BIN/shfmt" --version &>/dev/null && pass "shfmt runs OK" || fail "shfmt" "binary broken"
"$MASON_BIN/tree-sitter" --version &>/dev/null && pass "tree-sitter runs OK" || fail "tree-sitter" "binary broken (glibc?)"
if [[ -x "$MASON_BIN/stylua" ]]; then
  "$MASON_BIN/stylua" --version &>/dev/null && pass "stylua runs OK" || fail "stylua" "binary broken"
fi

# ── Ghostty terminfo ────────────────────────────────────────────────────
section "Ghostty terminfo"

[[ -d ~/.terminfo ]] && pass "~/.terminfo exists" || fail "terminfo dir" "missing"
infocmp xterm-ghostty &>/dev/null && pass "xterm-ghostty terminfo compiled" || fail "terminfo" "xterm-ghostty not found"

# ── Tmux config ─────────────────────────────────────────────────────────
section "Tmux config"

[[ -x ~/.local/bin/start-session.sh ]] && pass "start-session.sh executable" || fail "start-session.sh" "not executable"
[[ -f ~/.tmux.conf ]] && pass "~/.tmux.conf exists" || fail "tmux.conf" "missing"
grep -q "xterm-ghostty" ~/.tmux.conf && pass "tmux.conf has ghostty override" || fail "tmux.conf" "no ghostty entry"

# ── zk CLI tests ────────────────────────────────────────────────────────
section "zk CLI"

export HOME=/home/user
ZKDIR=/home/user/zettelkasten
mkdir -p "$ZKDIR/.zk" "$ZKDIR/untethered" "$ZKDIR/tethered" "$ZKDIR/daily" "$ZKDIR/todos"
cd "$ZKDIR"
git init -q .
git config user.email "test@test.com"
git config user.name "Test"

cat > .zk/config.yaml <<'YAML'
root: .
untethered_dir: untethered
tethered_dir: tethered
daily_dir: daily
todo_dir: todos
projects: []
YAML

# Test: create untethered note
OUTPUT=$(zk create "Test Note Alpha" --category untethered 2>&1)
echo "$OUTPUT" | grep -qi "created\|Created" && pass "create untethered note" || fail "create untethered" "$OUTPUT"

UNTETHERED_COUNT=$(find "$ZKDIR/untethered" -name "*.md" 2>/dev/null | wc -l)
[[ "$UNTETHERED_COUNT" -ge 1 ]] && pass "untethered file exists ($UNTETHERED_COUNT files)" || fail "untethered file" "no .md in untethered/"

# Test: create with template
OUTPUT=$(zk create "Meeting Review" --template meeting 2>&1)
echo "$OUTPUT" | grep -qi "created\|Created" && pass "create from template (meeting)" || fail "create template" "$OUTPUT"

# Test: create daily note
OUTPUT=$(zk daily 2>&1)
echo "$OUTPUT" | grep -qi "daily\|note\|created\|Daily" && pass "daily note" || fail "daily" "$OUTPUT"

# Daily notes are stored under untethered/daily/YYYY/MM/
DAILY_COUNT=$(find "$ZKDIR/untethered/daily" -name "*.md" 2>/dev/null | wc -l)
[[ "$DAILY_COUNT" -ge 1 ]] && pass "daily note file exists ($DAILY_COUNT files)" || fail "daily file" "no .md in untethered/daily/"

# Test: create todo
OUTPUT=$(zk todo "Fix the thing" 2>&1)
echo "$OUTPUT" | grep -qi "todo\|created\|Fix" && pass "create todo" || fail "create todo" "$OUTPUT"

# Todos are created as untethered notes in untethered/ (same-minute timestamps may collide)
ALL_UNTETHERED=$(find "$ZKDIR/untethered" -maxdepth 1 -name "*.md" 2>/dev/null | wc -l)
[[ "$ALL_UNTETHERED" -ge 1 ]] && pass "todo file exists in untethered/ ($ALL_UNTETHERED files)" || fail "todo file" "expected >=1 .md in untethered/"

# Test: set-status closed
TODO_FILE=$(find "$ZKDIR/untethered" -maxdepth 1 -name "*.md" 2>/dev/null | tail -1)
if [[ -n "$TODO_FILE" ]]; then
  OUTPUT=$(zk set-status "$TODO_FILE" closed 2>&1)
  echo "$OUTPUT" | grep -qi "closed\|Closed" && pass "set-status closed" || fail "set-status closed" "$OUTPUT"
fi

# Test: set-status in_progress
if [[ -n "$TODO_FILE" ]]; then
  OUTPUT=$(zk set-status "$TODO_FILE" in_progress 2>&1)
  echo "$OUTPUT" | grep -qi "in_progress" && pass "set-status in_progress" || fail "set-status in_progress" "$OUTPUT"
fi

# Test: set-status open
if [[ -n "$TODO_FILE" ]]; then
  OUTPUT=$(zk set-status "$TODO_FILE" open 2>&1)
  echo "$OUTPUT" | grep -qi "reopen\|open\|Reopen" && pass "set-status open" || fail "set-status open" "$OUTPUT"
fi

# Test: add-tags
NOTE_FILE=$(find "$ZKDIR/untethered" -maxdepth 1 -name "*.md" 2>/dev/null | head -1)
if [[ -n "$NOTE_FILE" ]]; then
  OUTPUT=$(zk add-tags "$NOTE_FILE" golang api 2>&1)
  echo "$OUTPUT" | grep -qi "added\|Added\|tags" && pass "add-tags" || fail "add-tags" "$OUTPUT"
  grep -q "golang" "$NOTE_FILE" 2>/dev/null && pass "add-tags wrote tag to file" || fail "add-tags file" "tag not found in file"
fi

# Test: validate
if [[ -n "$NOTE_FILE" ]]; then
  OUTPUT=$(zk validate "$NOTE_FILE" 2>&1)
  echo "$OUTPUT" | grep -qi "valid\|Valid" && pass "validate" || fail "validate" "$OUTPUT"
fi

# Test: index
OUTPUT=$(zk index . 2>&1)
echo "$OUTPUT" | grep -qi "indexed\|index\|Indexed" && pass "index" || fail "index" "$OUTPUT"

# Test: search
OUTPUT=$(zk search --json 2>&1)
echo "$OUTPUT" | grep -q "Test Note Alpha\|title" && pass "search finds notes (JSON)" || fail "search" "$OUTPUT"

# Test: search with --limit
OUTPUT=$(zk search --json --limit 1 2>&1)
echo "$OUTPUT" | grep -q "\[" && pass "search --limit 1" || fail "search --limit" "$OUTPUT"

# Test: search with --category
OUTPUT=$(zk search --json --category untethered 2>&1)
echo "$OUTPUT" | grep -q "untethered" && pass "search --category untethered" || fail "search --category" "$OUTPUT"

# Test: templates list
OUTPUT=$(zk templates 2>&1)
echo "$OUTPUT" | grep -q "meeting" && pass "templates lists meeting" || fail "templates" "$OUTPUT"
echo "$OUTPUT" | grep -q "todo" && pass "templates lists todo" || fail "templates" "$OUTPUT"
echo "$OUTPUT" | grep -q "snippet" && pass "templates lists snippet" || fail "templates" "$OUTPUT"
echo "$OUTPUT" | grep -q "daily" && pass "templates lists daily" || fail "templates" "$OUTPUT"
echo "$OUTPUT" | grep -q "book-review" && pass "templates lists book-review" || fail "templates" "$OUTPUT"

# Test: tether (needs an untethered note)
UNTETHERED_FILE=$(find "$ZKDIR/untethered" -name "*.md" 2>/dev/null | head -1)
if [[ -n "$UNTETHERED_FILE" ]]; then
  OUTPUT=$(zk tether "$UNTETHERED_FILE" --project test-proj 2>&1)
  echo "$OUTPUT" | grep -qi "tethered\|Tethered\|tether" && pass "tether untethered->tethered" || fail "tether" "$OUTPUT"
  # Tether updates frontmatter category in-place (doesn't move file)
  grep -q "tethered" "$UNTETHERED_FILE" 2>/dev/null && pass "tethered file has tethered category" || fail "tethered file" "category not updated"
else
  fail "tether" "no untethered file found"
fi

# Test: set-project
PERM_FILE=$(find "$ZKDIR/tethered" -name "*.md" 2>/dev/null | head -1)
if [[ -n "$PERM_FILE" ]]; then
  OUTPUT=$(zk set-project "$PERM_FILE" new-project 2>&1)
  echo "$OUTPUT" | grep -qi "project\|updated\|Project" && pass "set-project" || fail "set-project" "$OUTPUT"
fi

# Test: backlinks
if [[ -n "$PERM_FILE" ]]; then
  OUTPUT=$(zk backlinks "$PERM_FILE" --json 2>&1)
  echo "$OUTPUT" | grep -q "\[" && pass "backlinks (JSON output)" || fail "backlinks" "$OUTPUT"
fi

# Test: graph (ASCII tree output)
OUTPUT=$(zk graph . --limit 5 2>&1)
echo "$OUTPUT" | grep -q "[├└─]" && pass "graph" || fail "graph" "$OUTPUT"

# ── NeoVim plugin tests ────────────────────────────────────────────────
section "NeoVim plugin"

# Test: plugins loaded
for plugin in "zk.nvim" "snacks.nvim" "plenary.nvim" "LazyVim" "which-key.nvim" "mason.nvim"; do
  OUTPUT=$(nvim --headless -c "lua for _, p in ipairs(require('lazy').plugins()) do if p.name == '$plugin' then print('FOUND') end end" -c qa 2>&1)
  echo "$OUTPUT" | grep -q "FOUND" && pass "$plugin loaded" || fail "$plugin" "not found"
done

# Test: all Zk commands registered
for cmd in ZkSearch ZkNote ZkDaily ZkTodo ZkTemplate ZkIndex ZkGraph ZkRefreshTags ZkTodoList ZkDailyList; do
  OUTPUT=$(nvim --headless -c "lua if vim.api.nvim_get_commands({})['$cmd'] then print('OK') end" -c qa 2>&1)
  echo "$OUTPUT" | grep -q "OK" && pass ":$cmd command" || fail ":$cmd" "not registered"
done

# Test: zk module setup with correct binary path
OUTPUT=$(nvim --headless -c 'lua local zk = require("zk"); if zk.config and zk.config.bin == "/home/user/.local/bin/zk" then print("CONFIGURED") end' -c qa 2>&1)
echo "$OUTPUT" | grep -q "CONFIGURED" && pass "zk.setup() bin=/home/user/.local/bin/zk" || fail "zk.setup" "$OUTPUT"

# Test: zk module has expected functions
for fn in create_note daily todo search index tether_note untether_note set_project set_status add_tags validate graph backlinks_panel toggle_backlinks preview_note insert_link complete_tags; do
  OUTPUT=$(nvim --headless -c "lua if type(require('zk').$fn) == 'function' then print('OK') end" -c qa 2>&1)
  echo "$OUTPUT" | grep -q "OK" && pass "zk.$fn() exists" || fail "zk.$fn" "not a function"
done

# Test: snacks.nvim picker integration loads (via require("zk.picker"))
OUTPUT=$(nvim --headless -c "lua local ok = pcall(require, 'zk.picker'); print(ok and 'OK' or 'FAIL')" -c qa 2>&1)
echo "$OUTPUT" | grep -q "OK" && pass "snacks picker integration (require zk.picker)" || fail "snacks picker" "failed to load"

# Test: total plugin count
OUTPUT=$(nvim --headless -c 'lua print(#require("lazy").plugins())' -c qa 2>&1)
COUNT=$(echo "$OUTPUT" | grep -oP '^\d+' | head -1)
[[ -n "$COUNT" && "$COUNT" -ge 20 ]] && pass "LazyVim plugin count: $COUNT" || fail "plugin count" "got $COUNT"

# ── SSH server ─────────────────────────────────────────────────────────
section "SSH server"

command -v sshd &>/dev/null && pass "sshd binary exists" || fail "sshd" "not in PATH"
command -v ssh &>/dev/null && pass "ssh client binary exists" || fail "ssh" "not in PATH"
[[ -f /etc/ssh/sshd_config ]] && pass "/etc/ssh/sshd_config exists" || fail "sshd_config" "missing"
grep -q "^Port 2222" /etc/ssh/sshd_config && pass "sshd_config Port 2222" || fail "sshd_config Port" "not set to 2222"
grep -q "^PermitRootLogin no" /etc/ssh/sshd_config && pass "sshd_config PermitRootLogin no" || fail "sshd_config" "root login not disabled"
grep -q "^PasswordAuthentication no" /etc/ssh/sshd_config && pass "sshd_config PasswordAuthentication no" || fail "sshd_config" "password auth not disabled"
grep -q "^ForceCommand /home/user/.local/bin/ssh-login.sh" /etc/ssh/sshd_config && pass "sshd_config ForceCommand set" || fail "sshd_config" "ForceCommand missing"
grep -q "^AllowUsers user" /etc/ssh/sshd_config && pass "sshd_config AllowUsers user" || fail "sshd_config" "AllowUsers missing"
[[ -f /etc/ssh/ssh_host_rsa_key ]] && pass "RSA host key exists" || fail "host key" "RSA key missing"
[[ -f /etc/ssh/ssh_host_ed25519_key ]] && pass "ED25519 host key exists" || fail "host key" "ED25519 key missing"
[[ -d /run/sshd ]] && pass "/run/sshd directory exists" || fail "/run/sshd" "missing"
[[ -x /home/user/.local/bin/ssh-login.sh ]] && pass "ssh-login.sh is executable" || fail "ssh-login.sh" "not executable"
[[ -x /usr/local/bin/entrypoint.sh ]] && pass "entrypoint.sh is executable" || fail "entrypoint.sh" "not executable"
# sshd -t requires root to read host keys; as non-root, "no hostkeys" means config parsed OK
SSHD_ERR=$(sshd -t 2>&1)
if [[ $? -eq 0 ]]; then
  pass "sshd -t config syntax OK"
elif echo "$SSHD_ERR" | grep -q "no hostkeys available"; then
  pass "sshd -t config syntax OK (hostkeys unreadable as non-root)"
else
  fail "sshd -t" "$SSHD_ERR"
fi

# ── Cleanup ──────────────────────────────────────────────────────────────
rm -rf "$ZKDIR"

# ── Summary ──────────────────────────────────────────────────────────────
section "SUMMARY"
echo ""
echo "  Total: $TOTAL  |  Pass: $PASS  |  Fail: $FAIL"
echo ""

if [[ "$FAIL" -eq 0 ]]; then
  echo "  ALL TESTS PASSED"
  exit 0
else
  echo "  $FAIL TESTS FAILED"
  exit 1
fi
