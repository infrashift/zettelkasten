package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"zk/cli/internal/config"
	"zk/cli/internal/index"
)

// testIDs used across tests
const (
	testID1 = "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	testID2 = "20260213110000-ffffffff-1111-2222-3333-444444444444"
	testID3 = "20260213120000-11111111-2222-3333-4444-555555555555"
	testID4 = "20260213130000-22222222-3333-4444-5555-666666666666"
)

// setupTestZettelkasten creates a temp directory with test zettels and returns
// a configured ServerState and cleanup function.
func setupTestZettelkasten(t *testing.T) (*ServerState, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	untethered := filepath.Join(tmpDir, "untethered")
	tethered := filepath.Join(tmpDir, "tethered")
	if err := os.MkdirAll(untethered, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(tethered, 0755); err != nil {
		t.Fatal(err)
	}

	// Note 1: basic note
	writeFile(t, filepath.Join(untethered, testID1+".md"), `---
id: "`+testID1+`"
title: "Test Zettel"
project: "zettelkasten-cli"
category: "untethered"
tags:
  - "test"
  - "example"
created: "2026-02-13T10:45:00Z"
---

# Test Zettel

This is a valid zettel for testing.
`)

	// Note 2: child note with parent link and wiki-link back
	writeFile(t, filepath.Join(untethered, testID2+".md"), `---
id: "`+testID2+`"
title: "Child Zettel"
project: "zettelkasten-cli"
category: "untethered"
tags:
  - "child"
created: "2026-02-13T11:00:00Z"
parent: "`+testID1+`"
---

# Child Zettel

This zettel links to a parent for graph construction.

See also: [[`+testID1+`|Test Zettel]] for the parent note.
`)

	// Note 3: standalone note, no project
	writeFile(t, filepath.Join(untethered, testID3+".md"), `---
id: "`+testID3+`"
title: "Quick Idea"
category: "untethered"
tags:
  - "idea"
created: "2026-02-13T12:00:00Z"
---

# Quick Idea

A fleeting note captured without project context.
`)

	// Note 4: todo
	writeFile(t, filepath.Join(untethered, testID4+".md"), `---
id: "`+testID4+`"
title: "Fix the build"
type: "todo"
project: "zettelkasten-cli"
category: "untethered"
status: "open"
priority: "high"
due: "2026-02-15"
tags:
  - "todo"
  - "bug"
created: "2026-02-13T13:00:00Z"
---

# Fix the build

The CI pipeline is broken. Need to update Go version.
`)

	// Build config
	cfg := &config.Config{
		RootPath:  tmpDir,
		IndexPath: filepath.Join(tmpDir, ".zk_index"),
		TodosPath: filepath.Join(tmpDir, ".zk_todos"),
		Editor:    "nvim",
		Folders: config.Folders{
			Untethered: "untethered",
			Tethered:   "tethered",
			Tmp:        "tmp",
		},
	}

	// Create and populate index
	idx, err := index.OpenOrCreateIndex(cfg.IndexPath)
	if err != nil {
		t.Fatal(err)
	}

	// Index the test documents
	docs := []index.ZettelDoc{
		{
			ID: testID1, Title: "Test Zettel", Type: "note",
			Project: "zettelkasten-cli", Category: "untethered",
			Tags: []string{"test", "example"}, Body: "This is a valid zettel for testing.",
			FilePath: filepath.Join(untethered, testID1+".md"), Created: "2026-02-13T10:45:00Z",
		},
		{
			ID: testID2, Title: "Child Zettel", Type: "note",
			Project: "zettelkasten-cli", Category: "untethered",
			Tags: []string{"child"}, Body: "This zettel links to a parent.",
			FilePath: filepath.Join(untethered, testID2+".md"), Created: "2026-02-13T11:00:00Z",
		},
		{
			ID: testID3, Title: "Quick Idea", Type: "note",
			Category: "untethered",
			Tags: []string{"idea"}, Body: "A fleeting note captured without project context.",
			FilePath: filepath.Join(untethered, testID3+".md"), Created: "2026-02-13T12:00:00Z",
		},
		{
			ID: testID4, Title: "Fix the build", Type: "todo",
			Project: "zettelkasten-cli", Category: "untethered",
			Status: "open", Priority: "high", Due: "2026-02-15",
			Tags: []string{"todo", "bug"}, Body: "The CI pipeline is broken.",
			FilePath: filepath.Join(untethered, testID4+".md"), Created: "2026-02-13T13:00:00Z",
		},
	}
	for _, doc := range docs {
		if err := index.IndexDocument(idx, &doc); err != nil {
			t.Fatal(err)
		}
	}

	state := &ServerState{Config: cfg, Index: idx}

	cleanup := func() {
		idx.Close()
	}

	return state, cleanup
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// newTestClient creates an MCP server with the given state, connects an in-process client, and initializes it.
func newTestClient(t *testing.T, state *ServerState) *client.Client {
	t.Helper()

	s := server.NewMCPServer(
		"zettelkasten-test",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
	)
	registerTools(s, state)
	registerResources(s, state)

	c, err := client.NewInProcessClient(s)
	if err != nil {
		t.Fatalf("create in-process client: %v", err)
	}

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("start client: %v", err)
	}

	initReq := mcp.InitializeRequest{}
	initReq.Params.ClientInfo = mcp.Implementation{Name: "test", Version: "1.0.0"}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	if _, err := c.Initialize(ctx, initReq); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	t.Cleanup(func() { c.Close() })
	return c
}

func callTool(t *testing.T, c *client.Client, name string, args map[string]any) string {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args

	result, err := c.CallTool(context.Background(), req)
	if err != nil {
		t.Fatalf("call tool %s: %v", name, err)
	}
	if result.IsError {
		// Extract error text
		for _, c := range result.Content {
			if tc, ok := c.(mcp.TextContent); ok {
				t.Fatalf("tool %s returned error: %s", name, tc.Text)
			}
		}
		t.Fatalf("tool %s returned error", name)
	}

	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
	}
	t.Fatalf("tool %s returned no text content", name)
	return ""
}

