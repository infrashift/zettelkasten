---
title: User Commands
description: Complete reference for all :Zk user commands available in NeoVim.
---

The plugin registers user commands that you can run from NeoVim's command line (`:ZkCommand`). All commands support tab completion for flags and values.

## Note Creation

### `:ZkNote`

Create a new zettel. Defaults to untethered if no category is given.

```vim
" Create an untethered note (default)
:ZkNote

" Create an untethered note (explicit)
:ZkNote --category untethered

" Create a tethered note with a project
:ZkNote --category tethered --project my-project
```

**Flags:**

| Flag | Values | Description |
|------|--------|-------------|
| `--category` | `untethered`, `tethered` | Note category (default: untethered) |
| `--project` | any string | Project name (auto-detected from git if omitted) |

### `:ZkTemplate`

Create a note from a template. Opens a picker if no template name is given.

```vim
" Open template picker
:ZkTemplate

" Create meeting notes
:ZkTemplate meeting

" Create a feature spec for a project
:ZkTemplate feature --project my-project
```

**Available templates:** `meeting`, `book-review`, `snippet`, `project-idea`, `user-story`, `feature`, `daily`

---

## Search

### `:ZkSearch`

Search zettels. Opens a picker if available, otherwise prints results.
Use the bang (`!`) variant for live search that updates as you type.

```vim
" Search for a term
:ZkSearch authentication

" Live search (updates as you type)
:ZkSearch!

" Live search with an initial query
:ZkSearch! auth

" Filter by category and project
:ZkSearch --category tethered --project my-project

" Filter by type and status
:ZkSearch --type todo --status open

" Filter by priority and due date
:ZkSearch --priority high --due-before 2026-03-01

" Filter by tag
:ZkSearch --tag security

" Combine a query with filters
:ZkSearch auth --category tethered --tag security
```

**Flags:**

| Flag | Values | Description |
|------|--------|-------------|
| `--category` | `untethered`, `tethered` | Filter by category |
| `--project` | any string | Filter by project |
| `--type` | `note`, `todo`, `daily-note`, `issue` | Filter by note type |
| `--status` | `open`, `in_progress`, `closed` | Filter by status |
| `--priority` | `high`, `medium`, `low` | Filter by priority |
| `--due-before` | `YYYY-MM-DD` | Due before date |
| `--due-after` | `YYYY-MM-DD` | Due after date |
| `--tag` | any string | Filter by tag (can repeat) |

**Picker keymaps (when available):**

| Key | Action |
|-----|--------|
| `<CR>` | Open note |
| `<C-p>` | Preview in floating window |
| `<C-l>` | Insert `[[id]]` link at cursor |
| `<C-S-l>` | Insert `[[id\|title]]` link at cursor |

---

## Graph Visualization

### `:ZkGraph`

Generate an ASCII tree visualization of note relationships. Opens in a vertical split.

```vim
" Default graph (limit 10 nodes)
:ZkGraph

" Limit number of nodes
:ZkGraph --limit 20
:ZkGraph 20

" Start from a specific note
:ZkGraph --start 20260213143000-550e8400-e29b-41d4-a716-446655440000

" Set traversal depth
:ZkGraph --depth 3

" Combine options
:ZkGraph --start 20260213143000-550e8400-e29b-41d4-a716-446655440000 --depth 3 --limit 50
```

**Flags:**

| Flag | Values | Description |
|------|--------|-------------|
| `--limit` | number | Maximum nodes to display (default: 10) |
| `--start` | zettel ID | Start node for the graph |
| `--depth` | number | Traversal depth |

---

## Daily Notes

### `:ZkDaily`

Open or create a daily note. Daily notes are idempotent — running it twice on the same day opens the same file.

```vim
" Today's daily note
:ZkDaily

" Yesterday's note (for morning review)
:ZkDaily yesterday

" Specific date
:ZkDaily 2026-02-10
```

### `:ZkDailyList`

Browse daily notes in a picker. Use the bang variant to show only this week.

```vim
" Browse all daily notes
:ZkDailyList

" This week only
:ZkDailyList!
```

---

## Todos

### `:ZkTodo`

Create a new todo. The title is everything that isn't a flag.

```vim
" Simple todo
:ZkTodo Buy a notebook for handwritten zettel drafts

" With due date and priority
:ZkTodo Review meeting notes --due 2026-02-21 --priority high

" With project
:ZkTodo Fix login bug --project my-project --priority medium
```

**Flags:**

| Flag | Values | Description |
|------|--------|-------------|
| `--due` | `YYYY-MM-DD` | Due date |
| `--priority` | `high`, `medium`, `low` | Priority level |
| `--project` | any string | Project name |

### `:ZkTodoList`

Generate a markdown summary of todos and open it in a split.

```vim
" All open todos
:ZkTodoList

" Due today
:ZkTodoList today

" Due this week
:ZkTodoList week

" Project-specific list
:ZkTodoList my-project
```

---

## Index & Tags

### `:ZkIndex`

Index zettels for full-text search. Indexes the current directory by default.

```vim
" Index current directory
:ZkIndex

" Index specific path
:ZkIndex ~/zettelkasten/
```

### `:ZkRefreshTags`

Clear the tag cache and reload tags from the index.

```vim
:ZkRefreshTags
```

---

## Quick Reference

| Command | Description |
|---------|-------------|
| `:ZkNote` | Create new zettel |
| `:ZkTemplate [name]` | Create from template |
| `:ZkSearch[!] [query]` | Search zettels (! for live search) |
| `:ZkGraph` | Graph visualization |
| `:ZkDaily [date]` | Open daily note |
| `:ZkDailyList[!]` | Browse daily notes (! for this week) |
| `:ZkTodo [title]` | Create a todo |
| `:ZkTodoList [filter]` | Generate todo list markdown |
| `:ZkIndex [path]` | Index zettels |
| `:ZkRefreshTags` | Refresh tag cache |
