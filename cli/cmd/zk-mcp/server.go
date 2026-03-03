package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/mark3labs/mcp-go/server"

	"zk/cli/internal/config"
	"zk/cli/internal/graph"
)

// ServerState holds shared state for tool and resource handlers.
type ServerState struct {
	Config *config.Config
	Index  bleve.Index
}

// newServer creates a configured MCP server and returns it along with a cleanup function.
func newServer() (*server.MCPServer, func(), error) {
	cfg, err := config.Load("")
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}

	idxPath := cfg.IndexPath
	if idxPath == "" {
		idxPath = ".zk_index"
	}
	// Resolve relative index path against RootPath
	if !filepath.IsAbs(idxPath) && cfg.RootPath != "" {
		idxPath = filepath.Join(cfg.RootPath, idxPath)
	}

	idx, err := openIndexReadOnly(idxPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open index at %s: %w", idxPath, err)
	}

	state := &ServerState{Config: cfg, Index: idx}

	s := server.NewMCPServer(
		"zettelkasten",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
	)

	registerTools(s, state)
	registerResources(s, state)

	cleanup := func() {
		if idx != nil {
			idx.Close()
		}
	}

	return s, cleanup, nil
}

// openIndexReadOnly opens an existing Bleve index, or creates one if it doesn't exist.
func openIndexReadOnly(path string) (bleve.Index, error) {
	idx, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		// No index yet — that's fine for a fresh zettelkasten.
		// We still need a valid (empty) index for Search to work.
		return bleve.New(path, bleve.NewIndexMapping())
	}
	return idx, err
}

// resolveZettelPath resolves a zettel ID or file path to an absolute file path.
func resolveZettelPath(cfg *config.Config, idOrPath string) (string, error) {
	// If it looks like a file path, use it directly
	if strings.HasSuffix(idOrPath, ".md") || strings.Contains(idOrPath, "/") {
		if _, err := os.Stat(idOrPath); err != nil {
			return "", fmt.Errorf("file not found: %s", idOrPath)
		}
		return filepath.Abs(idOrPath)
	}

	// Otherwise, it's an ID — search standard directories first
	searchPaths := []string{
		filepath.Join(cfg.RootPath, cfg.Folders.Untethered, idOrPath+".md"),
		filepath.Join(cfg.RootPath, cfg.Folders.Tethered, idOrPath+".md"),
	}

	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			return filepath.Abs(p)
		}
	}

	// Recursive search as last resort
	var found string
	err := filepath.Walk(cfg.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == "ephemeral" {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(path, idOrPath+".md") {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil && err != filepath.SkipAll {
		return "", fmt.Errorf("error searching for zettel: %w", err)
	}

	if found != "" {
		return filepath.Abs(found)
	}

	return "", fmt.Errorf("zettel not found: %s", idOrPath)
}

// buildGraphFromPath walks the directory for .md files and builds a graph.
func buildGraphFromPath(path string) (*graph.Graph, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %q: %w", path, err)
	}

	var files []string
	if info.IsDir() {
		err = filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fi.IsDir() && (strings.HasPrefix(fi.Name(), ".") && fi.Name() != "." || fi.Name() == "ephemeral") {
				return filepath.SkipDir
			}
			if !fi.IsDir() && strings.HasSuffix(p, ".md") {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		files = []string{path}
	}

	g := graph.New()
	for _, filePath := range files {
		node, err := parseZettelForGraph(filePath)
		if err != nil {
			continue
		}
		g.AddNode(node)
	}

	if g.NodeCount() == 0 {
		return nil, fmt.Errorf("no valid zettels found in %s", path)
	}

	g.BuildRelationships()
	return g, nil
}

// parseZettelForGraph reads a markdown file and returns a graph node.
func parseZettelForGraph(filePath string) (*graph.Node, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	z, err := config.ParseFrontmatter(content)
	if err != nil {
		return nil, err
	}

	if z.ID == "" {
		return nil, fmt.Errorf("not a valid zettel: no id field")
	}

	body := extractBody(content)
	links := graph.ExtractLinks(body)

	absPath, _ := filepath.Abs(filePath)

	return &graph.Node{
		ID:       z.ID,
		Title:    z.Title,
		FilePath: absPath,
		Project:  z.Project,
		Category: z.Category,
		Parent:   z.Parent,
		Links:    links,
	}, nil
}

// extractBody extracts the markdown body after frontmatter.
func extractBody(content []byte) string {
	if !bytes.HasPrefix(content, []byte("---")) {
		return string(content)
	}

	rest := content[3:]
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	}

	endIdx := bytes.Index(rest, []byte("\n---"))
	if endIdx == -1 {
		return ""
	}

	body := rest[endIdx+4:]
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	}

	return string(body)
}

// rootPath returns the effective root path for walking zettels.
func (s *ServerState) rootPath() string {
	if s.Config.RootPath != "" {
		if _, err := os.Stat(s.Config.RootPath); err == nil {
			return s.Config.RootPath
		}
	}
	cwd, _ := os.Getwd()
	return cwd
}
