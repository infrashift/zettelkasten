package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"zk/internal/config"
	"zk/internal/graph"
	"zk/internal/index"
	"zk/internal/templates"
	"zk/internal/zettel"

	"github.com/spf13/cobra"
)

func main() {
	var noteType string
	var projectFlag string
	var categoryFlag string
	var tagsFlag []string
	var limitFlag int
	var jsonFlag bool
	var graphLimitFlag int
	var outputFlag string
	var templateFlag string

	// Link flags for create command
	var createLinkFlag []string
	var createLinkDailyFlag bool

	var rootCmd = &cobra.Command{Use: "zk"}

	var createCmd = &cobra.Command{
		Use:   "create [title]",
		Short: "Create a new zettel",
		Long: `Create a new zettel with optional template.

Examples:
  zk create "My Note"                           # Basic fleeting note
  zk create "Meeting Notes" --template meeting  # From template
  zk create "Feature Spec" -T feature -p myproj # Template with project`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			cwd, _ := os.Getwd()
			project := zettel.GetGitContext(cwd)
			if projectFlag != "" {
				project = projectFlag
			}
			title := args[0]

			// Generate ID
			id := time.Now().Format("200601021504")

			// Collect links
			var links []string
			if createLinkDailyFlag {
				// Add today's daily note ID
				dailyID := time.Now().Format("200601020000")
				links = append(links, dailyID)
			}
			links = append(links, createLinkFlag...)

			var content string
			var category string

			if templateFlag != "" {
				// Use template
				tmpl, err := templates.Get(templateFlag)
				if err != nil {
					return fmt.Errorf("template error: %w", err)
				}

				// Permanent templates require project
				if tmpl.Category == "permanent" && project == "" {
					return fmt.Errorf("template '%s' creates permanent notes which require a project; use --project flag", templateFlag)
				}

				// Use GenerateNoteWithOptions if we have links
				if len(links) > 0 {
					opts := &templates.TodoOptions{
						Links: links,
					}
					content, err = tmpl.GenerateNoteWithOptions(id, title, project, tagsFlag, opts)
				} else {
					content, err = tmpl.GenerateNote(id, title, project, tagsFlag)
				}
				if err != nil {
					return fmt.Errorf("failed to generate note from template: %w", err)
				}
				category = tmpl.Category

				fmt.Printf("Creating %s note from template '%s': %s", category, templateFlag, title)
			} else {
				// Basic note without template
				category = noteType
				content = generateBasicNote(id, title, project, category, tagsFlag, links)

				if category == "fleeting" && project == "" {
					fmt.Printf("Creating %s note: %s (no project context)", category, title)
				} else {
					fmt.Printf("Creating %s note: %s (Project: %s)", category, title, project)
				}
			}

			if len(links) > 0 {
				fmt.Printf("\nLinked to: %s", strings.Join(links, ", "))
			}

			// Determine output directory
			var outputDir string
			if category == "permanent" {
				outputDir = filepath.Join(cfg.RootPath, cfg.Folders.Permanent)
			} else {
				outputDir = filepath.Join(cfg.RootPath, cfg.Folders.Fleeting)
			}

			// Create directory if it doesn't exist
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			// Write file
			filePath := filepath.Join(outputDir, id+".md")
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			fmt.Printf("\nCreated: %s\n", filePath)

			// Auto-index the new note
			if err := autoIndex(cfg, filePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to index note: %v\n", err)
			}

			return nil
		},
	}

	var templatesCmd = &cobra.Command{
		Use:   "templates",
		Short: "List available note templates",
		Long: `List all available note templates.

Templates provide pre-defined structures for common note types like
meeting notes, book reviews, user stories, and more.

Use with: zk create "title" --template <name>`,
		Run: func(cmd *cobra.Command, args []string) {
			names := templates.List()
			sort.Strings(names)

			fmt.Println("Available templates:")
			fmt.Println()
			for _, name := range names {
				tmpl, _ := templates.Get(name)
				fmt.Printf("  %-15s %s [%s]\n", name, tmpl.Description, tmpl.Category)
			}
			fmt.Println("\nUsage: zk create \"title\" --template <name>")
		},
	}

	var indexCmd = &cobra.Command{
		Use:   "index [file_path]",
		Short: "Index a zettel for searching",
		Long: `Index a single zettel file or all zettels in a directory.

If given a file, indexes that file.
If given a directory, indexes all .md files recursively.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			idx, err := index.OpenOrCreateIndex(cfg.IndexPath)
			if err != nil {
				return fmt.Errorf("failed to open index: %w", err)
			}
			defer idx.Close()

			path := args[0]
			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("failed to stat path: %w", err)
			}

			var files []string
			if info.IsDir() {
				err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() && strings.HasSuffix(p, ".md") {
						files = append(files, p)
					}
					return nil
				})
				if err != nil {
					return fmt.Errorf("failed to walk directory: %w", err)
				}
			} else {
				files = []string{path}
			}

			indexed := 0
			for _, filePath := range files {
				if err := indexFile(idx, filePath); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to index %s: %v\n", filePath, err)
					continue
				}
				indexed++
			}

			fmt.Printf("Indexed %d file(s)\n", indexed)
			return nil
		},
	}

	var searchCmd = &cobra.Command{
		Use:   "search [query]",
		Short: "Search zettels",
		Long: `Search zettels by full-text query and/or field filters.

Examples:
  zk search "authentication"              # Full-text search
  zk search --project myproject           # Filter by project
  zk search --category permanent          # Filter by category
  zk search --tag golang --tag api        # Filter by tags (AND)
  zk search "auth" --project myproject    # Combined search`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			idx, err := index.OpenOrCreateIndex(cfg.IndexPath)
			if err != nil {
				return fmt.Errorf("failed to open index: %w", err)
			}
			defer idx.Close()

			opts := index.SearchOptions{
				Project:  projectFlag,
				Category: categoryFlag,
				Tags:     tagsFlag,
				Limit:    limitFlag,
			}

			if len(args) > 0 {
				opts.Query = strings.Join(args, " ")
			}

			results, total, err := index.Search(idx, opts)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			if jsonFlag {
				jsonOutput, err := index.SearchResultsToJSON(results)
				if err != nil {
					return fmt.Errorf("failed to format JSON: %w", err)
				}
				fmt.Println(jsonOutput)
			} else {
				fmt.Print(index.FormatSearchResults(results, total))
			}

			return nil
		},
	}

	var graphCmd = &cobra.Command{
		Use:   "graph [path]",
		Short: "Generate a graph visualization of zettels",
		Long: `Generate a Mermaid graph visualization showing relationships between zettels.

The output is a markdown file with:
- A Mermaid flowchart diagram
- A table of all nodes with links to files
- A list of relationships

Examples:
  zk graph .                     # Graph all zettels in current directory
  zk graph ~/zettelkasten        # Graph all zettels in specified directory
  zk graph . --limit 20          # Show up to 20 nodes (default: 10)
  zk graph . --output my-graph   # Custom output filename`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			path := args[0]
			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("failed to stat path: %w", err)
			}

			// Collect all markdown files
			var files []string
			if info.IsDir() {
				err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() && strings.HasSuffix(p, ".md") {
						// Skip files in the graph output directory
						if strings.Contains(p, cfg.GraphPath) {
							return nil
						}
						files = append(files, p)
					}
					return nil
				})
				if err != nil {
					return fmt.Errorf("failed to walk directory: %w", err)
				}
			} else {
				files = []string{path}
			}

			// Build the graph
			g := graph.New()
			for _, filePath := range files {
				node, err := parseZettelForGraph(filePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", filePath, err)
					continue
				}
				g.AddNode(node)
			}

			if g.NodeCount() == 0 {
				return fmt.Errorf("no valid zettels found")
			}

			// Build parent-child relationships
			g.BuildRelationships()

			// Get nodes up to the limit
			nodes := g.FindAllConnected(graphLimitFlag)
			edges := g.GetEdges(nodes)

			// Generate markdown
			absPath, _ := filepath.Abs(path)
			markdown := graph.GenerateMarkdown(nodes, edges, absPath, "")

			// Determine output path
			graphDir := cfg.GraphPath
			if !filepath.IsAbs(graphDir) {
				// Make relative to the scanned path if it's a directory
				if info.IsDir() {
					graphDir = filepath.Join(path, graphDir)
				} else {
					graphDir = filepath.Join(filepath.Dir(path), graphDir)
				}
			}

			// Create graph directory
			if err := os.MkdirAll(graphDir, 0755); err != nil {
				return fmt.Errorf("failed to create graph directory: %w", err)
			}

			// Ensure .gitignore exists and includes graph path
			if err := ensureGitignore(path, cfg.GraphPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not update .gitignore: %v\n", err)
			}

			// Determine output filename
			outputName := outputFlag
			if outputName == "" {
				outputName = fmt.Sprintf("graph-%s", time.Now().Format("20060102-150405"))
			}
			if !strings.HasSuffix(outputName, ".md") {
				outputName += ".md"
			}

			outputPath := filepath.Join(graphDir, outputName)
			if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
				return fmt.Errorf("failed to write graph file: %w", err)
			}

			fmt.Printf("Generated graph with %d nodes and %d edges\n", len(nodes), len(edges))
			fmt.Printf("Output: %s\n", outputPath)
			return nil
		},
	}

	var promoteCmd = &cobra.Command{
		Use:   "promote [file_path]",
		Short: "Promote a fleeting note to permanent",
		Long: `Promote a fleeting note to a permanent note.

Permanent notes require a project context. If the note doesn't have one,
you must provide it via --project flag or be in a git repository.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			z, err := config.ParseFrontmatter(content)
			if err != nil {
				return fmt.Errorf("failed to parse frontmatter: %w", err)
			}

			if z.Category == "permanent" {
				fmt.Println("Note is already permanent")
				return nil
			}

			project := z.Project
			if projectFlag != "" {
				project = projectFlag
			}
			if project == "" {
				cwd, _ := os.Getwd()
				project = zettel.GetGitContext(cwd)
			}
			if project == "" {
				return fmt.Errorf("permanent notes require a project context; use --project flag or run from a git repository")
			}

			newContent, err := updateFrontmatter(content, map[string]string{
				"category": "permanent",
				"project":  project,
			})
			if err != nil {
				return fmt.Errorf("failed to update frontmatter: %w", err)
			}

			if err := os.WriteFile(filePath, newContent, 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			fmt.Printf("Promoted to permanent note (Project: %s)\n", project)

			// Auto-index the updated note
			cfg, _ := config.Load("")
			if cfg != nil {
				if err := autoIndex(cfg, filePath); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to index note: %v\n", err)
				}
			}

			return nil
		},
	}

	var setProjectCmd = &cobra.Command{
		Use:   "set-project [file_path] [project]",
		Short: "Set or update the project for a zettel",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			project := args[1]

			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			newContent, err := updateFrontmatter(content, map[string]string{
				"project": project,
			})
			if err != nil {
				return fmt.Errorf("failed to update frontmatter: %w", err)
			}

			if err := os.WriteFile(filePath, newContent, 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			fmt.Printf("Updated project to: %s\n", project)

			// Auto-index the updated note
			cfg, _ := config.Load("")
			if cfg != nil {
				if err := autoIndex(cfg, filePath); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to index note: %v\n", err)
				}
			}

			return nil
		},
	}

	var backlinksCmd = &cobra.Command{
		Use:   "backlinks [id_or_file]",
		Short: "Find all notes that link to a zettel",
		Long: `Find all zettels that contain links to the specified zettel.

Links are detected in the format [[id]] or [[id|title]].

Examples:
  zk backlinks 202602131045           # Find backlinks by ID
  zk backlinks ./notes/202602131045.md  # Find backlinks by file path
  zk backlinks 202602131045 --json    # Output as JSON`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Determine the target ID
			targetID := args[0]
			targetTitle := ""

			// If it's a file path, extract the ID from frontmatter
			if strings.HasSuffix(targetID, ".md") {
				content, err := os.ReadFile(targetID)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				z, err := config.ParseFrontmatter(content)
				if err != nil {
					return fmt.Errorf("failed to parse frontmatter: %w", err)
				}
				targetID = z.ID
				targetTitle = z.Title
			}

			// Determine search path
			searchPath := cfg.RootPath
			if searchPath == "" {
				searchPath, _ = os.Getwd()
			}
			// If root_path doesn't exist, use current directory
			if _, err := os.Stat(searchPath); os.IsNotExist(err) {
				searchPath, _ = os.Getwd()
			}

			// Collect all markdown files
			var files []string
			err = filepath.Walk(searchPath, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && strings.HasSuffix(p, ".md") {
					// Skip graph output files
					if strings.Contains(p, cfg.GraphPath) {
						return nil
					}
					files = append(files, p)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to walk directory: %w", err)
			}

			// Find backlinks
			type Backlink struct {
				ID       string `json:"id"`
				Title    string `json:"title"`
				Project  string `json:"project"`
				Category string `json:"category"`
				FilePath string `json:"file_path"`
			}

			var backlinks []Backlink
			for _, filePath := range files {
				content, err := os.ReadFile(filePath)
				if err != nil {
					continue
				}

				// Check if this file links to the target
				body := extractBody(content)
				links := graph.ExtractLinks(body)

				for _, linkID := range links {
					if linkID == targetID {
						// This file links to our target
						z, err := config.ParseFrontmatter(content)
						if err != nil {
							continue
						}

						absPath, _ := filepath.Abs(filePath)
						backlinks = append(backlinks, Backlink{
							ID:       z.ID,
							Title:    z.Title,
							Project:  z.Project,
							Category: z.Category,
							FilePath: absPath,
						})
						break
					}
				}
			}

			if jsonFlag {
				output, err := backlinksToJSON(backlinks)
				if err != nil {
					return fmt.Errorf("failed to format JSON: %w", err)
				}
				fmt.Println(output)
			} else {
				if targetTitle != "" {
					fmt.Printf("Backlinks to: %s (%s)\n\n", targetTitle, targetID)
				} else {
					fmt.Printf("Backlinks to: %s\n\n", targetID)
				}

				if len(backlinks) == 0 {
					fmt.Println("No backlinks found")
				} else {
					for _, bl := range backlinks {
						fmt.Printf("  %s: %s [%s]\n", bl.ID, bl.Title, bl.Category)
						fmt.Printf("    %s\n\n", bl.FilePath)
					}
					fmt.Printf("Found %d backlink(s)\n", len(backlinks))
				}
			}

			return nil
		},
	}

	// Flags
	createCmd.Flags().StringVarP(&noteType, "type", "t", "fleeting", "Note category (fleeting or permanent)")
	createCmd.Flags().StringVarP(&projectFlag, "project", "p", "", "Project context (auto-detected from git if not provided)")
	createCmd.Flags().StringVarP(&templateFlag, "template", "T", "", "Use a template (see 'zk templates' for list)")
	createCmd.Flags().StringSliceVar(&tagsFlag, "tags", nil, "Additional tags (comma-separated)")
	createCmd.Flags().StringSliceVar(&createLinkFlag, "link", nil, "Link to zettel ID (can be repeated)")
	createCmd.Flags().BoolVar(&createLinkDailyFlag, "link-daily", false, "Link to today's daily note")

	promoteCmd.Flags().StringVarP(&projectFlag, "project", "p", "", "Project context for the permanent note")

	searchCmd.Flags().StringVarP(&projectFlag, "project", "p", "", "Filter by project")
	searchCmd.Flags().StringVarP(&categoryFlag, "category", "c", "", "Filter by category (fleeting/permanent)")
	searchCmd.Flags().StringSliceVarP(&tagsFlag, "tag", "T", nil, "Filter by tag (can be repeated)")
	searchCmd.Flags().IntVarP(&limitFlag, "limit", "l", 20, "Maximum number of results")
	searchCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output results as JSON")

	graphCmd.Flags().IntVarP(&graphLimitFlag, "limit", "l", 10, "Maximum number of nodes to display")
	graphCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output filename (default: graph-TIMESTAMP.md)")

	backlinksCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output results as JSON")

	// Daily notes flags
	var dateFlag string
	var yesterdayFlag bool
	var listFlag bool
	var weekFlag bool
	var monthFlag bool

	var dailyCmd = &cobra.Command{
		Use:   "daily",
		Short: "Create or open a daily note",
		Long: `Create or open a daily note for a specific date.

Daily notes provide a consistent place for capturing thoughts, tasks, and
reflections throughout the day. Each day has exactly one daily note.

Examples:
  zk daily                    # Today's daily note
  zk daily --yesterday        # Yesterday's daily note
  zk daily --date 2026-02-10  # Specific date
  zk daily --list             # List recent daily notes
  zk daily --list --week      # This week's daily notes
  zk daily --list --month     # This month's daily notes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Determine target date
			targetDate := time.Now()
			if yesterdayFlag {
				targetDate = targetDate.AddDate(0, 0, -1)
			} else if dateFlag != "" {
				parsed, err := time.Parse("2006-01-02", dateFlag)
				if err != nil {
					return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
				}
				targetDate = parsed
			}

			// Handle --list mode
			if listFlag {
				return listDailyNotes(cfg, weekFlag, monthFlag, jsonFlag)
			}

			// Create or open daily note
			return createOrOpenDaily(cfg, targetDate)
		},
	}

	dailyCmd.Flags().StringVar(&dateFlag, "date", "", "Target date (YYYY-MM-DD format)")
	dailyCmd.Flags().BoolVar(&yesterdayFlag, "yesterday", false, "Use yesterday's date")
	dailyCmd.Flags().BoolVar(&listFlag, "list", false, "List recent daily notes")
	dailyCmd.Flags().BoolVar(&weekFlag, "week", false, "Show this week's notes (with --list)")
	dailyCmd.Flags().BoolVar(&monthFlag, "month", false, "Show this month's notes (with --list)")
	dailyCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON (with --list)")

	// Todo-specific flags
	var dueFlag string
	var priorityFlag string
	var statusFlag string
	var overdueFlag bool
	var closedFlag bool
	var todayFlag bool
	var thisWeekFlag bool

	// Link flags for todo
	var linkFlag []string
	var linkDailyFlag bool

	var todoCmd = &cobra.Command{
		Use:   "todo [title]",
		Short: "Create a new todo",
		Long: `Create a new actionable todo item.

Todos are a special type of zettel with status tracking, due dates, and priority.
Todos can be linked to other zettels including daily notes.

Examples:
  zk todo "Fix login bug"                          # Basic todo
  zk todo "Update docs" --due 2026-02-20           # With due date
  zk todo "Critical fix" --priority high           # With priority
  zk todo "Feature work" --project myproj          # With project
  zk todo "Review PR" --link-daily                 # Link to today's daily note
  zk todo "Follow up" --link 202602131045          # Link to specific zettel`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			title := args[0]
			id := time.Now().Format("200601021504")

			cwd, _ := os.Getwd()
			project := zettel.GetGitContext(cwd)
			if projectFlag != "" {
				project = projectFlag
			}

			// Collect links
			var links []string
			if linkDailyFlag {
				// Add today's daily note ID
				dailyID := time.Now().Format("200601020000")
				links = append(links, dailyID)
			}
			links = append(links, linkFlag...)

			// Get the todo template
			tmpl, err := templates.Get("todo")
			if err != nil {
				return fmt.Errorf("failed to get todo template: %w", err)
			}

			// Set todo options
			todoOpts := &templates.TodoOptions{
				Status:   "open",
				Due:      dueFlag,
				Priority: priorityFlag,
				Links:    links,
			}

			// Generate the note with todo template
			content, err := tmpl.GenerateNoteWithOptions(id, title, project, tagsFlag, todoOpts)
			if err != nil {
				return fmt.Errorf("failed to generate todo: %w", err)
			}

			// Todos go in fleeting directory
			outputDir := filepath.Join(cfg.RootPath, cfg.Folders.Fleeting)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			filePath := filepath.Join(outputDir, id+".md")
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			fmt.Printf("Created todo: %s\n", title)
			if dueFlag != "" {
				fmt.Printf("Due: %s\n", dueFlag)
			}
			if len(links) > 0 {
				fmt.Printf("Linked to: %s\n", strings.Join(links, ", "))
			}
			fmt.Printf("File: %s\n", filePath)

			// Auto-index
			if err := autoIndex(cfg, filePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to index todo: %v\n", err)
			}

			return nil
		},
	}

	todoCmd.Flags().StringVarP(&projectFlag, "project", "p", "", "Project context")
	todoCmd.Flags().StringVar(&dueFlag, "due", "", "Due date (YYYY-MM-DD)")
	todoCmd.Flags().StringVar(&priorityFlag, "priority", "", "Priority (high, medium, low)")
	todoCmd.Flags().StringSliceVar(&tagsFlag, "tags", nil, "Additional tags")
	todoCmd.Flags().StringSliceVar(&linkFlag, "link", nil, "Link to zettel ID (can be repeated)")
	todoCmd.Flags().BoolVar(&linkDailyFlag, "link-daily", false, "Link to today's daily note")

	var doneCmd = &cobra.Command{
		Use:   "done [id_or_file]",
		Short: "Mark a todo as closed",
		Long: `Mark a todo as closed and set the completion date.

Examples:
  zk done 202602131045           # By ID
  zk done ./path/to/todo.md      # By file path`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			filePath, err := resolveZettelPath(cfg, args[0])
			if err != nil {
				return err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			z, err := config.ParseFrontmatter(content)
			if err != nil {
				return fmt.Errorf("failed to parse frontmatter: %w", err)
			}

			if !z.IsTodo() {
				return fmt.Errorf("note is not a todo (type: %s)", z.GetType())
			}

			if z.Status == "closed" {
				fmt.Println("Todo is already closed")
				return nil
			}

			completedDate := time.Now().Format("2006-01-02")
			newContent, err := updateFrontmatter(content, map[string]string{
				"status":    "closed",
				"completed": completedDate,
			})
			if err != nil {
				return fmt.Errorf("failed to update frontmatter: %w", err)
			}

			if err := os.WriteFile(filePath, newContent, 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			fmt.Printf("Closed todo: %s\n", z.Title)
			fmt.Printf("Completed: %s\n", completedDate)

			// Auto-index
			if err := autoIndex(cfg, filePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to index todo: %v\n", err)
			}

			return nil
		},
	}

	var reopenCmd = &cobra.Command{
		Use:   "reopen [id_or_file]",
		Short: "Reopen a closed todo",
		Long: `Reopen a previously closed todo.

Examples:
  zk reopen 202602131045         # By ID
  zk reopen ./path/to/todo.md    # By file path`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			filePath, err := resolveZettelPath(cfg, args[0])
			if err != nil {
				return err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			z, err := config.ParseFrontmatter(content)
			if err != nil {
				return fmt.Errorf("failed to parse frontmatter: %w", err)
			}

			if !z.IsTodo() {
				return fmt.Errorf("note is not a todo (type: %s)", z.GetType())
			}

			if z.Status != "closed" {
				fmt.Printf("Todo is already %s\n", z.Status)
				return nil
			}

			// Remove completed date and set status to open
			newContent, err := updateFrontmatterRemove(content, map[string]string{
				"status": "open",
			}, []string{"completed"})
			if err != nil {
				return fmt.Errorf("failed to update frontmatter: %w", err)
			}

			if err := os.WriteFile(filePath, newContent, 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			fmt.Printf("Reopened todo: %s\n", z.Title)

			// Auto-index
			if err := autoIndex(cfg, filePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to index todo: %v\n", err)
			}

			return nil
		},
	}

	var todosCmd = &cobra.Command{
		Use:   "todos [query]",
		Short: "List and search todos",
		Long: `List and search todos with various filters.

Examples:
  zk todos                        # All open todos
  zk todos --closed               # All closed todos
  zk todos --project myproj       # Filter by project
  zk todos --priority high        # Filter by priority
  zk todos --overdue              # Overdue todos
  zk todos --due today            # Due today
  zk todos --due this-week        # Due this week
  zk todos "fix bug"              # Full-text search in todos`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			idx, err := index.OpenOrCreateIndex(cfg.IndexPath)
			if err != nil {
				return fmt.Errorf("failed to open index: %w", err)
			}
			defer idx.Close()

			opts := index.SearchOptions{
				Type:     "todo",
				Project:  projectFlag,
				Tags:     tagsFlag,
				Limit:    limitFlag,
				Priority: priorityFlag,
			}

			// Status filter
			if closedFlag {
				opts.Status = "closed"
			} else if statusFlag != "" {
				opts.Status = statusFlag
			} else {
				// Default to open todos
				opts.Status = "open"
			}

			// Due date filters
			today := time.Now().Format("2006-01-02")
			if overdueFlag {
				opts.DueBefore = today
				opts.Status = "open" // Overdue only makes sense for open todos
			} else if todayFlag {
				opts.DueAfter = today
				opts.DueBefore = today
			} else if thisWeekFlag {
				opts.DueAfter = today
				endOfWeek := time.Now().AddDate(0, 0, 7-int(time.Now().Weekday()))
				opts.DueBefore = endOfWeek.Format("2006-01-02")
			}

			// Full-text query
			if len(args) > 0 {
				opts.Query = strings.Join(args, " ")
			}

			results, total, err := index.Search(idx, opts)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			if jsonFlag {
				jsonOutput, err := index.SearchResultsToJSON(results)
				if err != nil {
					return fmt.Errorf("failed to format JSON: %w", err)
				}
				fmt.Println(jsonOutput)
			} else {
				fmt.Print(index.FormatTodoResults(results, total))
			}

			return nil
		},
	}

	todosCmd.Flags().StringVarP(&projectFlag, "project", "p", "", "Filter by project")
	todosCmd.Flags().StringVar(&statusFlag, "status", "", "Filter by status (open, in_progress, closed)")
	todosCmd.Flags().BoolVar(&closedFlag, "closed", false, "Show closed todos")
	todosCmd.Flags().StringVar(&priorityFlag, "priority", "", "Filter by priority (high, medium, low)")
	todosCmd.Flags().BoolVar(&overdueFlag, "overdue", false, "Show overdue todos")
	todosCmd.Flags().BoolVar(&todayFlag, "today", false, "Show todos due today")
	todosCmd.Flags().BoolVar(&thisWeekFlag, "this-week", false, "Show todos due this week")
	todosCmd.Flags().StringSliceVarP(&tagsFlag, "tag", "T", nil, "Filter by tag")
	todosCmd.Flags().IntVarP(&limitFlag, "limit", "l", 50, "Maximum number of results")
	todosCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")

	var todoListCmd = &cobra.Command{
		Use:   "todo-list",
		Short: "Generate a markdown todo list",
		Long: `Generate a markdown file listing todos.

The generated file is placed in the configured todos_path directory.

Examples:
  zk todo-list                           # All open todos
  zk todo-list --project myproj          # Project-specific list
  zk todo-list --today                   # Due today
  zk todo-list --output my-todos.md      # Custom filename`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			idx, err := index.OpenOrCreateIndex(cfg.IndexPath)
			if err != nil {
				return fmt.Errorf("failed to open index: %w", err)
			}
			defer idx.Close()

			opts := index.SearchOptions{
				Type:    "todo",
				Project: projectFlag,
				Limit:   200, // Higher limit for list generation
			}

			// Default to open todos
			if closedFlag {
				opts.Status = "closed"
			} else {
				opts.Status = "open"
			}

			// Due date filters
			today := time.Now().Format("2006-01-02")
			if todayFlag {
				opts.DueAfter = today
				opts.DueBefore = today
			} else if thisWeekFlag {
				opts.DueAfter = today
				endOfWeek := time.Now().AddDate(0, 0, 7-int(time.Now().Weekday()))
				opts.DueBefore = endOfWeek.Format("2006-01-02")
			}

			results, _, err := index.Search(idx, opts)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			// Generate markdown
			markdown := generateTodoListMarkdown(results, projectFlag, todayFlag, thisWeekFlag)

			// Determine output path
			todosDir := cfg.TodosPath
			if !filepath.IsAbs(todosDir) {
				todosDir = filepath.Join(cfg.RootPath, todosDir)
			}

			if err := os.MkdirAll(todosDir, 0755); err != nil {
				return fmt.Errorf("failed to create todos directory: %w", err)
			}

			// Ensure .gitignore includes todos path
			if err := ensureGitignore(cfg.RootPath, cfg.TodosPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not update .gitignore: %v\n", err)
			}

			// Determine output filename
			outputName := outputFlag
			if outputName == "" {
				if projectFlag != "" {
					outputName = fmt.Sprintf("%s-todos.md", projectFlag)
				} else if todayFlag {
					outputName = fmt.Sprintf("todos-%s.md", today)
				} else {
					outputName = "todos.md"
				}
			}
			if !strings.HasSuffix(outputName, ".md") {
				outputName += ".md"
			}

			outputPath := filepath.Join(todosDir, outputName)
			if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
				return fmt.Errorf("failed to write todo list: %w", err)
			}

			fmt.Printf("Generated todo list with %d items\n", len(results))
			fmt.Printf("Output: %s\n", outputPath)

			return nil
		},
	}

	todoListCmd.Flags().StringVarP(&projectFlag, "project", "p", "", "Filter by project")
	todoListCmd.Flags().BoolVar(&todayFlag, "today", false, "Show todos due today")
	todoListCmd.Flags().BoolVar(&thisWeekFlag, "this-week", false, "Show todos due this week")
	todoListCmd.Flags().BoolVar(&closedFlag, "closed", false, "Include closed todos")
	todoListCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output filename")

	// Git workflow commands
	var helloCmd = &cobra.Command{
		Use:   "hello",
		Short: "Start your day by creating a dated branch",
		Long: `Start your zettelkasten session by creating a new branch from main.

The branch name will be today's date in YYYYMMDD format.

This command will:
1. Ensure you're on a clean working tree (no uncommitted changes)
2. Switch to main and pull latest changes
3. Create a new branch named with today's date

Examples:
  zk hello    # Creates branch like "20260213" from main`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Determine working directory (use zettelkasten root if it exists)
			workDir := cfg.RootPath
			if _, err := os.Stat(workDir); os.IsNotExist(err) {
				workDir, _ = os.Getwd()
			}

			// Check if we're in a git repository
			if !isGitRepo(workDir) {
				return fmt.Errorf("not a git repository: %s", workDir)
			}

			// Check for uncommitted changes
			hasChanges, err := hasUncommittedChanges(workDir)
			if err != nil {
				return fmt.Errorf("failed to check git status: %w", err)
			}
			if hasChanges {
				fmt.Println("Warning: You have uncommitted changes.")
				fmt.Println("")
				fmt.Println("View uncommitted changes with:")
				fmt.Println("  git status          # See changed files")
				fmt.Println("  git diff            # See unstaged changes")
				fmt.Println("  git diff --staged   # See staged changes")
				fmt.Println("")
				fmt.Println("Commit or stash your changes before running 'zk hello'.")
				return nil
			}

			// Generate branch name from today's date
			branchName := time.Now().Format("20060102")

			// Check if branch already exists
			exists, err := branchExists(workDir, branchName)
			if err != nil {
				return fmt.Errorf("failed to check branch: %w", err)
			}
			if exists {
				fmt.Printf("Warning: Branch '%s' already exists.\n", branchName)
				fmt.Println("")
				fmt.Println("To switch to this branch:")
				fmt.Printf("  git checkout %s\n", branchName)
				return nil
			}

			// Switch to main and pull
			if err := gitCheckout(workDir, "main"); err != nil {
				// Try "master" if "main" doesn't exist
				if err := gitCheckout(workDir, "master"); err != nil {
					return fmt.Errorf("failed to checkout main branch: %w", err)
				}
			}

			if err := gitPull(workDir); err != nil {
				// Pull might fail if no remote, that's okay
				fmt.Fprintf(os.Stderr, "Warning: git pull failed (no remote?): %v\n", err)
			}

			// Create and checkout new branch
			if err := gitCheckoutNewBranch(workDir, branchName); err != nil {
				return fmt.Errorf("failed to create branch: %w", err)
			}

			fmt.Printf("Good morning! Created branch: %s\n", branchName)
			fmt.Println("")
			fmt.Println("You're ready to capture ideas. When you're done, run:")
			fmt.Println("  zk goodbye")

			return nil
		},
	}

	var goodbyeCmd = &cobra.Command{
		Use:   "goodbye",
		Short: "End your day by committing and merging to main",
		Long: `End your zettelkasten session by committing changes and merging to main.

This command will:
1. Stage all changes in the current branch
2. Commit with a message including today's date
3. Switch to main
4. Merge the dated branch into main
5. Delete the dated branch

Examples:
  zk goodbye    # Commits changes and merges to main`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Determine working directory
			workDir := cfg.RootPath
			if _, err := os.Stat(workDir); os.IsNotExist(err) {
				workDir, _ = os.Getwd()
			}

			// Check if we're in a git repository
			if !isGitRepo(workDir) {
				return fmt.Errorf("not a git repository: %s", workDir)
			}

			// Get current branch
			currentBranch, err := getCurrentBranch(workDir)
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}

			// Verify we're on a dated branch (YYYYMMDD format)
			if !isDateBranch(currentBranch) {
				fmt.Printf("Warning: Current branch '%s' is not a dated branch (YYYYMMDD format).\n", currentBranch)
				fmt.Println("")
				fmt.Println("The 'goodbye' command is meant to be used with branches created by 'zk hello'.")
				fmt.Println("")
				fmt.Println("To commit manually:")
				fmt.Println("  git add -A")
				fmt.Println("  git commit -m \"Your message\"")
				return nil
			}

			// Check if there are changes to commit
			hasChanges, err := hasUncommittedChanges(workDir)
			if err != nil {
				return fmt.Errorf("failed to check git status: %w", err)
			}

			if hasChanges {
				// Stage all changes
				if err := gitAddAll(workDir); err != nil {
					return fmt.Errorf("failed to stage changes: %w", err)
				}

				// Commit with dated message
				commitMsg := fmt.Sprintf("Notes for %s", currentBranch)
				if err := gitCommit(workDir, commitMsg); err != nil {
					return fmt.Errorf("failed to commit: %w", err)
				}
				fmt.Printf("Committed changes: %s\n", commitMsg)
			} else {
				fmt.Println("No changes to commit.")
			}

			// Switch to main
			mainBranch := "main"
			if err := gitCheckout(workDir, mainBranch); err != nil {
				// Try "master" if "main" doesn't exist
				mainBranch = "master"
				if err := gitCheckout(workDir, mainBranch); err != nil {
					return fmt.Errorf("failed to checkout main branch: %w", err)
				}
			}

			// Merge the dated branch
			if err := gitMerge(workDir, currentBranch); err != nil {
				return fmt.Errorf("failed to merge branch: %w", err)
			}
			fmt.Printf("Merged '%s' into '%s'\n", currentBranch, mainBranch)

			// Delete the dated branch
			if err := gitDeleteBranch(workDir, currentBranch); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to delete branch '%s': %v\n", currentBranch, err)
			} else {
				fmt.Printf("Deleted branch: %s\n", currentBranch)
			}

			fmt.Println("")
			fmt.Println("Goodbye! Your notes have been merged to main.")

			return nil
		},
	}

	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(templatesCmd)
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(graphCmd)
	rootCmd.AddCommand(promoteCmd)
	rootCmd.AddCommand(setProjectCmd)
	rootCmd.AddCommand(backlinksCmd)
	rootCmd.AddCommand(dailyCmd)
	rootCmd.AddCommand(todoCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(reopenCmd)
	rootCmd.AddCommand(todosCmd)
	rootCmd.AddCommand(todoListCmd)
	rootCmd.AddCommand(helloCmd)
	rootCmd.AddCommand(goodbyeCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// parseZettelForGraph parses a zettel file and returns a graph node
func parseZettelForGraph(filePath string) (*graph.Node, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	z, err := config.ParseFrontmatter(content)
	if err != nil {
		return nil, err
	}

	// Extract links from body
	body := extractBody(content)
	links := graph.ExtractLinks(body)

	absPath, _ := filepath.Abs(filePath)

	return &graph.Node{
		ID:       z.ID,
		Title:    z.Title,
		FilePath: absPath,
		Project:  z.Project,
		Category: z.Category,
		Parent:   z.Parent,
		Links:    links,
	}, nil
}

// ensureGitignore ensures the .gitignore file includes the graph path
func ensureGitignore(basePath string, graphPath string) error {
	info, err := os.Stat(basePath)
	if err != nil {
		return err
	}

	var gitignorePath string
	if info.IsDir() {
		gitignorePath = filepath.Join(basePath, ".gitignore")
	} else {
		gitignorePath = filepath.Join(filepath.Dir(basePath), ".gitignore")
	}

	// Read existing .gitignore or create new
	var lines []string
	content, err := os.ReadFile(gitignorePath)
	if err == nil {
		lines = strings.Split(string(content), "\n")
	}

	// Check if graph path is already ignored
	graphPattern := graphPath
	if !strings.HasPrefix(graphPattern, "/") {
		graphPattern = "/" + graphPattern
	}
	graphPatternAlt := graphPath + "/"

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == graphPath || trimmed == graphPattern || trimmed == graphPatternAlt {
			return nil // Already ignored
		}
	}

	// Add graph path to .gitignore
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Add newline if file doesn't end with one
	if len(content) > 0 && content[len(content)-1] != '\n' {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}

	// Add comment and pattern
	entry := fmt.Sprintf("\n# Generated graph visualizations\n%s/\n", graphPath)
	if _, err := f.WriteString(entry); err != nil {
		return err
	}

	return nil
}

// indexFile indexes a single markdown file
func indexFile(idx interface{ Index(string, interface{}) error }, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	z, err := config.ParseFrontmatter(content)
	if err != nil {
		return err
	}

	body := extractBody(content)

	// Default type to "note" if not set
	docType := z.GetType()

	doc := &index.ZettelDoc{
		ID:        z.ID,
		Title:     z.Title,
		Type:      docType,
		Project:   z.Project,
		Category:  z.Category,
		Tags:      z.Tags,
		Links:     z.Links,
		Body:      body,
		FilePath:  filePath,
		Created:   z.Created.Format("2006-01-02T15:04:05Z07:00"),
		Status:    z.Status,
		Due:       z.Due,
		Completed: z.Completed,
		Priority:  z.Priority,
	}

	return idx.Index(z.ID, doc)
}

// extractBody extracts the markdown body after frontmatter
func extractBody(content []byte) string {
	if !bytes.HasPrefix(content, []byte("---")) {
		return string(content)
	}

	rest := content[3:]
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	}

	endIdx := bytes.Index(rest, []byte("\n---"))
	if endIdx == -1 {
		return ""
	}

	body := rest[endIdx+4:]
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	}

	return string(body)
}

