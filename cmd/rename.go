package cmd

import (
	"os"
	"path/filepath"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var renameDir string

var renameCmd = &cobra.Command{
	Use:   "rename [old_name] [new_name]",
	Short: "Rename a guide and update all references to it",
	Example: `# Rename a guide in the graph
qst rename "Old Name" "New Name" -d my_graph/

# Rename specifying file names or paths
qst rename "Old Name.md" "New Name.md"
qst rename my_graph/Old\ Name.md "New Name"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		oldName := resolveGuideAndDir(cmd, args[0], &renameDir)
		newName := cleanGuideName(args[1])

		guides, err := graph.ParseGraph(renameDir)
		if err != nil {
			cmd.PrintErrf("Error parsing graph: %v\n", err)
			exitFunc(1)
			return
		}

		if _, exists := guides[oldName]; !exists {
			cmd.PrintErrf("Error: guide '%s' does not exist in the graph\n", oldName)
			exitFunc(1)
			return
		}
		if _, exists := guides[newName]; exists {
			cmd.PrintErrf("Error: guide '%s' already exists in the graph\n", newName)
			exitFunc(1)
			return
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
					cmd.PrintErrf("Failed to update guide '%s': %v\n", guide.ID, err)
					exitFunc(1)
					return
				}
				updatedCount++
			}
		}

		// Update references in manifest.yaml tours
		config, err := graph.ParseConfig(renameDir)
		if err == nil {
			configUpdated := false
			for i, tour := range config.Tours {
				for j, g := range tour.Guides {
					if g == oldName {
						config.Tours[i].Guides[j] = newName
						configUpdated = true
						updatedCount++
					}
				}
			}
			if configUpdated {
				err = graph.UpdateConfig(renameDir, config)
				if err != nil {
					cmd.PrintErrf("Failed to update manifest.yaml: %v\n", err)
					exitFunc(1)
					return
				}
			}
		}

		// Finally, rename the file itself
		oldPath := guides[oldName].Path
		newPath := filepath.Join(filepath.Dir(oldPath), newName+".md")
		err = os.Rename(oldPath, newPath)
		if err != nil {
			cmd.PrintErrf("Failed to rename file: %v\n", err)
			exitFunc(1)
			return
		}

		cmd.Printf("Successfully renamed '%s' to '%s' and updated %d references.\n", oldName, newName, updatedCount)
	},
}

func init() {
	renameCmd.Flags().StringVarP(&renameDir, "dir", "d", ".", "Knowledge graph directory")
	rootCmd.AddCommand(renameCmd)
}
