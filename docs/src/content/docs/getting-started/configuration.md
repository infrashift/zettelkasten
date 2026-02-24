---
title: Configuration
description: Configure the Zettelkasten CLI with a CUE config file.
---

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

## Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `root_path` | `~/zettelkasten` | Root directory for all notes |
| `index_path` | `.zk_index` | Location of Bleve search index |
| `editor` | `nvim` | Editor for opening notes |
| `folders.untethered` | `untethered` | Subdirectory for untethered notes |
| `folders.tethered` | `tethered` | Subdirectory for tethered notes |
| `folders.tmp` | `tmp` | Subdirectory for temporary notes |