// updateFrontmatter updates specific fields in YAML frontmatter
func updateFrontmatter(content []byte, updates map[string]string) ([]byte, error) {
	if !bytes.HasPrefix(content, []byte("---")) {
		return nil, fmt.Errorf("no frontmatter found")
	}

	rest := content[3:]
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	}

	endIdx := bytes.Index(rest, []byte("\n---"))
	if endIdx == -1 {
		return nil, fmt.Errorf("frontmatter closing not found")
	}

	frontmatter := rest[:endIdx]
	afterFrontmatter := rest[endIdx+4:]

	var newFrontmatter bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(frontmatter))
	updatedFields := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		updated := false

		for field, value := range updates {
			if strings.HasPrefix(line, field+":") {
				newFrontmatter.WriteString(fmt.Sprintf("%s: %q\n", field, value))
				updatedFields[field] = true
				updated = true
				break
			}
		}

		if !updated {
			newFrontmatter.WriteString(line + "\n")
		}
	}

	for field, value := range updates {
		if !updatedFields[field] {
			newFrontmatter.WriteString(fmt.Sprintf("%s: %q\n", field, value))
		}
	}

	var result bytes.Buffer
	result.WriteString("---\n")
	result.Write(newFrontmatter.Bytes())
	result.WriteString("---")
	result.Write(afterFrontmatter)

	return result.Bytes(), nil
}

