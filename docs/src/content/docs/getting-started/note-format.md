---
title: Note Format
description: YAML frontmatter schema and fields for Zettelkasten notes.
---

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

## Frontmatter Fields

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

## Todo-specific fields

When `type: "todo"`:

| Field | Required | Format | Description |
|-------|----------|--------|-------------|
| `status` | Yes | `open`, `in_progress`, `closed` | Task status |
| `due` | No | `YYYY-MM-DD` | Due date |
| `completed` | No | `YYYY-MM-DD` | Completion date (set automatically) |
| `priority` | No | `high`, `medium`, `low` | Task priority |

:::note
Untethered notes can be created without a project context for quick idea capture. When tethering to a tethered note, a project is required.
:::
