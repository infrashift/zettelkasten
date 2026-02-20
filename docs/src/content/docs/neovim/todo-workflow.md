---
title: Todo Workflow
description: A comprehensive guide to managing todos with zk directly from NeoVim.
---

A comprehensive guide to managing todos with `zk` directly from your editor. Todos are a special type of zettel designed for actionable tasks with status tracking, due dates, priorities, and links to other notes.

---

## What Are Todos?

Todos in `zk` are not just checkboxes - they're full zettels with:
- **Status tracking** - Open, in progress, or closed
- **Due dates** - Target completion dates
- **Priority levels** - High, medium, or low
- **Links** - Connections to other zettels, including daily notes
- **Full-text search** - Todos are indexed and searchable
- **Rich content** - Descriptions, acceptance criteria, notes

**Key characteristics:**
- **Type: todo** - A distinct zettel type for actionable tasks
- **Linkable** - Can link to any other zettel (notes, daily notes, other todos)
- **Queryable** - Filter by status, priority, due date, project
- **Persistent** - Never lost, always searchable, even when closed

---

## Why Todos as Zettels?

### Traditional Task Managers

| Feature | Task Managers | zk Todos |
|---------|--------------|----------|
| Rich descriptions | Limited | Full markdown |
| Links to knowledge | None | Native linking |
| Searchable history | Varies | Always indexed |
| Context preservation | Lost on completion | Permanent record |
| Integration | Separate app | In your editor |

### The zk Advantage

1. **Context lives with the task** - Add notes, links, and details directly
2. **Tasks link to knowledge** - Connect todos to related notes
3. **History is preserved** - Closed todos remain searchable
4. **No context switching** - Manage tasks in your editor

---

## The Todo Structure

`zk` creates todos with this frontmatter:

```markdown
---
id: "202602131045"
title: "Fix authentication bug"
type: "todo"
project: "my-project"
category: "fleeting"
status: "open"
due: "2026-02-20"
priority: "high"
links:
  - "202602130000"
tags:
  - "todo"
  - "bug"
created: "2026-02-13T10:45:00Z"
---

# Fix authentication bug

## Description

Users are getting logged out randomly after 10 minutes.

## Acceptance Criteria

- [ ] Identify root cause
- [ ] Write failing test
- [ ] Implement fix
- [ ] Verify in staging

## Notes

Might be related to session timeout configuration.

## Related

- [[202602120930|Session Management Architecture]]
```

### Status Values

| Status | Meaning | Icon |
|--------|---------|------|
| `open` | Not started | `[ ]` |
| `in_progress` | Being worked on | `[~]` |
| `closed` | Completed | `[x]` |

### Priority Values

| Priority | When to Use |
|----------|-------------|
| `high` | Urgent, blocking other work |
| `medium` | Important but not urgent |
| `low` | Nice to have, do when possible |

---

## NeoVim Commands

### Creating Todos

```vim
" Basic todo
:ZkTodo Fix the login bug

" With due date
:ZkTodo Update documentation --due 2026-02-20

" With priority
:ZkTodo Critical security fix --priority high

" With project
:ZkTodo Refactor auth module --project my-project

" Linked to today's daily note
:ZkTodo Review PR feedback --link-daily

" Linked to specific zettel
:ZkTodo Follow up on meeting --link 202602131045

" Multiple options
:ZkTodo Fix auth --due 2026-02-20 --priority high --link-daily
```

### Quick Todo Linked to Daily

```vim
" Creates todo automatically linked to today's daily note
:ZkTodoDaily Review the pull request
```

### Managing Todo Status

```vim
" Mark as done (works on current buffer or by ID)
:ZkDone
:ZkDone 202602131045
:ZkDone ./path/to/todo.md

" Reopen a closed todo
:ZkReopen
:ZkReopen 202602131045
```

### Browsing Todos

```vim
" Open todos (default)
:ZkTodos

" Closed todos
:ZkTodos!

" Filter by project
:ZkTodos my-project

" Due today
:ZkTodos today

" Due this week
:ZkTodos week

" Overdue
:ZkTodos overdue
```

