# zk - Zettelkasten CLI

A fast, opinionated command-line tool for managing a Zettelkasten note-taking system. Built with Go, validated with CUE schemas, and integrated with NeoVim.

---

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

To improve this section, I’ve refined the flow to distinguish between **Pre-requisites** (the reading) and the **Quick Start** (the action). I also polished the note to sound more like a seasoned engineer sharing a "pro-tip" rather than an afterthought.

---

## Getting Started

Follow these steps to transition from theory to a fully functional, containerized Zettelkasten environment.

### 1. Research & Methodology

Before diving into the code, understand the "why" behind the workflow:

* [Why Zettelkasten?](https://infrashift.github.io/zettelkasten/methodology/why-zettelkasten/) – Core philosophy and benefits.
* [The Tutorial](https://infrashift.github.io/zettelkasten/tutorial/) – A deep dive into the specific implementation used here.

### 2. Quick Start

Deploy the environment in three commands:

1. **Clone the repository:**
```bash
git clone https://github.com/infrashift/zettelkasten.git && cd zettelkasten

```


2. **Spin up the environment:**
Build and run the OCI container. This image comes pre-configured with `tmux`, `LazyVim`, the `Zettelkasten` CLI, and the `ZK` Neovim plugin.
3. **Execute Workflows:**
Open Neovim within the container and follow the interactive workflows defined in the tutorial.

---

> [!TIP]
> **The Default Workspace:** > The container launches a `tmux` session partitioned into three panes: **LazyVim**, a standard **Terminal**, and **Claude Code**.
> To ease the transition from IDEs like VSCode, mouse support is enabled by default in LazyVim. While this helps you get moving immediately, the true power of this stack is unlocked through **Vim motions**—it's worth the investment to learn them!

---

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

---

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

---

## Project Structure

```
zettelkasten/
├── cli/
│   ├── cmd/zk/          # CLI entry point
│   ├── internal/
│   │   ├── config/      # CUE schemas and config loading
│   │   ├── graph/       # Note relationship graph
│   │   ├── index/       # Bleve search index
│   │   └── zettel/      # Note utilities
│   └── testdata/        # Test fixtures
├── lua/zk/              # NeoVim plugin
├── plugin/              # NeoVim autoload
├── doc/                 # NeoVim help
├── after/               # NeoVim ftplugin
├── go.mod, go.sum       # Go module (stays at root)
└── Makefile
```

---

## License

Apache 2.0
