# Development Skills & Rules

## Go Architecture
- Use the **Internal Pattern**: Logic lives in `cli/internal/`, binaries in `cli/cmd/`.
- **CUE Validation**: All YAML frontmatter must be unified with the `#Zettel` schema in `cli/internal/config`.
- **Bleve Mapping**: Use the `keyword` analyzer for `id`, `project`, and `category`. Use the `en` analyzer for `body`.

## Neovim Integration
- **Async Only**: Never block the UI. Use `plenary.job`.
- **snacks.nvim**: Stream JSON-lines from the CLI to the snacks.nvim picker.
- **Ephemeral Files**: MOC (Map of Content) files should be written to a `tmp/` directory which is git-ignored.