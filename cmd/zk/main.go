package main

import (
	"fmt"
	"os"
	"zk/internal/config"
	"zk/internal/zettel"

	"github.com/spf13/cobra"
)

func main() {
	var noteType string

	var rootCmd = &cobra.Command{Use: "zk"}

	var createCmd = &cobra.Command{
		Use:   "create [title]",
		Short: "Create a new zettel",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.Load()
			cwd, _ := os.Getwd()

			project := zettel.GetGitContext(cwd)
			title := args[0]

			fmt.Printf("Creating %s note: %s (Project: %s)\n", noteType, title, project)
			// Implementation for writing markdown file...
		},
	}

	var indexCmd = &cobra.Command{
		Use:   "index [file_path]",
		Short: "Update the search index for a specific file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.Load()
			idx, _ := index.OpenOrCreateIndex(cfg.IndexPath)
			defer idx.Close()

			// 1. Read the Markdown file
			// 2. Parse YAML and Body (Validated by CUE)
			// 3. doc := index.ZettelDoc{...}
			// 4. index.UpsertNote(idx, doc)
		},
	}

	createCmd.Flags().StringVarP(&noteType, "type", "t", "fleeting", "Note category")
	rootCmd.AddCommand(createCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