// backlinksToJSON converts backlinks slice to JSON
func backlinksToJSON(backlinks interface{}) (string, error) {
	data, err := json.MarshalIndent(backlinks, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// createOrOpenDaily creates or opens a daily note for the given date
func createOrOpenDaily(cfg *config.Config, targetDate time.Time) error {
	// Daily notes use a special ID format: YYYYMMDD0000
	id := targetDate.Format("200601020000")
	dateStr := targetDate.Format("2006-01-02")
	dayName := targetDate.Format("Monday")
	title := fmt.Sprintf("%s %s", dateStr, dayName)

	// Daily notes go into a "daily" subdirectory within fleeting
	dailyDir := filepath.Join(cfg.RootPath, cfg.Folders.Fleeting, "daily",
		targetDate.Format("2006"), targetDate.Format("01"))

	filePath := filepath.Join(dailyDir, targetDate.Format("2006-01-02")+".md")

	// Check if the daily note already exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, just report it
		fmt.Printf("Daily note: %s\n", filePath)
		return nil
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dailyDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Get the daily template
	tmpl, err := templates.Get("daily")
	if err != nil {
		return fmt.Errorf("failed to get daily template: %w", err)
	}

	// Generate the note with the daily template
	content, err := tmpl.GenerateNote(id, title, "", nil)
	if err != nil {
		return fmt.Errorf("failed to generate daily note: %w", err)
	}

	// Write the file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Created daily note: %s\n", filePath)

	// Auto-index the new daily note
	if err := autoIndex(cfg, filePath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to index note: %v\n", err)
	}

	return nil
}

// listDailyNotes lists recent daily notes
func listDailyNotes(cfg *config.Config, weekOnly, monthOnly, asJSON bool) error {
	dailyDir := filepath.Join(cfg.RootPath, cfg.Folders.Fleeting, "daily")

	// Check if daily directory exists
	if _, err := os.Stat(dailyDir); os.IsNotExist(err) {
		if asJSON {
			fmt.Println("[]")
		} else {
			fmt.Println("No daily notes found")
		}
		return nil
	}

	// Calculate date range
	now := time.Now()
	var startDate time.Time
	if weekOnly {
		// Start of this week (Sunday)
		daysUntilSunday := int(now.Weekday())
		startDate = now.AddDate(0, 0, -daysUntilSunday)
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	} else if monthOnly {
		// Start of this month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	} else {
		// Default: last 14 days
		startDate = now.AddDate(0, 0, -14)
	}

	type DailyNote struct {
		Date     string `json:"date"`
		Title    string `json:"title"`
		FilePath string `json:"file_path"`
	}

	var notes []DailyNote

	// Walk the daily directory
	err := filepath.Walk(dailyDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Parse date from filename (YYYY-MM-DD.md)
		baseName := strings.TrimSuffix(filepath.Base(path), ".md")
		fileDate, err := time.Parse("2006-01-02", baseName)
		if err != nil {
			// Not a daily note filename format, skip
			return nil
		}

		// Check if within date range
		if fileDate.Before(startDate) {
			return nil
		}

		// Read title from file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		z, err := config.ParseFrontmatter(content)
		title := baseName + " " + fileDate.Format("Monday")
		if err == nil && z.Title != "" {
			title = z.Title
		}

		absPath, _ := filepath.Abs(path)
		notes = append(notes, DailyNote{
			Date:     baseName,
			Title:    title,
			FilePath: absPath,
		})

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to list daily notes: %w", err)
	}

	// Sort by date (newest first)
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Date > notes[j].Date
	})

	if asJSON {
		data, err := json.MarshalIndent(notes, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		if len(notes) == 0 {
			fmt.Println("No daily notes found in the specified range")
			return nil
		}

		rangeName := "recent"
		if weekOnly {
			rangeName = "this week"
		} else if monthOnly {
			rangeName = "this month"
		}
		fmt.Printf("Daily notes (%s):\n\n", rangeName)

		for _, note := range notes {
			fmt.Printf("  %s: %s\n", note.Date, note.Title)
		}
		fmt.Printf("\nFound %d daily note(s)\n", len(notes))
	}

	return nil
}

// autoIndex indexes a single file silently (for auto-indexing on create/update)
func autoIndex(cfg *config.Config, filePath string) error {
	idx, err := index.OpenOrCreateIndex(cfg.IndexPath)
	if err != nil {
		return err
	}
	defer idx.Close()

	return indexFile(idx, filePath)
}

// generateBasicNote creates a basic note without a template
func generateBasicNote(id, title, project, category string, tags []string, links []string) string {
	now := time.Now()

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("id: %q\n", id))
	buf.WriteString(fmt.Sprintf("title: %q\n", title))

	if project != "" {
		buf.WriteString(fmt.Sprintf("project: %q\n", project))
	}

	buf.WriteString(fmt.Sprintf("category: %q\n", category))

	// Links to other zettels
	if len(links) > 0 {
		buf.WriteString("links:\n")
		for _, link := range links {
			buf.WriteString(fmt.Sprintf("  - %q\n", link))
		}
	}

	buf.WriteString("tags:\n")
	if len(tags) > 0 {
		for _, tag := range tags {
			buf.WriteString(fmt.Sprintf("  - %q\n", tag))
		}
	} else {
		buf.WriteString("  - \"\"\n")
	}

	buf.WriteString(fmt.Sprintf("created: %q\n", now.Format(time.RFC3339)))
	buf.WriteString("---\n\n")
	buf.WriteString(fmt.Sprintf("# %s\n\n", title))

	return buf.String()
}

