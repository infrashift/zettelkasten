# zk - Zettelkasten CLI

A fast, opinionated command-line tool for managing a Zettelkasten note-taking system. Built with Go, validated with CUE schemas, and integrated with NeoVim.

## Features

- **CUE-validated schemas** - Frontmatter is validated against strict CUE schemas
- **Git-aware** - Automatically detects project context from git repositories
- **Bleve full-text search** - Fast, local search index with structured field queries
- **Graph visualization** - Mermaid diagrams showing note relationships
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
# Create a fleeting note (project context optional)
zk create "My first idea" --type fleeting

# Create a fleeting note with explicit project
zk create "Project-specific idea" --type fleeting --project my-project

# Set project on an existing note
zk set-project path/to/note.md my-project

# Promote a fleeting note to permanent (requires project)
zk promote path/to/note.md --project my-project

# Index notes for searching
zk index path/to/notes/       # Index a directory
zk index path/to/note.md      # Index a single file

# Search notes
zk search "authentication"              # Full-text search
zk search --project my-project          # Filter by project
zk search --category permanent          # Filter by category
zk search --tag golang --tag api        # Filter by tags
zk search "auth" --project my-project   # Combined search
zk search --json                        # JSON output for tooling

# Generate graph visualization
zk graph path/to/notes/                 # Generate Mermaid graph
zk graph . --limit 20                   # Custom node limit
zk graph . --output my-graph.md         # Custom output filename

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
zk todo "Review PR" --link-daily        # Link to today's daily note
zk todo "Follow up" --link 202602131045 # Link to specific zettel
zk todos                                # List open todos
zk todos --project my-project           # Filter by project
zk todos --overdue                      # Show overdue todos
zk todos --today                        # Due today
zk todos --this-week                    # Due this week
zk todos --closed                       # Show closed todos
zk done 202602131045                    # Mark todo as done
zk reopen 202602131045                  # Reopen a closed todo
zk todo-list                            # Generate todo list markdown
zk todo-list --project my-project       # Project-specific list

# Linking zettels
zk create "Research notes" --link-daily          # Link note to today's daily
zk create "Follow-up" --link 202602131045        # Link to specific zettel
zk create "Multi-link" --link 123 --link 456     # Multiple links

# Git workflow (dated branches)
zk hello                                # Start day: create YYYYMMDD branch from main
zk goodbye                              # End day: commit and merge to main
```

## Configuration

Create a config file at `~/.config/zk/config.cue`:

```cue
root_path:  "~/zettelkasten"
index_path: ".zk_index"
graph_path: ".zk_graphs"
todos_path: ".zk_todos"
editor:     "nvim"
folders: {
    fleeting:  "fleeting"
    permanent: "permanent"
    tmp:       "tmp"
}
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `root_path` | `~/zettelkasten` | Root directory for all notes |
| `index_path` | `.zk_index` | Location of Bleve search index |
| `graph_path` | `.zk_graphs` | Location of generated graph files |
| `todos_path` | `.zk_todos` | Location of generated todo lists |
| `editor` | `nvim` | Editor for opening notes |
| `folders.fleeting` | `fleeting` | Subdirectory for fleeting notes |
| `folders.permanent` | `permanent` | Subdirectory for permanent notes |
| `folders.tmp` | `tmp` | Subdirectory for temporary notes |

## Note Format

Notes use YAML frontmatter validated against a CUE schema:

```markdown
---
id: "202602131045"
title: "My Note Title"
project: "my-project"
category: "fleeting"
links:
  - "202602130900"
  - "202602131200"
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
| `type` | No | `note`, `todo`, or `dailynote` | Zettel type (default: `note`) |
| `project` | Fleeting: No, Permanent: Yes | Non-empty string | Project context (auto-detected from git) |
| `category` | Yes | `fleeting` or `permanent` | Note category |
| `links` | No | List of 12-digit IDs | Links to other zettels |
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

**Note:** Fleeting notes can be created without a project context for quick idea capture. When promoting to permanent, a project is required.

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
