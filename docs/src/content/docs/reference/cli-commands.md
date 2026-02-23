---
title: CLI Commands
description: Quick reference for all zk CLI commands, flags, and workflows.
---

Quick reference for all `zk` commands.

---

## Note Creation

### Create an untethered note (no project required)
```bash
zk create "Quick idea about X" --category untethered
zk create "My note title"  # Default is untethered
```

### Create an untethered note with project
```bash
zk create "Project idea" --category untethered --project my-project
zk create "Project idea" --category untethered -p my-project
```

### Create a tethered note (project required)
```bash
zk create "Refined concept" --category tethered --project my-project
```

---

## Note Management

### Set project on a note
```bash
zk set-project path/to/note.md my-project
```

### Tether an untethered note (promote to tethered)
```bash
# With explicit project
zk tether path/to/note.md --project my-project
zk tether path/to/note.md -p my-project

# Auto-detect from git (if in a repo)
zk tether path/to/note.md
```

### Untether a tethered note (demote to untethered)
```bash
# Reverts a tethered note back to untethered
zk untether path/to/note.md
```

---

## Indexing

### Index a single file
```bash
zk index path/to/note.md
```

### Index all notes in a directory
```bash
zk index ~/zettelkasten/
zk index .
```

---

## Searching

### Full-text search
```bash
zk search "authentication patterns"
zk search "golang error handling"
```

### Filter by project
```bash
zk search --project my-project
zk search -p my-project
zk search "query" --project my-project  # Combined
```

### Filter by category
```bash
zk search --category untethered
zk search --category tethered
zk search -c tethered
```

### Filter by tags
```bash
zk search --tag golang
zk search --tag golang --tag api  # AND (must have both)
zk search -T golang -T testing
```

### Limit results
```bash
zk search --limit 10
zk search -l 50
```

### JSON output (for scripting/tooling)
```bash
zk search --json
zk search "query" --json | jq '.[] | .title'
```

### Combined filters
```bash
zk search "authentication" --project my-project --category tethered --tag security
```

---

## Graph Visualization

### Generate a graph from a directory
```bash
zk graph ~/zettelkasten/
zk graph .
```

### Limit the number of nodes
```bash
zk graph . --limit 20
zk graph . -l 5
```

### Custom output filename
```bash
zk graph . --output my-graph.md
zk graph . -o project-graph.md
```

The graph command generates a Markdown file with:
- Mermaid flowchart diagram
- Node table with links to files
- Relationship listing

Graph files are saved to `.zk_graphs/` (configurable) and auto-added to `.gitignore`.

---

## Note Templates

List and use note templates for structured content.

### List available templates
```bash
zk templates
```

### Create a note from a template
```bash
zk create "Sprint Planning Q1" --template meeting
zk create "Clean Code Review" --template book-review
zk create "OAuth Implementation" --template feature
zk create "Add Login Flow" --template user-story
zk create "CLI Tool Idea" --template project-idea
zk create "Go Error Handling" --template snippet
```

Available templates:
- `meeting` - Meeting notes with attendees and action items
- `book-review` - Book review with rating and key takeaways
- `snippet` - Code snippet with context and explanation
- `project-idea` - Project idea with goals and next steps
- `user-story` - User story in standard format with acceptance criteria
- `feature` - Feature specification with requirements and design notes
- `daily` - Daily note for thoughts, tasks, and reflections
- `todo` - Actionable task with status tracking
- `issue` - Issue tracking like GitHub (bug, enhancement, question)

---

## Backlinks

Find all notes that link to a specific zettel.

### Find backlinks by ID
```bash
zk backlinks 202602131045
```

### Find backlinks by file path
```bash
zk backlinks ./notes/202602131045.md
```

### JSON output
```bash
zk backlinks 202602131045 --json
zk backlinks 202602131045 --json | jq '.[].title'
```

---

## Daily Notes

Create and manage daily notes for capturing thoughts, tasks, and reflections.

### Create or open today's daily note
```bash
zk daily
```

