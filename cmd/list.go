package cmd

import (
	"fmt"
	"os"

	"github.com/Carter907/quest-cli/internal/graph"
	"github.com/spf13/cobra"
)

var (
	listDir   string
	listScope string
	listTags  []string
)

var listCmd = &cobra.Command{
	Use:     "list [directory]",
	Aliases: []string{"ls"},
	Short:   "List guides in the knowledge graph",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			listDir = args[0]
		}
		guides, err := graph.ParseGraph(listDir)
		if err != nil {
			fmt.Printf("Error parsing graph: %v\n", err)
			os.Exit(1)
		}

		for id, guide := range guides {
			if listScope != "" && guide.Metadata.Scope != listScope {
				continue
			}
			if len(listTags) > 0 {
				hasAllTags := true
				guideTags := make(map[string]bool)
				for _, t := range guide.Metadata.Tags {
					guideTags[t] = true
				}
				for _, t := range listTags {
					if !guideTags[t] {
						hasAllTags = false
						break
					}
				}
				if !hasAllTags {
					continue
				}
			}

			fmt.Println(id)
		}
	},
}

func init() {
	listCmd.Flags().StringVarP(&listDir, "dir", "d", ".", "Knowledge graph directory")
	listCmd.Flags().StringVar(&listScope, "scope", "", "Filter by scope")
	listCmd.Flags().StringSliceVar(&listTags, "tags", nil, "Filter by tags (comma separated)")

	rootCmd.AddCommand(listCmd)
}
