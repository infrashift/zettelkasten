---
title: ZK MCP Server Tutorial
description: Set up the ZK MCP Server with Claude Code and use AI to search, read, and navigate your zettelkasten.
---

This tutorial walks you through setting up the ZK MCP Server and using Claude
Code as the MCP client to search, read, and navigate your zettelkasten through
natural language.

By the end you will have Claude Code connected to your knowledge graph, able to
search by concept, read notes, follow backlinks, traverse the graph, and review
todos — all without you manually opening files or running CLI commands.

## Prerequisites

- A zettelkasten with at least a few notes (follow the
  [NeoVim ZK Plugin Tutorial](/zettelkasten/tutorial/) to create one, or use an
  existing zettelkasten)
- Notes indexed with `zk index` (the NeoVim plugin does this on save)
- [Claude Code](https://claude.ai/download) installed
- Go 1.22+ (to build `zk-mcp`)

## 1. Build and Install the MCP Server

From the zettelkasten-cli repository root:

```bash
make build-mcp
```

This produces a `zk-mcp` binary in the current directory. Install it to your
PATH:

```bash
make install-mcp
```

Verify it works:

```bash
zk-mcp --help 2>/dev/null; echo $?
# Should print 0 (the server exits cleanly when stdin closes immediately)
```

If you are using the devcontainer, `zk-mcp` is already installed at
`/home/user/.local/bin/zk-mcp`.

## 2. Configure Claude Code

Claude Code needs to know about the MCP server. Add it to your settings.

**Option A: Project-level config** (recommended — keeps the config with your
project):

Create or edit `.claude/settings.json` in your project root:

```json
{
  "mcpServers": {
    "zettelkasten": {
      "command": "zk-mcp"
    }
  }
}
```

**Option B: User-level config** (available in all projects):

Edit `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "zettelkasten": {
      "command": "zk-mcp"
    }
  }
}
```

If `zk-mcp` is not on your PATH, use the full path to the binary:

```json
{
  "mcpServers": {
    "zettelkasten": {
      "command": "/home/user/.local/bin/zk-mcp"
    }
  }
}
```

**In the devcontainer**, create the config at
`/home/user/.claude/settings.json`:

```json
{
  "mcpServers": {
    "zettelkasten": {
      "command": "/home/user/.local/bin/zk-mcp"
    }
  }
}
```

## 3. Verify the Connection

Start (or restart) Claude Code. The MCP server starts automatically in the
background. Verify the tools are available by asking:

```
What zettelkasten tools do you have access to?
```

Claude should list six tools:

- `zk_search` — full-text search with metadata filters
- `zk_read` — read a zettel's frontmatter and body
- `zk_backlinks` — find notes linking to a zettel
- `zk_graph` — BFS traversal of the knowledge graph
- `zk_todos` — list and filter todos
- `zk_list_templates` — list available note templates

If the tools don't appear, check that `zk-mcp` runs without error:

```bash
echo '{}' | zk-mcp 2>&1
```

Common issues:
- The `zk-mcp` binary is not on PATH — use the full path in the config
- No config file found — `zk-mcp` uses the same `~/.config/zk/config.cue` as
  the CLI. Run with defaults if no config exists
- Index doesn't exist — run `zk index ~/zettelkasten` first

## 4. Make Sure Your Index Is Current

The MCP server reads from the Bleve search index. If your notes have changed
since the last index, update it:

```bash
zk index ~/zettelkasten
```

The NeoVim plugin runs `zk index` on every save, so if you edit through NeoVim
the index stays current. But if you create or modify notes outside NeoVim (or
through Claude Code itself), reindex manually.

---

## Using the Tools

The following sections show real prompts and explain what happens behind the
scenes. Try each one against your own zettelkasten.

### Search for notes

Ask Claude to search your knowledge base:

```
Search my zettelkasten for notes about "knowledge management"
```

Claude calls `zk_search` with `query: "knowledge management"`. The response
includes each matching note's ID, title, project, category, tags, and a
relevance score. Claude will summarize the results for you.

You can also filter by metadata:

```
Find all tethered notes tagged "architecture"
```

Claude calls `zk_search` with `category: "tethered"` and `tags: "architecture"`.

```
Show me all notes in the zettelkasten-cli project
```

Claude calls `zk_search` with `project: "zettelkasten-cli"`.

### Read a note

Once Claude finds a note, ask it to read the full content:

```
Read that note and summarize it
```

Claude calls `zk_read` with the ID from the previous search result. It gets
back the complete frontmatter (type, project, category, tags, creation date,
parent, status) and the full markdown body. Claude can then summarize, answer
questions about the content, or use it as context for other tasks.

You can also read by ID directly:

```
Read zettel 20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
```

### Find backlinks

Discover what links TO a note:

```
What notes link to this one?
```

Claude calls `zk_backlinks` with the note's ID. The response lists every note
that contains a `[[id]]` or `[[id|title]]` wiki-link pointing to the target.
Each backlink includes the linking note's ID, title, project, and category.

This is valuable for understanding how an idea connects to the broader knowledge
graph — you see not just what the note links to, but what links back.

### Traverse the graph

Explore the neighborhood around a note:

```
Show me the graph around this note, 3 hops deep
```

Claude calls `zk_graph` with `start_id`, `depth: 3`, and a reasonable `limit`.
The response includes a list of nodes (with their IDs, titles, projects, and
relationship data) and edges (from, to, and relationship type: parent, child, or
link).

This gives Claude a structural understanding of how ideas cluster in your
zettelkasten.

You can also ask for the global graph:

```
Show me the 20 most connected notes in my zettelkasten
```

Claude calls `zk_graph` without a `start_id`, which starts from the
most-connected node and traverses outward.

### Track todos

```
What are my open todos?
```

Claude calls `zk_todos` with `status: "open"`. The response includes each
todo's title, project, priority, due date, and status.

Filter further:

```
Show me overdue high-priority todos
```

Claude calls `zk_todos` with `overdue: true` and `priority: "high"`. Only open
or in-progress todos past their due date are returned.

```
What todos are in the zettelkasten-cli project?
```

Claude calls `zk_todos` with `project: "zettelkasten-cli"`.

### List templates

```
What note templates are available?
```

Claude calls `zk_list_templates` and gets back each template's name,
description, category, type, and default tags. This is useful when you're about
to ask Claude to help you create a note — it knows what structures are available.

---

## Combining Tools

The real power of the MCP server is combining tools in a single conversation.
Here are some practical workflows.

### Research a topic

```
I'm working on the authentication system. Find all notes related to
"authentication", read the most relevant ones, and show me how they connect
in the graph.
```

Claude will:
1. Call `zk_search` to find matching notes
2. Call `zk_read` on the top results to understand the content
3. Call `zk_graph` centered on the most relevant note to show connections
4. Synthesize everything into a coherent summary

### Project review

```
Give me a status report for the zettelkasten-cli project: what notes exist,
what todos are open, and what's overdue.
```

Claude will:
1. Call `zk_search` with `project: "zettelkasten-cli"` for all project notes
2. Call `zk_todos` with `project: "zettelkasten-cli"` for open tasks
3. Call `zk_todos` with `overdue: true` and `project: "zettelkasten-cli"`
4. Present a structured project overview

### Understand a note's context

```
Read this note and show me everything connected to it — backlinks, graph
neighbors, and related todos.
```

Claude will:
1. Call `zk_read` for the note content
2. Call `zk_backlinks` to find incoming links
3. Call `zk_graph` for the broader neighborhood
4. Call `zk_todos` filtered by the note's project
5. Present the full context

### Explore an unfamiliar zettelkasten

If you are working with someone else's zettelkasten (or revisiting your own
after a long break):

