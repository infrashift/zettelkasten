---
title: Configuration
description: Configure the Zettelkasten CLI with a CUE config file.
---

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

## Configuration Options

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
