package index

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search/query"
)

// ZettelDoc is the structure we feed into the index
type ZettelDoc struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Type     string   `json:"type"` // "note", "todo", "daily-note", or "issue"
	Project  string   `json:"project"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Body     string   `json:"body"`
	FilePath string   `json:"file_path"`
	Created  string   `json:"created"`

	// Todo-specific fields
	Status    string `json:"status,omitempty"`    // "open", "in_progress", "closed"
	Due       string `json:"due,omitempty"`       // YYYY-MM-DD
	Completed string `json:"completed,omitempty"` // YYYY-MM-DD
	Priority  string `json:"priority,omitempty"`  // "high", "medium", "low"
}

// SearchResult represents a single search hit
type SearchResult struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Type     string   `json:"type"`
	Project  string   `json:"project"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	FilePath string   `json:"file_path"`
	Score    float64  `json:"score"`
	Snippet  string   `json:"snippet,omitempty"`

	// Todo-specific fields
	Status    string `json:"status,omitempty"`
	Due       string `json:"due,omitempty"`
	Completed string `json:"completed,omitempty"`
	Priority  string `json:"priority,omitempty"`
}

// SearchOptions configures search behavior
type SearchOptions struct {
	Query    string   // Full-text query
	Project  string   // Filter by project
	Category string   // Filter by category (untethered/tethered)
	Type     string   // Filter by type (note/todo)
	Tags     []string // Filter by tags (AND)
	Limit    int      // Max results (default 20)
	Offset   int      // Pagination offset

	// Todo-specific filters
	Status       string // Filter by status (open/in_progress/closed)
	Priority     string // Filter by priority (high/medium/low)
	DueBefore    string // Filter todos due before this date (YYYY-MM-DD)
	DueAfter     string // Filter todos due after this date (YYYY-MM-DD)
	CompletedAfter  string // Filter todos completed after this date
	CompletedBefore string // Filter todos completed before this date
}

func CreateMapping() mapping.IndexMapping {
	indexMapping := bleve.NewIndexMapping()

	// Full-text fields with English stemming
	textMapping := bleve.NewTextFieldMapping()
	textMapping.Analyzer = "en"
	indexMapping.DefaultMapping.AddFieldMappingsAt("body", textMapping)
	indexMapping.DefaultMapping.AddFieldMappingsAt("title", textMapping)

	// Keyword fields (exact match, no tokenization)
	keywordMapping := bleve.NewTextFieldMapping()
	keywordMapping.Analyzer = "keyword"
	indexMapping.DefaultMapping.AddFieldMappingsAt("id", keywordMapping)
	indexMapping.DefaultMapping.AddFieldMappingsAt("type", keywordMapping)
	indexMapping.DefaultMapping.AddFieldMappingsAt("project", keywordMapping)
	indexMapping.DefaultMapping.AddFieldMappingsAt("category", keywordMapping)
	indexMapping.DefaultMapping.AddFieldMappingsAt("tags", keywordMapping)
	indexMapping.DefaultMapping.AddFieldMappingsAt("file_path", keywordMapping)
	indexMapping.DefaultMapping.AddFieldMappingsAt("created", keywordMapping)

	// Todo-specific keyword fields
	indexMapping.DefaultMapping.AddFieldMappingsAt("status", keywordMapping)
	indexMapping.DefaultMapping.AddFieldMappingsAt("priority", keywordMapping)
	indexMapping.DefaultMapping.AddFieldMappingsAt("due", keywordMapping)
	indexMapping.DefaultMapping.AddFieldMappingsAt("completed", keywordMapping)

	return indexMapping
}

func OpenOrCreateIndex(path string) (bleve.Index, error) {
	idx, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		return bleve.New(path, CreateMapping())
	}
	return idx, err
}

// IndexDocument adds or updates a document in the index
func IndexDocument(idx bleve.Index, doc *ZettelDoc) error {
	return idx.Index(doc.ID, doc)
}

