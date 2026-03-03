package main

import (
	"strings"
	"testing"
)

func TestGenerateBasicNote_Untethered(t *testing.T) {
	note := generateBasicNote("20260201100000-abc", "My Idea", "", "untethered", nil)

	if !strings.Contains(note, `category: "untethered"`) {
		t.Error("Expected category: untethered")
	}
	if !strings.Contains(note, `type: "note"`) {
		t.Error("Expected type: note")
	}
	if !strings.Contains(note, `title: "My Idea"`) {
		t.Error("Expected title")
	}
	if strings.Contains(note, "project:") {
		t.Error("Untethered note without project should not have project field")
	}
	if !strings.Contains(note, "tags: []") {
		t.Error("Expected empty tags as tags: []")
	}
	if !strings.Contains(note, "# My Idea") {
		t.Error("Expected markdown heading")
	}
}

func TestGenerateBasicNote_Tethered(t *testing.T) {
	note := generateBasicNote("20260201100000-abc", "Project Note", "backend", "tethered", nil)

	if !strings.Contains(note, `category: "tethered"`) {
		t.Error("Expected category: tethered")
	}
	if !strings.Contains(note, `project: "backend"`) {
		t.Error("Expected project: backend")
	}
}

func TestGenerateBasicNote_WithTags(t *testing.T) {
	note := generateBasicNote("20260201100000-abc", "Tagged Note", "", "untethered", []string{"golang", "api"})

	if !strings.Contains(note, `- "golang"`) {
		t.Error("Expected golang tag")
	}
	if !strings.Contains(note, `- "api"`) {
		t.Error("Expected api tag")
	}
	if strings.Contains(note, "tags: []") {
		t.Error("Should not have empty tags when tags provided")
	}
}

func TestGenerateBasicNote_EmptyTags(t *testing.T) {
	note := generateBasicNote("20260201100000-abc", "No Tags", "", "untethered", []string{})

	if !strings.Contains(note, "tags: []") {
		t.Error("Expected tags: [] for empty tag slice")
	}
}

func TestGenerateBasicNote_NilTags(t *testing.T) {
	note := generateBasicNote("20260201100000-abc", "Nil Tags", "", "untethered", nil)

	if !strings.Contains(note, "tags: []") {
		t.Error("Expected tags: [] for nil tag slice")
	}
}

func TestGenerateBasicNote_HasFrontmatter(t *testing.T) {
	note := generateBasicNote("20260201100000-abc", "Test", "", "untethered", nil)

	if !strings.HasPrefix(note, "---\n") {
		t.Error("Expected frontmatter to start with ---")
	}
	// Count --- delimiters (should be exactly 2)
	count := strings.Count(note, "---\n")
	if count != 2 {
		t.Errorf("Expected 2 frontmatter delimiters, got %d", count)
	}
}

func TestAddFrontmatterTags_AppendsToExisting(t *testing.T) {
	content := []byte(`---
id: "20260201100000-abc"
title: "Test Note"
tags:
  - "golang"
created: "2026-02-01T10:00:00Z"
---

# Test Note
`)
	result, err := addFrontmatterTags(content, []string{"api", "security"})
	if err != nil {
		t.Fatalf("addFrontmatterTags failed: %v", err)
	}

	s := string(result)
	if !strings.Contains(s, `- "golang"`) {
		t.Error("Expected existing golang tag to be preserved")
	}
	if !strings.Contains(s, `- "api"`) {
		t.Error("Expected api tag to be added")
	}
	if !strings.Contains(s, `- "security"`) {
		t.Error("Expected security tag to be added")
	}
	if !strings.Contains(s, "# Test Note") {
		t.Error("Expected body content to be preserved")
	}
}

func TestAddFrontmatterTags_EmptyInlineArray(t *testing.T) {
	content := []byte(`---
id: "20260201100000-abc"
title: "Test Note"
tags: []
created: "2026-02-01T10:00:00Z"
---

# Test Note
`)
	result, err := addFrontmatterTags(content, []string{"newtag"})
	if err != nil {
		t.Fatalf("addFrontmatterTags failed: %v", err)
	}

	s := string(result)
	if !strings.Contains(s, `- "newtag"`) {
		t.Error("Expected newtag to be added")
	}
	if strings.Contains(s, "tags: []") {
		t.Error("Should not keep inline empty array when adding tags")
	}
	// Verify the YAML is valid by checking tags: followed by list items
	if !strings.Contains(s, "tags:\n  - \"newtag\"") {
		t.Errorf("Expected tags:\\n  - \"newtag\" but got:\n%s", s)
	}
	// Verify created field is preserved and not broken
	if !strings.Contains(s, "created:") {
		t.Error("Expected created field to be preserved")
	}
}

func TestAddFrontmatterTags_NoTagsField(t *testing.T) {
	content := []byte(`---
id: "20260201100000-abc"
title: "Test Note"
created: "2026-02-01T10:00:00Z"
---

# Test Note
`)
	result, err := addFrontmatterTags(content, []string{"first-tag"})
	if err != nil {
		t.Fatalf("addFrontmatterTags failed: %v", err)
	}

	s := string(result)
	if !strings.Contains(s, "tags:") {
		t.Error("Expected tags: field to be added")
	}
	if !strings.Contains(s, `- "first-tag"`) {
		t.Error("Expected first-tag to be added")
	}
}

func TestAddFrontmatterTags_NoFrontmatter(t *testing.T) {
	content := []byte("# Just a heading\n\nSome content.\n")
	_, err := addFrontmatterTags(content, []string{"tag"})
	if err == nil {
		t.Error("Expected error for content without frontmatter")
	}
}

func TestUpdateFrontmatter_UpdatesExistingField(t *testing.T) {
	content := []byte(`---
id: "20260201100000-abc"
title: "Old Title"
category: "untethered"
---

# Old Title
`)
	result, err := updateFrontmatter(content, map[string]string{
		"category": "tethered",
	})
	if err != nil {
		t.Fatalf("updateFrontmatter failed: %v", err)
	}

	s := string(result)
	if !strings.Contains(s, `category: "tethered"`) {
		t.Error("Expected category to be updated to tethered")
	}
	if strings.Contains(s, `category: "untethered"`) {
		t.Error("Old category value should be replaced")
	}
}

func TestUpdateFrontmatter_AddsNewField(t *testing.T) {
	content := []byte(`---
id: "20260201100000-abc"
title: "Test"
---

# Test
`)
	result, err := updateFrontmatter(content, map[string]string{
		"project": "my-proj",
	})
	if err != nil {
		t.Fatalf("updateFrontmatter failed: %v", err)
	}

	s := string(result)
	if !strings.Contains(s, `project: "my-proj"`) {
		t.Error("Expected project field to be added")
	}
}

func TestUpdateFrontmatterRemove_RemovesField(t *testing.T) {
	content := []byte(`---
id: "20260201100000-abc"
title: "Test"
completed: "2026-02-01"
status: "closed"
---

# Test
`)
	result, err := updateFrontmatterRemove(content, map[string]string{
		"status": "open",
	}, []string{"completed"})
	if err != nil {
		t.Fatalf("updateFrontmatterRemove failed: %v", err)
	}

	s := string(result)
	if !strings.Contains(s, `status: "open"`) {
		t.Error("Expected status to be updated to open")
	}
	if strings.Contains(s, "completed") {
		t.Error("Expected completed field to be removed")
	}
}
