# NeoVim Plugin Installation

The `zk` NeoVim plugin provides seamless integration for creating notes directly from your editor.

## Prerequisites

- NeoVim 0.9+
- [plenary.nvim](https://github.com/nvim-lua/plenary.nvim) (required dependency)
- [snacks.nvim](https://github.com/folke/snacks.nvim) (optional, for picker UI)
- `zk` binary installed and in your `$PATH`

## Installation

### Using lazy.nvim

```lua
{
    "infrashift/zettelkasten-cli",
    dependencies = {
        "nvim-lua/plenary.nvim",
        "folke/snacks.nvim",  -- Optional, for picker UI
    },
    config = function()
        require("zk").setup({
            bin = "zk",  -- Path to zk binary
        })
    end,
}
```

### Using packer.nvim

```lua
use {
    "infrashift/zettelkasten-cli",
    requires = { "nvim-lua/plenary.nvim" },
    config = function()
        require("zk").setup({
            bin = "zk",
        })
    end,
}
```

### Using vim-plug

```vim
Plug 'nvim-lua/plenary.nvim'
Plug 'infrashift/zettelkasten-cli'

" In your init.lua or after/plugin:
lua require("zk").setup({ bin = "zk" })
```

### Manual Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/infrashift/zettelkasten-cli.git ~/.local/share/nvim/site/pack/plugins/start/zettelkasten-cli
   ```

2. Ensure plenary.nvim is also installed

3. Add to your `init.lua`:
   ```lua
   require("zk").setup({
       bin = "zk",
   })
   ```

## Configuration

### Basic Setup

```lua
require("zk").setup({
    bin = "zk",  -- Default: "zk"
})
```

### Custom Binary Path

```lua
require("zk").setup({
    bin = "/usr/local/bin/zk",
})
```

### With Development Build

```lua
require("zk").setup({
    bin = vim.fn.expand("~/projects/zettelkasten-cli/zk"),
})
```

### Full Setup with Tag Completion

```lua
require("zk").setup({
    bin = "zk",
})

-- Enable tag completion in markdown files
require("zk").setup_tag_completion()

-- Register nvim-cmp source (optional, if using nvim-cmp)
require("zk").setup_cmp()
```

## User Commands

The plugin provides user commands that work in command mode (`:ZkCommand`):

| Command | Description |
|---------|-------------|
| `:ZkDaily [date]` | Open daily note (today, "yesterday", or YYYY-MM-DD) |
| `:ZkDailyList[!]` | Browse daily notes (! for this week only) |
| `:ZkNote [category] [project]` | Create new zettel |
| `:ZkTemplate [name]` | Create from template (picker if no name) |
| `:ZkSearch[!] [query]` | Search zettels (! for live search) |
| `:ZkGraph [limit]` | Show graph tree in scratch buffer |
| `:ZkIndex [path]` | Index zettels for searching |
| `:ZkRefreshTags` | Refresh tag cache |

**Examples:**
```vim
:ZkDaily                    " Today's daily note
:ZkDaily yesterday          " Yesterday's note
:ZkTemplate meeting         " Create meeting notes
:ZkSearch! authentication   " Live search for "authentication"
```

## Filetype Settings

When editing a zettel file (markdown with `id:` in frontmatter), the plugin
automatically provides buffer-local keymaps:

| Keymap | Description |
|--------|-------------|
| `<C-x><C-t>` | Tag completion (insert mode) |
| `<localleader>l` | Insert link |
| `<localleader>L` | Insert link with title |
| `<localleader>p` | Preview note |
| `<localleader>b` | Toggle backlinks |
| `<localleader>a` | Add tags to current note |
| `<localleader>v` | Validate frontmatter against CUE schema |
| `<localleader>T` | Tether current note |
| `<localleader>U` | Untether current note |

## Help Documentation

Full documentation is available via `:help zk`.

## Usage

### Create a Note

```lua
-- Create an untethered note (no project required)
require("zk").create_note("untethered")

-- Create an untethered note with explicit project
require("zk").create_note("untethered", "my-project")

-- Create a tethered note (project required)
require("zk").create_note("tethered", "my-project")
```

### Tether a Note

```lua
-- Tether current file (move to tethered, auto-detect project from git)
require("zk").tether_note()

-- Tether current file with explicit project
require("zk").tether_note(nil, "my-project")

-- Tether specific file
require("zk").tether_note("/path/to/note.md", "my-project")
```

### Untether a Note

```lua
-- Untether current file (move to untethered, auto-detect project from git)
require("zk").untether_note()

-- Untether current file with explicit project
require("zk").untether_note(nil, "my-project")

-- Untether specific file
require("zk").untether_note("/path/to/note.md", "my-project")
```

### Set Project

```lua
-- Set project on current file (prompts for project name)
require("zk").set_project()

-- Set project on current file with explicit value
require("zk").set_project(nil, "my-project")

-- Set project on specific file
require("zk").set_project("/path/to/note.md", "my-project")
```

### Search

```lua
-- Basic search (prints results)
require("zk").search("authentication")

-- Search with filters
require("zk").search("auth", {
    project = "my-project",
    category = "tethered",
    tags = { "security" },
    limit = 10,
})

-- Search with callback
require("zk").search("query", {
    on_results = function(results)
        for _, r in ipairs(results) do
            print(r.title .. " - " .. r.file_path)
        end
    end,
})
```

### Index

```lua
-- Index current directory
require("zk").index()

-- Index specific path
require("zk").index("~/zettelkasten/")
```

### Graph Visualization

```lua
-- Show graph tree for current directory
require("zk").graph()

-- Show with options
require("zk").graph({
    path = "~/zettelkasten/",
    limit = 20,               -- Max nodes to display
    depth = 3,                -- Max BFS hops
    start = "<id>",           -- Center on a zettel
})
```

The ASCII tree opens in a scratch buffer. Press `q` to close.

### Preview Note in Floating Window

```lua
-- Preview current file in floating window
require("zk").preview_note()

-- Preview specific file
require("zk").preview_note("/path/to/note.md")

-- Preview by ID (searches index first)
require("zk").preview_by_id("202602131045")
```

**Floating window keymaps:**
- `q` or `<Esc>` - Close the preview
- `<CR>` - Open the note in the current buffer

### Insert Links Between Notes

Insert `[[id]]` or `[[id|title]]` style links at the cursor position:

```lua
-- Open picker to search and insert link
require("zk").link_picker()

-- Insert [[id|title]] format by default
require("zk").link_picker({ include_title = true })

-- Directly insert link by ID
require("zk").insert_link("202602131045")

-- Insert link with title
require("zk").insert_link("202602131045", "My Note Title", true)

-- Prompt for ID and insert
require("zk").insert_link_prompt()        -- [[id]] format
require("zk").insert_link_prompt(true)    -- [[id|title]] format
```

### Tag Completion

Complete tags from your zettel collection while editing frontmatter:

```lua
-- Enable tag completion for markdown files
require("zk").setup_tag_completion()

-- Manual tag completion (in insert mode in tags section)
-- Press <C-x><C-t> to trigger

-- Get all tags (async)
require("zk").get_tags(function(tags)
    print("Found " .. #tags .. " tags")
end)

-- Refresh tag cache
require("zk").refresh_tags()
```

**With nvim-cmp:**

```lua
-- In your cmp setup
local cmp = require("cmp")
require("zk").setup_cmp()  -- Register zk_tags source

cmp.setup({
    sources = cmp.config.sources({
        { name = "nvim_lsp" },
        { name = "zk_tags" },  -- Add this source
        { name = "buffer" },
    }),
})
```

The completion automatically activates when you're in the `tags:` section of YAML frontmatter.

### Backlinks Panel

View all notes that link to the current note:

```lua
-- Open floating backlinks panel (right side)
require("zk").backlinks_panel()

-- Toggle the panel
require("zk").toggle_backlinks()

-- Open in a split instead
require("zk").backlinks_split()
require("zk").backlinks_split({ position = "left" })
require("zk").backlinks_split({ position = "bottom" })

-- Get backlinks for a specific ID
require("zk").backlinks_panel({ id = "202602131045" })

-- Get backlinks programmatically
require("zk").get_backlinks("202602131045", function(backlinks)
    for _, bl in ipairs(backlinks) do
        print(bl.title .. " -> " .. bl.file_path)
    end
end)
```

**Panel keymaps:**
- `<CR>` or `o` - Open selected note
- `p` - Preview selected note in floating window
- `q` or `<Esc>` - Close panel

### Note Templates

Create notes from predefined templates:

```lua
-- Create a note from a template (prompts for title)
require("zk").create_from_template("meeting")
require("zk").create_from_template("user-story")
require("zk").create_from_template("feature")

-- With explicit project
require("zk").create_from_template("meeting", "my-project")

-- Open picker to select a template
require("zk").template_picker()
require("zk").template_picker({ project = "my-project" })

-- Get template info
local template = require("zk").get_template("meeting")
print(template.description)  -- "Meeting notes with attendees and action items"
```

**Available templates:**
- `meeting` - Meeting notes with attendees and action items
- `book-review` - Book review with rating and key takeaways
- `snippet` - Code snippet with context and explanation
- `project-idea` - Project idea with goals and next steps
- `user-story` - User story in standard format with acceptance criteria
- `feature` - Feature specification with requirements and design notes
- `daily` - Daily note for thoughts, tasks, and reflections

### Daily Notes

Create and manage daily notes:

```lua
-- Open or create today's daily note
require("zk").daily()

-- Open yesterday's daily note (for morning review)
require("zk").daily({ date = "yesterday" })

-- Open a specific date's daily note
require("zk").daily({ date = "2026-02-10" })

-- Browse daily notes with picker
require("zk").daily_picker()
require("zk").daily_picker({ week = true })   -- This week only
require("zk").daily_picker({ month = true })  -- This month only

-- List daily notes programmatically (async)
require("zk").list_daily(function(notes)
    for _, note in ipairs(notes) do
        print(note.date .. ": " .. note.title)
    end
end)

-- List daily notes synchronously
local notes = require("zk").list_daily_sync()
local this_week = require("zk").list_daily_sync({ week = true })
```

Daily notes are idempotent - running `daily()` multiple times on the same day opens the same file.

See [ZK-DAILYNOTES-WORKFLOW-IN-NEOVIM.md](ZK-DAILYNOTES-WORKFLOW-IN-NEOVIM.md) for a comprehensive guide to daily note workflows.

### Picker Integration

If you have snacks.nvim installed, you get a powerful search UI:

```lua
-- Search all zettels
require("zk.picker").search()

-- Live search (updates as you type)
require("zk.picker").live_search()

-- Search with filters
require("zk.picker").search({
    project = "my-project",
    category = "tethered",
})
```

**Picker keymaps:**
- `<CR>` - Open note in current buffer
- `<C-p>` - Open note in floating preview window
- `<C-l>` - Insert link as `[[id]]` at cursor
- `<C-S-l>` - Insert link as `[[id|title]]` at cursor

### Keybindings

Add these to your NeoVim configuration:

```lua
-- Quick untethered note (no project)
vim.keymap.set("n", "<leader>zf", function()
    require("zk").create_note("untethered")
end, { desc = "Create untethered note" })

-- Quick tethered note (will use git project)
vim.keymap.set("n", "<leader>zp", function()
    require("zk").create_note("tethered")
end, { desc = "Create tethered note" })

-- Tether current note
vim.keymap.set("n", "<leader>zT", function()
    require("zk").tether_note()
end, { desc = "Tether current note" })

-- Untether current note
vim.keymap.set("n", "<leader>zU", function()
    require("zk").untether_note()
end, { desc = "Untether current note" })

-- Set project on current note
vim.keymap.set("n", "<leader>zs", function()
    require("zk").set_project()
end, { desc = "Set project" })

-- Search with picker
vim.keymap.set("n", "<leader>zz", function()
    require("zk.picker").search()
end, { desc = "Search zettels" })

-- Live search
vim.keymap.set("n", "<leader>z/", function()
    require("zk.picker").live_search()
end, { desc = "Live search zettels" })

-- Index current directory
vim.keymap.set("n", "<leader>zi", function()
    require("zk").index()
end, { desc = "Index zettels" })

-- Show graph tree
vim.keymap.set("n", "<leader>zg", function()
    require("zk").graph()
end, { desc = "Show graph tree" })

-- Preview current note in floating window
vim.keymap.set("n", "<leader>zv", function()
    require("zk").preview_note()
end, { desc = "Preview note" })

-- Preview note by ID
vim.keymap.set("n", "<leader>zV", function()
    require("zk").preview_by_id()
end, { desc = "Preview by ID" })

-- Insert link (opens picker)
vim.keymap.set("n", "<leader>zl", function()
    require("zk").link_picker()
end, { desc = "Insert link" })

-- Insert link with title
vim.keymap.set("n", "<leader>zL", function()
    require("zk").link_picker({ include_title = true })
end, { desc = "Insert link with title" })

-- Toggle backlinks panel
vim.keymap.set("n", "<leader>zb", function()
    require("zk").toggle_backlinks()
end, { desc = "Toggle backlinks" })

-- Open backlinks in split
vim.keymap.set("n", "<leader>zB", function()
    require("zk").backlinks_split()
end, { desc = "Backlinks split" })

-- Create note from template (opens picker)
vim.keymap.set("n", "<leader>zt", function()
    require("zk").template_picker()
end, { desc = "Create from template" })

-- Quick meeting notes
vim.keymap.set("n", "<leader>zm", function()
    require("zk").create_from_template("meeting")
end, { desc = "Create meeting notes" })

-- Today's daily note
vim.keymap.set("n", "<leader>zd", function()
    require("zk").daily()
end, { desc = "Today's daily note" })

-- Yesterday's daily note (morning review)
vim.keymap.set("n", "<leader>zD", function()
    require("zk").daily({ date = "yesterday" })
end, { desc = "Yesterday's daily note" })

-- Browse daily notes
vim.keymap.set("n", "<leader>zw", function()
    require("zk").daily_picker()
end, { desc = "Browse daily notes" })
```

### With Which-Key

```lua
local wk = require("which-key")
wk.register({
    z = {
        name = "Zettelkasten",
        f = { function() require("zk").create_note("untethered") end, "Untethered note" },
        p = { function() require("zk").create_note("tethered") end, "Tethered note" },
        T = { function() require("zk").tether_note() end, "Tether current note" },
        U = { function() require("zk").untether_note() end, "Untether current note" },
        s = { function() require("zk").set_project() end, "Set project" },
        g = { function() require("zk").graph() end, "Show graph tree" },
        v = { function() require("zk").preview_note() end, "Preview note" },
        V = { function() require("zk").preview_by_id() end, "Preview by ID" },
        l = { function() require("zk").link_picker() end, "Insert link" },
        L = { function() require("zk").link_picker({ include_title = true }) end, "Insert link with title" },
        b = { function() require("zk").toggle_backlinks() end, "Toggle backlinks" },
        B = { function() require("zk").backlinks_split() end, "Backlinks split" },
        t = { function() require("zk").template_picker() end, "Create from template" },
        m = { function() require("zk").create_from_template("meeting") end, "Meeting notes" },
        d = { function() require("zk").daily() end, "Today's daily" },
        D = { function() require("zk").daily({ date = "yesterday" }) end, "Yesterday's daily" },
        w = { function() require("zk").daily_picker() end, "Browse daily notes" },
    },
}, { prefix = "<leader>" })
```

## How It Works

1. When you call `create_note()`, NeoVim prompts for a note title
2. The plugin invokes `zk create "title" --category <category>`
3. The `zk` binary detects your current git project automatically
4. A confirmation message appears on success

## Troubleshooting

### "zk: command not found"

Ensure the `zk` binary is in your `$PATH`:

```bash
# Check if zk is accessible
which zk

# If not, add to PATH or specify full path in setup
require("zk").setup({
    bin = "/full/path/to/zk",
})
```

### "No module named 'plenary'"

Install plenary.nvim:

```lua
-- lazy.nvim
{ "nvim-lua/plenary.nvim" }

-- packer.nvim
use "nvim-lua/plenary.nvim"
```

### Notes Not Created

1. Check that `zk create` works from terminal
2. Verify you're in a git repository (for project detection)
3. Check NeoVim messages with `:messages`

## API Reference

### `setup(opts)`

Initialize the plugin with configuration options.

```lua
require("zk").setup({
    bin = "zk",  -- Path to zk binary (default: "zk")
})
```

### `create_note(note_category, project)`

Create a new note with the specified category.

**Parameters:**
- `note_category` (string): Either `"untethered"` or `"tethered"`
- `project` (string, optional): Project context. If nil, auto-detected from git.

**Behavior:**
1. Prompts for note title via `vim.fn.input()`
2. Executes `zk create` asynchronously via plenary.job
3. Prints success/failure message

```lua
require("zk").create_note("untethered")
require("zk").create_note("tethered", "my-project")
```

### `tether_note(file_path, project)`

Tether an untethered note (move to tethered).

**Parameters:**
- `file_path` (string, optional): Path to the note. Defaults to current buffer.
- `project` (string, optional): Project context. If nil, auto-detected from git.

**Behavior:**
1. Executes `zk tether` asynchronously
2. Reloads the buffer if the tethered file is currently open
3. Prints success/failure message

```lua
require("zk").tether_note()  -- Current file, auto-detect project
require("zk").tether_note(nil, "my-project")  -- Current file, explicit project
```

### `untether_note(file_path, project)`

Untether a tethered note (move to untethered).

**Parameters:**
- `file_path` (string, optional): Path to the note. Defaults to current buffer.
- `project` (string, optional): Project context. If nil, auto-detected from git.

**Behavior:**
1. Executes `zk untether` asynchronously
2. Reloads the buffer if the untethered file is currently open
3. Prints success/failure message

```lua
require("zk").untether_note()  -- Current file, auto-detect project
require("zk").untether_note(nil, "my-project")  -- Current file, explicit project
```

### `set_project(file_path, project)`

Set or update the project for a zettel.

**Parameters:**
- `file_path` (string, optional): Path to the note. Defaults to current buffer.
- `project` (string, optional): Project name. If nil, prompts for input.

**Behavior:**
1. Prompts for project name if not provided
2. Executes `zk set-project` asynchronously
3. Reloads the buffer if the modified file is currently open
4. Prints success/failure message

```lua
require("zk").set_project()  -- Current file, prompt for project
require("zk").set_project(nil, "my-project")  -- Current file, explicit project
```

### `search(query, opts)`

Search zettels with optional filters.

**Parameters:**
- `query` (string, optional): Full-text search query.
- `opts` (table, optional):
  - `project` (string): Filter by project
  - `category` (string): Filter by category
  - `tags` (table): Filter by tags (AND logic)
  - `limit` (number): Max results
  - `on_results` (function): Callback receiving results array

```lua
require("zk").search("authentication")
require("zk").search("query", { project = "my-project", on_results = function(r) ... end })
```

### `index(path)`

Index zettels for searching.

**Parameters:**
- `path` (string, optional): Path to index. Defaults to current directory.

```lua
require("zk").index()  -- Index cwd
require("zk").index("~/zettelkasten/")
```

### `graph(opts)`

Show an ASCII tree of note relationships in a scratch buffer.

**Parameters:**
- `opts` (table, optional):
  - `path` (string): Path to scan. Defaults to current directory.
  - `limit` (number): Maximum nodes to display. Defaults to 10.
  - `depth` (number): Maximum BFS hops from start. Defaults to 0 (unlimited).
  - `start` (string): Center graph on a specific zettel ID.

**Behavior:**
1. Executes `zk graph` asynchronously
2. Opens the ASCII tree output in a scratch buffer (`buftype=nofile`)
3. Press `q` to close the buffer

```lua
require("zk").graph()  -- Graph cwd with defaults
require("zk").graph({ limit = 20, path = "~/zettelkasten/" })
```

### `preview_note(file_path)`

Preview a note in a floating window.

**Parameters:**
- `file_path` (string, optional): Path to the note. Defaults to current buffer.

**Behavior:**
1. Reads file content
2. Opens a centered floating window with rounded border
3. Sets up keymaps: `q`/`<Esc>` to close, `<CR>` to open in buffer

**Returns:** Table with `buf` (buffer handle) and `win` (window handle)

```lua
require("zk").preview_note()  -- Preview current file
require("zk").preview_note("/path/to/note.md")
```

### `preview_by_id(id)`

Preview a note by its ID (searches the index).

**Parameters:**
- `id` (string, optional): 12-digit zettel ID. Prompts for input if nil.

**Behavior:**
1. Searches for note with matching ID
2. Opens floating preview if found

```lua
require("zk").preview_by_id("202602131045")
require("zk").preview_by_id()  -- Prompts for ID
```

### `insert_link(id, title, include_title)`

Insert a zettel link at the cursor position.

**Parameters:**
- `id` (string): The 12-digit zettel ID
- `title` (string, optional): Note title for `[[id|title]]` format
- `include_title` (boolean): If true and title provided, uses `[[id|title]]` format

**Behavior:**
1. Formats link as `[[id]]` or `[[id|title]]`
2. Inserts at cursor position
3. Moves cursor to end of inserted link

```lua
require("zk").insert_link("202602131045")  -- Inserts [[202602131045]]
require("zk").insert_link("202602131045", "My Note", true)  -- Inserts [[202602131045|My Note]]
```

### `insert_link_prompt(include_title)`

Prompt for a zettel ID and insert a link.

**Parameters:**
- `include_title` (boolean, optional): If true, searches for title and uses `[[id|title]]` format

```lua
require("zk").insert_link_prompt()       -- Prompts for ID, inserts [[id]]
require("zk").insert_link_prompt(true)   -- Prompts for ID, inserts [[id|title]]
```

### `link_picker(opts)`

Open a picker to search and insert a link. Requires snacks.nvim.

**Parameters:**
- `opts` (table, optional):
  - `include_title` (boolean): Default format for `<CR>` action
  - `query` (string): Initial search query

**Picker keymaps:**
- `<CR>` - Insert link (format depends on `include_title` option)
- `<C-t>` - Insert link with title `[[id|title]]`

```lua
require("zk").link_picker()  -- Opens picker, <CR> inserts [[id]]
require("zk").link_picker({ include_title = true })  -- <CR> inserts [[id|title]]
```

### `get_tags(callback)`

Get all unique tags from indexed zettels (async).

**Parameters:**
- `callback` (function): Called with sorted list of tags

```lua
require("zk").get_tags(function(tags)
    for _, tag in ipairs(tags) do
        print(tag)
    end
end)
```

### `get_tags_sync()`

Get all tags synchronously. Uses a 60-second cache.

**Returns:** Sorted list of tag strings

```lua
local tags = require("zk").get_tags_sync()
```

### `refresh_tags()`

Clear the tag cache and reload tags.

```lua
require("zk").refresh_tags()
```

### `complete_tags()`

Trigger manual tag completion at cursor using vim's completion menu.

```lua
require("zk").complete_tags()
```

### `setup_tag_completion()`

Set up automatic tag completion for markdown files.

**Behavior:**
1. Sets `omnifunc` for markdown files
2. Adds `<C-x><C-t>` keymap for tag completion
3. Only completes when cursor is in frontmatter `tags:` section

```lua
require("zk").setup_tag_completion()
```

### `setup_cmp()`

Register nvim-cmp source for tag completion.

**Returns:** `true` if nvim-cmp is available, `false` otherwise

```lua
if require("zk").setup_cmp() then
    print("nvim-cmp source registered")
end
```

### `get_backlinks(id_or_file, callback)`

Get all notes that link to the specified zettel (async).

**Parameters:**
- `id_or_file` (string): Zettel ID or file path
- `callback` (function): Called with list of backlink objects

**Backlink object fields:**
- `id`: Zettel ID
- `title`: Note title
- `project`: Project name
- `category`: "untethered" or "tethered"
- `file_path`: Absolute path to file

```lua
require("zk").get_backlinks("202602131045", function(backlinks)
    print("Found " .. #backlinks .. " backlinks")
end)
```

### `get_backlinks_sync(id_or_file)`

Get backlinks synchronously.

**Returns:** List of backlink objects

```lua
local backlinks = require("zk").get_backlinks_sync("202602131045")
```

### `backlinks_panel(opts)`

Open a floating backlinks panel for the current note.

**Parameters:**
- `opts` (table, optional):
  - `id` (string): Zettel ID to show backlinks for
  - `file` (string): File path to show backlinks for

**Panel keymaps:**
- `<CR>` / `o` - Open selected note
- `p` - Preview in floating window
- `q` / `<Esc>` - Close panel

**Returns:** Table with `buf`, `win`, and `backlinks`

```lua
require("zk").backlinks_panel()  -- Current note
require("zk").backlinks_panel({ id = "202602131045" })
```

### `backlinks_split(opts)`

Open backlinks in a split window.

**Parameters:**
- `opts` (table, optional):
  - `position` (string): "right" (default), "left", "bottom", or "top"
  - `id` (string): Zettel ID
  - `file` (string): File path

```lua
require("zk").backlinks_split()
require("zk").backlinks_split({ position = "bottom" })
```

### `toggle_backlinks(opts)`

Toggle the floating backlinks panel.

```lua
require("zk").toggle_backlinks()
```

### `templates`

Table containing template metadata. Available templates: `meeting`, `book-review`, `snippet`, `project-idea`, `user-story`, `feature`.

```lua
for name, meta in pairs(require("zk").templates) do
    print(name .. ": " .. meta.description)
end
```

### `get_template(name)`

Get template metadata by name.

**Parameters:**
- `name` (string): Template name

**Returns:** Table with `name`, `description`, `category`, `tags` or `nil` if not found

```lua
local tmpl = require("zk").get_template("meeting")
print(tmpl.description)  -- "Meeting notes with attendees and action items"
print(tmpl.category)     -- "untethered"
```

### `create_from_template(template_name, project)`

Create a note from a template.

**Parameters:**
- `template_name` (string): Template name (e.g., `"meeting"`, `"user-story"`)
- `project` (string, optional): Project context. If nil, auto-detected from git.

**Behavior:**
1. Prompts for note title
2. Executes `zk create --template <name>` asynchronously
3. Prints success/failure message

```lua
require("zk").create_from_template("meeting")
require("zk").create_from_template("feature", "my-project")
```

### `template_picker(opts)`

Open a picker to select a template and create a note. Requires snacks.nvim.

**Parameters:**
- `opts` (table, optional):
  - `project` (string): Project context for the new note

**Behavior:**
1. Displays all available templates with descriptions
2. On selection, prompts for note title
3. Creates note using selected template

```lua
require("zk").template_picker()
require("zk").template_picker({ project = "my-project" })
```

### `daily(opts)`

Create or open a daily note. Daily notes are idempotent - the same file is returned for the same date.

**Parameters:**
- `opts` (table, optional):
  - `date` (string): Date in `YYYY-MM-DD` format, or `"yesterday"` for yesterday

**Behavior:**
1. Determines target date (defaults to today)
2. Creates daily note if it doesn't exist
3. Opens the daily note in the current buffer

```lua
require("zk").daily()                       -- Today
require("zk").daily({ date = "yesterday" }) -- Yesterday
require("zk").daily({ date = "2026-02-10" }) -- Specific date
```

### `list_daily(opts, callback)`

Get daily notes asynchronously.

**Parameters:**
- `opts` (table, optional):
  - `week` (boolean): Show only this week's notes
  - `month` (boolean): Show only this month's notes
- `callback` (function): Called with list of daily note objects

**Daily note object fields:**
- `date`: Date string (YYYY-MM-DD)
- `title`: Note title
- `file_path`: Absolute path to file

```lua
require("zk").list_daily({ week = true }, function(notes)
    for _, note in ipairs(notes) do
        print(note.date .. ": " .. note.file_path)
    end
end)
```

### `list_daily_sync(opts)`

Get daily notes synchronously.

**Parameters:**
- `opts` (table, optional): Same as `list_daily`

**Returns:** List of daily note objects

```lua
local notes = require("zk").list_daily_sync()
local this_week = require("zk").list_daily_sync({ week = true })
```

### `daily_picker(opts)`

Open a picker to browse daily notes. Requires snacks.nvim.

**Parameters:**
- `opts` (table, optional):
  - `week` (boolean): Show only this week's notes
  - `month` (boolean): Show only this month's notes

```lua
require("zk").daily_picker()
require("zk").daily_picker({ week = true })
require("zk").daily_picker({ month = true })
```

## Picker API

Requires snacks.nvim to be installed.

### `require("zk.picker").search(opts)`

Open picker with all indexed zettels.

### `require("zk.picker").live_search(opts)`

Open picker with live search (results update as you type).

### `require("zk.picker").untethered(opts)`

Browse only untethered notes.

### `require("zk.picker").tethered(opts)`

Browse only tethered notes.

**Common opts:**
- `project` (string): Filter by project
- `category` (string): Filter by category
- `tags` (table): Filter by tags
- `limit` (number): Max results

### `require("zk.picker").insert_link(opts)`

Open picker specifically for inserting links. Same as `require("zk").link_picker(opts)`.

**Opts:**
- `include_title` (boolean): Use `[[id|title]]` format by default

## Future Features

- Project completion
