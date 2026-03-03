---
title: MCP Server
description: Expose your zettelkasten to AI agents via the Model Context Protocol.
---

The `zk-mcp` binary turns your zettelkasten into a structured knowledge source for any MCP-compatible AI client. Instead of the agent reading raw markdown files and guessing at structure, it gets typed tools for search, graph traversal, backlinks, and todo tracking — all backed by the same Bleve index and graph engine the CLI uses.

## Why MCP?

A zettelkasten is more than a directory of files. It has:

- **Structured metadata** — project, category, tags, status, priority, due dates
- **A full-text search index** — Bleve, already maintained by `zk index`
- **A relationship graph** — parent-child links, wiki-links, backlinks

An MCP server exposes all of this as queryable tools. An agent can search by concept, traverse relationships, understand project context, and track tasks — without parsing raw markdown or guessing at conventions.

## Installation

Build from source (requires Go 1.22+):

```bash
# From the repo root
make build-mcp

# Install to $GOPATH/bin
make install-mcp
```

The container image includes `zk-mcp` pre-installed.

## Client Configuration

### Claude Code

Add to your project-level `.claude/settings.json` or `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "zettelkasten": {
      "command": "zk-mcp"
    }
  }
}
```

If `zk-mcp` is not on your `PATH`, use the full path:

```json
{
  "mcpServers": {
    "zettelkasten": {
      "command": "/home/user/.local/bin/zk-mcp"
    }
  }
}
```

### Other MCP Clients

Any client that supports stdio transport can use `zk-mcp`. The server reads configuration from the same `~/.config/zk/config.cue` as the CLI and opens the Bleve index at startup.

## Available Tools

All tools are read-only. They return JSON.

### zk_search

Full-text search with metadata filters.

| Parameter | Type | Description |
|-----------|------|-------------|
| `query` | string | Full-text search query |
| `project` | string | Filter by project name |
| `category` | string | `untethered` or `tethered` |
| `type` | string | `note`, `todo`, `daily-note`, or `issue` |
| `tags` | string | Comma-separated, AND logic |
| `limit` | number | Max results (default 20) |

```
Search my zettelkasten for notes about "graph traversal"
```

### zk_read

Read a single zettel. Returns parsed frontmatter fields plus the markdown body.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | yes | Zettel ID or file path |

### zk_backlinks

Find all notes that link TO a given zettel via `[[id]]` wiki-links.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | yes | Zettel ID or file path |

### zk_graph

BFS traversal of the knowledge graph neighborhood.

| Parameter | Type | Description |
|-----------|------|-------------|
| `start_id` | string | Starting zettel ID (omit for most-connected node) |
| `limit` | number | Max nodes to return (default 10) |
| `depth` | number | Max BFS hops (default unlimited) |

### zk_todos

List and filter todo zettels.

| Parameter | Type | Description |
|-----------|------|-------------|
| `query` | string | Full-text search within todos |
| `status` | string | `open`, `in_progress`, or `closed` |
| `priority` | string | `high`, `medium`, or `low` |
| `project` | string | Filter by project name |
| `overdue` | boolean | Only show todos past their due date |
| `limit` | number | Max results (default 20) |

### zk_list_templates

List all available note templates. Takes no parameters.

## Available Resources

Resources provide context that clients can read on demand.

| Resource | URI | Description |
|----------|-----|-------------|
| Config | `zk://config` | Current zettelkasten configuration |
| Stats | `zk://stats` | Note counts by type/category/project, open todo count |
| Zettel | `zk://zettel/{id}` | Read a single zettel by ID |
| Project | `zk://project/{name}/notes` | List all notes for a project |

## Tutorial: Using with Claude Code

This walkthrough shows how an agent uses the MCP tools to navigate a zettelkasten.

### 1. Verify tools are available

After adding `zk-mcp` to your MCP config, restart Claude Code. You should see the zettelkasten tools in the tool list. You can verify with:

```
What zettelkasten tools do you have access to?
```

The agent should list all six `zk_*` tools.

### 2. Search for notes

Ask the agent to search:

```
Search my zettelkasten for notes about "authentication"
```

The agent calls `zk_search` with `query: "authentication"` and gets back structured results with IDs, titles, projects, tags, and relevance scores.

### 3. Read a specific note

Once the agent finds a relevant result, it can read the full note:

```
Read that note and summarize it
```

The agent calls `zk_read` with the ID from the search results. It gets the complete frontmatter (type, project, category, tags, dates) and the markdown body.

### 4. Explore relationships

The agent can follow the knowledge graph:

```
What notes link to this one? Show me the graph neighborhood.
```

This triggers `zk_backlinks` to find incoming links, then `zk_graph` to show the broader neighborhood — parent-child relationships, wiki-links, and reverse links.

### 5. Review todos

```
Show me all high-priority open todos for the zettelkasten-cli project
```

The agent calls `zk_todos` with `status: "open"`, `priority: "high"`, `project: "zettelkasten-cli"`.

### 6. Get project context

```
What's the current state of the zettelkasten-cli project?
```

The agent can combine `zk_search` (filtered by project), `zk_todos` (open items), and the `zk://stats` resource to give a comprehensive project overview.

## Architecture

`zk-mcp` is a separate binary at `cli/cmd/zk-mcp/` that imports the same internal Go packages as the CLI:

- **`config`** — loads `~/.config/zk/config.cue`
- **`index`** — opens the Bleve index for full-text search
- **`graph`** — builds the relationship graph on-demand from the filesystem
- **`templates`** — lists available note templates

The server communicates over stdio using JSON-RPC 2.0 (the MCP transport protocol). Config is loaded once at startup. The Bleve index is opened once (thread-safe for reads). The graph is built fresh per request so it always reflects the current filesystem state.

## Keeping the Index Current

The MCP server reads from the same Bleve index as `zk search`. If your index is stale, search results will be too. Run:

```bash
zk index
```

The NeoVim plugin runs `zk index` automatically on save, so if you primarily edit through NeoVim, the index stays current.
