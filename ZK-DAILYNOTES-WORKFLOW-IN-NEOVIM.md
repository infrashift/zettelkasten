# Daily Notes Workflow in NeoVim

A comprehensive guide to managing daily notes with `zk` directly from your editor. This document covers the daily, weekly, and monthly review workflows with NeoVim-specific commands and keybindings.

---

## What Are Daily Notes?

A daily note is a dated entry that serves as your primary capture point for a given day. Rather than deciding "what type of note is this?" every time you have a thought, you write in today's daily note first. Later, you review and extract tethered insights.

**Key characteristics:**
- **One per day** - Each date has exactly one daily note
- **Idempotent** - Running `:ZkDaily` twice returns the same file
- **Temporal anchor** - Creates a timeline of your thinking
- **Low friction** - No categorization decisions needed during capture
- **Type: daily-note** - A distinct zettel type for daily captures

---

## Why Daily Notes Matter

### 1. Reduces Cognitive Overhead

When an idea strikes, you don't want to spend mental energy deciding:
- Is this untethered or tethered?
- What project does this belong to?
- What should I title this?

Daily notes eliminate these decisions. Just write. Sort it out later.

### 2. Creates a Personal Changelog

Daily notes become a searchable history of your work and thoughts:
- "What was I working on last Tuesday?"
- "When did I first encounter this bug?"
- "What did I learn during that project?"

### 3. Enables Spaced Repetition

The daily review practice naturally resurfaces ideas:
- Morning: Review yesterday's note
- Extract anything valuable into tethered notes
- Unprocessed ideas get a second chance

### 4. Supports GTD-Style Workflows

Daily notes function as an "inbox" in Getting Things Done methodology:
- Capture everything during the day
- Process and organize during review
- Tether valuable insights to projects

### 5. Integrates with Todos

Daily notes can be linked to todos, creating a natural connection between your daily capture and actionable tasks.

---

## The Daily Note Structure

`zk` creates daily notes with this template:

```markdown
---
id: "202602130000"
title: "2026-02-13 Friday"
type: "daily-note"
category: "untethered"
tags:
  - "daily"
created: "2026-02-13T00:00:00Z"
---

# 2026-02-13 Friday

## Morning

-

## Tasks

- [ ]

## Notes



## End of Day



## Links Created

-
```

### Section Purposes

| Section | Purpose |
|---------|---------|
| **Morning** | Intentions, priorities, energy level, gratitude |
| **Tasks** | Today's action items (can sync with task manager) |
| **Notes** | Free-form capture throughout the day |
| **End of Day** | Reflection, wins, blockers, tomorrow's focus |
| **Links Created** | Track zettels created today for review |

---

## NeoVim Commands

### Create or Open Daily Note

```vim
" Today's daily note
:ZkDaily

" Yesterday's note (for morning review)
:ZkDaily yesterday

" Specific date
:ZkDaily 2026-02-10
```

### Browse Daily Notes

```vim
" Browse recent daily notes with Telescope
:ZkDailyList

" Browse this week's notes only
:ZkDailyList!
```

### Lua API

```lua
-- Open today's daily note
require("zk").daily()

-- Open yesterday's note
require("zk").daily({ date = "yesterday" })

-- Open specific date
require("zk").daily({ date = "2026-02-10" })

-- Browse daily notes with Telescope
require("zk").daily_picker()

-- List daily notes this week
require("zk").daily_picker({ week = true })

-- List daily notes this month
require("zk").daily_picker({ month = true })

-- Get daily notes synchronously (for scripting)
local notes = require("zk").list_daily_sync({ week = true })
```

---

## Daily Workflow in NeoVim

### Morning Routine (5-10 minutes)

1. **Open today's note:**
   ```vim
   :ZkDaily
   ```

2. **Review yesterday (in a split):**
   ```vim
   :ZkDaily yesterday
   ```
   Or use the picker to see this week:
   ```vim
   :ZkDailyList!
   ```

3. **Extract insights from yesterday:**
   - Identify anything worth keeping as a tethered note
   - Create new notes with `:ZkNote` or `:ZkTemplate`
   - Link back to the daily note if relevant

4. **Set today's intentions:**
   - Fill in the Morning section
   - List 3-5 tasks

### Throughout the Day

Capture everything in your daily note. Press your keymap to jump there instantly:

```lua
vim.keymap.set("n", "<leader>zd", "<cmd>ZkDaily<cr>", { desc = "Today's daily" })
```

### End of Day (5 minutes)

1. Review the Notes section
2. Fill in the End of Day reflection
3. List any notes you created in Links Created
4. Mark completed tasks

---

## Linking Todos to Daily Notes

One of the most powerful features is linking todos to daily notes. This creates a connection between your actionable items and the day you captured them.

### Create a Todo from the Daily Note

```vim
:ZkTodo Fix the authentication bug
```

Or with the Lua API:
```lua
require("zk").todo({ title = "Fix the auth bug" })
```

### Create Todo with Full Options