### Open yesterday's daily note
```bash
zk daily --yesterday
```

### Open a specific date's daily note
```bash
zk daily --date 2026-02-10
```

### List recent daily notes
```bash
zk daily --list              # Last 14 days
zk daily --list --week       # This week
zk daily --list --month      # This month
zk daily --list --json       # JSON output
```

Daily notes are stored in `untethered/daily/YYYY/MM/YYYY-MM-DD.md`.

See [Daily Notes Workflow](/zettelkasten-cli/neovim/daily-notes-workflow/) for a comprehensive guide to daily note workflows.

---

## Todo Management

### Create a todo
```bash
zk todo "Fix login bug"
zk todo "Update docs" --project my-project
```

### Create with due date and priority
```bash
zk todo "Critical fix" --due 2026-02-20 --priority high
zk todo "Nice to have" --priority low
```

### List todos
```bash
zk todos                        # List open todos
zk todos --project my-project   # Filter by project
zk todos --overdue              # Overdue todos
zk todos --today                # Due today
zk todos --this-week            # Due this week
zk todos --closed               # Show closed todos
zk todos --json                 # JSON output
```

### Manage todo status
```bash
zk done 202602131045            # Mark as closed
zk reopen 202602131045          # Reopen a closed todo
```

### Generate todo list markdown
```bash
zk todo-list                    # Generate todo list
zk todo-list --project my-project
zk todo-list --output my-todos.md
```

See [Todo Workflow](/zettelkasten-cli/neovim/todo-workflow/) for a comprehensive guide to todo workflows.

---

## Git Workflow

Manage your zettelkasten git repository with dated branches.

### Start your day (hello)
```bash
zk hello
```

This command:
- Checks out the `main` branch
- Pulls latest changes
- Creates a new branch named with today's date (`YYYYMMDD` format, e.g., `20260213`)
- Warns if the branch already exists (does nothing)
- Warns if there are uncommitted changes (does nothing)

### End your day (goodbye)
```bash
zk goodbye
```

This command:
- Commits all changes with a timestamped message
- Checks out `main`
- Merges the dated branch into `main`
- Deletes the dated branch after merge
- Warns if not on a date branch (does nothing)
- Warns if there are no changes to commit (still merges)

### Typical daily workflow
```bash
# Morning
zk hello                    # Create today's branch

# Throughout the day
zk create "Meeting notes" --template meeting
zk todo "Follow up on action item"
zk daily                    # Add to daily note

# End of day
zk goodbye                  # Commit and merge to main
```

---

## Help

### General help
```bash
zk --help
zk -h
```

### Command-specific help
```bash
zk create --help
zk index --help
```

---

## Flags Reference

### `zk create`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--category` | `-c` | `untethered` | Note category (`untethered` or `tethered`) |
| `--project` | `-p` | (auto-detect) | Project context |
| `--template` | - | - | Use a note template (e.g., `meeting`, `user-story`) |
| `--help` | `-h` | - | Show help |

### `zk templates`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--help` | `-h` | - | Show help |

### `zk tether`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--project` | `-p` | (auto-detect) | Project context for tethered note |
| `--help` | `-h` | - | Show help |

### `zk untether`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--help` | `-h` | - | Show help |

### `zk set-project`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--help` | `-h` | - | Show help |

### `zk index`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--help` | `-h` | - | Show help |

### `zk search`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--project` | `-p` | - | Filter by project |
| `--category` | `-c` | - | Filter by category (`untethered`/`tethered`) |
| `--tag` | `-T` | - | Filter by tag (repeatable, AND logic) |
| `--limit` | `-l` | `20` | Maximum number of results |
| `--json` | - | `false` | Output as JSON |
| `--help` | `-h` | - | Show help |

### `zk graph`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--limit` | `-l` | `10` | Maximum number of nodes to display |
| `--output` | `-o` | `graph-TIMESTAMP.md` | Output filename |
| `--help` | `-h` | - | Show help |

### `zk backlinks`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--json` | - | `false` | Output as JSON |
| `--help` | `-h` | - | Show help |

