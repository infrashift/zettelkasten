package graph

import (
	"strings"
	"testing"
)

func buildTestGraph() *Graph {
	// A -> B -> C -> D (linear chain via links)
	// A -> E (branch)
	// F is isolated
	g := New()
	g.AddNode(&Node{ID: "A", Title: "Node A", Category: "untethered", Links: []string{"B", "E"}})
	g.AddNode(&Node{ID: "B", Title: "Node B", Category: "untethered", Links: []string{"C"}})
	g.AddNode(&Node{ID: "C", Title: "Node C", Category: "tethered", Links: []string{"D"}})
	g.AddNode(&Node{ID: "D", Title: "Node D", Category: "tethered"})
	g.AddNode(&Node{ID: "E", Title: "Node E", Category: "untethered"})
	g.AddNode(&Node{ID: "F", Title: "Node F", Category: "untethered"})
	return g
}

func TestFindConnected_NoDepthLimit(t *testing.T) {
	g := buildTestGraph()
	nodes := g.FindConnected("A", 100, 0)
	if len(nodes) != 5 {
		t.Errorf("Expected 5 connected nodes from A (all except F), got %d", len(nodes))
	}
}

func TestFindConnected_DepthOne(t *testing.T) {
	g := buildTestGraph()
	// Depth 1 from A: A itself (depth 0), then B and E (depth 1)
	nodes := g.FindConnected("A", 100, 1)
	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes at depth 1 from A (A, B, E), got %d", len(nodes))
	}
	ids := nodeIDs(nodes)
	for _, expected := range []string{"A", "B", "E"} {
		if !ids[expected] {
			t.Errorf("Expected node %s in depth-1 results", expected)
		}
	}
}

func TestFindConnected_DepthTwo(t *testing.T) {
	g := buildTestGraph()
	// Depth 2 from A: A(0), B(1), E(1), C(2)
	nodes := g.FindConnected("A", 100, 2)
	if len(nodes) != 4 {
		t.Errorf("Expected 4 nodes at depth 2 from A, got %d", len(nodes))
	}
	ids := nodeIDs(nodes)
	if ids["D"] {
		t.Error("Node D should not be in depth-2 results (it's 3 hops from A)")
	}
}

func TestFindConnected_LimitOverridesDepth(t *testing.T) {
	g := buildTestGraph()
	// Depth unlimited but limit 2
	nodes := g.FindConnected("A", 2, 0)
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes with limit=2, got %d", len(nodes))
	}
}

func TestFindConnected_ReverseLinks(t *testing.T) {
	g := New()
	// B links to A, so starting from A should find B via reverse link
	g.AddNode(&Node{ID: "A", Title: "Target"})
	g.AddNode(&Node{ID: "B", Title: "Linker", Links: []string{"A"}})
	nodes := g.FindConnected("A", 100, 0)
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes (A + reverse-linked B), got %d", len(nodes))
	}
}

func TestFindConnected_ParentChild(t *testing.T) {
	g := New()
	g.AddNode(&Node{ID: "parent", Title: "Parent"})
	g.AddNode(&Node{ID: "child", Title: "Child", Parent: "parent"})
	g.BuildRelationships()
	nodes := g.FindConnected("parent", 100, 0)
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes (parent + child), got %d", len(nodes))
	}
}

func TestFindAllConnected_PicksMostConnected(t *testing.T) {
	g := buildTestGraph()
	// A has 2 links (most connections), so FindAllConnected should start from A
	nodes := g.FindAllConnected(100, 0)
	if len(nodes) != 5 {
		t.Errorf("Expected 5 connected nodes, got %d", len(nodes))
	}
	// F is isolated and should not be included
	ids := nodeIDs(nodes)
	if ids["F"] {
		t.Error("Isolated node F should not be in results")
	}
}

func TestExtractLinks(t *testing.T) {
	content := `Some text [[20260213143000-550e8400-e29b-41d4-a716-446655440000]] and
also [[20260214100000-6ba7b810-9dad-11d1-80b4-00c04fd430c8|My Title]] here.
Duplicate [[20260213143000-550e8400-e29b-41d4-a716-446655440000]] should be deduplicated.`

	links := ExtractLinks(content)
	if len(links) != 2 {
		t.Errorf("Expected 2 unique links, got %d: %v", len(links), links)
	}
}

func TestExtractLinks_NoLinks(t *testing.T) {
	links := ExtractLinks("No links in this content")
	if len(links) != 0 {
		t.Errorf("Expected 0 links, got %d", len(links))
	}
}

// --- GenerateASCIITree tests ---

func TestGenerateASCIITree_EmptyNodes(t *testing.T) {
	result := GenerateASCIITree(nil, nil, "")
	if result != "(no nodes found)" {
		t.Errorf("Expected '(no nodes found)', got %q", result)
	}
}

func TestGenerateASCIITree_SingleNode(t *testing.T) {
	nodes := []*Node{{ID: "A", Title: "Only Node"}}
	result := GenerateASCIITree(nodes, nil, "A")
	if !strings.Contains(result, "A Only Node") {
		t.Errorf("Expected root line 'A Only Node', got:\n%s", result)
	}
	// Single node, no tree connectors
	if strings.Contains(result, "├") || strings.Contains(result, "└") {
		t.Errorf("Single node should have no tree connectors, got:\n%s", result)
	}
}

