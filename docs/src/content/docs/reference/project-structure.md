---
title: Project Structure
description: Development commands and source code layout for the Zettelkasten CLI.
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

## Source Layout

```
zettelkasten/
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
