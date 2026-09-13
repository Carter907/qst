package cmd

import (
	"fmt"
	"os"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var statsDir string

var statsCmd = &cobra.Command{
	Use:   "stats [directory]",
	Short: "Display statistics about the knowledge graph",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			statsDir = args[0]
		}
		guides, err := graph.ParseGraph(statsDir)
		if err != nil {
			fmt.Printf("Error parsing graph: %v\n", err)
			os.Exit(1)
		}

		totalGuides := len(guides)
		totalEdges := 0
		emptyGuides := 0
		orphanedGuides := 0

		// First, collect all targets to determine in-degree
		inDegree := make(map[string]int)
		for _, guide := range guides {
			for _, p := range guide.Metadata.Prerequisites {
				inDegree[p]++
				totalEdges++
			}
			for _, s := range guide.Metadata.SubGuides {
				inDegree[s.Guide]++
				totalEdges++
			}
		}

		for id, guide := range guides {
			if !guide.HasContent {
				emptyGuides++
			}
			// It's orphaned if it has no incoming edges and no outgoing edges
			outDegree := len(guide.Metadata.Prerequisites) + len(guide.Metadata.SubGuides)
			if inDegree[id] == 0 && outDegree == 0 {
				orphanedGuides++
			}
		}

		fmt.Printf("Graph Statistics for %s\n", statsDir)
		fmt.Println("--------------------------------")
		fmt.Printf("Total Guides:      %d\n", totalGuides)
		fmt.Printf("Total Edges:       %d\n", totalEdges)
		fmt.Printf("Empty Guides:      %d\n", emptyGuides)
		fmt.Printf("Orphaned Guides:   %d\n", orphanedGuides)
	},
}

func init() {
	statsCmd.Flags().StringVarP(&statsDir, "dir", "d", ".", "Knowledge graph directory")
	rootCmd.AddCommand(statsCmd)
}
