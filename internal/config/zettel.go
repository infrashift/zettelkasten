package config

import (
	"bytes"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Zettel represents a zettelkasten note's frontmatter.
type Zettel struct {
	ID       string    `yaml:"id"`
	Title    string    `yaml:"title"`
	Type     string    `yaml:"type,omitempty"`     // "note" (default), "todo", "daily-note", or "issue"
	Project  string    `yaml:"project,omitempty"`
	Category string    `yaml:"category"`
	Tags     []string  `yaml:"tags"`
	Created  time.Time `yaml:"created"`
	Parent   string    `yaml:"parent,omitempty"`

	// Todo-specific fields (only used when Type == "todo")
	Status    string `yaml:"status,omitempty"`    // "open", "in_progress", "closed"
	Due       string `yaml:"due,omitempty"`       // YYYY-MM-DD
	Completed string `yaml:"completed,omitempty"` // YYYY-MM-DD (set when closed)
	Priority  string `yaml:"priority,omitempty"`  // "high", "medium", "low"
}

// IsTodo returns true if this zettel is a todo item.
func (z *Zettel) IsTodo() bool {
	return z.Type == "todo"
}

// IsDailyNote returns true if this zettel is a daily note.
func (z *Zettel) IsDailyNote() bool {
	return z.Type == "daily-note"
}

// GetType returns the zettel type, defaulting to "note" if not set.
func (z *Zettel) GetType() string {
	if z.Type == "" {
		return "note"
	}
	return z.Type
}

// ParseFrontmatter extracts and parses YAML frontmatter from markdown content.
// Frontmatter must be enclosed between "---" markers at the start of the file.
func ParseFrontmatter(content []byte) (*Zettel, error) {
	frontmatter, err := ExtractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	var z Zettel
	if err := yaml.Unmarshal(frontmatter, &z); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &z, nil
}

// ExtractFrontmatter extracts raw YAML frontmatter bytes from markdown content.
func ExtractFrontmatter(content []byte) ([]byte, error) {
	// Check for opening delimiter
	if !bytes.HasPrefix(content, []byte("---")) {
		return nil, fmt.Errorf("frontmatter must start with ---")
	}

	// Find the closing delimiter
	rest := content[3:]
	// Skip newline after opening ---
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	} else if len(rest) > 1 && rest[0] == '\r' && rest[1] == '\n' {
		rest = rest[2:]
	}

	endIdx := bytes.Index(rest, []byte("\n---"))
	if endIdx == -1 {
		// Try with \r\n
		endIdx = bytes.Index(rest, []byte("\r\n---"))
		if endIdx == -1 {
			return nil, fmt.Errorf("frontmatter closing --- not found")
		}
	}

	return rest[:endIdx], nil
}

// ValidateZettel validates a Zettel struct against the CUE schema and business rules.
func ValidateZettel(z *Zettel) error {
	// Convert to YAML and validate against CUE schema
	data, err := yaml.Marshal(z)
	if err != nil {
		return fmt.Errorf("failed to marshal zettel: %w", err)
	}

	if err := ValidateZettelYAML(data); err != nil {
		return err
	}

	// Additional business rule validation for todos
	if z.IsTodo() {
		if err := validateTodoFields(z); err != nil {
			return err
		}
	}

	return nil
}

// validateTodoFields validates todo-specific business rules.
func validateTodoFields(z *Zettel) error {
	// Status is required for todos
	if z.Status == "" {
		return fmt.Errorf("todo requires status field (open, in_progress, or closed)")
	}

	// Validate status values
	switch z.Status {
	case "open", "in_progress", "closed":
		// Valid
	default:
		return fmt.Errorf("invalid status '%s': must be open, in_progress, or closed", z.Status)
	}

	// Validate priority if provided
	if z.Priority != "" {
		switch z.Priority {
		case "high", "medium", "low":
			// Valid
		default:
			return fmt.Errorf("invalid priority '%s': must be high, medium, or low", z.Priority)
		}
	}

	return nil
}

// ParseAndValidate extracts frontmatter from content and validates it against the schema.
func ParseAndValidate(content []byte) (*Zettel, error) {
	frontmatter, err := ExtractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	// Validate raw YAML against schema
	if err := ValidateZettelYAML(frontmatter); err != nil {
		return nil, err
	}

	// Parse into struct
	var z Zettel
	if err := yaml.Unmarshal(frontmatter, &z); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &z, nil
}
