package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "create a new knowledge graph directory",
	Long:  "You start a new knowledge graph with the new command. You only have to specify the name of the knowledge graph; a directory will be created. If you don't specify a directory, the current directory will be used. You can optionally include a starter guide template.",
	Example: `# Specify New Directory
qst new learn-cpp

# Use Current Directory
qst new

# Create with starter guide
qst new --starter`,
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) != 0 {
			dir = args[0]
		}

		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			cmd.PrintErrf("Failed to create knowledge graph: %v\n", err)
			exitFunc(1)
			return
		}

		if createStarter {
			starterPath := filepath.Join(dir, "starter.md")
			if _, err := os.Stat(starterPath); os.IsNotExist(err) {
				starterContent := `---
prerequisites: []
sub_guides: []
scope: definition
tags: ["example"]
---

## Starter Guide

This is an example guide. Replace this content with your own knowledge!
`
				err = os.WriteFile(starterPath, []byte(starterContent), 0o644)
				if err != nil {
					cmd.PrintErrf("Failed to create starter guide: %v\n", err)
					exitFunc(1)
					return
				}
			} else {
				cmd.Println("Sample guide already exists, skipping write...")
			}
		}

		metaFilePath := filepath.Join(dir, "manifest.yaml")
		if _, err := os.Stat(metaFilePath); os.IsNotExist(err) {
			metaFileContent := fmt.Sprintf(`title: Your New Knowledge
description: What can you say about this knowledge?
scopes:
  - definition
  - description
  - explanation
  - lesson
adherences:
  - strict
  - detailed
  - introductory
  - vague
relaxed_subguides: false
require_subguides: true
strict_coverage: %t
tours:`, newStrictCoverage)

			err = os.WriteFile(metaFilePath, []byte(metaFileContent), 0o644)
			if err != nil {
				cmd.PrintErrf("Failed to write metafile: %v\n", err)
				exitFunc(1)
				return
			}

		} else {
			cmd.Println("Metafile already exists, skipping write...")
		}
		cmd.Printf("Initialized knowledge graph in %s\n", dir)
	},
}

var (
	createStarter     bool
	newStrictCoverage bool
)

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().BoolVarP(&createStarter, "starter", "s", false, "Initialize with a starter markdown guide")
	newCmd.Flags().BoolVar(&newStrictCoverage, "strict-coverage", false, "Enable strict coverage mode in manifest.yaml")
}