### Generating Todo Lists

```vim
" Generate markdown todo list
:ZkTodoList

" Project-specific list
:ZkTodoList my-project

" Due today
:ZkTodoList today

" Due this week
:ZkTodoList week
```

---

## Lua API

### Creating Todos

```lua
-- Basic todo
require("zk").todo({ title = "Fix the bug" })

-- With options
require("zk").todo({
    title = "Update documentation",
    due = "2026-02-20",
    priority = "high",
    project = "my-project",
})

-- Linked to daily note
require("zk").todo({
    title = "Review PR",
    link_daily = true,
})

-- Linked to specific zettels
require("zk").todo({
    title = "Follow up",
    links = { "202602131045", "202602131100" },
})

-- Convenience: todo linked to daily
require("zk").todo_daily({ title = "Quick task" })
```

### Managing Status

```lua
-- Mark current buffer's todo as done
require("zk").done()

-- Mark by ID or path
require("zk").done("202602131045")
require("zk").done("./path/to/todo.md")

-- Reopen a todo
require("zk").reopen()
require("zk").reopen("202602131045")
```

### Browsing Todos

```lua
-- Open todos (Telescope picker)
require("zk").todo_picker()

-- Closed todos
require("zk").todo_picker({ closed = true })

-- Filtered
require("zk").todo_picker({
    project = "my-project",
    priority = "high",
})

-- Due this week
require("zk").todo_picker({ this_week = true })

-- Overdue
require("zk").todo_picker({ overdue = true })

-- Get todos synchronously
local todos = require("zk").todos_sync({ this_week = true })
```

### Generating Lists

```lua
require("zk").todo_list()
require("zk").todo_list({ project = "my-project" })
require("zk").todo_list({ today = true })
```

---

## Daily Workflow

### Morning Review

1. **Check overdue todos:**
   ```vim
   :ZkTodos overdue
   ```

2. **Review today's todos:**
   ```vim
   :ZkTodos today
   ```

3. **Plan from daily note:**
   Open today's daily (`:ZkDaily`) and create linked todos for tasks.

### Throughout the Day

**Capture task from idea:**
```vim
:ZkTodoDaily Investigate the memory leak
```

**Start working on a task:**
Open the todo, review context, optionally mark as in_progress.

**Complete a task:**
```vim
:ZkDone
```

### End of Day

1. **Review open todos:**
   ```vim
   :ZkTodos
   ```

2. **Update due dates if needed** (edit the frontmatter)

3. **Generate tomorrow's list:**
   ```vim
   :ZkTodoList today
   ```

---

## Linking Strategies

### Link Todo to Daily Note

When you capture a todo from your daily workflow:
```vim
:ZkTodo Research caching strategies --link-daily
```

The todo's frontmatter will include:
```yaml
links:
  - "202602130000"  # Today's daily note
```

### Link Todo to Related Notes

When a todo relates to existing knowledge:
```vim
:ZkTodo Implement OAuth --link 202602101200
```

### Multiple Links

```vim
:ZkTodo Update auth --link 202602101200 --link 202602111430
```

Or in Lua:
```lua
require("zk").todo({
    title = "Update auth",
    links = { "202602101200", "202602111430" },
})
```

### Find Backlinks to Todos

To see what notes link TO a todo:
```vim
:ZkBacklinks
```

---

## Telescope Picker Features

When using `:ZkTodos`, the Telescope picker provides:

### Display Format
```
[ ] Fix authentication bug (due: 2026-02-20) [high]
[~] Update documentation (due: 2026-02-15) [medium]
[x] Review PR feedback [low]
```

### Keybindings in Picker

| Key | Action |
|-----|--------|
| `<CR>` | Open the todo |
| `d` | Mark as done |
| `r` | Reopen |
| `<C-p>` / `<C-n>` | Navigate |
| `<Esc>` | Close picker |

---

## Recommended Keymaps

