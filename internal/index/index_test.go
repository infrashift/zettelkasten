package index

import (
	"os"
	"testing"
)

// setupTestIndex creates a temporary Bleve index with test documents
func setupTestIndex(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "zk-index-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	indexPath := tmpDir + "/test.bleve"
	idx, err := OpenOrCreateIndex(indexPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create index: %v", err)
	}

	docs := []ZettelDoc{
		{
			ID:       "20260201100000-aaa",
			Title:    "Authentication Design",
			Type:     "note",
			Project:  "backend",
			Category: "tethered",
			Tags:     []string{"auth", "design"},
			Body:     "OAuth2 implementation with JWT tokens",
			FilePath: "/notes/auth-design.md",
			Created:  "2026-02-01T10:00:00Z",
		},
		{
			ID:       "20260202100000-bbb",
			Title:    "Fix login bug",
			Type:     "todo",
			Project:  "backend",
			Category: "untethered",
			Tags:     []string{"bug", "auth"},
			Body:     "Users getting logged out after 10 minutes",
			FilePath: "/notes/fix-login.md",
			Created:  "2026-02-02T10:00:00Z",
			Status:   "open",
			Priority: "high",
			Due:      "2026-02-15",
		},
		{
			ID:       "20260203100000-ccc",
			Title:    "Update API docs",
			Type:     "todo",
			Project:  "docs",
			Category: "untethered",
			Tags:     []string{"docs"},
			Body:     "API documentation is outdated",
			FilePath: "/notes/update-docs.md",
			Created:  "2026-02-03T10:00:00Z",
			Status:   "closed",
			Priority: "medium",
			Due:      "2026-02-10",
			Completed: "2026-02-09",
		},
		{
			ID:       "20260204100000-ddd",
			Title:    "Daily 2026-02-04",
			Type:     "daily-note",
			Project:  "",
			Category: "untethered",
			Tags:     []string{"daily"},
			Body:     "Morning standup notes and planning",
			FilePath: "/notes/daily-2026-02-04.md",
			Created:  "2026-02-04T08:00:00Z",
		},
		{
			ID:       "20260205100000-eee",
			Title:    "Security audit findings",
			Type:     "issue",
			Project:  "backend",
			Category: "tethered",
			Tags:     []string{"security", "audit"},
			Body:     "Found XSS vulnerability in search endpoint",
			FilePath: "/notes/security-audit.md",
			Created:  "2026-02-05T10:00:00Z",
			Status:   "open",
			Priority: "high",
			Due:      "2026-02-20",
		},
		{
			ID:       "20260206100000-fff",
			Title:    "Refactor database layer",
			Type:     "todo",
			Project:  "backend",
			Category: "tethered",
			Tags:     []string{"refactor", "database"},
			Body:     "Move to repository pattern",
			FilePath: "/notes/refactor-db.md",
			Created:  "2026-02-06T10:00:00Z",
			Status:   "open",
			Priority: "low",
			Due:      "2026-03-01",
		},
	}

	for _, doc := range docs {
		d := doc
		if err := IndexDocument(idx, &d); err != nil {
			idx.Close()
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to index document %s: %v", doc.ID, err)
		}
	}

	idx.Close()

	return indexPath, func() {
		os.RemoveAll(tmpDir)
	}
}

