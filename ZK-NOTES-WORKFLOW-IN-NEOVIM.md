# Zettelkasten Workflow in NeoVim

A hands-on tutorial for using `zk` to build a personal knowledge base directly from your editor.

---

## Why This Matters

You're a developer. You read documentation, solve problems, attend meetings, and have ideas throughout the day. Where does all that knowledge go?

**The problem:**
- Notes scattered across apps, files, and sticky notes
- Good ideas forgotten within days
- No way to connect related thoughts
- Context lost when switching projects

**The solution:**
- One system for capturing everything
- Notes that link to each other
- Searchable, indexed knowledge
- Works from your editor - no context switching

`zk` brings the Zettelkasten method into NeoVim, so your knowledge base grows while you work.

---

## Before You Begin

**Recommended reading:** [WHY-ZETTELKASTEN.md](WHY-ZETTELKASTEN.md)

Understanding the methodology will help you get more value from the tool. The key concepts are:

- **Fleeting notes** - Quick captures, unrefined thoughts
- **Permanent notes** - Refined insights, linked to other notes
- **Atomic notes** - One idea per note
- **Links** - Connections between notes create emergent structure

---

## User Story

> *As a developer working on multiple projects, I want to capture ideas and insights as I work, so that I can build a searchable knowledge base without leaving my editor.*

This tutorial will show you how to:
1. Capture a quick idea (fleeting note)
2. Refine and promote it (permanent note)
3. Search and connect your notes
4. Use templates for structured content
5. Build daily review habits

---

## Prerequisites

Ensure you have:
- NeoVim 0.9+ with the zk plugin installed
- The `zk` CLI binary in your PATH
- A zettelkasten directory (default: `~/zettelkasten/`)

Verify your setup:
```vim
:ZkTemplates
```

You should see a list of available templates.

---

## Part 1: Basic Workflow (Without Templates)

### Step 1: Capture a Fleeting Note

You're reading code and notice an interesting pattern. Capture it immediately.

**From NeoVim:**
```vim
:ZkNew fleeting
```

Or using the Lua API:
```vim
:lua require("zk").create_note("fleeting")
```

You'll be prompted for a title:
```
Note Title: Interesting error handling pattern in auth module
```

The plugin creates a new note and shows you the path:
```
Created: ~/zettelkasten/fleeting/202602131423.md
```

**What just happened:**
- A new markdown file was created with a unique timestamp ID
- YAML frontmatter was generated with metadata
- The note is categorized as "fleeting" (quick capture)
- No project context required - just capture the thought

### Step 2: View Your Note

Open the note to add content:
```vim
:ZkSearch
```

Find your note in the Telescope picker and press `<CR>` to open it.

The file looks like this:
```markdown
---
id: "202602131423"
title: "Interesting error handling pattern in auth module"
category: "fleeting"
tags:
  - ""
created: "2026-02-13T14:23:45-06:00"
---

# Interesting error handling pattern in auth module

```

### Step 3: Add Content

Fill in your thoughts while they're fresh:

```markdown
---
id: "202602131423"
title: "Interesting error handling pattern in auth module"
category: "fleeting"
tags:
  - "go"
  - "error-handling"
created: "2026-02-13T14:23:45-06:00"
---

# Interesting error handling pattern in auth module

Noticed in the auth service that errors are wrapped with context
at each layer, making debugging much easier.

Instead of just returning the error, each function adds context:

    return fmt.Errorf("authenticate user %s: %w", userID, err)

This creates a chain like:
    "authenticate user 123: validate token: parse claims: invalid signature"

Worth adopting in other services.
```

Save the file (`:w`).

### Step 4: Index Your Notes

Before you can search, index your notes:
```vim
:ZkIndex ~/zettelkasten/
```

Or index just the current directory:
```vim
:ZkIndex
```

### Step 5: Promote to Permanent

A week later, you've used this pattern successfully. Time to promote the note.

Open the fleeting note, then:
```vim
:ZkPromote
```

If you're in a git repository, the project is auto-detected. Otherwise:
```vim
:ZkSetProject my-project
:ZkPromote
```

The note's frontmatter is updated:
```yaml
category: "permanent"
project: "my-project"
```

**Why promote?**
- Permanent notes represent validated knowledge
- They require project context (accountability)
- They're the building blocks of your knowledge base