```vim
:ZkTodo Fix auth bug --priority high --due 2026-02-15
```

### From the Daily Note

When reviewing your daily note, you can:
1. Identify actionable items in the Notes section
2. Create todos: `:ZkTodo`
3. The todo can be linked to related notes

---

## Weekly Review Workflow

Weekly reviews help you step back and see patterns.

### List This Week's Notes

```vim
:ZkDailyList!
```

Or via CLI:
```bash
zk daily --list --week
```

### Weekly Review in NeoVim

1. **Open the daily picker for this week:**
   ```vim
   :ZkDailyList!
   ```

2. **Scan each day's notes:**
   - Press `<CR>` to open a note
   - Use `<C-n>` and `<C-p>` to navigate the picker

3. **Create synthesis notes:**
   ```vim
   :ZkTemplate meeting
   ```
   Use the meeting template for weekly summaries.

### Weekly Review Checklist

1. **Scan all daily notes** from the week
2. **Identify patterns:**
   - What topics kept coming up?
   - What questions remain unanswered?
   - What projects progressed? Stalled?
3. **Create synthesis notes:**
   - Weekly summary note
   - Tethered notes for recurring themes
4. **Review todos** via `:ZkTodoList`

---

## Monthly Review Workflow

Monthly reviews are for strategic thinking.

### List This Month's Notes

```lua
require("zk").daily_picker({ month = true })
```

Or via CLI:
```bash
zk daily --list --month
```

### Monthly Review Checklist

1. **Skim all daily notes** (focus on End of Day sections)
2. **Review weekly summaries** if you created them
3. **Assess projects:**
   - What shipped?
   - What's blocked?
   - What should be abandoned?
4. **Review completed todos** via `:ZkTodoList`
5. **Set next month's themes**

---

## Recommended Keymaps

Add these to your NeoVim configuration:

```lua
-- Daily notes
vim.keymap.set("n", "<leader>zd", "<cmd>ZkDaily<cr>", { desc = "Today's daily" })
vim.keymap.set("n", "<leader>zD", "<cmd>ZkDaily yesterday<cr>", { desc = "Yesterday's daily" })
vim.keymap.set("n", "<leader>zw", "<cmd>ZkDailyList!<cr>", { desc = "This week's dailies" })

-- Quick todo
vim.keymap.set("n", "<leader>zT", "<cmd>ZkTodo<cr>", { desc = "New todo" })
```

---

## File Organization

Daily notes are stored in a structured hierarchy:

```
~/zettelkasten/
└── untethered/
    └── daily/
        └── 2026/
            ├── 01/
            │   ├── 2026-01-01.md
            │   ├── 2026-01-02.md
            │   └── ...
            └── 02/
                ├── 2026-02-01.md
                └── ...
```

This structure:
- Groups notes by year and month
- Uses human-readable filenames (YYYY-MM-DD.md)
- Makes manual browsing easy
- Supports standard file system tools

---

## Integration with Zettelkasten

Daily notes complement the Zettelkasten method:

| Zettelkasten Concept | Daily Notes Role |
|---------------------|------------------|
| **Untethered notes** | Daily note = primary untethered inbox |
| **Tethered notes** | Extract from daily during review |
| **Links** | Track in "Links Created" section |
| **Todos** | Create todos from daily captures |
| **Projects** | Daily notes capture project-related thoughts |

### The Flow

```
Daily Note (capture)
    |
Morning Review (extract)
    |
Tethered Note (refine)    <->    Todo (actionable)
    |                              |
Graph (connect)             Done (complete)
```

---

## Quick Reference

| Action | Command | Keymap Suggestion |
|--------|---------|-------------------|
| Today's daily | `:ZkDaily` | `<leader>zd` |
| Yesterday's daily | `:ZkDaily yesterday` | `<leader>zD` |
| Browse dailies | `:ZkDailyList` | |
| This week's dailies | `:ZkDailyList!` | `<leader>zw` |
| Add tags | `\a` (buffer-local) | Auto-mapped on zettels |
| Validate frontmatter | `\v` (buffer-local) | Auto-mapped on zettels |
| New todo | `:ZkTodo` | `<leader>zT` |

---

## Tips for Success

1. **Lower the bar** - A one-line daily note is still valuable
2. **Time-box reviews** - 5 minutes daily, 15 minutes weekly
3. **Don't backfill** - If you miss a day, just start fresh
4. **Use the search** - Your daily notes are searchable; trust the system
5. **Experiment** - Modify the template to fit your needs

---

## Related Documentation

- [ZK-NOTES-WORKFLOW-IN-NEOVIM.md](ZK-NOTES-WORKFLOW-IN-NEOVIM.md) - General note-taking workflow
- [ZK-TODO-WORKFLOW-IN-NEOVIM.md](ZK-TODO-WORKFLOW-IN-NEOVIM.md) - Todo management workflow
- [ZK-CLI-COMMANDS-CHEATSHEET.md](ZK-CLI-COMMANDS-CHEATSHEET.md) - Complete CLI reference
