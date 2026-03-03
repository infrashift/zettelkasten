package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"zk/cli/internal/config"
	"zk/cli/internal/graph"
	"zk/cli/internal/index"
	"zk/cli/internal/templates"
)

func registerTools(s *server.MCPServer, state *ServerState) {
	s.AddTool(searchTool(), state.handleSearch)
	s.AddTool(readTool(), state.handleRead)
	s.AddTool(backlinksTool(), state.handleBacklinks)
	s.AddTool(graphTool(), state.handleGraph)
	s.AddTool(todosTool(), state.handleTodos)
	s.AddTool(listTemplatesTool(), state.handleListTemplates)
}

// --- Tool definitions ---

func searchTool() mcp.Tool {
	return mcp.NewTool("zk_search",
		mcp.WithDescription("Full-text search across zettelkasten notes with metadata filters"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("query", mcp.Description("Full-text search query")),
		mcp.WithString("project", mcp.Description("Filter by project name")),
		mcp.WithString("category", mcp.Description("Filter by category (untethered or tethered)")),
		mcp.WithString("type", mcp.Description("Filter by type (note, todo, daily-note, issue)")),
		mcp.WithString("tags", mcp.Description("Comma-separated tags to filter by (AND logic)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default 20)")),
	)
}

func readTool() mcp.Tool {
	return mcp.NewTool("zk_read",
		mcp.WithDescription("Read a zettel — returns parsed frontmatter and markdown body"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("id", mcp.Required(), mcp.Description("Zettel ID or file path")),
	)
}

func backlinksTool() mcp.Tool {
	return mcp.NewTool("zk_backlinks",
		mcp.WithDescription("Find all notes that link TO a given zettel"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("id", mcp.Required(), mcp.Description("Zettel ID or file path")),
	)
}

func graphTool() mcp.Tool {
	return mcp.NewTool("zk_graph",
		mcp.WithDescription("BFS traversal of the knowledge graph neighborhood around a zettel"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("start_id", mcp.Description("Starting zettel ID (omit for most-connected node)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of nodes to return (default 10)")),
		mcp.WithNumber("depth", mcp.Description("Maximum BFS depth (default unlimited)")),
	)
}

func todosTool() mcp.Tool {
	return mcp.NewTool("zk_todos",
		mcp.WithDescription("List and filter todo zettels by status, priority, project, and dates"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("query", mcp.Description("Full-text search within todos")),
		mcp.WithString("status", mcp.Description("Filter by status: open, in_progress, closed")),
		mcp.WithString("priority", mcp.Description("Filter by priority: high, medium, low")),
		mcp.WithString("project", mcp.Description("Filter by project name")),
		mcp.WithBoolean("overdue", mcp.Description("If true, only show todos past their due date")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default 20)")),
	)
}

func listTemplatesTool() mcp.Tool {
	return mcp.NewTool("zk_list_templates",
		mcp.WithDescription("List all available zettel note templates"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

// --- Tool handlers ---

func (s *ServerState) handleSearch(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	opts := index.SearchOptions{
		Query:    request.GetString("query", ""),
		Project:  request.GetString("project", ""),
		Category: request.GetString("category", ""),
		Type:     request.GetString("type", ""),
		Limit:    request.GetInt("limit", 20),
	}

	tagsStr := request.GetString("tags", "")
	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				opts.Tags = append(opts.Tags, t)
			}
		}
	}

	results, total, err := index.Search(s.Index, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	type searchResponse struct {
		Total   int                  `json:"total"`
		Results []index.SearchResult `json:"results"`
	}

	data, err := json.Marshal(searchResponse{Total: total, Results: results})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal failed: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (s *ServerState) handleRead(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	idOrPath, err := request.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	filePath, err := resolveZettelPath(s.Config, idOrPath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
	}

	z, err := config.ParseFrontmatter(content)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse frontmatter: %v", err)), nil
	}

	body := extractBody(content)

	type readResponse struct {
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

	resp := readResponse{
		ID:        z.ID,
		Title:     z.Title,
		Type:      z.GetType(),
		Project:   z.Project,
		Category:  z.Category,
		Tags:      z.Tags,
		Created:   z.Created.Format(time.RFC3339),
		Parent:    z.Parent,
		Status:    z.Status,
		Due:       z.Due,
		Completed: z.Completed,
		Priority:  z.Priority,
		FilePath:  filePath,
		Body:      body,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal failed: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (s *ServerState) handleBacklinks(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	idOrPath, err := request.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Resolve ID from path if needed
	targetID := idOrPath
	if strings.HasSuffix(idOrPath, ".md") || strings.Contains(idOrPath, "/") {
		filePath, err := resolveZettelPath(s.Config, idOrPath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
		}
		z, err := config.ParseFrontmatter(content)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to parse frontmatter: %v", err)), nil
		}
		targetID = z.ID
	}

	searchPath := s.rootPath()

	var files []string
	err = filepath.Walk(searchPath, func(p string, fi os.FileInfo, err error) error {
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
		return mcp.NewToolResultError(fmt.Sprintf("failed to walk directory: %v", err)), nil
	}

	type Backlink struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Project  string `json:"project,omitempty"`
		Category string `json:"category"`
		FilePath string `json:"file_path"`
	}

	var backlinks []Backlink
	for _, fp := range files {
		content, err := os.ReadFile(fp)
		if err != nil {
			continue
		}

		body := extractBody(content)
		links := graph.ExtractLinks(body)

		for _, linkID := range links {
			if linkID == targetID {
				z, err := config.ParseFrontmatter(content)
				if err != nil {
					continue
				}
				absPath, _ := filepath.Abs(fp)
				backlinks = append(backlinks, Backlink{
					ID:       z.ID,
					Title:    z.Title,
					Project:  z.Project,
					Category: z.Category,
					FilePath: absPath,
				})
				break
			}
		}
	}

	type backlinksResponse struct {
		TargetID  string     `json:"target_id"`
		Count     int        `json:"count"`
		Backlinks []Backlink `json:"backlinks"`
	}

	data, err := json.Marshal(backlinksResponse{
		TargetID:  targetID,
		Count:     len(backlinks),
		Backlinks: backlinks,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal failed: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (s *ServerState) handleGraph(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startID := request.GetString("start_id", "")
	limit := request.GetInt("limit", 10)
	depth := request.GetInt("depth", 0)

	searchPath := s.rootPath()
	g, err := buildGraphFromPath(searchPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to build graph: %v", err)), nil
	}

	var nodes []*graph.Node
	if startID != "" {
		if g.GetNode(startID) == nil {
			return mcp.NewToolResultError(fmt.Sprintf("start node %q not found in graph", startID)), nil
		}
		nodes = g.FindConnected(startID, limit, depth)
	} else {
		nodes = g.FindAllConnected(limit, depth)
	}

	edges := g.GetEdges(nodes)

	type nodeJSON struct {
		ID       string   `json:"id"`
		Title    string   `json:"title"`
		Project  string   `json:"project,omitempty"`
		Category string   `json:"category"`
		FilePath string   `json:"file_path"`
		Parent   string   `json:"parent,omitempty"`
		Links    []string `json:"links,omitempty"`
		Children []string `json:"children,omitempty"`
	}

	type edgeJSON struct {
		From  string `json:"from"`
		To    string `json:"to"`
		Label string `json:"label"`
	}

	type graphResponse struct {
		NodeCount int        `json:"node_count"`
		EdgeCount int        `json:"edge_count"`
		Nodes     []nodeJSON `json:"nodes"`
		Edges     []edgeJSON `json:"edges"`
	}

	respNodes := make([]nodeJSON, 0, len(nodes))
	for _, n := range nodes {
		respNodes = append(respNodes, nodeJSON{
			ID:       n.ID,
			Title:    n.Title,
			Project:  n.Project,
			Category: n.Category,
			FilePath: n.FilePath,
			Parent:   n.Parent,
			Links:    n.Links,
			Children: n.Children,
		})
	}

	respEdges := make([]edgeJSON, 0, len(edges))
	for _, e := range edges {
		respEdges = append(respEdges, edgeJSON{
			From:  e.From,
			To:    e.To,
			Label: e.Label,
		})
	}

	data, err := json.Marshal(graphResponse{
		NodeCount: len(respNodes),
		EdgeCount: len(respEdges),
		Nodes:     respNodes,
		Edges:     respEdges,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal failed: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (s *ServerState) handleTodos(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	opts := index.SearchOptions{
		Query:    request.GetString("query", ""),
		Type:     "todo",
		Status:   request.GetString("status", ""),
		Priority: request.GetString("priority", ""),
		Project:  request.GetString("project", ""),
		Limit:    request.GetInt("limit", 20),
	}

	overdue := request.GetBool("overdue", false)
	if overdue {
		today := time.Now().Format("2006-01-02")
		opts.DueBefore = today
		if opts.Status == "" {
			// Only show open/in_progress if overdue and no explicit status filter
			// We can't do OR in the Bleve query easily, so we filter after.
			// Instead, exclude closed by not setting status filter — we'll
			// filter closed items out of the results below.
		}
	}

	results, total, err := index.Search(s.Index, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	// Post-filter: remove closed todos when overdue filter is active
	if overdue {
		filtered := results[:0]
		for _, r := range results {
			if r.Status != "closed" {
				filtered = append(filtered, r)
			}
		}
		results = filtered
		total = len(results)
	}

	type todosResponse struct {
		Total   int                  `json:"total"`
		Results []index.SearchResult `json:"results"`
	}

	data, err := json.Marshal(todosResponse{Total: total, Results: results})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal failed: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (s *ServerState) handleListTemplates(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type templateInfo struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Category    string   `json:"category"`
		Type        string   `json:"type"`
		Tags        []string `json:"tags"`
	}

	names := templates.List()
	sort.Strings(names)

	infos := make([]templateInfo, 0, len(names))
	for _, name := range names {
		t, err := templates.Get(name)
		if err != nil {
			continue
		}
		typ := t.Type
		if typ == "" {
			typ = "note"
		}
		infos = append(infos, templateInfo{
			Name:        t.Name,
			Description: t.Description,
			Category:    t.Category,
			Type:        typ,
			Tags:        t.Tags,
		})
	}

	data, err := json.Marshal(infos)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal failed: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}