// --- Tests ---

func TestListTools(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	result, err := c.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}

	expectedTools := map[string]bool{
		"zk_search":         false,
		"zk_read":           false,
		"zk_backlinks":      false,
		"zk_graph":          false,
		"zk_todos":          false,
		"zk_list_templates": false,
	}

	for _, tool := range result.Tools {
		if _, ok := expectedTools[tool.Name]; ok {
			expectedTools[tool.Name] = true
		}
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("expected tool %q not found", name)
		}
	}

	if len(result.Tools) != 6 {
		t.Errorf("expected 6 tools, got %d", len(result.Tools))
	}
}

func TestListResources(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	result, err := c.ListResources(context.Background(), mcp.ListResourcesRequest{})
	if err != nil {
		t.Fatalf("list resources: %v", err)
	}

	if len(result.Resources) != 2 {
		t.Errorf("expected 2 static resources, got %d", len(result.Resources))
	}

	tmplResult, err := c.ListResourceTemplates(context.Background(), mcp.ListResourceTemplatesRequest{})
	if err != nil {
		t.Fatalf("list resource templates: %v", err)
	}

	if len(tmplResult.ResourceTemplates) != 2 {
		t.Errorf("expected 2 resource templates, got %d", len(tmplResult.ResourceTemplates))
	}
}

func TestSearchTool(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	text := callTool(t, c, "zk_search", map[string]any{
		"query": "zettel",
	})

	var resp struct {
		Total   int                  `json:"total"`
		Results []index.SearchResult `json:"results"`
	}
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Total == 0 {
		t.Error("expected at least one search result")
	}
}

