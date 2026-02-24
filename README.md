# zk - Zettelkasten CLI

A fast, opinionated command-line tool for managing a Zettelkasten note-taking system. Built with Go, validated with CUE schemas, and integrated with NeoVim.

## Features

- **CUE-validated schemas** - Frontmatter is validated against strict CUE schemas
- **Git-aware** - Automatically detects project context from git repositories
- **Bleve full-text search** - Fast, local search index with structured field queries
- **Graph visualization** - ASCII tree showing note relationships
- **Backlinks discovery** - Find all notes that link to any zettel
- **Note templates** - Built-in templates for meetings, user stories, features, and more
- **Daily notes** - Idempotent daily capture with review workflows
- **Todo management** - Task tracking with status, due dates, priority, and generated lists
- **NeoVim integration** - Lua plugin for seamless note creation from your editor
- **Hierarchical notes** - Support for parent-child relationships between zettels
- **Universal linking** - Link any zettel to any other zettel (todos, notes, daily notes)
- **Git workflow** - Dated branch management with `hello` and `goodbye` commands

## Installation

### Prerequisites

- Go 1.22+
- CUE (for schema validation)

### Build from source

```bash
git clone https://github.com/infrashift/zettelkasten-cli.git
cd zettelkasten-cli
make build
make install  # Installs to $GOPATH/bin
```

## Quick Start

```bash
# Create an untethered note (project context optional)
zk create "My first idea" --category untethered

# Create an untethered note with explicit project
zk create "Project-specific idea" --category untethered --project my-project

# Set project on an existing note
zk set-project path/to/note.md my-project

# Tether an untethered note to a project (requires project)
zk tether path/to/note.md --project my-project

# Index notes for searching
zk index path/to/notes/       # Index a directory
zk index path/to/note.md      # Index a single file

# Search notes
zk search "authentication"              # Full-text search
zk search --project my-project          # Filter by project
zk search --category tethered           # Filter by category
zk search --tag golang --tag api        # Filter by tags
zk search --type todo --status open     # Filter by type and status
zk search --priority high               # Filter by priority
zk search --due-before 2026-03-01       # Filter by due date
zk search "auth" --project my-project   # Combined search
zk search --json                        # JSON output for tooling

# Show graph visualization
zk graph path/to/notes/                 # ASCII tree of relationships
zk graph . --limit 20                   # Custom node limit
zk graph . --start <id>                 # Center on a specific zettel

# Find backlinks (notes that link to a zettel)
zk backlinks 202602131045               # By ID
zk backlinks path/to/note.md            # By file path
zk backlinks 202602131045 --json        # JSON output

# Create notes from templates
zk templates                            # List available templates
zk create "Sprint Planning" --template meeting
zk create "Login Feature" --template user-story
zk create "OAuth2 Integration" --template feature
zk create "Login broken on mobile" --template issue

# Daily notes
zk daily                                # Today's daily note
zk daily --yesterday                    # Yesterday (for morning review)
zk daily --date 2026-02-10              # Specific date
zk daily --list                         # List recent daily notes

# Todo management
zk todo "Fix login bug"                 # Create a todo
zk todo "Update docs" --due 2026-02-20  # With due date
zk todo "Critical fix" --priority high  # With priority
zk todos                                # List open todos
zk todos --project my-project           # Filter by project
zk todos --overdue                      # Show overdue todos
zk todos --today                        # Due today
zk todos --this-week                    # Due this week
zk todos --closed                       # Show closed todos
zk set-status 202602131045 closed       # Mark todo as closed
zk set-status 202602131045 in_progress  # Set to in progress
zk set-status 202602131045 open         # Reopen a todo
zk todo-list                            # Generate todo list markdown
zk todo-list --project my-project       # Project-specific list

# Add tags to a zettel
zk add-tags path/to/note.md golang api  # Add one or more tags

# Validate frontmatter
zk validate path/to/note.md             # Validate against CUE schema

# Git workflow (dated branches)
zk hello                                # Start day: create YYYYMMDD branch from main
zk goodbye                              # End day: commit and merge to main
```

