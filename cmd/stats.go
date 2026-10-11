package cmd

import (
	"strings"

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

		config, err := graph.ParseConfig(statsDir)
		if err != nil {
			cmd.PrintErrf("Error loading config: %v\n", err)
			exitFunc(1)
			return
		}

		guides, err := graph.ParseGraph(statsDir)
		if err != nil {
			cmd.PrintErrf("Error parsing graph: %v\n", err)
			exitFunc(1)
			return
		}

		totalGuides := len(guides)
		prereqEdges := 0
		subguideEdges := 0
		emptyGuides := 0
		orphanedGuides := 0

		inDegree := make(map[string]int)
		outDegree := make(map[string]int)
		scopeCounts := make(map[string]int)

		for id, guide := range guides {
			if !guide.HasContent {
				emptyGuides++
			}
			scopeCounts[guide.Metadata.Scope]++

			for _, p := range guide.Metadata.Prerequisites {
				if _, exists := guides[p]; exists {
					inDegree[id]++
					outDegree[p]++
					prereqEdges++
				}
			}
			for _, s := range guide.Metadata.SubGuides {
				if _, exists := guides[s.Guide]; exists {
					inDegree[s.Guide]++
					outDegree[id]++
					subguideEdges++
				}
			}
		}

		for id := range guides {
			if inDegree[id] == 0 && outDegree[id] == 0 {
				orphanedGuides++
			}
		}

		totalEdges := prereqEdges + subguideEdges

		title := config.Title
		if title == "" {
			title = statsDir
		}

		cmd.Printf("Graph Statistics: %s (%s)\n", title, statsDir)
		cmd.Println(strings.Repeat("=", 60))

		// 1. Overview
		cmd.Println("Overview:")
		cmd.Printf("  Total Guides:          %d\n", totalGuides)
		cmd.Printf("  Prerequisite Edges:    %d\n", prereqEdges)
		cmd.Printf("  Subguide Edges:        %d\n", subguideEdges)
		cmd.Printf("  Total Edges:           %d\n", totalEdges)
		cmd.Printf("  Empty Guides:          %d\n", emptyGuides)
		cmd.Printf("  Orphaned Guides:       %d\n", orphanedGuides)

		// 2. Scope Distribution
		cmd.Println("\nScope Distribution:")
		seenScopes := make(map[string]bool)
		for _, s := range config.Scopes {
			seenScopes[s] = true
			count := scopeCounts[s]
			pct := 0.0
			if totalGuides > 0 {
				pct = (float64(count) / float64(totalGuides)) * 100.0
			}
			cmd.Printf("  • %-18s %3d (%5.1f%%)\n", s+":", count, pct)
		}
		for s, count := range scopeCounts {
			if !seenScopes[s] {
				pct := 0.0
				if totalGuides > 0 {
					pct = (float64(count) / float64(totalGuides)) * 100.0
				}
				name := s
				if name == "" {
					name = "(unscoped)"
				}
				cmd.Printf("  • %-18s %3d (%5.1f%%)\n", name+":", count, pct)
			}
		}

		// 3. Graph Depth & Learning Paths
		longestChain := findLongestPrereqChain(guides)
		subDepth := findMaxSubguideDepth(guides)

		cmd.Println("\nGraph Depth & Learning Paths:")
		prereqDepth := len(longestChain)
		if prereqDepth == 0 {
			cmd.Println("  Prerequisite Depth:    0 guides")
		} else {
			cmd.Printf("  Prerequisite Depth:    %d guides\n", prereqDepth)
			if prereqDepth <= 5 {
				cmd.Printf("  Longest Chain:         %s\n", strings.Join(longestChain, " -> "))
			} else {
				start := longestChain[0]
				mid := longestChain[1]
				pen := longestChain[len(longestChain)-2]
				end := longestChain[len(longestChain)-1]
				cmd.Printf("  Longest Chain:         %s -> %s -> ... -> %s -> %s\n", start, mid, pen, end)
			}
		}
		cmd.Printf("  Subguide Max Depth:    %d levels\n", subDepth)

		// 4. Tour Coverage
		tourGuides := make(map[string]bool)
		for _, tour := range config.Tours {
			for _, g := range tour.Guides {
				if _, exists := guides[g]; exists {
					tourGuides[g] = true
				}
			}
		}
		touredCount := len(tourGuides)
		untouredCount := totalGuides - touredCount
		tourPct := 0.0
		if totalGuides > 0 {
			tourPct = (float64(touredCount) / float64(totalGuides)) * 100.0
		}

		cmd.Println("\nTour Coverage:")
		cmd.Printf("  Total Tours:           %d\n", len(config.Tours))
		cmd.Printf("  Guides in Tours:       %d/%d (%5.1f%%)\n", touredCount, totalGuides, tourPct)
		cmd.Printf("  Untoured Guides:       %d\n", untouredCount)

		// 5. Content & Subguide Coverage
		totalBodyLines := 0
		totalCoveredLines := 0
		totalUnmappedLines := 0
		totalOverlappingLines := 0

		for _, guide := range guides {
			report, err := graph.AnalyzeGuideCoverage(guide, config)
			if err == nil && report.TotalBodyLines > 0 {
				totalBodyLines += report.TotalBodyLines
				totalCoveredLines += report.CoveredLines
				totalUnmappedLines += report.UnmappedLines
				totalOverlappingLines += report.OverlappingLines
			}
		}

		if totalBodyLines > 0 {
			covPct := (float64(totalCoveredLines) / float64(totalBodyLines)) * 100.0
			cmd.Println("\nContent & Subguide Coverage:")
			cmd.Printf("  Total Body Lines:      %d lines\n", totalBodyLines)
			cmd.Printf("  Mapped Lines:          %d (%5.1f%%)\n", totalCoveredLines, covPct)
			cmd.Printf("  Unmapped Lines:        %d\n", totalUnmappedLines)
			cmd.Printf("  Overlapping Lines:     %d\n", totalOverlappingLines)
		}
	},
}

