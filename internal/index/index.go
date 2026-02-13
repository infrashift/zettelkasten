package index

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

// ZettelDoc is the structure we feed into the index
type ZettelDoc struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Project  string   `json:"project"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Body     string   `json:"body"`
}

func CreateMapping() mapping.IndexMapping {
	indexMapping := bleve.NewIndexMapping()

	// 1. Unstructured Full-Text (Markdown Body)
	bodyMapping := bleve.NewTextFieldMapping()
	bodyMapping.Analyzer = "en" // Uses English stemming (e.g., "running" matches "run")
	indexMapping.DefaultMapping.AddFieldMappingsAt("body", bodyMapping)

	// 2. Structured Fields (Project, Category, Tags)
	// We use the "keyword" analyzer so they aren't tokenized/broken up
	keywordMapping := bleve.NewTextFieldMapping()
	keywordMapping.Analyzer = "keyword"

	indexMapping.DefaultMapping.AddFieldMappingsAt("project", keywordMapping)
	indexMapping.DefaultMapping.AddFieldMappingsAt("category", keywordMapping)
	indexMapping.DefaultMapping.AddFieldMappingsAt("tags", keywordMapping)

	return indexMapping
}

func OpenOrCreateIndex(path string) (bleve.Index, error) {
	idx, err := bleve.Open(path)
	if err == bleve.ErrIndexClosed || err == bleve.ErrIndexNotFound {
		return bleve.New(path, CreateMapping())
	}
	return idx, err
}
