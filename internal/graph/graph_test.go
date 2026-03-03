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

	// Should not infinite-loop; each node appears as a tree entry exactly once.
	// Use "ID Title" pattern to avoid matching cross-link annotations like "(also → Node C)".
	if strings.Count(result, "A Node A") != 1 {
		t.Errorf("Node A should appear as tree entry once, got:\n%s", result)
	}
	if strings.Count(result, "B Node B") != 1 {
		t.Errorf("Node B should appear as tree entry once, got:\n%s", result)
	}
	if strings.Count(result, "C Node C") != 1 {
		t.Errorf("Node C should appear as tree entry once, got:\n%s", result)
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

func TestGenerateASCIITree_ForwardLinkMarker(t *testing.T) {
	g := New()
	// note2 links to both note1 and note3
	g.AddNode(&Node{ID: "note1", Title: "Note 1"})
	g.AddNode(&Node{ID: "note2", Title: "Note 2", Links: []string{"note1", "note3"}})
	g.AddNode(&Node{ID: "note3", Title: "Note 3"})
	nodes := g.FindConnected("note1", 100, 0)
	edges := g.GetEdges(nodes)

	result := GenerateASCIITree(nodes, edges, "note1")

	// note2 links TO note1 → reverse edge → should show ←
	if !strings.Contains(result, "note2 Note 2 ← [link]") {
		t.Errorf("Expected note2 with ← marker (reverse link to note1), got:\n%s", result)
	}
	// note2 links TO note3 → forward edge from tree-parent note2 → should show →
	if !strings.Contains(result, "note3 Note 3 → [link]") {
		t.Errorf("Expected note3 with → marker (forward link from note2), got:\n%s", result)
	}
}

func TestGenerateASCIITree_CrossLinks(t *testing.T) {
	g := New()
	// note2 links to note1 and note4
	// note3 links to note1 and note2
	// note5 links to note2
	// Replicates the user scenario where note3's link to note2 is a cross-edge
	g.AddNode(&Node{ID: "note1", Title: "Note 1"})
	g.AddNode(&Node{ID: "note2", Title: "Note 2", Links: []string{"note1", "note4"}})
	g.AddNode(&Node{ID: "note3", Title: "Note 3", Links: []string{"note1", "note2"}})
	g.AddNode(&Node{ID: "note4", Title: "Note 4"})
	g.AddNode(&Node{ID: "note5", Title: "Note 5", Links: []string{"note2"}})
	nodes := g.FindConnected("note1", 100, 0)
	edges := g.GetEdges(nodes)

	result := GenerateASCIITree(nodes, edges, "note1")

	// note3 links to note2, but note3 is placed under note1 in the tree.
	// The note3→note2 edge is a cross-link and should appear as annotation.
	if !strings.Contains(result, "(also → Note 2)") {
		t.Errorf("Expected cross-link annotation '(also → Note 2)' on note3, got:\n%s", result)
	}
}

// --- GenerateMermaid tests ---

func TestGenerateMermaid_Empty(t *testing.T) {
	result := GenerateMermaid(nil, nil)
	if result != "graph LR;" {
		t.Errorf("Expected 'graph LR;', got %q", result)
	}
}

func TestGenerateMermaid_SingleNode(t *testing.T) {
	nodes := []*Node{{ID: "A", Title: "Only Node"}}
	result := GenerateMermaid(nodes, nil)
	expected := "graph LR;\n  A[\"Only Node\"];"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGenerateMermaid_LinearChain(t *testing.T) {
	g := New()
	g.AddNode(&Node{ID: "A", Title: "Node A", Links: []string{"B"}})
	g.AddNode(&Node{ID: "B", Title: "Node B", Links: []string{"C"}})
	g.AddNode(&Node{ID: "C", Title: "Node C"})
	nodes := g.FindConnected("A", 100, 0)
	edges := g.GetEdges(nodes)

	result := GenerateMermaid(nodes, edges)

	if !strings.Contains(result, "graph LR;") {
		t.Errorf("Expected 'graph LR;' header, got:\n%s", result)
	}
	// Node declarations with titles
	if !strings.Contains(result, `A["Node A"];`) {
		t.Errorf("Expected node declaration 'A[\"Node A\"];', got:\n%s", result)
	}
	if !strings.Contains(result, `B["Node B"];`) {
		t.Errorf("Expected node declaration 'B[\"Node B\"];', got:\n%s", result)
	}
	if !strings.Contains(result, `C["Node C"];`) {
		t.Errorf("Expected node declaration 'C[\"Node C\"];', got:\n%s", result)
	}
	// Edge lines
	if !strings.Contains(result, "A --> B;") {
		t.Errorf("Expected 'A --> B;' edge, got:\n%s", result)
	}
	if !strings.Contains(result, "B --> C;") {
		t.Errorf("Expected 'B --> C;' edge, got:\n%s", result)
	}
}

func TestGenerateMermaid_AllEdgesPresent(t *testing.T) {
	// Cross-links that the ASCII tree lost are now visible
	g := New()
	g.AddNode(&Node{ID: "note1", Title: "Note 1"})
	g.AddNode(&Node{ID: "note2", Title: "Note 2", Links: []string{"note1", "note4"}})
	g.AddNode(&Node{ID: "note3", Title: "Note 3", Links: []string{"note1", "note2"}})
	g.AddNode(&Node{ID: "note4", Title: "Note 4"})
	nodes := g.FindConnected("note1", 100, 0)
	edges := g.GetEdges(nodes)

	result := GenerateMermaid(nodes, edges)

	// All node declarations with titles
	if !strings.Contains(result, `note1["Note 1"];`) {
		t.Errorf("Expected node declaration for note1, got:\n%s", result)
	}
	if !strings.Contains(result, `note4["Note 4"];`) {
		t.Errorf("Expected node declaration for note4, got:\n%s", result)
	}

	// All 4 edges must be present — no cross-link loss
	expectedEdges := []string{
		"note2 --> note1;",
		"note2 --> note4;",
		"note3 --> note1;",
		"note3 --> note2;",
	}
	for _, expected := range expectedEdges {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected edge %q in output, got:\n%s", expected, result)
		}
	}
}

func TestGenerateMermaid_DisconnectedNodes(t *testing.T) {
	g := buildTestGraph()
	allNodes := g.AllNodes()
	edges := g.GetEdges(allNodes)

	result := GenerateMermaid(allNodes, edges)

	// F has no edges but should still appear as a declared node with title
	if !strings.Contains(result, `F["Node F"];`) {
		t.Errorf("Expected node declaration 'F[\"Node F\"];' in output, got:\n%s", result)
	}
}

func TestGenerateMermaid_DeterministicOutput(t *testing.T) {
	g := New()
	g.AddNode(&Node{ID: "C", Title: "Node C", Links: []string{"A"}})
	g.AddNode(&Node{ID: "A", Title: "Node A", Links: []string{"B"}})
	g.AddNode(&Node{ID: "B", Title: "Node B"})
	nodes := g.FindConnected("A", 100, 0)
	edges := g.GetEdges(nodes)

	result1 := GenerateMermaid(nodes, edges)
	result2 := GenerateMermaid(nodes, edges)

	if result1 != result2 {
		t.Errorf("Output should be deterministic.\nFirst:  %q\nSecond: %q", result1, result2)
	}
}

// --- TransformLinks tests ---

func TestTransformLinks_BareLink_InSet(t *testing.T) {
	nodeMap := map[string]*Node{
		"20260213143000-550e8400-e29b-41d4-a716-446655440000": {
			ID:    "20260213143000-550e8400-e29b-41d4-a716-446655440000",
			Title: "My Note",
		},
	}
	content := "See [[20260213143000-550e8400-e29b-41d4-a716-446655440000]] for details."
	result := TransformLinks(content, nodeMap)
	expected := "See [My Note](20260213143000-550e8400-e29b-41d4-a716-446655440000.md) for details."
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestTransformLinks_DisplayLink_InSet(t *testing.T) {
	nodeMap := map[string]*Node{
		"20260213143000-550e8400-e29b-41d4-a716-446655440000": {
			ID:    "20260213143000-550e8400-e29b-41d4-a716-446655440000",
			Title: "My Note",
		},
	}
	content := "See [[20260213143000-550e8400-e29b-41d4-a716-446655440000|Custom Display]] for details."
	result := TransformLinks(content, nodeMap)
	expected := "See [Custom Display](20260213143000-550e8400-e29b-41d4-a716-446655440000.md) for details."
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestTransformLinks_NotInSet(t *testing.T) {
	nodeMap := map[string]*Node{}
	content := "See [[20260213143000-550e8400-e29b-41d4-a716-446655440000]] for details."
	result := TransformLinks(content, nodeMap)
	if result != content {
		t.Errorf("Expected content unchanged, got %q", result)
	}
}

func TestTransformLinks_Mixed(t *testing.T) {
	nodeMap := map[string]*Node{
		"20260213143000-550e8400-e29b-41d4-a716-446655440000": {
			ID:    "20260213143000-550e8400-e29b-41d4-a716-446655440000",
			Title: "Note A",
		},
	}
	content := "Link to [[20260213143000-550e8400-e29b-41d4-a716-446655440000]] and [[20260214100000-6ba7b810-9dad-11d1-80b4-00c04fd430c8|External]]."
	result := TransformLinks(content, nodeMap)
	expected := "Link to [Note A](20260213143000-550e8400-e29b-41d4-a716-446655440000.md) and [[20260214100000-6ba7b810-9dad-11d1-80b4-00c04fd430c8|External]]."
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestTransformLinks_NoTitle(t *testing.T) {
	id := "20260213143000-550e8400-e29b-41d4-a716-446655440000"
	nodeMap := map[string]*Node{
		id: {ID: id, Title: ""},
	}
	content := "See [[20260213143000-550e8400-e29b-41d4-a716-446655440000]] here."
	result := TransformLinks(content, nodeMap)
	expected := "See [" + id + "](" + id + ".md) here."
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestTransformLinks_NoLinks(t *testing.T) {
	nodeMap := map[string]*Node{
		"20260213143000-550e8400-e29b-41d4-a716-446655440000": {
			ID:    "20260213143000-550e8400-e29b-41d4-a716-446655440000",
			Title: "Note A",
		},
	}
	content := "No links in this content at all."
	result := TransformLinks(content, nodeMap)
	if result != content {
		t.Errorf("Expected content unchanged, got %q", result)
	}
}

func nodeIDs(nodes []*Node) map[string]bool {
	ids := make(map[string]bool)
	for _, n := range nodes {
		ids[n.ID] = true
	}
	return ids
}