// DeleteDocument removes a document from the index
func DeleteDocument(idx bleve.Index, id string) error {
	return idx.Delete(id)
}

// Search performs a search with the given options
func Search(idx bleve.Index, opts SearchOptions) ([]SearchResult, int, error) {
	if opts.Limit <= 0 {
		opts.Limit = 20
	}

	// Build the query
	var queries []query.Query

	// Full-text query on body and title
	if opts.Query != "" {
		// Search across title and body
		titleQuery := bleve.NewMatchQuery(opts.Query)
		titleQuery.SetField("title")
		titleQuery.SetBoost(2.0) // Boost title matches

		bodyQuery := bleve.NewMatchQuery(opts.Query)
		bodyQuery.SetField("body")

		textQuery := bleve.NewDisjunctionQuery(titleQuery, bodyQuery)
		queries = append(queries, textQuery)
	}

	// Project filter
	if opts.Project != "" {
		projectQuery := bleve.NewTermQuery(opts.Project)
		projectQuery.SetField("project")
		queries = append(queries, projectQuery)
	}

	// Category filter
	if opts.Category != "" {
		categoryQuery := bleve.NewTermQuery(opts.Category)
		categoryQuery.SetField("category")
		queries = append(queries, categoryQuery)
	}

	// Tags filter (AND - must have all tags)
	for _, tag := range opts.Tags {
		tagQuery := bleve.NewTermQuery(tag)
		tagQuery.SetField("tags")
		queries = append(queries, tagQuery)
	}

	// Type filter
	if opts.Type != "" {
		typeQuery := bleve.NewTermQuery(opts.Type)
		typeQuery.SetField("type")
		queries = append(queries, typeQuery)
	}

	// Status filter (for todos)
	if opts.Status != "" {
		statusQuery := bleve.NewTermQuery(opts.Status)
		statusQuery.SetField("status")
		queries = append(queries, statusQuery)
	}

	// Priority filter (for todos)
	if opts.Priority != "" {
		priorityQuery := bleve.NewTermQuery(opts.Priority)
		priorityQuery.SetField("priority")
		queries = append(queries, priorityQuery)
	}

	// Due date range filters
	if opts.DueAfter != "" || opts.DueBefore != "" {
		min := opts.DueAfter
		max := opts.DueBefore
		if min == "" {
			min = "0000-00-00"
		}
		if max == "" {
			max = "9999-99-99"
		}
		dueQuery := bleve.NewTermRangeInclusiveQuery(min, max, &[]bool{true}[0], &[]bool{true}[0])
		dueQuery.SetField("due")
		queries = append(queries, dueQuery)
	}

	// Completed date range filters
	if opts.CompletedAfter != "" || opts.CompletedBefore != "" {
		min := opts.CompletedAfter
		max := opts.CompletedBefore
		if min == "" {
			min = "0000-00-00"
		}
		if max == "" {
			max = "9999-99-99"
		}
		completedQuery := bleve.NewTermRangeInclusiveQuery(min, max, &[]bool{true}[0], &[]bool{true}[0])
		completedQuery.SetField("completed")
		queries = append(queries, completedQuery)
	}

	// Combine all queries with AND
	var finalQuery query.Query
	if len(queries) == 0 {
		finalQuery = bleve.NewMatchAllQuery()
	} else if len(queries) == 1 {
		finalQuery = queries[0]
	} else {
		finalQuery = bleve.NewConjunctionQuery(queries...)
	}

	// Create search request
	searchRequest := bleve.NewSearchRequest(finalQuery)
	searchRequest.Size = opts.Limit
	searchRequest.From = opts.Offset
	searchRequest.Fields = []string{"id", "title", "type", "project", "category", "tags", "file_path", "status", "due", "completed", "priority"}

	// Add highlighting for body matches
	if opts.Query != "" {
		searchRequest.Highlight = bleve.NewHighlight()
	}

	// Execute search
	searchResult, err := idx.Search(searchRequest)
	if err != nil {
		return nil, 0, fmt.Errorf("search failed: %w", err)
	}

	// Convert results
	results := make([]SearchResult, 0, len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		result := SearchResult{
			ID:    hit.ID,
			Score: hit.Score,
		}

		// Extract fields
		if title, ok := hit.Fields["title"].(string); ok {
			result.Title = title
		}
		if typ, ok := hit.Fields["type"].(string); ok {
			result.Type = typ
		}
		if project, ok := hit.Fields["project"].(string); ok {
			result.Project = project
		}
		if category, ok := hit.Fields["category"].(string); ok {
			result.Category = category
		}
		if filePath, ok := hit.Fields["file_path"].(string); ok {
			result.FilePath = filePath
		}
		if tags, ok := hit.Fields["tags"].([]interface{}); ok {
			for _, t := range tags {
				if s, ok := t.(string); ok {
					result.Tags = append(result.Tags, s)
				}
			}
		}
		// Extract todo-specific fields
		if status, ok := hit.Fields["status"].(string); ok {
			result.Status = status
		}
		if due, ok := hit.Fields["due"].(string); ok {
			result.Due = due
		}
		if completed, ok := hit.Fields["completed"].(string); ok {
			result.Completed = completed
		}
		if priority, ok := hit.Fields["priority"].(string); ok {
			result.Priority = priority
		}

		// Extract highlight snippet
		if len(hit.Fragments) > 0 {
			for _, fragments := range hit.Fragments {
				if len(fragments) > 0 {
					result.Snippet = fragments[0]
					break
				}
			}
		}

		results = append(results, result)
	}

	return results, int(searchResult.Total), nil
}