// resolveZettelPath resolves a zettel ID or file path to an absolute file path
func resolveZettelPath(cfg *config.Config, idOrPath string) (string, error) {
	// If it looks like a file path, use it directly
	if strings.HasSuffix(idOrPath, ".md") || strings.Contains(idOrPath, "/") {
		if _, err := os.Stat(idOrPath); err != nil {
			return "", fmt.Errorf("file not found: %s", idOrPath)
		}
		return filepath.Abs(idOrPath)
	}

	// Otherwise, it's an ID - search for it
	searchPaths := []string{
		filepath.Join(cfg.RootPath, cfg.Folders.Fleeting, idOrPath+".md"),
		filepath.Join(cfg.RootPath, cfg.Folders.Permanent, idOrPath+".md"),
	}

	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			return filepath.Abs(p)
		}
	}

	// Try recursive search as last resort
	var found string
	err := filepath.Walk(cfg.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, idOrPath+".md") {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil && err != filepath.SkipAll {
		return "", fmt.Errorf("error searching for zettel: %w", err)
	}

	if found != "" {
		return filepath.Abs(found)
	}

	return "", fmt.Errorf("zettel not found: %s", idOrPath)
}

// updateFrontmatterRemove updates fields and removes specified fields from frontmatter
func updateFrontmatterRemove(content []byte, updates map[string]string, remove []string) ([]byte, error) {
	if !bytes.HasPrefix(content, []byte("---")) {
		return nil, fmt.Errorf("no frontmatter found")
	}

	rest := content[3:]
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	}

	endIdx := bytes.Index(rest, []byte("\n---"))
	if endIdx == -1 {
		return nil, fmt.Errorf("frontmatter closing not found")
	}

	frontmatter := rest[:endIdx]
	afterFrontmatter := rest[endIdx+4:]

	// Convert remove slice to set for easy lookup
	removeSet := make(map[string]bool)
	for _, field := range remove {
		removeSet[field] = true
	}

	var newFrontmatter bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(frontmatter))
	updatedFields := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		skip := false

		// Check if this line should be removed
		for field := range removeSet {
			if strings.HasPrefix(line, field+":") {
				skip = true
				break
			}
		}

		if skip {
			continue
		}

		updated := false
		for field, value := range updates {
			if strings.HasPrefix(line, field+":") {
				newFrontmatter.WriteString(fmt.Sprintf("%s: %q\n", field, value))
				updatedFields[field] = true
				updated = true
				break
			}
		}

		if !updated {
			newFrontmatter.WriteString(line + "\n")
		}
	}

	// Add any fields that weren't already present
	for field, value := range updates {
		if !updatedFields[field] {
			newFrontmatter.WriteString(fmt.Sprintf("%s: %q\n", field, value))
		}
	}

	var result bytes.Buffer
	result.WriteString("---\n")
	result.Write(newFrontmatter.Bytes())
	result.WriteString("---")
	result.Write(afterFrontmatter)

	return result.Bytes(), nil
}

