package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"zk/cli/internal/config"
	"zk/cli/internal/index"
)

func registerResources(s *server.MCPServer, state *ServerState) {
	// Static resources
	s.AddResource(configResource(), state.handleConfigResource)
	s.AddResource(statsResource(), state.handleStatsResource)

	// Resource templates
	s.AddResourceTemplate(zettelResourceTemplate(), state.handleZettelResource)
	s.AddResourceTemplate(projectResourceTemplate(), state.handleProjectResource)
}

// --- Resource definitions ---

func configResource() mcp.Resource {
	return mcp.NewResource(
		"zk://config",
		"Zettelkasten Configuration",
		mcp.WithResourceDescription("Current zettelkasten configuration (root path, folders, index path)"),
		mcp.WithMIMEType("application/json"),
	)
}

func statsResource() mcp.Resource {
	return mcp.NewResource(
		"zk://stats",
		"Zettelkasten Statistics",
		mcp.WithResourceDescription("Note counts by type, category, and project; open and overdue todo counts"),
		mcp.WithMIMEType("application/json"),
	)
}

func zettelResourceTemplate() mcp.ResourceTemplate {
	return mcp.NewResourceTemplate(
		"zk://zettel/{id}",
		"Zettel Note",
		mcp.WithTemplateDescription("Read a single zettel by ID — returns frontmatter and body"),
		mcp.WithTemplateMIMEType("application/json"),
	)
}

func projectResourceTemplate() mcp.ResourceTemplate {
	return mcp.NewResourceTemplate(
		"zk://project/{name}/notes",
		"Project Notes",
		mcp.WithTemplateDescription("List all notes belonging to a project"),
		mcp.WithTemplateMIMEType("application/json"),
	)
}

// --- Resource handlers ---

func (s *ServerState) handleConfigResource(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	data, err := json.MarshalIndent(s.Config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "zk://config",
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

func (s *ServerState) handleStatsResource(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	searchPath := s.rootPath()

	type stats struct {
		TotalNotes int            `json:"total_notes"`
		ByType     map[string]int `json:"by_type"`
		ByCategory map[string]int `json:"by_category"`
		ByProject  map[string]int `json:"by_project"`
		OpenTodos  int            `json:"open_todos"`
	}

	st := stats{
		ByType:     make(map[string]int),
		ByCategory: make(map[string]int),
		ByProject:  make(map[string]int),
	}

	err := filepath.Walk(searchPath, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() && (strings.HasPrefix(fi.Name(), ".") && fi.Name() != "." || fi.Name() == "ephemeral") {
			return filepath.SkipDir
		}
		if fi.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}

		content, err := os.ReadFile(p)
		if err != nil {
			return nil
		}

		z, err := config.ParseFrontmatter(content)
		if err != nil {
			return nil
		}
		if z.ID == "" {
			return nil
		}

		st.TotalNotes++
		st.ByType[z.GetType()]++
		st.ByCategory[z.Category]++
		if z.Project != "" {
			st.ByProject[z.Project]++
		}
		if z.IsTodo() && z.Status != "closed" {
			st.OpenTodos++
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal stats: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "zk://stats",
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

func (s *ServerState) handleZettelResource(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	// Extract ID from "zk://zettel/{id}"
	id := strings.TrimPrefix(uri, "zk://zettel/")
	if id == "" || id == uri {
		return nil, fmt.Errorf("invalid zettel URI: %s", uri)
	}

	filePath, err := resolveZettelPath(s.Config, id)
	if err != nil {
		return nil, fmt.Errorf("zettel not found: %w", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	z, err := config.ParseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	body := extractBody(content)

	type zettelResource struct {
		ID        string   `json:"id"`
		Title     string   `json:"title"`
		Type      string   `json:"type"`
		Project   string   `json:"project,omitempty"`
		Category  string   `json:"category"`
		Tags      []string `json:"tags"`
		Created   string   `json:"created"`
		Parent    string   `json:"parent,omitempty"`
		Status    string   `json:"status,omitempty"`
		Due       string   `json:"due,omitempty"`
		Completed string   `json:"completed,omitempty"`
		Priority  string   `json:"priority,omitempty"`
		FilePath  string   `json:"file_path"`
		Body      string   `json:"body"`
	}

	resp := zettelResource{
		ID:        z.ID,
		Title:     z.Title,
		Type:      z.GetType(),
		Project:   z.Project,
		Category:  z.Category,
		Tags:      z.Tags,
		Created:   z.Created.Format("2006-01-02T15:04:05Z07:00"),
		Parent:    z.Parent,
		Status:    z.Status,
		Due:       z.Due,
		Completed: z.Completed,
		Priority:  z.Priority,
		FilePath:  filePath,
		Body:      body,
	}

	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal zettel: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

func (s *ServerState) handleProjectResource(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	// Extract project name from "zk://project/{name}/notes"
	trimmed := strings.TrimPrefix(uri, "zk://project/")
	name := strings.TrimSuffix(trimmed, "/notes")
	if name == "" || name == trimmed {
		return nil, fmt.Errorf("invalid project URI: %s", uri)
	}

	opts := index.SearchOptions{
		Project: name,
		Limit:   100,
	}

	results, total, err := index.Search(s.Index, opts)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	type projectResponse struct {
		Project string               `json:"project"`
		Total   int                  `json:"total"`
		Notes   []index.SearchResult `json:"notes"`
	}

	data, err := json.MarshalIndent(projectResponse{
		Project: name,
		Total:   total,
		Notes:   results,
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal project notes: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}
