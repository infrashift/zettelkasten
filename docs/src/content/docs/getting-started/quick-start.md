---
title: Quick Start
description: Get up and running with the Zettelkasten CLI in minutes.
---

```bash
# Create an untethered note (project context optional)
zk create "My first idea" --category untethered

# Create an untethered note with explicit project
zk create "Project-specific idea" --category untethered --project my-project

# Set project on an existing note
zk set-project path/to/note.md my-project

# Tether an untethered note (requires project)
zk tether path/to/note.md --project my-project

# Untether a tethered note
zk untether path/to/note.md

# Add tags to a zettel
zk add-tags path/to/note.md golang          # Add a single tag
zk add-tags path/to/note.md golang api      # Add multiple tags

# Validate frontmatter against CUE schema
zk validate path/to/note.md

# Index notes for searching
zk index path/to/notes/       # Index a directory
zk index path/to/note.md      # Index a single file

# Search notes
zk search "authentication"              # Full-text search
zk search --project my-project          # Filter by project
zk search --category tethered           # Filter by category
zk search --tag golang --tag api        # Filter by tags
zk search "auth" --project my-project   # Combined search
zk search --json                        # JSON output for tooling

# Generate graph visualization
zk graph path/to/notes/                 # Show ASCII tree graph
zk graph . --limit 20                   # Custom node limit
zk graph . --start <note-id>            # Start from a specific note

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

# Git workflow (dated branches)
zk hello                                # Start day: create YYYYMMDD branch from main
zk goodbye                              # End day: commit and merge to main
```
