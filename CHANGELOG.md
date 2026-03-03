# Changelog

All notable changes to this project are documented here.

## 2026-03-02

### Changed
- **Moved Go CLI source into `cli/` subdirectory** — `cmd/`, `internal/`, and `testdata/` now live under `cli/` to cleanly separate the Go CLI from the NeoVim plugin files at the repo root. `go.mod` remains at root; import paths gain a `cli/` segment.

### Added
- **Graph export to portable markdown** (`zk export`, `:ZkExport`)
  - Runs the same BFS query as `zk graph` (--limit, --depth, --start)
  - Copies notes into `<root>/ephemeral/` with `[[id|title]]` links converted to `[title](id.md)` relative links
  - Ephemeral directory has a `.gitignore` that ignores all exported files
  - `TransformLinks()` function in `internal/graph` with full test coverage
- **Shared graph helpers** (`buildGraphFromPath`, `queryGraph`) extracted from `graphCmd` for reuse by `exportCmd`
- **Ephemeral directory skip** added to all file walkers (index, graph, backlinks, resolveZettelPath)

### Changed
- **Replaced telescope.nvim with snacks.nvim** for picker UI
  - Removed `lua/zk/telescope.lua`
  - Updated `lua/zk/init.lua` with snacks.nvim picker integration
  - Updated plugin installation docs and all workflow guides
- **Refactored `graphCmd`** to use shared `buildGraphFromPath`/`queryGraph` helpers
- **Container config** updated for snacks.nvim dependency

### Fixed
- **Docs site 404s**: internal links used `/zettelkasten-cli/` base path instead of `/zettelkasten/` (matching the GitHub repo name)
- **Docs repo references**: clone URLs, plugin specs, and directory names corrected from `zettelkasten-cli` to `zettelkasten`
