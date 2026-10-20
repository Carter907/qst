package cmd

import (
	"strings"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var (
	deleteDir    string
	deleteDryRun bool
)

var deleteCmd = &cobra.Command{
	Use:     "delete [guide]",
	Aliases: []string{"rm"},
	Short:   "Delete a guide and cascade removal of all references to it",
	Long: `Delete removes a guide file from the knowledge graph and automatically prunes
all incoming references to it from other guides' prerequisites, subguides, and manifest tours.`,
	Example: `# Delete a guide and clean up all references across the graph
qst delete "Old Concept" -d my_graph/

# Delete specifying file name or path
qst delete "Old Concept.md" -d my_graph/
qst delete my_graph/Old\ Concept.md

# Preview which references would be pruned without deleting
qst delete "Old Concept" -d my_graph/ --dry-run

# Using the rm alias
qst rm "Old Concept"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		guideName := resolveGuideAndDir(cmd, args[0], &deleteDir)

		guides, err := graph.ParseGraph(deleteDir)
		if err != nil {
			cmd.PrintErrf("Error parsing graph: %v\n", err)
			exitFunc(1)
			return
		}

		targetGuide, exists := guides[guideName]
		if !exists {
			cmd.PrintErrf("Error: guide '%s' does not exist in graph\n", guideName)
			exitFunc(1)
			return
		}

		if deleteDryRun {
			prereqRefs := []string{}
			subRefs := []string{}
			for id, g := range guides {
				if id == guideName {
					continue
				}
				for _, p := range g.Metadata.Prerequisites {
					if p == guideName {
						prereqRefs = append(prereqRefs, id)
					}
				}
				for _, s := range g.Metadata.SubGuides {
					if s.Guide == guideName {
						subRefs = append(subRefs, id)
					}
				}
			}

			tourRefs := []string{}
			config, err := graph.ParseConfig(deleteDir)
			if err == nil {
				for _, t := range config.Tours {
					for _, g := range t.Guides {
						if g == guideName {
							tourRefs = append(tourRefs, t.Name)
							break
						}
					}
				}
			}

			totalRefs := len(prereqRefs) + len(subRefs) + len(tourRefs)
			cmd.Printf("Dry run: Would delete '%s' (%s)\n", guideName, targetGuide.Path)
			cmd.Printf("References to prune (%d total):\n", totalRefs)
			if len(prereqRefs) > 0 {
				cmd.Printf("  • Prerequisites in %d guides: %s\n", len(prereqRefs), strings.Join(prereqRefs, ", "))
			}
			if len(subRefs) > 0 {
				cmd.Printf("  • Subguides in %d guides: %s\n", len(subRefs), strings.Join(subRefs, ", "))
			}
			if len(tourRefs) > 0 {
				cmd.Printf("  • Tours (%d): %s\n", len(tourRefs), strings.Join(tourRefs, ", "))
			}
			if totalRefs == 0 {
				cmd.Println("  (No incoming references found)")
			}
			return
		}

		prunedCount, err := graph.DeleteGuide(deleteDir, guideName)
		if err != nil {
			cmd.PrintErrf("Failed to delete guide '%s': %v\n", guideName, err)
			exitFunc(1)
			return
		}

		cmd.Printf("Successfully deleted guide '%s' and pruned %d references.\n", guideName, prunedCount)
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteDir, "dir", "d", ".", "Knowledge graph directory")
	deleteCmd.Flags().BoolVar(&deleteDryRun, "dry-run", false, "Preview references that would be pruned without deleting")
	rootCmd.AddCommand(deleteCmd)
}