func TestSearchByType(t *testing.T) {
	indexPath, cleanup := setupTestIndex(t)
	defer cleanup()

	idx, err := OpenOrCreateIndex(indexPath)
	if err != nil {
		t.Fatalf("Failed to open index: %v", err)
	}
	defer idx.Close()

	tests := []struct {
		name     string
		typeVal  string
		wantMin  int
		wantIDs  []string
		dontWant []string
	}{
		{
			name:    "filter by type=todo",
			typeVal: "todo",
			wantMin: 3,
			wantIDs: []string{"20260202100000-bbb", "20260203100000-ccc", "20260206100000-fff"},
		},
		{
			name:    "filter by type=note",
			typeVal: "note",
			wantMin: 1,
			wantIDs: []string{"20260201100000-aaa"},
		},
		{
			name:    "filter by type=daily-note",
			typeVal: "daily-note",
			wantMin: 1,
			wantIDs: []string{"20260204100000-ddd"},
		},
		{
			name:    "filter by type=issue",
			typeVal: "issue",
			wantMin: 1,
			wantIDs: []string{"20260205100000-eee"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, _, err := Search(idx, SearchOptions{
				Type:  tt.typeVal,
				Limit: 50,
			})
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			if len(results) < tt.wantMin {
				t.Errorf("Expected at least %d results, got %d", tt.wantMin, len(results))
			}
			resultIDs := map[string]bool{}
			for _, r := range results {
				resultIDs[r.ID] = true
			}
			for _, id := range tt.wantIDs {
				if !resultIDs[id] {
					t.Errorf("Expected result ID %s not found", id)
				}
			}
		})
	}
}