func TestSearchByProject(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	text := callTool(t, c, "zk_search", map[string]any{
		"project": "zettelkasten-cli",
	})

	var resp struct {
		Total   int                  `json:"total"`
		Results []index.SearchResult `json:"results"`
	}
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// ID1, ID2, ID4 have project=zettelkasten-cli
	if resp.Total != 3 {
		t.Errorf("expected 3 results for project filter, got %d", resp.Total)
	}
}

func TestReadTool(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	text := callTool(t, c, "zk_read", map[string]any{
		"id": testID1,
	})

	var resp struct {
		ID       string   `json:"id"`
		Title    string   `json:"title"`
		Type     string   `json:"type"`
		Project  string   `json:"project"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
		Body     string   `json:"body"`
		FilePath string   `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.ID != testID1 {
		t.Errorf("expected id %q, got %q", testID1, resp.ID)
	}
	if resp.Title != "Test Zettel" {
		t.Errorf("expected title %q, got %q", "Test Zettel", resp.Title)
	}
	if resp.Type != "note" {
		t.Errorf("expected type %q, got %q", "note", resp.Type)
	}
	if !strings.Contains(resp.Body, "valid zettel for testing") {
		t.Errorf("expected body to contain test content, got %q", resp.Body)
	}
}

func TestReadToolByPath(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	filePath := filepath.Join(state.Config.RootPath, "untethered", testID1+".md")
	text := callTool(t, c, "zk_read", map[string]any{
		"id": filePath,
	})

	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.ID != testID1 {
		t.Errorf("expected id %q, got %q", testID1, resp.ID)
	}
}

func TestBacklinksTool(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	// Note 2 links to Note 1 via [[testID1|Test Zettel]]
	text := callTool(t, c, "zk_backlinks", map[string]any{
		"id": testID1,
	})

	var resp struct {
		TargetID  string `json:"target_id"`
		Count     int    `json:"count"`
		Backlinks []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"backlinks"`
	}
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.TargetID != testID1 {
		t.Errorf("expected target_id %q, got %q", testID1, resp.TargetID)
	}
	if resp.Count != 1 {
		t.Errorf("expected 1 backlink, got %d", resp.Count)
	}
	if resp.Count > 0 && resp.Backlinks[0].ID != testID2 {
		t.Errorf("expected backlink from %q, got %q", testID2, resp.Backlinks[0].ID)
	}
}

func TestGraphTool(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	text := callTool(t, c, "zk_graph", map[string]any{
		"start_id": testID1,
		"limit":    10,
	})

	var resp struct {
		NodeCount int `json:"node_count"`
		EdgeCount int `json:"edge_count"`
		Nodes     []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"nodes"`
		Edges []struct {
			From  string `json:"from"`
			To    string `json:"to"`
			Label string `json:"label"`
		} `json:"edges"`
	}
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.NodeCount < 2 {
		t.Errorf("expected at least 2 nodes in graph, got %d", resp.NodeCount)
	}

	// Should have the start node
	foundStart := false
	for _, n := range resp.Nodes {
		if n.ID == testID1 {
			foundStart = true
		}
	}
	if !foundStart {
		t.Error("expected start node in graph results")
	}
}

func TestTodosTool(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	text := callTool(t, c, "zk_todos", map[string]any{
		"status": "open",
	})

	var resp struct {
		Total   int                  `json:"total"`
		Results []index.SearchResult `json:"results"`
	}
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.Total != 1 {
		t.Errorf("expected 1 open todo, got %d", resp.Total)
	}
	if resp.Total > 0 {
		if resp.Results[0].Status != "open" {
			t.Errorf("expected status 'open', got %q", resp.Results[0].Status)
		}
	}
}

func TestListTemplatesTool(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	text := callTool(t, c, "zk_list_templates", nil)

	var templates []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Category    string `json:"category"`
		Type        string `json:"type"`
	}
	if err := json.Unmarshal([]byte(text), &templates); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(templates) == 0 {
		t.Error("expected at least one template")
	}

	// Check a known template exists
	found := false
	for _, tmpl := range templates {
		if tmpl.Name == "meeting" {
			found = true
			if tmpl.Description == "" {
				t.Error("expected meeting template to have a description")
			}
		}
	}
	if !found {
		t.Error("expected 'meeting' template")
	}
}

