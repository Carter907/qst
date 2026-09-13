package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var renameDir string

var renameCmd = &cobra.Command{
	Use:   "rename [old_name] [new_name]",
	Short: "Rename a guide and update all references to it",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		oldName := args[0]
		newName := args[1]

		guides, err := graph.ParseGraph(renameDir)
		if err != nil {
			fmt.Printf("Error parsing graph: %v\n", err)
			os.Exit(1)
		}

		if _, exists := guides[oldName]; !exists {
			fmt.Printf("Error: guide '%s' does not exist in the graph\n", oldName)
			os.Exit(1)
		}
		if _, exists := guides[newName]; exists {
			fmt.Printf("Error: guide '%s' already exists in the graph\n", newName)
			os.Exit(1)
		}

		// Iterate and update references
		updatedCount := 0
		for _, guide := range guides {
			needsUpdate := false
			meta := guide.Metadata

			for i, p := range meta.Prerequisites {
				if p == oldName {
					meta.Prerequisites[i] = newName
					needsUpdate = true
				}
			}

			for i, s := range meta.SubGuides {
				if s.Guide == oldName {
					meta.SubGuides[i].Guide = newName
					needsUpdate = true
				}
			}

			if needsUpdate {
				err = graph.UpdateGuideMetadata(guide.Path, meta)
				if err != nil {
					fmt.Printf("Failed to update guide '%s': %v\n", guide.ID, err)
					os.Exit(1)
				}
				updatedCount++
			}
		}

		// Finally, rename the file itself
		oldPath := guides[oldName].Path
		newPath := filepath.Join(filepath.Dir(oldPath), newName+".md")
		err = os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Printf("Failed to rename file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully renamed '%s' to '%s' and updated %d references.\n", oldName, newName, updatedCount)
	},
}

func init() {
	renameCmd.Flags().StringVarP(&renameDir, "dir", "d", ".", "Knowledge graph directory")
	rootCmd.AddCommand(renameCmd)
}