### `zk daily`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--date` | - | - | Target date (YYYY-MM-DD format) |
| `--yesterday` | - | `false` | Use yesterday's date |
| `--list` | - | `false` | List recent daily notes |
| `--week` | - | `false` | Show this week's notes (with --list) |
| `--month` | - | `false` | Show this month's notes (with --list) |
| `--json` | - | `false` | Output as JSON (with --list) |
| `--help` | `-h` | - | Show help |

### `zk todo`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--project` | `-p` | (auto-detect) | Project context |
| `--due` | - | - | Due date (YYYY-MM-DD format) |
| `--priority` | - | - | Priority (`high`, `medium`, `low`) |
| `--help` | `-h` | - | Show help |

### `zk todos`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--project` | `-p` | - | Filter by project |
| `--overdue` | - | `false` | Show overdue todos |
| `--today` | - | `false` | Show todos due today |
| `--this-week` | - | `false` | Show todos due this week |
| `--closed` | - | `false` | Show closed todos |
| `--json` | - | `false` | Output as JSON |
| `--help` | `-h` | - | Show help |

### `zk done`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--help` | `-h` | - | Show help |

### `zk reopen`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--help` | `-h` | - | Show help |

### `zk todo-list`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--project` | `-p` | - | Filter by project |
| `--output` | `-o` | `todos-TIMESTAMP.md` | Output filename |
| `--help` | `-h` | - | Show help |

### `zk hello`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--help` | `-h` | - | Show help |

### `zk goodbye`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--help` | `-h` | - | Show help |

---

## Note ID Format

Notes use a 12-digit timestamp ID: `YYYYMMDDHHMM`

Examples:
- `202602131045` = February 13, 2026 at 10:45
- `202601010900` = January 1, 2026 at 09:00

---

## Frontmatter Template

```yaml
---
id: "YYYYMMDDHHMM"
title: "Note Title"
project: "project-name"
category: "untethered"  # or "tethered"
tags:
  - "tag1"
  - "tag2"
created: "YYYY-MM-DDTHH:MM:SSZ"
parent: "YYYYMMDDHHMM"  # optional
---
```

---

## Directory Structure

Default layout (configurable):

```
~/zettelkasten/
├── untethered/         # Quick captures, ideas
│   └── 202602131045.md
├── tethered/           # Refined, linked notes
│   └── 202602131100.md
└── tmp/                # Scratch space
```

---

## Environment

`zk` automatically detects:
- **Git project**: Uses repo name as `project` field
- **Working directory**: For relative paths

---

## Common Workflows

### Capture an idea quickly (no project context)
```bash
zk create "Random thought while commuting"
```

### Capture while working on a project
```bash
cd ~/projects/my-project
zk create "Interesting pattern in auth flow"
# Project auto-detected from git
```

### Add project context later
```bash
zk set-project ~/zettelkasten/untethered/202602131045.md my-project
```

### Tether an untethered note
```bash
# Review and refine the untethered note, then tether
zk tether ~/zettelkasten/untethered/202602131045.md --project my-project

# Or if in a git repo, project is auto-detected
cd ~/projects/my-project
zk tether ~/zettelkasten/untethered/202602131045.md
```

### Untether a tethered note
```bash
# Revert a tethered note back to untethered
zk untether ~/zettelkasten/tethered/202602131100.md
```

### Build a note chain
```bash
# Create parent
zk create "Main concept" --category tethered -p my-project
# Note the ID (e.g., 202602131045)

# Create child (manually add parent field to frontmatter)
zk create "Sub-concept" --category tethered -p my-project
# Edit to add: parent: "202602131045"
```

---

## Makefile Targets

```bash
make build            # Build binary
make test             # Run Go tests + CUE validation
make lint             # Run go vet + staticcheck
make fmt              # Format Go + CUE files
make install          # Install to $GOPATH/bin
make clean            # Remove binary and index
make tidy             # Run go mod tidy
make ui-check         # Lint Lua plugin
make integration-test # Test NeoVim integration
```
