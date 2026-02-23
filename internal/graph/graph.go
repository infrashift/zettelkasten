package graph

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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

// FindConnected finds all nodes connected to the given node ID up to the limit.
// Uses BFS to find the closest related nodes first.
func (g *Graph) FindConnected(startID string, limit int) []*Node {
	if limit <= 0 {
		limit = 10
	}

	startNode := g.nodes[startID]
	if startNode == nil {
		return nil
	}

	visited := make(map[string]bool)
	result := make([]*Node, 0, limit)
	queue := []string{startID}

	for len(queue) > 0 && len(result) < limit {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		node := g.nodes[current]
		if node == nil {
			continue
		}

		result = append(result, node)

		// Add connected nodes to queue
		// Parent
		if node.Parent != "" && !visited[node.Parent] {
			queue = append(queue, node.Parent)
		}
		// Children
		for _, childID := range node.Children {
			if !visited[childID] {
				queue = append(queue, childID)
			}
		}
		// Explicit links
		for _, linkID := range node.Links {
			if !visited[linkID] {
				queue = append(queue, linkID)
			}
		}
	}

	return result
}

// FindAllConnected returns all connected components starting from any node.
func (g *Graph) FindAllConnected(limit int) []*Node {
	if limit <= 0 {
		limit = 10
	}

	if len(g.nodes) == 0 {
		return nil
	}

	// Start from the first node we find
	var startID string
	for id := range g.nodes {
		startID = id
		break
	}

	return g.FindConnected(startID, limit)
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

// GenerateMermaid generates a Mermaid flowchart diagram for the given nodes.
func GenerateMermaid(nodes []*Node, edges []Edge) string {
	if len(nodes) == 0 {
		return "```mermaid\nflowchart TD\n    empty[No nodes found]\n```"
	}

	var sb strings.Builder
	sb.WriteString("```mermaid\nflowchart TD\n")

	// Define nodes with titles
	for _, node := range nodes {
		title := node.Title
		if title == "" {
			title = node.ID
		}
		// Escape quotes and special characters for Mermaid
		title = strings.ReplaceAll(title, "\"", "'")
		title = strings.ReplaceAll(title, "\n", " ")

		// Use different shapes for categories
		switch node.Category {
		case "tethered":
			sb.WriteString(fmt.Sprintf("    %s[[\"%s\"]]\n", node.ID, title))
		case "untethered":
			sb.WriteString(fmt.Sprintf("    %s(\"%s\")\n", node.ID, title))
		default:
			sb.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", node.ID, title))
		}
	}

	sb.WriteString("\n")

	// Define edges
	for _, edge := range edges {
		switch edge.Label {
		case "parent":
			sb.WriteString(fmt.Sprintf("    %s -->|parent| %s\n", edge.From, edge.To))
		case "link":
			sb.WriteString(fmt.Sprintf("    %s -.->|link| %s\n", edge.From, edge.To))
		default:
			sb.WriteString(fmt.Sprintf("    %s --> %s\n", edge.From, edge.To))
		}
	}

	// Add styling
	sb.WriteString("\n")
	sb.WriteString("    classDef tethered fill:#90EE90,stroke:#228B22\n")
	sb.WriteString("    classDef untethered fill:#FFE4B5,stroke:#FFA500\n")

	// Apply classes
	var tetheredIDs, untetheredIDs []string
	for _, node := range nodes {
		switch node.Category {
		case "tethered":
			tetheredIDs = append(tetheredIDs, node.ID)
		case "untethered":
			untetheredIDs = append(untetheredIDs, node.ID)
		}
	}

	if len(tetheredIDs) > 0 {
		sb.WriteString(fmt.Sprintf("    class %s tethered\n", strings.Join(tetheredIDs, ",")))
	}
	if len(untetheredIDs) > 0 {
		sb.WriteString(fmt.Sprintf("    class %s untethered\n", strings.Join(untetheredIDs, ",")))
	}

	sb.WriteString("```")
	return sb.String()
}

// GenerateMarkdown generates a complete markdown file with the graph visualization.
func GenerateMarkdown(nodes []*Node, edges []Edge, rootPath string, startNodeID string) string {
	var sb strings.Builder

	// Header
	sb.WriteString("# Zettelkasten Graph\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	if startNodeID != "" {
		sb.WriteString(fmt.Sprintf("Starting from: `%s`\n\n", startNodeID))
	}

	sb.WriteString(fmt.Sprintf("Showing %d nodes\n\n", len(nodes)))

	// Legend
	sb.WriteString("## Legend\n\n")
	sb.WriteString("- **Rounded rectangle** `(title)`: Untethered note\n")
	sb.WriteString("- **Stadium shape** `[[title]]`: Tethered note\n")
	sb.WriteString("- **Solid arrow** `-->`: Parent relationship\n")
	sb.WriteString("- **Dashed arrow** `-.->`: Link reference\n\n")

	// Mermaid diagram
	sb.WriteString("## Graph\n\n")
	sb.WriteString(GenerateMermaid(nodes, edges))
	sb.WriteString("\n\n")

	// Node list with links
	sb.WriteString("## Nodes\n\n")
	sb.WriteString("| ID | Title | Category | Project | File |\n")
	sb.WriteString("|-----|-------|----------|---------|------|\n")

	for _, node := range nodes {
		title := node.Title
		if title == "" {
			title = "(untitled)"
		}

		filePath := node.FilePath
		if rootPath != "" && filePath != "" {
			// Make path relative to root for cleaner display
			if rel, err := filepath.Rel(rootPath, filePath); err == nil {
				filePath = rel
			}
		}

		fileLink := ""
		if filePath != "" {
			fileLink = fmt.Sprintf("[%s](%s)", filepath.Base(filePath), filePath)
		}

		project := node.Project
		if project == "" {
			project = "-"
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			node.ID, title, node.Category, project, fileLink))
	}

	sb.WriteString("\n")

	// Relationships section
	if len(edges) > 0 {
		sb.WriteString("## Relationships\n\n")
		for _, edge := range edges {
			fromNode := findNodeByID(nodes, edge.From)
			toNode := findNodeByID(nodes, edge.To)

			fromTitle := edge.From
			toTitle := edge.To
			if fromNode != nil && fromNode.Title != "" {
				fromTitle = fromNode.Title
			}
			if toNode != nil && toNode.Title != "" {
				toTitle = toNode.Title
			}

			sb.WriteString(fmt.Sprintf("- **%s** ─(%s)→ **%s**\n", fromTitle, edge.Label, toTitle))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func findNodeByID(nodes []*Node, id string) *Node {
	for _, n := range nodes {
		if n.ID == id {
			return n
		}
	}
	return nil
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