```lua
-- Todo management
vim.keymap.set("n", "<leader>zt", "<cmd>ZkTodo<cr>", { desc = "New todo" })
vim.keymap.set("n", "<leader>zT", "<cmd>ZkTodoDaily<cr>", { desc = "Todo linked to daily" })
vim.keymap.set("n", "<leader>zs", "<cmd>ZkTodos<cr>", { desc = "Browse todos" })
vim.keymap.set("n", "<leader>zS", "<cmd>ZkTodos!<cr>", { desc = "Browse closed todos" })
vim.keymap.set("n", "<leader>zx", "<cmd>ZkDone<cr>", { desc = "Mark todo done" })
vim.keymap.set("n", "<leader>zo", "<cmd>ZkTodos overdue<cr>", { desc = "Overdue todos" })

-- Quick filters
vim.keymap.set("n", "<leader>ztt", "<cmd>ZkTodos today<cr>", { desc = "Due today" })
vim.keymap.set("n", "<leader>ztw", "<cmd>ZkTodos week<cr>", { desc = "Due this week" })
```

---

## Integration with Daily Notes

Todos and daily notes work together naturally:

### Capture Pattern

```
Daily Note (capture thoughts)
    ↓
Identify actionable item
    ↓
Create linked todo (:ZkTodoDaily)
    ↓
Todo links back to daily (context preserved)
```

### Review Pattern

```
Morning: Open daily note (:ZkDaily)
    ↓
Check linked todos from yesterday
    ↓
Review today's todos (:ZkTodos today)
    ↓
Plan the day
```

### The Todo Lifecycle

```
Created (open)
    ↓
Working (in_progress) [optional]
    ↓
Completed (closed)
    ↓
Still searchable (permanent record)
```

---

## Generated Todo Lists

The `:ZkTodoList` command generates a markdown file you can review or share:

```markdown
# Todo List

Generated: 2026-02-13 15:30

## High Priority

- [ ] **Fix authentication bug** [[202602131045]]
  - Due: 2026-02-15
  - Project: my-project

## Medium Priority

- [ ] **Update documentation** [[202602131100]]
  - Due: 2026-02-20

## Other

- [ ] **Refactor tests** [[202602131200]]

---

Total: 3 todos
```

### Generated List Location

Lists are saved to the configured `todos_path` (default: `.zk_todos/`):
```
~/zettelkasten/
└── .zk_todos/
    ├── todos.md
    ├── my-project-todos.md
    └── todos-2026-02-13.md
```

These are gitignored by default.

---

## Quick Reference

| Action | Command | Keymap Suggestion |
|--------|---------|-------------------|
| New todo | `:ZkTodo [title]` | `<leader>zt` |
| Todo linked to daily | `:ZkTodoDaily [title]` | `<leader>zT` |
| Browse open todos | `:ZkTodos` | `<leader>zs` |
| Browse closed todos | `:ZkTodos!` | `<leader>zS` |
| Mark done | `:ZkDone` | `<leader>zx` |
| Reopen | `:ZkReopen` | |
| Overdue | `:ZkTodos overdue` | `<leader>zo` |
| Due today | `:ZkTodos today` | `<leader>ztt` |
| Due this week | `:ZkTodos week` | `<leader>ztw` |
| Generate list | `:ZkTodoList` | |

---

## Tips for Success

1. **Use --link-daily** - Always link todos to daily notes for context
2. **Set due dates** - Helps prioritize and enables filtering
3. **Use priorities sparingly** - Reserve "high" for truly urgent items
4. **Don't delete, close** - Closed todos are a record of accomplishment
5. **Add descriptions** - Your future self will thank you
6. **Link related notes** - Build context around tasks
7. **Review regularly** - Use `:ZkTodos overdue` to stay on track

---

## Related Documentation

- [Notes Workflow](/zettelkasten-cli/neovim/notes-workflow/) - General note-taking workflow
- [Daily Notes Workflow](/zettelkasten-cli/neovim/daily-notes-workflow/) - Daily notes workflow
- [CLI Commands](/zettelkasten-cli/reference/cli-commands/) - Complete CLI reference
