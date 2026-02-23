package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSchema(t *testing.T) {
	schema, err := LoadSchema()
	if err != nil {
		t.Fatalf("LoadSchema() error = %v", err)
	}

	if schema.Err() != nil {
		t.Fatalf("schema has error: %v", schema.Err())
	}
}

func TestLoadZettelSchema(t *testing.T) {
	schema, err := LoadZettelSchema()
	if err != nil {
		t.Fatalf("LoadZettelSchema() error = %v", err)
	}

	if schema.Err() != nil {
		t.Fatalf("schema has error: %v", schema.Err())
	}
}

func TestLoadDefaultConfig(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check defaults
	home, _ := os.UserHomeDir()
	expectedRoot := filepath.Join(home, "zettelkasten")
	if cfg.RootPath != expectedRoot {
		t.Errorf("RootPath = %q, want %q", cfg.RootPath, expectedRoot)
	}

	if cfg.IndexPath != ".zk_index" {
		t.Errorf("IndexPath = %q, want %q", cfg.IndexPath, ".zk_index")
	}

	if cfg.Editor != "nvim" {
		t.Errorf("Editor = %q, want %q", cfg.Editor, "nvim")
	}

	if cfg.Folders.Untethered != "untethered" {
		t.Errorf("Folders.Untethered = %q, want %q", cfg.Folders.Untethered, "untethered")
	}

	if cfg.Folders.Tethered != "tethered" {
		t.Errorf("Folders.Tethered = %q, want %q", cfg.Folders.Tethered, "tethered")
	}

	if cfg.Folders.Tmp != "tmp" {
		t.Errorf("Folders.Tmp = %q, want %q", cfg.Folders.Tmp, "tmp")
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.cue")

	configContent := `
root_path: "/custom/path"
editor: "vim"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.RootPath != "/custom/path" {
		t.Errorf("RootPath = %q, want %q", cfg.RootPath, "/custom/path")
	}

	if cfg.Editor != "vim" {
		t.Errorf("Editor = %q, want %q", cfg.Editor, "vim")
	}

	// Defaults should still apply for unset values
	if cfg.IndexPath != ".zk_index" {
		t.Errorf("IndexPath = %q, want %q", cfg.IndexPath, ".zk_index")
	}
}

func TestValidateZettelYAML_Valid(t *testing.T) {
	validYAML := `
id: "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
title: "Test Zettel"
project: "test-project"
category: "untethered"
tags:
  - "test"
created: "2026-02-13T10:45:00Z"
`
	if err := ValidateZettelYAML([]byte(validYAML)); err != nil {
		t.Errorf("ValidateZettelYAML() error = %v", err)
	}
}

func TestValidateZettelYAML_ValidWithParent(t *testing.T) {
	validYAML := `
id: "20260213110000-ffffffff-1111-2222-3333-444444444444"
title: "Child Zettel"
project: "test-project"
category: "tethered"
tags:
  - "child"
created: "2026-02-13T11:00:00Z"
parent: "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
`
	if err := ValidateZettelYAML([]byte(validYAML)); err != nil {
		t.Errorf("ValidateZettelYAML() error = %v", err)
	}
}

func TestValidateZettelYAML_InvalidID(t *testing.T) {
	invalidYAML := `
id: "12345"
title: "Test"
project: "test"
category: "untethered"
tags: []
created: "2026-02-13T10:45:00Z"
`
	if err := ValidateZettelYAML([]byte(invalidYAML)); err == nil {
		t.Error("ValidateZettelYAML() expected error for invalid ID, got nil")
	}
}

func TestValidateZettelYAML_EmptyTitle(t *testing.T) {
	invalidYAML := `
id: "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
title: ""
project: "test"
category: "untethered"
tags: []
created: "2026-02-13T10:45:00Z"
`
	if err := ValidateZettelYAML([]byte(invalidYAML)); err == nil {
		t.Error("ValidateZettelYAML() expected error for empty title, got nil")
	}
}

func TestValidateZettelYAML_UntetheredWithoutProject(t *testing.T) {
	// Untethered notes can omit project entirely
	validYAML := `
id: "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
title: "Quick Idea"
category: "untethered"
tags: []
created: "2026-02-13T10:45:00Z"
`
	if err := ValidateZettelYAML([]byte(validYAML)); err != nil {
		t.Errorf("ValidateZettelYAML() untethered without project should be valid, got error = %v", err)
	}
}

func TestValidateZettelYAML_TetheredRequiresProject(t *testing.T) {
	// Tethered notes must have a non-empty project
	invalidYAML := `
id: "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
title: "Test"
category: "tethered"
tags: []
created: "2026-02-13T10:45:00Z"
`
	if err := ValidateZettelYAML([]byte(invalidYAML)); err == nil {
		t.Error("ValidateZettelYAML() expected error for tethered note without project, got nil")
	}
}

func TestValidateZettelYAML_TetheredEmptyProject(t *testing.T) {
	// Tethered notes cannot have empty project
	invalidYAML := `
id: "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
title: "Test"
project: ""
category: "tethered"
tags: []
created: "2026-02-13T10:45:00Z"
`
	if err := ValidateZettelYAML([]byte(invalidYAML)); err == nil {
		t.Error("ValidateZettelYAML() expected error for tethered note with empty project, got nil")
	}
}

func TestValidateZettelYAML_InvalidCategory(t *testing.T) {
	invalidYAML := `
id: "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
title: "Test"
project: "test"
category: "draft"
tags: []
created: "2026-02-13T10:45:00Z"
`
	if err := ValidateZettelYAML([]byte(invalidYAML)); err == nil {
		t.Error("ValidateZettelYAML() expected error for invalid category, got nil")
	}
}

func TestValidateZettelYAML_MissingCreated(t *testing.T) {
	invalidYAML := `
id: "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
title: "Test"
project: "test"
category: "untethered"
tags: []
`
	if err := ValidateZettelYAML([]byte(invalidYAML)); err == nil {
		t.Error("ValidateZettelYAML() expected error for missing created, got nil")
	}
}

func TestExtractFrontmatter(t *testing.T) {
	content := []byte(`---
id: "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
title: "Test"
---

# Content here
`)
	frontmatter, err := ExtractFrontmatter(content)
	if err != nil {
		t.Fatalf("ExtractFrontmatter() error = %v", err)
	}

	expected := `id: "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
title: "Test"`
	if string(frontmatter) != expected {
		t.Errorf("ExtractFrontmatter() = %q, want %q", string(frontmatter), expected)
	}
}

func TestExtractFrontmatter_NoOpeningDelimiter(t *testing.T) {
	content := []byte(`id: "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
title: "Test"
---
`)
	_, err := ExtractFrontmatter(content)
	if err == nil {
		t.Error("ExtractFrontmatter() expected error for missing opening delimiter")
	}
}

func TestExtractFrontmatter_NoClosingDelimiter(t *testing.T) {
	content := []byte(`---
id: "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
title: "Test"
`)
	_, err := ExtractFrontmatter(content)
	if err == nil {
		t.Error("ExtractFrontmatter() expected error for missing closing delimiter")
	}
}

func TestParseFrontmatter(t *testing.T) {
	content := []byte(`---
id: "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
title: "Test Zettel"
project: "test-project"
category: "untethered"
tags:
  - "test"
created: "2026-02-13T10:45:00Z"
---

# Content
`)
	z, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("ParseFrontmatter() error = %v", err)
	}

	if z.ID != "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Errorf("ID = %q, want %q", z.ID, "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	}

	if z.Title != "Test Zettel" {
		t.Errorf("Title = %q, want %q", z.Title, "Test Zettel")
	}

	if z.Project != "test-project" {
		t.Errorf("Project = %q, want %q", z.Project, "test-project")
	}

	if z.Category != "untethered" {
		t.Errorf("Category = %q, want %q", z.Category, "untethered")
	}

	if len(z.Tags) != 1 || z.Tags[0] != "test" {
		t.Errorf("Tags = %v, want [test]", z.Tags)
	}
}

func TestParseAndValidate_ValidFile(t *testing.T) {
	// Read the valid test file
	content, err := os.ReadFile("../../testdata/valid_zettel.md")
	if err != nil {
		t.Skipf("testdata file not found: %v", err)
	}

	z, err := ParseAndValidate(content)
	if err != nil {
		t.Fatalf("ParseAndValidate() error = %v", err)
	}

	if z.ID != "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Errorf("ID = %q, want %q", z.ID, "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	}
}

func TestParseAndValidate_ValidFileWithParent(t *testing.T) {
	content, err := os.ReadFile("../../testdata/valid_zettel_with_parent.md")
	if err != nil {
		t.Skipf("testdata file not found: %v", err)
	}

	z, err := ParseAndValidate(content)
	if err != nil {
		t.Fatalf("ParseAndValidate() error = %v", err)
	}

	if z.Parent != "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Errorf("Parent = %q, want %q", z.Parent, "20260213104500-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	}
}

func TestParseAndValidate_InvalidFile(t *testing.T) {
	content, err := os.ReadFile("../../testdata/invalid_zettel.md")
	if err != nil {
		t.Skipf("testdata file not found: %v", err)
	}

	_, err = ParseAndValidate(content)
	if err == nil {
		t.Error("ParseAndValidate() expected error for invalid file, got nil")
	}
}

func TestParseAndValidate_UntetheredWithoutProject(t *testing.T) {
	content, err := os.ReadFile("../../testdata/valid_fleeting_no_project.md")
	if err != nil {
		t.Skipf("testdata file not found: %v", err)
	}

	z, err := ParseAndValidate(content)
	if err != nil {
		t.Fatalf("ParseAndValidate() error = %v", err)
	}

	if z.Category != "untethered" {
		t.Errorf("Category = %q, want %q", z.Category, "untethered")
	}

	if z.Project != "" {
		t.Errorf("Project = %q, want empty string", z.Project)
	}
}
