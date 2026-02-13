# Claude Code Strategy
- **Mode**: Architect & Implementer.
- **Process**:
  1. Compile CUE schema and run `cue vet` for all config changes.
  2. Implement Go logic with 100% test coverage in `internal/`.
  3. Verify CLI output matches the JSON schema expected by Lua.
  4. Implement Lua integration.
- **Constraints**: No external dependencies beyond: `cobra`, `bleve`, `cue`, `yaml.v3`, `plenary.nvim`, and `telescope.nvim`.