func TestSearchByStatus(t *testing.T) {
	indexPath, cleanup := setupTestIndex(t)
	defer cleanup()

	idx, err := OpenOrCreateIndex(indexPath)
	if err != nil {
		t.Fatalf("Failed to open index: %v", err)
	}
	defer idx.Close()

	// Search for open todos
	results, _, err := Search(idx, SearchOptions{
		Status: "open",
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	for _, r := range results {
		if r.Status != "open" {
			t.Errorf("Expected status=open, got %q for %s", r.Status, r.ID)
		}
	}

	if len(results) < 2 {
		t.Errorf("Expected at least 2 open results, got %d", len(results))
	}

	// Search for closed todos
	results, _, err = Search(idx, SearchOptions{
		Status: "closed",
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 closed result, got %d", len(results))
	}
	if len(results) > 0 && results[0].ID != "20260203100000-ccc" {
		t.Errorf("Expected closed result ID 20260203100000-ccc, got %s", results[0].ID)
	}
}

func TestSearchByPriority(t *testing.T) {
	indexPath, cleanup := setupTestIndex(t)
	defer cleanup()

	idx, err := OpenOrCreateIndex(indexPath)
	if err != nil {
		t.Fatalf("Failed to open index: %v", err)
	}
	defer idx.Close()

	results, _, err := Search(idx, SearchOptions{
		Priority: "high",
		Limit:    50,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 high-priority results, got %d", len(results))
	}

	for _, r := range results {
		if r.Priority != "high" {
			t.Errorf("Expected priority=high, got %q for %s", r.Priority, r.ID)
		}
	}
}

func TestSearchByDueDateRange(t *testing.T) {
	indexPath, cleanup := setupTestIndex(t)
	defer cleanup()

	idx, err := OpenOrCreateIndex(indexPath)
	if err != nil {
		t.Fatalf("Failed to open index: %v", err)
	}
	defer idx.Close()

	// Due before 2026-02-16 should get login bug (2026-02-15) and docs (2026-02-10)
	results, _, err := Search(idx, SearchOptions{
		DueBefore: "2026-02-16",
		Limit:     50,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	resultIDs := map[string]bool{}
	for _, r := range results {
		resultIDs[r.ID] = true
	}

	if !resultIDs["20260202100000-bbb"] {
		t.Error("Expected login bug (due 2026-02-15) in results")
	}
	if !resultIDs["20260203100000-ccc"] {
		t.Error("Expected docs todo (due 2026-02-10) in results")
	}

	// Due after 2026-02-16 should get security audit (2026-02-20) and refactor (2026-03-01)
	results, _, err = Search(idx, SearchOptions{
		DueAfter: "2026-02-16",
		Limit:    50,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	resultIDs = map[string]bool{}
	for _, r := range results {
		resultIDs[r.ID] = true
	}

	if !resultIDs["20260205100000-eee"] {
		t.Error("Expected security audit (due 2026-02-20) in results")
	}
	if !resultIDs["20260206100000-fff"] {
		t.Error("Expected refactor (due 2026-03-01) in results")
	}

	// Range: due between 2026-02-14 and 2026-02-21
	results, _, err = Search(idx, SearchOptions{
		DueAfter:  "2026-02-14",
		DueBefore: "2026-02-21",
		Limit:     50,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	resultIDs = map[string]bool{}
	for _, r := range results {
		resultIDs[r.ID] = true
	}

	if !resultIDs["20260202100000-bbb"] {
		t.Error("Expected login bug (due 2026-02-15) in range results")
	}
	if !resultIDs["20260205100000-eee"] {
		t.Error("Expected security audit (due 2026-02-20) in range results")
	}
	if resultIDs["20260203100000-ccc"] {
		t.Error("Did not expect docs todo (due 2026-02-10) in range results")
	}
	if resultIDs["20260206100000-fff"] {
		t.Error("Did not expect refactor (due 2026-03-01) in range results")
	}
}

func TestSearchCombinedFilters(t *testing.T) {
	indexPath, cleanup := setupTestIndex(t)
	defer cleanup()

	idx, err := OpenOrCreateIndex(indexPath)
	if err != nil {
		t.Fatalf("Failed to open index: %v", err)
	}
	defer idx.Close()

	// type=todo + status=open + project=backend
	results, _, err := Search(idx, SearchOptions{
		Type:    "todo",
		Status:  "open",
		Project: "backend",
		Limit:   50,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	resultIDs := map[string]bool{}
	for _, r := range results {
		resultIDs[r.ID] = true
	}

	if !resultIDs["20260202100000-bbb"] {
		t.Error("Expected login bug in results")
	}
	if !resultIDs["20260206100000-fff"] {
		t.Error("Expected refactor in results")
	}
	// closed doc todo should NOT appear
	if resultIDs["20260203100000-ccc"] {
		t.Error("Did not expect closed docs todo in results")
	}
	// issue should NOT appear (type=issue, not todo)
	if resultIDs["20260205100000-eee"] {
		t.Error("Did not expect issue in todo results")
	}

	// type=todo + priority=high + status=open
	results, _, err = Search(idx, SearchOptions{
		Type:     "todo",
		Priority: "high",
		Status:   "open",
		Limit:    50,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result (open high-priority todo), got %d", len(results))
	}
	if len(results) > 0 && results[0].ID != "20260202100000-bbb" {
		t.Errorf("Expected login bug, got %s", results[0].ID)
	}
}

func TestSearchFullTextWithFilters(t *testing.T) {
	indexPath, cleanup := setupTestIndex(t)
	defer cleanup()

	idx, err := OpenOrCreateIndex(indexPath)
	if err != nil {
		t.Fatalf("Failed to open index: %v", err)
	}
	defer idx.Close()

	// Full-text "auth" + type=note
	results, _, err := Search(idx, SearchOptions{
		Query: "authentication",
		Type:  "note",
		Limit: 50,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least 1 result for 'authentication' + type=note")
	}
	if len(results) > 0 && results[0].ID != "20260201100000-aaa" {
		t.Errorf("Expected auth design note, got %s", results[0].ID)
	}

	// Full-text "auth" + type=todo should not return the note
	results, _, err = Search(idx, SearchOptions{
		Query: "authentication",
		Type:  "todo",
		Limit: 50,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	for _, r := range results {
		if r.ID == "20260201100000-aaa" {
			t.Error("Did not expect note in todo-filtered results")
		}
	}
}

func TestSearchMatchAll(t *testing.T) {
	indexPath, cleanup := setupTestIndex(t)
	defer cleanup()

	idx, err := OpenOrCreateIndex(indexPath)
	if err != nil {
		t.Fatalf("Failed to open index: %v", err)
	}
	defer idx.Close()

	// No filters should return all documents
	results, total, err := Search(idx, SearchOptions{
		Limit: 50,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if total != 6 {
		t.Errorf("Expected 6 total results, got %d", total)
	}
	if len(results) != 6 {
		t.Errorf("Expected 6 results, got %d", len(results))
	}
}