// generateTodoListMarkdown generates a markdown todo list from search results
func generateTodoListMarkdown(results []index.SearchResult, project string, today, thisWeek bool) string {
	var buf bytes.Buffer

	// Title
	title := "Todo List"
	if project != "" {
		title = fmt.Sprintf("Todo List: %s", project)
	} else if today {
		title = fmt.Sprintf("Todos Due: %s", time.Now().Format("2006-01-02"))
	} else if thisWeek {
		title = "Todos Due This Week"
	}

	buf.WriteString(fmt.Sprintf("# %s\n\n", title))
	buf.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04")))

	if len(results) == 0 {
		buf.WriteString("No todos found.\n")
		return buf.String()
	}

	// Group by priority
	high := make([]index.SearchResult, 0)
	medium := make([]index.SearchResult, 0)
	low := make([]index.SearchResult, 0)
	noPriority := make([]index.SearchResult, 0)

	for _, r := range results {
		switch r.Priority {
		case "high":
			high = append(high, r)
		case "medium":
			medium = append(medium, r)
		case "low":
			low = append(low, r)
		default:
			noPriority = append(noPriority, r)
		}
	}

	writeTodoGroup := func(title string, todos []index.SearchResult) {
		if len(todos) == 0 {
			return
		}
		buf.WriteString(fmt.Sprintf("## %s\n\n", title))
		for _, todo := range todos {
			checkbox := "[ ]"
			if todo.Status == "closed" {
				checkbox = "[x]"
			} else if todo.Status == "in_progress" {
				checkbox = "[~]"
			}
			buf.WriteString(fmt.Sprintf("- %s **%s** [[%s]]\n", checkbox, todo.Title, todo.ID))
			if todo.Due != "" {
				buf.WriteString(fmt.Sprintf("  - Due: %s\n", todo.Due))
			}
			if todo.Project != "" {
				buf.WriteString(fmt.Sprintf("  - Project: %s\n", todo.Project))
			}
		}
		buf.WriteString("\n")
	}

	writeTodoGroup("High Priority", high)
	writeTodoGroup("Medium Priority", medium)
	writeTodoGroup("Low Priority", low)
	writeTodoGroup("Other", noPriority)

	buf.WriteString(fmt.Sprintf("---\n\nTotal: %d todos\n", len(results)))

	return buf.String()
}