func findLongestPrereqChain(guides map[string]graph.Guide) []string {
	memo := make(map[string][]string)
	visiting := make(map[string]bool)

	var getChain func(id string) []string
	getChain = func(id string) []string {
		if chain, ok := memo[id]; ok {
			return chain
		}
		if visiting[id] {
			return []string{id}
		}
		visiting[id] = true
		defer func() { visiting[id] = false }()

		guide, ok := guides[id]
		if !ok {
			return nil
		}

		var bestParentChain []string
		for _, p := range guide.Metadata.Prerequisites {
			if _, exists := guides[p]; !exists {
				continue
			}
			pChain := getChain(p)
			if len(pChain) > len(bestParentChain) {
				bestParentChain = pChain
			}
		}

		res := append(append([]string(nil), bestParentChain...), id)
		memo[id] = res
		return res
	}

	var longest []string
	for id := range guides {
		chain := getChain(id)
		if len(chain) > len(longest) {
			longest = chain
		}
	}
	return longest
}

func findMaxSubguideDepth(guides map[string]graph.Guide) int {
	hasSubguides := false
	for _, g := range guides {
		if len(g.Metadata.SubGuides) > 0 {
			hasSubguides = true
			break
		}
	}
	if !hasSubguides {
		return 0
	}

	memo := make(map[string]int)
	visiting := make(map[string]bool)

	var getDepth func(id string) int
	getDepth = func(id string) int {
		if d, ok := memo[id]; ok {
			return d
		}
		if visiting[id] {
			return 1
		}
		visiting[id] = true
		defer func() { visiting[id] = false }()

		guide, ok := guides[id]
		if !ok {
			return 0
		}

		maxChildDepth := 0
		for _, s := range guide.Metadata.SubGuides {
			if _, exists := guides[s.Guide]; !exists {
				continue
			}
			childDepth := getDepth(s.Guide)
			if childDepth > maxChildDepth {
				maxChildDepth = childDepth
			}
		}

		res := 1 + maxChildDepth
		memo[id] = res
		return res
	}

	maxDepth := 0
	for id := range guides {
		d := getDepth(id)
		if d > maxDepth {
			maxDepth = d
		}
	}
	return maxDepth
}

func init() {
	statsCmd.Flags().StringVarP(&statsDir, "dir", "d", ".", "Knowledge graph directory")
	rootCmd.AddCommand(statsCmd)
}
