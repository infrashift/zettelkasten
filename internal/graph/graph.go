package graph

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Node represents a zettel in the graph.
type Node struct {
	ID       string
	Title    string
	FilePath string
	Project  string
	Category string
	Parent   string   // Optional parent zettel ID
	Children []string // Child zettel IDs
	Links    []string // Links found in body content via [[id]] or [[id|title]] scanning (not from frontmatter)
}

// Graph represents the zettelkasten link graph.
type Graph struct {
	nodes map[string]*Node
}

// New creates a new empty graph.
func New() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
	}
}

// AddNode adds a node to the graph.
func (g *Graph) AddNode(node *Node) {
	g.nodes[node.ID] = node
}

// GetNode returns a node by ID.
func (g *Graph) GetNode(id string) *Node {
	return g.nodes[id]
}

// NodeCount returns the number of nodes in the graph.
func (g *Graph) NodeCount() int {
	return len(g.nodes)
}

// AllNodes returns all nodes in the graph.
func (g *Graph) AllNodes() []*Node {
	nodes := make([]*Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		nodes = append(nodes, n)
	}
	return nodes
}

// BuildRelationships processes parent fields and builds bidirectional relationships.
func (g *Graph) BuildRelationships() {
	for _, node := range g.nodes {
		// Add this node as a child of its parent
		if node.Parent != "" {
			if parent, ok := g.nodes[node.Parent]; ok {
				parent.Children = append(parent.Children, node.ID)
			}
		}
	}
}

// bfsItem tracks a node ID and its distance from the start node.
type bfsItem struct {
	id    string
	depth int
}

// FindConnected finds all nodes connected to the given node ID up to the limit.
// Uses BFS to find the closest related nodes first.
// If maxDepth > 0, limits traversal to that many hops from the start node.
func (g *Graph) FindConnected(startID string, limit int, maxDepth int) []*Node {
	if limit <= 0 {
		limit = 10
	}

	startNode := g.nodes[startID]
	if startNode == nil {
		return nil
	}

	visited := make(map[string]bool)
	result := make([]*Node, 0, limit)
	queue := []bfsItem{{id: startID, depth: 0}}

	for len(queue) > 0 && len(result) < limit {
		current := queue[0]
		queue = queue[1:]

		if visited[current.id] {
			continue
		}
		visited[current.id] = true

		node := g.nodes[current.id]
		if node == nil {
			continue
		}

		result = append(result, node)

		// Stop expanding if we've reached max depth
		if maxDepth > 0 && current.depth >= maxDepth {
			continue
		}

		nextDepth := current.depth + 1

		// Add connected nodes to queue
		// Parent
		if node.Parent != "" && !visited[node.Parent] {
			queue = append(queue, bfsItem{id: node.Parent, depth: nextDepth})
		}
		// Children
		for _, childID := range node.Children {
			if !visited[childID] {
				queue = append(queue, bfsItem{id: childID, depth: nextDepth})
			}
		}
		// Explicit links
		for _, linkID := range node.Links {
			if !visited[linkID] {
				queue = append(queue, bfsItem{id: linkID, depth: nextDepth})
			}
		}
		// Reverse links (notes that link TO this node)
		for _, other := range g.nodes {
			if visited[other.ID] {
				continue
			}
			for _, linkID := range other.Links {
				if linkID == node.ID {
					queue = append(queue, bfsItem{id: other.ID, depth: nextDepth})
					break
				}
			}
		}
	}

	return result
}

// FindAllConnected returns all connected components starting from any node.
func (g *Graph) FindAllConnected(limit int, maxDepth int) []*Node {
	if limit <= 0 {
		limit = 10
	}

	if len(g.nodes) == 0 {
		return nil
	}

	// Start from the node with the most connections for a meaningful graph
	var startID string
	maxConns := -1
	for id, node := range g.nodes {
		conns := len(node.Links) + len(node.Children)
		if node.Parent != "" {
			conns++
		}
		if conns > maxConns {
			maxConns = conns
			startID = id
		}
	}

	return g.FindConnected(startID, limit, maxDepth)
}

// Edge represents a directed edge in the graph.
type Edge struct {
	From  string
	To    string
	Label string // "parent", "child", "link"
}

// GetEdges returns all edges for the given nodes.
func (g *Graph) GetEdges(nodes []*Node) []Edge {
	nodeSet := make(map[string]bool)
	for _, n := range nodes {
		nodeSet[n.ID] = true
	}

	var edges []Edge
	seen := make(map[string]bool)

	for _, node := range nodes {
		// Parent edge (child -> parent)
		if node.Parent != "" && nodeSet[node.Parent] {
			key := node.ID + "->" + node.Parent
			if !seen[key] {
				edges = append(edges, Edge{From: node.ID, To: node.Parent, Label: "parent"})
				seen[key] = true
			}
		}

		// Link edges
		for _, linkID := range node.Links {
			if nodeSet[linkID] {
				key := node.ID + "->" + linkID
				if !seen[key] {
					edges = append(edges, Edge{From: node.ID, To: linkID, Label: "link"})
					seen[key] = true
				}
			}
		}
	}

	return edges
}