func TestConfigResource(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	req := mcp.ReadResourceRequest{}
	req.Params.URI = "zk://config"

	result, err := c.ReadResource(context.Background(), req)
	if err != nil {
		t.Fatalf("read config resource: %v", err)
	}

	if len(result.Contents) == 0 {
		t.Fatal("expected config resource content")
	}

	text := result.Contents[0].(mcp.TextResourceContents).Text

	var cfg config.Config
	if err := json.Unmarshal([]byte(text), &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if cfg.RootPath != state.Config.RootPath {
		t.Errorf("expected root_path %q, got %q", state.Config.RootPath, cfg.RootPath)
	}
}

func TestStatsResource(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	req := mcp.ReadResourceRequest{}
	req.Params.URI = "zk://stats"

	result, err := c.ReadResource(context.Background(), req)
	if err != nil {
		t.Fatalf("read stats resource: %v", err)
	}

	if len(result.Contents) == 0 {
		t.Fatal("expected stats resource content")
	}

	text := result.Contents[0].(mcp.TextResourceContents).Text

	var stats struct {
		TotalNotes int            `json:"total_notes"`
		ByType     map[string]int `json:"by_type"`
		OpenTodos  int            `json:"open_todos"`
	}
	if err := json.Unmarshal([]byte(text), &stats); err != nil {
		t.Fatalf("unmarshal stats: %v", err)
	}

	if stats.TotalNotes != 4 {
		t.Errorf("expected 4 total notes, got %d", stats.TotalNotes)
	}
	if stats.OpenTodos != 1 {
		t.Errorf("expected 1 open todo, got %d", stats.OpenTodos)
	}
}

func TestZettelResource(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	req := mcp.ReadResourceRequest{}
	req.Params.URI = "zk://zettel/" + testID1

	result, err := c.ReadResource(context.Background(), req)
	if err != nil {
		t.Fatalf("read zettel resource: %v", err)
	}

	if len(result.Contents) == 0 {
		t.Fatal("expected zettel resource content")
	}

	text := result.Contents[0].(mcp.TextResourceContents).Text

	var zettel struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.Unmarshal([]byte(text), &zettel); err != nil {
		t.Fatalf("unmarshal zettel: %v", err)
	}

	if zettel.ID != testID1 {
		t.Errorf("expected id %q, got %q", testID1, zettel.ID)
	}
	if zettel.Title != "Test Zettel" {
		t.Errorf("expected title %q, got %q", "Test Zettel", zettel.Title)
	}
}

func TestProjectResource(t *testing.T) {
	state, cleanup := setupTestZettelkasten(t)
	defer cleanup()
	c := newTestClient(t, state)

	req := mcp.ReadResourceRequest{}
	req.Params.URI = "zk://project/zettelkasten-cli/notes"

	result, err := c.ReadResource(context.Background(), req)
	if err != nil {
		t.Fatalf("read project resource: %v", err)
	}

	if len(result.Contents) == 0 {
		t.Fatal("expected project resource content")
	}

	text := result.Contents[0].(mcp.TextResourceContents).Text

	var resp struct {
		Project string `json:"project"`
		Total   int    `json:"total"`
	}
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("unmarshal project: %v", err)
	}

	if resp.Project != "zettelkasten-cli" {
		t.Errorf("expected project %q, got %q", "zettelkasten-cli", resp.Project)
	}
	if resp.Total != 3 {
		t.Errorf("expected 3 notes for project, got %d", resp.Total)
	}
}