// SearchResultsToJSON converts search results to JSON string
func SearchResultsToJSON(results []SearchResult) (string, error) {
	data, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatSearchResults formats results for terminal display
func FormatSearchResults(results []SearchResult, total int) string {
	if len(results) == 0 {
		return "No results found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d results:\n\n", total))

	for i, r := range results {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, r.Title))
		sb.WriteString(fmt.Sprintf("   ID: %s | Category: %s", r.ID, r.Category))
		if r.Project != "" {
			sb.WriteString(fmt.Sprintf(" | Project: %s", r.Project))
		}
		sb.WriteString("\n")
		if len(r.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("   Tags: %s\n", strings.Join(r.Tags, ", ")))
		}
		if r.FilePath != "" {
			sb.WriteString(fmt.Sprintf("   File: %s\n", r.FilePath))
		}
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("   ...%s...\n", r.Snippet))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatTodoResults formats todo results for terminal display
func FormatTodoResults(results []SearchResult, total int) string {
	if len(results) == 0 {
		return "No todos found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d todos:\n\n", total))

	for i, r := range results {
		// Status indicator
		statusIcon := "[ ]"
		switch r.Status {
		case "in_progress":
			statusIcon = "[~]"
		case "closed":
			statusIcon = "[x]"
		}

		sb.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, statusIcon, r.Title))
		sb.WriteString(fmt.Sprintf("   ID: %s | Status: %s", r.ID, r.Status))
		if r.Priority != "" {
			sb.WriteString(fmt.Sprintf(" | Priority: %s", r.Priority))
		}
		if r.Project != "" {
			sb.WriteString(fmt.Sprintf(" | Project: %s", r.Project))
		}
		sb.WriteString("\n")
		if r.Due != "" {
			sb.WriteString(fmt.Sprintf("   Due: %s\n", r.Due))
		}
		if r.Completed != "" {
			sb.WriteString(fmt.Sprintf("   Completed: %s\n", r.Completed))
		}
		if r.FilePath != "" {
			sb.WriteString(fmt.Sprintf("   File: %s\n", r.FilePath))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