// GenerateASCIITree renders nodes and edges as a Unicode box-drawing tree.
//
// The root is startNodeID (if provided and present), otherwise the most-connected
// node. Children are sorted by ID for stable output. Reverse links (incoming
// edges from other nodes) are marked with "←". Disconnected nodes appear as
// separate single-line roots at the bottom.
func GenerateASCIITree(nodes []*Node, edges []Edge, startNodeID string) string {
	if len(nodes) == 0 {
		return "(no nodes found)"
	}

	// Build lookup map
	nodeByID := make(map[string]*Node)
	for _, n := range nodes {
		nodeByID[n.ID] = n
	}

	// Build adjacency: outgoing children per node, with edge label + direction
	type child struct {
		id      string
		label   string
		reverse bool // true when the other node links TO this one
	}
	outgoing := make(map[string][]child) // parent → children (forward edges)
	incoming := make(map[string][]child) // target → sources  (reverse edges)

	for _, e := range edges {
		outgoing[e.From] = append(outgoing[e.From], child{id: e.To, label: e.Label})
		incoming[e.To] = append(incoming[e.To], child{id: e.From, label: e.Label, reverse: true})
	}

	// Pick root
	rootID := startNodeID
	if rootID == "" || nodeByID[rootID] == nil {
		// Most-connected node
		maxConns := -1
		for _, n := range nodes {
			conns := len(outgoing[n.ID]) + len(incoming[n.ID])
			if conns > maxConns || (conns == maxConns && n.ID < rootID) {
				maxConns = conns
				rootID = n.ID
			}
		}
	}

	// BFS spanning tree from root
	type treeChild struct {
		id      string
		label   string
		reverse bool
	}
	treeChildren := make(map[string][]treeChild)
	visited := make(map[string]bool)

	queue := []string{rootID}
	visited[rootID] = true

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		// Collect all neighbours: forward edges + reverse edges
		var neighbours []treeChild
		for _, c := range outgoing[cur] {
			if !visited[c.id] {
				neighbours = append(neighbours, treeChild{id: c.id, label: c.label})
			}
		}
		for _, c := range incoming[cur] {
			if !visited[c.id] {
				neighbours = append(neighbours, treeChild{id: c.id, label: c.label, reverse: true})
			}
		}

		// Sort by ID for deterministic output
		sort.Slice(neighbours, func(i, j int) bool {
			return neighbours[i].id < neighbours[j].id
		})

		for _, nb := range neighbours {
			if !visited[nb.id] {
				visited[nb.id] = true
				treeChildren[cur] = append(treeChildren[cur], nb)
				queue = append(queue, nb.id)
			}
		}
	}

	// Render tree
	var sb strings.Builder
	nodeTitle := func(id string) string {
		if n := nodeByID[id]; n != nil && n.Title != "" {
			return n.Title
		}
		return id
	}

	// Root line (no label)
	sb.WriteString(fmt.Sprintf("%s %s\n", rootID, nodeTitle(rootID)))

	// Recursive render
	var render func(parentID string, prefix string)
	render = func(parentID string, prefix string) {
		children := treeChildren[parentID]
		for i, ch := range children {
			isLast := i == len(children)-1
			connector := "├── "
			childPrefix := "│   "
			if isLast {
				connector = "└── "
				childPrefix = "    "
			}

			label := fmt.Sprintf(" [%s]", ch.label)
			marker := ""
			if ch.reverse {
				marker = " ←"
			}

			sb.WriteString(fmt.Sprintf("%s%s%s %s%s%s\n",
				prefix, connector, ch.id, nodeTitle(ch.id), marker, label))

			render(ch.id, prefix+childPrefix)
		}
	}
	render(rootID, "")

	// Disconnected nodes as separate roots
	for _, n := range nodes {
		if !visited[n.ID] {
			sb.WriteString(fmt.Sprintf("%s %s\n", n.ID, nodeTitle(n.ID)))
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

// ExtractLinks finds all [[id]] or [[id|title]] style links in content.
var linkPattern = regexp.MustCompile(`\[\[([0-9]{14}-[0-9a-f-]{36})(?:\|[^\]]+)?\]\]`)

func ExtractLinks(content string) []string {
	matches := linkPattern.FindAllStringSubmatch(content, -1)
	var links []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) >= 2 {
			id := match[1]
			if !seen[id] {
				links = append(links, id)
				seen[id] = true
			}
		}
	}

	return links
}
