# Todo Workflow in NeoVim

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
category: "untethered"
status: "open"
due: "2026-02-20"
priority: "high"
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

" Multiple options
:ZkTodo Fix auth --due 2026-02-20 --priority high
```

### Managing Todo Status

Use the `\s` keymap (buffer-local on todo zettels) to open a status picker:

```
\s  →  Pick: open / in_progress / closed
```

This calls `zk set-status` under the hood.

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

-- Linked to specific zettels
require("zk").todo({
    title = "Follow up",
    links = { "202602131045", "202602131100" },
})
```

### Managing Status

```lua
-- Set status on current buffer's todo
require("zk").set_status("closed")
require("zk").set_status("in_progress")
require("zk").set_status("open")

-- Set status on a specific file
require("zk").set_status("closed", "./path/to/todo.md")
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

1. **Open today's daily note:**
   ```vim
   :ZkDaily
   ```

2. **Review and plan** from your daily note — create todos for tasks.

### Throughout the Day

**Capture task from idea:**
```vim
:ZkTodo Investigate the memory leak
```

**Start working on a task:**
Open the todo, review context, optionally mark as in_progress.

**Complete a task:**
Press `\s` and select `closed` from the picker.

### End of Day

1. **Update due dates if needed** (edit the frontmatter)

2. **Generate tomorrow's list:**
   ```vim
   :ZkTodoList today
   ```

---

## Linking Strategies

### Link Todo to Related Notes

When a todo relates to existing knowledge:
```vim
:ZkTodo Implement OAuth
```

### Find Backlinks to Todos

To see what notes link TO a todo, press `\b` to toggle the backlinks panel.

---

## Recommended Keymaps

```lua
-- Todo management
vim.keymap.set("n", "<leader>zt", "<cmd>ZkTodo<cr>", { desc = "New todo" })
-- \s is automatically mapped on todo zettels (status picker)
```

---

## Integration with Daily Notes

Todos and daily notes work together naturally:

### Capture Pattern

```
Daily Note (capture thoughts)
    |
Identify actionable item
    |
Create todo (:ZkTodo)
    |
Todo tracks the task
```

### Review Pattern

```
Morning: Open daily note (:ZkDaily)
    |
Check todos from yesterday
    |
Plan the day
```

### The Todo Lifecycle

```
Created (open)
    |
Working (in_progress) [optional]
    |
Completed (closed)
    |
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

---

## Quick Reference

| Action | Command / Keymap | Description |
|--------|-----------------|-------------|
| New todo | `:ZkTodo [title]` | Create a new todo |
| Set status | `\s` (buffer-local) | Pick open / in_progress / closed |
| Add tags | `\a` (buffer-local) | Prompt for tags and add to frontmatter |
| Validate frontmatter | `\v` (buffer-local) | Validate frontmatter against CUE schema |
| Generate list | `:ZkTodoList` | Markdown summary of open todos |
| Add tags (CLI) | `zk add-tags <file> <tag1> [tag2...]` | Add tags to a zettel |
| Validate (CLI) | `zk validate <file>` | Validate zettel frontmatter |

---

## Tips for Success

1. **Set due dates** - Helps prioritize and enables filtering
2. **Use priorities sparingly** - Reserve "high" for truly urgent items
3. **Don't delete, close** - Closed todos are a record of accomplishment
4. **Add descriptions** - Your future self will thank you
5. **Link related notes** - Build context around tasks
6. **Review regularly** - Use `:ZkTodoList` to stay on track

---

## Related Documentation

- [ZK-NOTES-WORKFLOW-IN-NEOVIM.md](ZK-NOTES-WORKFLOW-IN-NEOVIM.md) - General note-taking workflow
- [ZK-DAILYNOTES-WORKFLOW-IN-NEOVIM.md](ZK-DAILYNOTES-WORKFLOW-IN-NEOVIM.md) - Daily notes workflow
- [ZK-CLI-COMMANDS-CHEATSHEET.md](ZK-CLI-COMMANDS-CHEATSHEET.md) - Complete CLI reference