## Configuration

Create a config file at `~/.config/zk/config.cue`:

```cue
root_path:  "~/zettelkasten"
index_path: ".zk_index"
editor:     "nvim"
folders: {
    untethered: "untethered"
    tethered:   "tethered"
    tmp:        "tmp"
}
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `root_path` | `~/zettelkasten` | Root directory for all notes |
| `index_path` | `.zk_index` | Location of Bleve search index |
| `editor` | `nvim` | Editor for opening notes |
| `folders.untethered` | `untethered` | Subdirectory for untethered notes |
| `folders.tethered` | `tethered` | Subdirectory for tethered notes |
| `folders.tmp` | `tmp` | Subdirectory for temporary notes |

## Note Format

Notes use YAML frontmatter validated against a CUE schema:

```markdown
---
id: "202602131045"
title: "My Note Title"
project: "my-project"
category: "untethered"
tags:
  - "idea"
  - "research"
created: "2026-02-13T10:45:00Z"
parent: "202602130900"  # Optional
---

# My Note Title

Your note content here...
```

### Frontmatter Fields

| Field | Required | Format | Description |
|-------|----------|--------|-------------|
| `id` | Yes | 12 digits (YYYYMMDDHHMM) | Unique timestamp identifier |
| `title` | Yes | Non-empty string | Note title |
| `type` | No | `note`, `todo`, or `daily-note` | Zettel type (default: `note`) |
| `project` | Untethered: No, Tethered: Yes | Non-empty string | Project context (auto-detected from git) |
| `category` | Yes | `untethered` or `tethered` | Note category |
| `tags` | Yes | List of non-empty strings | Categorization tags |
| `created` | Yes | ISO 8601 timestamp | Creation timestamp |
| `parent` | No | 12 digits | Parent zettel ID for hierarchies |

**Todo-specific fields** (when `type: "todo"`):

| Field | Required | Format | Description |
|-------|----------|--------|-------------|
| `status` | Yes | `open`, `in_progress`, `closed` | Task status |
| `due` | No | `YYYY-MM-DD` | Due date |
| `completed` | No | `YYYY-MM-DD` | Completion date (set automatically) |
| `priority` | No | `high`, `medium`, `low` | Task priority |

**Note:** Untethered notes can be created without a project context for quick idea capture. When tethering to a project, a project is required.

## NeoVim Integration

See [NEOVIM-PLUGIN-INSTALL.md](NEOVIM-PLUGIN-INSTALL.md) for installation instructions.

### Workflow Guides

- [ZK-NOTES-WORKFLOW-IN-NEOVIM.md](ZK-NOTES-WORKFLOW-IN-NEOVIM.md) - General note-taking workflow
- [ZK-DAILYNOTES-WORKFLOW-IN-NEOVIM.md](ZK-DAILYNOTES-WORKFLOW-IN-NEOVIM.md) - Daily notes and review workflows
- [ZK-TODO-WORKFLOW-IN-NEOVIM.md](ZK-TODO-WORKFLOW-IN-NEOVIM.md) - Todo management workflow

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Format code
make fmt

# Run NeoVim plugin integration tests
make integration-test

# Check Lua syntax
make ui-check
```

## Project Structure

```
zettelkasten-cli/
├── cmd/zk/              # CLI entry point
├── internal/
│   ├── config/          # CUE schemas and config loading
│   ├── graph/           # Note relationship graph
│   ├── index/           # Bleve search index
│   └── zettel/          # Note utilities
├── lua/zk/              # NeoVim plugin
├── test/                # Integration tests
└── testdata/            # Test fixtures
```

## Philosophy

This tool follows the Zettelkasten method. See [WHY-ZETTELKASTEN.md](WHY-ZETTELKASTEN.md) for background on the methodology.

## License

MIT