### Step 6: Link Notes Together

Later, you write another note about Go best practices. Link back to your error handling note.

While editing, insert a link:
```vim
:ZkInsertLink
```

Use Telescope to find the error handling note, press `<CR>`, and a link is inserted:
```markdown
See also [[202602131423]] for error wrapping patterns.
```

Or include the title:
```vim
:ZkInsertLink!
```
```markdown
See [[202602131423|Interesting error handling pattern in auth module]].
```

### Step 7: Find Backlinks

Open your error handling note and see what links to it:
```vim
:ZkBacklinks
```

A panel shows all notes that reference this one - your knowledge graph is forming.

---

## Part 2: Working with Templates

Templates provide structure for common note types. They save time and ensure consistency.

### Available Templates

```vim
:ZkTemplates
```

Output:
```
Available templates:

  book-review     Book review with rating and key takeaways [permanent]
  daily           Daily note for thoughts, tasks, and reflections [fleeting]
  feature         Feature specification with requirements [fleeting]
  meeting         Meeting notes with attendees and action items [fleeting]
  project-idea    Project idea with goals and next steps [fleeting]
  snippet         Code snippet with context and explanation [fleeting]
  user-story      User story with acceptance criteria [fleeting]
```

### Example: Meeting Notes

You're about to join a sprint planning meeting.

```vim
:ZkTemplate meeting
```

Enter the title:
```
Note Title: Sprint 14 Planning
```

The note opens with a pre-filled structure:

```markdown
---
id: "202602131500"
title: "Sprint 14 Planning"
category: "fleeting"
tags:
  - "meeting"
created: "2026-02-13T15:00:00-06:00"
---

# Sprint 14 Planning

**Date:** 2026-02-13
**Attendees:**
-

## Agenda

1.

## Discussion



## Action Items

- [ ]

## Next Steps

```

Fill it in during the meeting. You have a consistent format every time.

### Example: User Story

Product wants a new feature. Capture it properly:

```vim
:ZkTemplate user-story
```

Title: `Add password reset flow`

```markdown
---
id: "202602131530"
title: "Add password reset flow"
category: "fleeting"
tags:
  - "user-story"
  - "requirements"
created: "2026-02-13T15:30:00-06:00"
---

# Add password reset flow

## User Story

**As a** registered user
**I want** to reset my password via email
**So that** I can regain access if I forget my credentials

## Acceptance Criteria

- [ ] Given a valid email, When I request reset, Then I receive an email within 5 minutes
- [ ] Given an invalid token, When I try to reset, Then I see an error message
- [ ] Given a valid token, When I set a new password, Then I can log in immediately

## Priority

- [x] Must Have
- [ ] Should Have
- [ ] Could Have
- [ ] Won't Have (this time)

## Technical Details

- Use existing email service
- Tokens expire after 1 hour
- Rate limit: 3 requests per hour per email

## Related Stories

- [[202602131100|User authentication epic]]
```

### Example: Code Snippet

You found a useful snippet you want to remember:

```vim
:ZkTemplate snippet
```

Title: `Go context with timeout pattern`

The template includes sections for:
- The code itself
- Language and context
- When to use it
- Related snippets

### Using the Template Picker

Don't remember template names? Use the picker:

```vim
:ZkTemplate
```

Telescope shows all templates with descriptions. Select one and continue.

---

## Part 3: Search and Discovery

### Full-Text Search

Find notes containing specific terms:
```vim
:ZkSearch authentication
```

### Live Search

Search as you type:
```vim
:ZkSearch!
```

### Filter by Category

Browse only fleeting notes (your inbox):
```vim
:ZkFleeting
```

Browse permanent notes (your knowledge base):
```vim
:ZkPermanent
```

### Graph Visualization

See how your notes connect:
```vim
:ZkGraph 20
```

This generates a Mermaid diagram showing up to 20 nodes and their relationships.

---

## Part 4: The Promote Workflow

The core Zettelkasten workflow is:

```
Capture (fleeting) → Review → Refine → Promote (permanent)
```

### When to Promote

Promote a fleeting note when:
- You've validated the idea through use
- You've refined the content for clarity
- It connects meaningfully to other notes
- It's worth keeping long-term

### How to Promote

1. Open the fleeting note
2. Review and refine the content
3. Add links to related notes
4. Set the project context (if not already set)
5. Run `:ZkPromote`

