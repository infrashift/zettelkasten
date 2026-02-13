# PRD: zk - Zettelkasten CLI for Neovim

## 1. Overview
A high-performance Zettelkasten management system consisting of a Go CLI (`zk`) and a Neovim integration. The system prioritizes data integrity, structured search, and interconnected knowledge.

## 2. Technical Stack
- **CLI**: Go 1.22+ (using `cobra`)
- **Indexing**: Bleve (Full-text + Structured)
- **Validation**: CUE (Schema-first metadata)
- **Integration**: Neovim (Lua), Telescope, Plenary.job
- **Version Control**: Git (Used as the storage and backup engine)

## 3. Core Requirements
- **Automated Metadata**: Extract Git project context during note creation.
- **Categorization**: Support 'fleeting' and 'permanent' note lifecycles.
- **Verification**: Use CUE to validate YAML frontmatter before indexing.
- **Search**: Structured search via Telescope (e.g., `project:my-repo title:config`).
- **Graphing**: Generate local relationship graphs (Forward/Backlinks) in Markdown/Mermaid.