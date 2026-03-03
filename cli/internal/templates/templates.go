package templates

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"
	"time"
)

//go:embed meeting.md
var meetingTemplate string

//go:embed book-review.md
var bookReviewTemplate string

//go:embed snippet.md
var snippetTemplate string

//go:embed project-idea.md
var projectIdeaTemplate string

//go:embed user-story.md
var userStoryTemplate string

//go:embed feature.md
var featureTemplate string

//go:embed daily.md
var dailyTemplate string

//go:embed todo.md
var todoTemplate string

//go:embed issue.md
var issueTemplate string

// Template represents a note template
type Template struct {
	Name        string
	Description string
	Category    string
	Type        string   // "note" (default) or "todo"
	Tags        []string
	Body        string
}

// TodoOptions holds todo-specific frontmatter options
type TodoOptions struct {
	Status   string // "open", "in_progress", "closed"
	Due      string // YYYY-MM-DD
	Priority string // "high", "medium", "low"
}

// TemplateData holds the data for template rendering
type TemplateData struct {
	ID      string
	Title   string
	Date    string
	Project string
}

// BuiltinTemplates returns all built-in templates
var BuiltinTemplates = map[string]*Template{
	"meeting": {
		Name:        "meeting",
		Description: "Meeting notes with attendees and action items",
		Category:    "untethered",
		Tags:        []string{"meeting"},
		Body:        meetingTemplate,
	},
	"book-review": {
		Name:        "book-review",
		Description: "Book review with rating and key takeaways",
		Category:    "tethered",
		Tags:        []string{"book", "review"},
		Body:        bookReviewTemplate,
	},
	"snippet": {
		Name:        "snippet",
		Description: "Code snippet with context and explanation",
		Category:    "untethered",
		Tags:        []string{"code", "snippet"},
		Body:        snippetTemplate,
	},
	"project-idea": {
		Name:        "project-idea",
		Description: "Project idea with goals and next steps",
		Category:    "untethered",
		Tags:        []string{"idea", "project"},
		Body:        projectIdeaTemplate,
	},
	"user-story": {
		Name:        "user-story",
		Description: "User story in standard format with acceptance criteria",
		Category:    "untethered",
		Tags:        []string{"user-story", "requirements"},
		Body:        userStoryTemplate,
	},
	"feature": {
		Name:        "feature",
		Description: "Feature specification with requirements and design notes",
		Category:    "untethered",
		Tags:        []string{"feature", "spec"},
		Body:        featureTemplate,
	},
	"daily": {
		Name:        "daily",
		Description: "Daily note for capturing thoughts, tasks, and reflections",
		Category:    "untethered",
		Type:        "daily-note",
		Tags:        []string{"daily"},
		Body:        dailyTemplate,
	},
	"todo": {
		Name:        "todo",
		Description: "Actionable task with status tracking",
		Category:    "untethered",
		Type:        "todo",
		Tags:        []string{"todo"},
		Body:        todoTemplate,
	},
	"issue": {
		Name:        "issue",
		Description: "Issue tracking like GitHub (bug, enhancement, question)",
		Category:    "untethered",
		Type:        "issue",
		Tags:        []string{"issue"},
		Body:        issueTemplate,
	},
}

// Get returns a template by name
func Get(name string) (*Template, error) {
	t, ok := BuiltinTemplates[name]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", name)
	}
	return t, nil
}

// List returns all available template names
func List() []string {
	names := make([]string, 0, len(BuiltinTemplates))
	for name := range BuiltinTemplates {
		names = append(names, name)
	}
	return names
}

// Render renders a template with the given data
func (t *Template) Render(data TemplateData) (string, error) {
	tmpl, err := template.New(t.Name).Parse(t.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// GenerateFrontmatter generates YAML frontmatter for a template
func (t *Template) GenerateFrontmatter(id, title, project string, extraTags []string) string {
	return t.GenerateFrontmatterWithOptions(id, title, project, extraTags, nil)
}

// GenerateFrontmatterWithOptions generates YAML frontmatter with optional todo-specific fields
func (t *Template) GenerateFrontmatterWithOptions(id, title, project string, extraTags []string, todoOpts *TodoOptions) string {
	now := time.Now()

	// Merge template tags with extra tags
	allTags := make([]string, 0, len(t.Tags)+len(extraTags))
	allTags = append(allTags, t.Tags...)
	for _, tag := range extraTags {
		// Avoid duplicates
		found := false
		for _, existing := range allTags {
			if existing == tag {
				found = true
				break
			}
		}
		if !found {
			allTags = append(allTags, tag)
		}
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("id: %q\n", id))
	buf.WriteString(fmt.Sprintf("title: %q\n", title))

	// Add type if not "note" (default)
	if t.Type != "" && t.Type != "note" {
		buf.WriteString(fmt.Sprintf("type: %q\n", t.Type))
	}

	if project != "" {
		buf.WriteString(fmt.Sprintf("project: %q\n", project))
	}

	buf.WriteString(fmt.Sprintf("category: %q\n", t.Category))

	// Todo-specific fields
	if t.Type == "todo" {
		status := "open"
		if todoOpts != nil && todoOpts.Status != "" {
			status = todoOpts.Status
		}
		buf.WriteString(fmt.Sprintf("status: %q\n", status))

		if todoOpts != nil && todoOpts.Due != "" {
			buf.WriteString(fmt.Sprintf("due: %q\n", todoOpts.Due))
		}

		if todoOpts != nil && todoOpts.Priority != "" {
			buf.WriteString(fmt.Sprintf("priority: %q\n", todoOpts.Priority))
		}
	}

	buf.WriteString("tags:\n")
	for _, tag := range allTags {
		buf.WriteString(fmt.Sprintf("  - %q\n", tag))
	}

	buf.WriteString(fmt.Sprintf("created: %q\n", now.Format(time.RFC3339)))
	buf.WriteString("---\n\n")

	return buf.String()
}

// GenerateNote generates a complete note from a template
func (t *Template) GenerateNote(id, title, project string, extraTags []string) (string, error) {
	return t.GenerateNoteWithOptions(id, title, project, extraTags, nil)
}

// GenerateNoteWithOptions generates a complete note with optional todo-specific fields
func (t *Template) GenerateNoteWithOptions(id, title, project string, extraTags []string, todoOpts *TodoOptions) (string, error) {
	frontmatter := t.GenerateFrontmatterWithOptions(id, title, project, extraTags, todoOpts)

	data := TemplateData{
		ID:      id,
		Title:   title,
		Date:    time.Now().Format("2006-01-02"),
		Project: project,
	}

	body, err := t.Render(data)
	if err != nil {
		return "", err
	}

	return frontmatter + body, nil
}