```
What are the main topics in this zettelkasten? Show me the most connected
notes and the overall structure.
```

Claude will:
1. Call `zk_graph` with a high limit to see the graph structure
2. Call `zk_search` to identify clusters of related notes
3. Call `zk_todos` to see what work is tracked
4. Map out the major themes and their connections

---

## Using in the Devcontainer

The devcontainer ships both `zk` and `zk-mcp` pre-installed. The tmux layout
includes a Claude Code pane at the bottom-right. To connect them:

1. Switch to the Claude Code pane (click it or `Ctrl-b` then arrow keys)

2. Create the MCP config:

```bash
mkdir -p ~/.claude
cat > ~/.claude/settings.json << 'EOF'
{
  "mcpServers": {
    "zettelkasten": {
      "command": "/home/user/.local/bin/zk-mcp"
    }
  }
}
EOF
```

3. Restart Claude Code (type `/exit` then start it again)

4. Verify: ask Claude what zettelkasten tools it has

Now you have three panes working together:
- **NeoVim** (left) — edit notes, use ZK plugin commands
- **Bash** (top-right) — run CLI commands, git workflows
- **Claude Code** (bottom-right) — AI-assisted search, reading, and navigation
  through the MCP server

This setup lets you ask Claude questions about your notes while editing them.
For example, while writing a new note in NeoVim, ask Claude in the other pane:

```
What notes in my zettelkasten are related to "distributed systems"?
Read the top result and suggest what I should link to from my current note.
```

## Quick Reference

### Tools

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `zk_search` | Full-text search + filters | `query`, `project`, `category`, `type`, `tags`, `limit` |
| `zk_read` | Read note content | `id` (required) |
| `zk_backlinks` | Incoming links | `id` (required) |
| `zk_graph` | Graph traversal | `start_id`, `limit`, `depth` |
| `zk_todos` | Todo tracking | `status`, `priority`, `project`, `overdue`, `limit` |
| `zk_list_templates` | List templates | (none) |

### Resources

| URI | Description |
|-----|-------------|
| `zk://config` | Current configuration |
| `zk://stats` | Note counts and todo stats |
| `zk://zettel/{id}` | Read a zettel by ID |
| `zk://project/{name}/notes` | List notes for a project |

### Troubleshooting

| Problem | Fix |
|---------|-----|
| Tools don't appear | Check `zk-mcp` path in settings.json, restart Claude Code |
| Search returns no results | Run `zk index ~/zettelkasten` to rebuild the index |
| "zettel not found" errors | Verify `root_path` in `~/.config/zk/config.cue` points to your zettelkasten |
| Stale search results | Reindex: `zk index ~/zettelkasten` |