### What Changes

- `category` changes from `"fleeting"` to `"permanent"`
- `project` is set (required for permanent notes)
- The file stays in place (no movement required)

---

## Recommended Keymaps

Add these to your NeoVim configuration:

```lua
-- Daily notes
vim.keymap.set("n", "<leader>zd", "<cmd>ZkDaily<cr>", { desc = "Today's daily" })
vim.keymap.set("n", "<leader>zD", "<cmd>ZkDaily yesterday<cr>", { desc = "Yesterday's daily" })

-- Note creation
vim.keymap.set("n", "<leader>zf", function()
    require("zk").create_note("fleeting")
end, { desc = "New fleeting note" })

vim.keymap.set("n", "<leader>zt", "<cmd>ZkTemplate<cr>", { desc = "New from template" })

-- Search
vim.keymap.set("n", "<leader>zz", "<cmd>ZkSearch<cr>", { desc = "Search notes" })
vim.keymap.set("n", "<leader>z/", "<cmd>ZkSearch!<cr>", { desc = "Live search" })

-- Navigation
vim.keymap.set("n", "<leader>zb", "<cmd>ZkBacklinks<cr>", { desc = "Backlinks" })
vim.keymap.set("n", "<leader>zl", "<cmd>ZkInsertLink<cr>", { desc = "Insert link" })
vim.keymap.set("n", "<leader>zg", "<cmd>ZkGraph<cr>", { desc = "Graph" })

-- Management
vim.keymap.set("n", "<leader>zP", "<cmd>ZkPromote<cr>", { desc = "Promote note" })
vim.keymap.set("n", "<leader>zi", "<cmd>ZkIndex<cr>", { desc = "Index notes" })
```

---

## Quick Reference

| Action | Command | Keymap Suggestion |
|--------|---------|-------------------|
| New fleeting note | `:ZkNew fleeting` | `<leader>zf` |
| New from template | `:ZkTemplate` | `<leader>zt` |
| Today's daily | `:ZkDaily` | `<leader>zd` |
| Search notes | `:ZkSearch` | `<leader>zz` |
| Live search | `:ZkSearch!` | `<leader>z/` |
| Insert link | `:ZkInsertLink` | `<leader>zl` |
| Show backlinks | `:ZkBacklinks` | `<leader>zb` |
| Promote note | `:ZkPromote` | `<leader>zP` |
| Generate graph | `:ZkGraph` | `<leader>zg` |
| Index notes | `:ZkIndex` | `<leader>zi` |

---

## Next Steps

### Build a Daily Habit

Daily notes are the foundation of a sustainable practice. They provide:
- A consistent place to capture thoughts
- A natural review rhythm
- A timeline of your thinking

**Read:** [ZK-DAILYNOTES-WORKFLOW-IN-NEOVIM.md](ZK-DAILYNOTES-WORKFLOW-IN-NEOVIM.md) for a comprehensive guide to daily, weekly, and monthly review workflows in NeoVim.

### Manage Tasks with Todos

Todos are a special type of zettel for actionable tasks. They integrate with daily notes and provide:
- Status tracking (open, in progress, closed)
- Due dates and priorities
- Links to related notes

**Read:** [ZK-TODO-WORKFLOW-IN-NEOVIM.md](ZK-TODO-WORKFLOW-IN-NEOVIM.md) for the complete todo workflow guide.

### Master the CLI

The NeoVim plugin wraps the `zk` CLI. For scripting, automation, or terminal use:

**Read:** [ZK-CLI-COMMANDS-CHEATSHEET.md](ZK-CLI-COMMANDS-CHEATSHEET.md) for the complete command reference.

### Get Help

Full documentation is available in NeoVim:
```vim
:help zk
```

CLI help:
```bash
zk --help
zk create --help
zk daily --help
```

---

## Summary

You've learned how to:

1. **Capture** - Create fleeting notes for quick ideas
2. **Refine** - Add content, tags, and links
3. **Promote** - Elevate validated insights to permanent notes
4. **Search** - Find and connect your knowledge
5. **Template** - Use structured formats for consistency

The key is consistency. Capture freely, review regularly, and promote deliberately. Your knowledge base will grow organically as you work.

---

*"The best time to plant a tree was 20 years ago. The second best time is now."*

Start with one note today.
