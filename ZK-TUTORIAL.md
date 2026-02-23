# Zettelkasten CLI Tutorial

This tutorial walks you through building the devcontainer, connecting via SSH,
and using the full Zettelkasten workflow: creating notes, linking them together,
tethering untethered thoughts into tethered knowledge, searching, generating
graphs, and managing todos.

## Prerequisites

- [Podman](https://podman.io/) (or Docker, substituting `docker` for `podman`)
- An SSH client
- A terminal emulator (Ghostty recommended for full color support)

## 1. Build the Container

Clone the repository and build the image:

```bash
git clone https://github.com/infrashift/zettelkasten-cli.git
cd zettelkasten-cli

podman build -f Containerfiles/Containerfile -t zk-devcontainer:latest .
```

The build compiles the `zk` Go binary, installs NeoVim, Claude Code, tmux, and
all plugins. It takes a few minutes the first time.

## 2. Start the Container

The container runs an SSH server on port 2222. You need to mount an
`authorized_keys` file so you can authenticate.

**Using the included test key** (for local development only):

```bash
podman run -d --name zk-dev \
  -p 2222:2222 \
  -v ./Containerfiles/config/test_ssh_key.pub:/home/user/.ssh/authorized_keys:ro,Z \
  zk-devcontainer:latest
```

**Using your own SSH key** (recommended):

```bash
podman run -d --name zk-dev \
  -p 2222:2222 \
  -v ~/.ssh/id_ed25519.pub:/home/user/.ssh/authorized_keys:ro,Z \
  zk-devcontainer:latest
```

Verify the container started:

```bash
podman logs zk-dev
# Should show: Starting sshd on port 2222...
#              Server listening on 0.0.0.0 port 2222.
```

## 3. Connect via SSH

**With the test key:**

```bash
ssh -p 2222 -i ./Containerfiles/config/test_ssh_key \
  -o StrictHostKeyChecking=no user@localhost
```

**With your own key:**

```bash
ssh -p 2222 user@localhost
```

You land in a **tmux session** with three panes:

```
+---------------------------+------------+
|                           |   bash     |
|         NeoVim            +------------+
|                           |   claude   |
+---------------------------+------------+
```

- **Left pane**: NeoVim (your editor)
- **Top-right pane**: Bash shell
- **Bottom-right pane**: Claude Code AI assistant

Navigate between panes with your mouse (mouse mode is enabled) or with tmux
keybindings (`Ctrl-b` then arrow keys).

If you disconnect and reconnect, SSH reattaches to the same tmux session
automatically. Your editor state, shell history, and Claude conversation are
preserved.

## 4. Where Are Notes Stored?

All notes live under `~/zettelkasten/` inside the container, organized as:

```
~/zettelkasten/
  untethered/                  Untethered notes (quick captures)
    <id>.md
    daily/                     Daily notes
      2026/
        02/
          <id>.md
  tethered/                    Tethered notes (refined knowledge)
    <id>.md
  .zk_index/                   Full-text search index (auto-managed)
  .zk_graphs/                  Generated graph visualizations
  .zk_todos/                   Generated todo list files
```

**Untethered notes** are quick captures: ideas, meeting notes, snippets. They
don't require a project.

**Tethered notes** are refined, rewritten knowledge that you've decided to
keep. They require a project tag.

**Daily notes** are date-stamped journal entries stored under `untethered/daily/`
in a `YYYY/MM/` directory hierarchy.

Each note has an ID in the format `YYYYMMDDHHmmss-UUIDv4`, which combines a
timestamp for natural sorting with a UUID to prevent collisions.

To **persist notes across container restarts**, mount a host directory:

```bash
podman run -d --name zk-dev \
  -p 2222:2222 \
  -v ~/.ssh/id_ed25519.pub:/home/user/.ssh/authorized_keys:ro,Z \
  -v ~/zettelkasten:/home/user/zettelkasten:Z \
  zk-devcontainer:latest
```

## 5. Initialize Your Zettelkasten

SSH into the container and switch to the bash pane (click it or press `Ctrl-b`
then right-arrow). Initialize the zettelkasten directory:

```bash
mkdir -p ~/zettelkasten
cd ~/zettelkasten
git init
git config user.email "you@example.com"
git config user.name "Your Name"
```

The `zk` tool auto-creates the folder structure (`untethered/`, `tethered/`,
etc.) on first use.

## 6. Start Your Day with a Daily Note

Switch to the NeoVim pane and create your first daily note:

```vim
:ZkDaily
```

NeoVim opens today's daily note. It has YAML frontmatter at the top and
sections for your morning plan, tasks, notes, and end-of-day reflection:

```markdown
---
id: "20260219143000-a1b2c3d4-..."
title: "Daily Note - 2026-02-19"
type: daily-note
category: untethered
tags:
  - daily
created: "2026-02-19T14:30:00Z"
---

# Daily Note - 2026-02-19

## Morning
- [ ] ...

## Tasks
- [ ] ...

## Notes


## End of Day


## Links Created
```

Fill in your morning plan. Add a few tasks. Save the file with `:w`. The note
is automatically indexed for search.

You can also view yesterday's daily note:

```vim
:ZkDaily yesterday
```

Or browse all daily notes with a Telescope picker:

```vim
:ZkDailyList
```

## 7. Create an Untethered Note

Capture a quick idea. In NeoVim:

```vim
:ZkNew
```

You are prompted for a title. Type `Learning Zettelkasten Method` and press
Enter. A new untethered note opens with frontmatter and an empty body. Write
some content:

```markdown
The Zettelkasten method is a personal knowledge management system
developed by Niklas Luhmann. Key principles:

- One idea per note (atomicity)
- Write in your own words (elaboration)
- Connect notes to each other (linking)
- Use fleeting (untethered) notes for captures, permanent (tethered) notes for refined ideas
```

Save with `:w`.

## 8. Create a Note from a Template

Templates give notes structure. List available templates:

```vim
:ZkTemplates
```

This shows: `meeting`, `book-review`, `snippet`, `project-idea`, `user-story`,
`feature`, `daily`, `todo`, `issue`.

Create a meeting note:

```vim
:ZkTemplate meeting
```

Enter a title like `Team Standup - Knowledge Management`. The template provides
sections for attendees, agenda, discussion, action items, and next steps. Fill
it in and save.

Create a code snippet note:

```vim
:ZkTemplate snippet
```

Title it `Bash - Find Files by Extension`. Fill in the code block with a useful
snippet and save.

## 9. Link Notes Together

Notes become powerful when connected. Open the meeting note you just created,
place your cursor where you want a link, and insert one:

```vim
\l
```

(`\` is the local leader key followed by `l`)

A Telescope picker opens showing all your notes. Select `Learning Zettelkasten
Method` and press Enter. A `[[id]]` link is inserted at your cursor.

To insert a link that shows the title (more readable):

```vim
\L
```

This inserts `[[id|Learning Zettelkasten Method]]` instead.

You can also insert links from the Telescope search picker. Run `:ZkSearch`,
highlight a note, and press `Ctrl-l` to insert its link at the cursor.

## 10. Create More Related Notes

To build a meaningful graph, create several connected notes. Here is a suggested
set. For each one, use `:ZkNew`, write content, and link to related notes using
`\l`:

**Note 2: "Atomic Notes"**
```
An atomic note contains exactly one idea, fully developed.
This makes notes reusable across contexts.
```
Link to: `Learning Zettelkasten Method`

**Note 3: "Linking as Thinking"**
```
The act of linking notes forces you to articulate the relationship
between ideas. This is where insight happens.
```
Link to: `Learning Zettelkasten Method`, `Atomic Notes`

**Note 4: "Progressive Summarization"**
```
Layer highlights on top of notes over time. Bold the most important
passages, then highlight the bold, then write a summary.
```
Link to: `Atomic Notes`

**Note 5: "Untethered vs Tethered Notes"**
```
Untethered notes are raw captures. Tethered notes are refined ideas
you've rewritten in your own words with context and connections.
```
Link to: `Learning Zettelkasten Method`, `Atomic Notes`

**Note 6: "Graph Thinking"**
```
A zettelkasten is a graph of ideas, not a hierarchy.
Any note can connect to any other note.
```
Link to: `Linking as Thinking`, `Untethered vs Tethered Notes`

After creating these, you have a small web of interconnected notes.

## 11. View Backlinks

Open the `Learning Zettelkasten Method` note. Several notes link to it. View
its backlinks:

```vim
\b
```

A floating panel appears at the top-right showing every note that references
this one. From the panel:

- Press `Enter` or `o` to open a backlink
- Press `p` to preview it in a floating window
- Press `q` or `Esc` to close the panel

Toggle the panel on and off with `\b`.

## 12. Preview a Note

To peek at a note without leaving your current file:

```vim
\p
```

A floating preview window shows the current note's rendered content. Inside the
preview, press `Enter` to open it for editing or `q` to close.

You can also preview any note by ID:

```vim
:ZkPreview <id>
```

## 13. Tether a Note

The `Learning Zettelkasten Method` note has been refined and linked. Tether it
to make it a tethered note. With the note open:

```vim
\P
```

Or use the command:

```vim
:ZkTether
```

You are prompted for a project name. Enter `knowledge-management`. The note's
frontmatter updates: `category` changes from `untethered` to `tethered` and the
`project` field is set.

You can also untether a note (move it back to untethered status):

```vim
:ZkUntether
```

You can also set or change a project on any note:

```vim
:ZkSetProject my-project
```

## 14. Search Your Notes

### Quick search

```vim
:ZkSearch zettelkasten
```

A Telescope picker shows matching notes. Select one and press Enter to open it,
or `Ctrl-p` to preview.

### Live search (updates as you type)

```vim
:ZkSearch!
```

The `!` (bang) enables live search mode. Start typing and results filter in
real-time.

### Filter by category

Browse only untethered notes:

```vim
:ZkUntethered
```

Browse only tethered notes:

```vim
:ZkTethered
```

### CLI search (from the bash pane)

Switch to the bash pane and search from the command line:

```bash
cd ~/zettelkasten

# Full-text search
zk search "atomic notes"

# JSON output (for scripting)
zk search --json --limit 5

# Filter by category
zk search --category tethered

# Filter by tag
zk search --tag daily
```

## 15. Generate a Graph

Visualize how your notes connect. In NeoVim:

```vim
:ZkGraph 20
```

This generates a Mermaid flowchart of up to 20 connected notes and opens it in
a new buffer. The graph uses:

- **Rounded rectangles** (orange) for untethered notes
- **Stadium shapes** (green) for tethered notes
- **Solid arrows** for parent relationships
- **Dashed arrows** for link references

The file is saved to `~/zettelkasten/.zk_graphs/` and includes a legend, a
table of all nodes, and a list of relationships.

From the CLI:

```bash
zk graph ~/zettelkasten --limit 20
```

To render the Mermaid diagram visually, paste the code block into
[mermaid.live](https://mermaid.live) or use a Mermaid-compatible markdown
viewer.

## 16. Manage Todos

### Create a todo

```vim
:ZkTodo Buy a notebook for handwritten zettel drafts
```

Or with a due date and priority:

```vim
:ZkTodo Review meeting notes --due 2026-02-21 --priority high
```

### Browse todos

```vim
:ZkTodos
```

A Telescope picker shows open todos. From the picker:

- Press `Enter` to open a todo
- Press `d` (in normal mode) to mark it done
- Press `r` (in normal mode) to reopen it

Filter todos:

```vim
:ZkTodos overdue        " Show overdue todos
:ZkTodos today          " Due today
:ZkTodos week           " Due this week
:ZkTodos!               " Show closed/completed todos
```

### Mark a todo done

With the todo open in the editor:

```vim
:ZkDone
```

### Reopen a todo

```vim
:ZkReopen
```

### Generate a todo list

Create a markdown summary of your todos:

```vim
:ZkTodoList
```

This generates a file in `.zk_todos/` and opens it in a split.

## 17. Use Templates for Structured Notes

Beyond the meeting and snippet templates shown earlier, try these:

**Book review** (creates a tethered note, requires a project):

```vim
:ZkTemplate book-review --project reading-list
```

Sections: author, rating, summary, key takeaways, favorite quotes.

**Project idea:**

```vim
:ZkTemplate project-idea
```

Sections: problem statement, proposed solution, goals, non-goals, success
metrics.

**Feature spec:**

```vim
:ZkTemplate feature
```

Sections: requirements, design, API changes, testing strategy, rollout plan.

**Issue (bug report or enhancement):**

```vim
:ZkTemplate issue
```

Sections: type (bug/enhancement/question), description, steps to reproduce,
expected vs actual behavior.

## 18. Tag Completion

When editing a note's YAML frontmatter, place your cursor in the `tags:`
section and press `Ctrl-x Ctrl-t` in insert mode. This triggers tag
autocompletion from all tags used across your zettelkasten.

Refresh the tag cache if you've added new tags externally:

```vim
:ZkRefreshTags
```

## 19. Index Your Notes

Notes are automatically indexed when you save them in NeoVim. To manually
reindex everything:

```vim
:ZkIndex
```

Or from the CLI:

```bash
zk index ~/zettelkasten
```

## 20. Day Start / Day End (Git Workflow)

The CLI includes two convenience commands for a daily git workflow:

**Start your day:**

```bash
zk hello
```

This pulls the latest `main` branch and creates a new branch named with today's
date (e.g., `20260219`).

**End your day:**

```bash
zk goodbye
```

This stages all changes, commits with a dated message, merges to main, and
cleans up the dated branch.

## Quick Reference

### NeoVim Commands

| Command | Description |
|---|---|
| `:ZkDaily` | Open/create today's daily note |
| `:ZkDaily yesterday` | Open yesterday's daily note |
| `:ZkDailyList` | Browse daily notes |
| `:ZkNew` | Create a new untethered note |
| `:ZkNew tethered` | Create a new tethered note |
| `:ZkTemplate [name]` | Create from template |
| `:ZkTemplates` | List all templates |
| `:ZkSearch [query]` | Search notes |
| `:ZkSearch!` | Live search (updates as you type) |
| `:ZkUntethered` | Browse untethered notes |
| `:ZkTethered` | Browse tethered notes |
| `:ZkTether` | Tether current note (make tethered) |
| `:ZkUntether` | Untether current note (make untethered) |
| `:ZkSetProject [name]` | Set project on current note |
| `:ZkBacklinks` | Show backlinks panel |
| `:ZkGraph [limit]` | Generate Mermaid graph |
| `:ZkInsertLink` | Insert `[[id]]` link via picker |
| `:ZkPreview` | Preview current note |
| `:ZkTodo [title]` | Create a todo |
| `:ZkTodos` | Browse todos |
| `:ZkDone` | Mark todo as done |
| `:ZkReopen` | Reopen a closed todo |
| `:ZkTodoList` | Generate todo list markdown |
| `:ZkIndex` | Reindex all notes |
| `:ZkRefreshTags` | Refresh tag cache |

### Keybindings (in zettel markdown files)

| Key | Action |
|---|---|
| `\l` | Insert `[[id]]` link |
| `\L` | Insert `[[id\|title]]` link |
| `\b` | Toggle backlinks panel |
| `\p` | Preview current note |
| `\P` | Tether note |
| `Ctrl-x Ctrl-t` | Tag completion (insert mode) |

### Telescope Picker Keys

| Key | Action |
|---|---|
| `Enter` | Open selected note |
| `Ctrl-p` | Preview selected note |
| `Ctrl-l` | Insert link to selected note |

### CLI Commands

| Command | Description |
|---|---|
| `zk create [title]` | Create a note |
| `zk daily` | Create/open daily note |
| `zk todo [title]` | Create a todo |
| `zk todos` | List todos |
| `zk done [file]` | Mark todo done |
| `zk search [query]` | Search notes |
| `zk graph [path]` | Generate graph |
| `zk tether [file]` | Tether a note (make tethered) |
| `zk untether [file]` | Untether a note (make untethered) |
| `zk backlinks [file]` | Show backlinks |
| `zk index [path]` | Index notes |
| `zk templates` | List templates |
| `zk hello` | Start-of-day git workflow |
| `zk goodbye` | End-of-day git workflow |

## Container Management

```bash
# Stop the container
podman stop zk-dev

# Start it again (notes and tmux session preserved)
podman start zk-dev

# SSH reconnects to the same tmux session
ssh -p 2222 user@localhost

# Remove the container
podman rm -f zk-dev

# Custom SSH port
podman run -d --name zk-dev \
  -e SSH_PORT=3333 -p 3333:3333 \
  -v ~/.ssh/id_ed25519.pub:/home/user/.ssh/authorized_keys:ro,Z \
  -v ~/zettelkasten:/home/user/zettelkasten:Z \
  zk-devcontainer:latest
```