// Git helper functions for hello/goodbye commands

// isGitRepo checks if the given path is inside a git repository
func isGitRepo(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	return cmd.Run() == nil
}

// hasUncommittedChanges checks if there are uncommitted changes
func hasUncommittedChanges(path string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// branchExists checks if a branch exists
func branchExists(path, branch string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	cmd.Dir = path
	err := cmd.Run()
	if err != nil {
		// Branch doesn't exist
		return false, nil
	}
	return true, nil
}

// getCurrentBranch returns the current branch name
func getCurrentBranch(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// isDateBranch checks if a branch name matches YYYYMMDD format
func isDateBranch(branch string) bool {
	if len(branch) != 8 {
		return false
	}
	_, err := time.Parse("20060102", branch)
	return err == nil
}

// gitCheckout switches to an existing branch
func gitCheckout(path, branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = path
	return cmd.Run()
}

// gitCheckoutNewBranch creates and switches to a new branch
func gitCheckoutNewBranch(path, branch string) error {
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = path
	return cmd.Run()
}

// gitPull pulls latest changes from remote
func gitPull(path string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = path
	return cmd.Run()
}

// gitAddAll stages all changes
func gitAddAll(path string) error {
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = path
	return cmd.Run()
}

// gitCommit creates a commit with the given message
func gitCommit(path, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = path
	return cmd.Run()
}

// gitMerge merges a branch into the current branch
func gitMerge(path, branch string) error {
	cmd := exec.Command("git", "merge", branch, "--no-edit")
	cmd.Dir = path
	return cmd.Run()
}

// gitDeleteBranch deletes a local branch
func gitDeleteBranch(path, branch string) error {
	cmd := exec.Command("git", "branch", "-d", branch)
	cmd.Dir = path
	return cmd.Run()
}