func TestGenerateASCIITree_LinearChain(t *testing.T) {
	g := New()
	g.AddNode(&Node{ID: "A", Title: "Node A", Links: []string{"B"}})
	g.AddNode(&Node{ID: "B", Title: "Node B", Links: []string{"C"}})
	g.AddNode(&Node{ID: "C", Title: "Node C"})
	nodes := g.FindConnected("A", 100, 0)
	edges := g.GetEdges(nodes)

	result := GenerateASCIITree(nodes, edges, "A")

	// Root should be A
	if !strings.HasPrefix(result, "A Node A") {
		t.Errorf("Expected tree to start with 'A Node A', got:\n%s", result)
	}
	// Should contain B and C with tree connectors
	if !strings.Contains(result, "B Node B") {
		t.Errorf("Expected 'B Node B' in tree, got:\n%s", result)
	}
	if !strings.Contains(result, "C Node C") {
		t.Errorf("Expected 'C Node C' in tree, got:\n%s", result)
	}
	// Should have tree-drawing characters
	if !strings.Contains(result, "└── ") {
		t.Errorf("Expected tree-drawing characters, got:\n%s", result)
	}
	// Should have [link] labels
	if !strings.Contains(result, "[link]") {
		t.Errorf("Expected [link] label, got:\n%s", result)
	}
}

func TestGenerateASCIITree_WithReverseLinks(t *testing.T) {
	g := New()
	// B links TO A, so from A's perspective B is a reverse link
	g.AddNode(&Node{ID: "A", Title: "Target"})
	g.AddNode(&Node{ID: "B", Title: "Linker", Links: []string{"A"}})
	nodes := g.FindConnected("A", 100, 0)
	edges := g.GetEdges(nodes)

	result := GenerateASCIITree(nodes, edges, "A")

	// B should show ← marker for reverse link
	if !strings.Contains(result, "←") {
		t.Errorf("Expected ← marker for reverse link, got:\n%s", result)
	}
	if !strings.Contains(result, "B Linker") {
		t.Errorf("Expected 'B Linker' in tree, got:\n%s", result)
	}
}

func TestGenerateASCIITree_ParentChildEdges(t *testing.T) {
	g := New()
	g.AddNode(&Node{ID: "P", Title: "Parent"})
	g.AddNode(&Node{ID: "C", Title: "Child", Parent: "P"})
	g.BuildRelationships()
	nodes := g.FindConnected("P", 100, 0)
	edges := g.GetEdges(nodes)

	result := GenerateASCIITree(nodes, edges, "P")

	if !strings.HasPrefix(result, "P Parent") {
		t.Errorf("Expected tree to start with 'P Parent', got:\n%s", result)
	}
	// Child should appear with [parent] label (reverse of C->P parent edge)
	if !strings.Contains(result, "C Child") {
		t.Errorf("Expected 'C Child' in tree, got:\n%s", result)
	}
	if !strings.Contains(result, "[parent]") {
		t.Errorf("Expected [parent] label, got:\n%s", result)
	}
}

func TestGenerateASCIITree_WithStartNode(t *testing.T) {
	g := buildTestGraph()
	nodes := g.FindConnected("B", 100, 0)
	edges := g.GetEdges(nodes)

	result := GenerateASCIITree(nodes, edges, "B")

	// Should start from B, not A
	if !strings.HasPrefix(result, "B Node B") {
		t.Errorf("Expected tree to start with 'B Node B', got:\n%s", result)
	}
}

func TestGenerateASCIITree_CycleHandling(t *testing.T) {
	g := New()
	// A -> B -> C -> A (cycle)
	g.AddNode(&Node{ID: "A", Title: "Node A", Links: []string{"B"}})
	g.AddNode(&Node{ID: "B", Title: "Node B", Links: []string{"C"}})
	g.AddNode(&Node{ID: "C", Title: "Node C", Links: []string{"A"}})
	nodes := g.FindConnected("A", 100, 0)
	edges := g.GetEdges(nodes)

	result := GenerateASCIITree(nodes, edges, "A")

	// Should not infinite-loop; each node appears exactly once
	if strings.Count(result, "Node A") != 1 {
		t.Errorf("Node A should appear exactly once, got:\n%s", result)
	}
	if strings.Count(result, "Node B") != 1 {
		t.Errorf("Node B should appear exactly once, got:\n%s", result)
	}
	if strings.Count(result, "Node C") != 1 {
		t.Errorf("Node C should appear exactly once, got:\n%s", result)
	}
}

func TestGenerateASCIITree_DisconnectedNodes(t *testing.T) {
	g := buildTestGraph()
	// FindConnected from A won't include F, but we can add all nodes
	allNodes := g.AllNodes()
	edges := g.GetEdges(allNodes)

	result := GenerateASCIITree(allNodes, edges, "A")

	// F is disconnected - should appear as separate root at bottom
	if !strings.Contains(result, "F Node F") {
		t.Errorf("Expected disconnected node 'F Node F', got:\n%s", result)
	}
}

func nodeIDs(nodes []*Node) map[string]bool {
	ids := make(map[string]bool)
	for _, n := range nodes {
		ids[n.ID] = true
	}
	return ids
}
