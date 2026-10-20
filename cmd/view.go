package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Carter907/qst/internal/graph"
	"github.com/Carter907/qst/internal/ssg"
	"github.com/spf13/cobra"
)

var viewGuide string

var viewCmd = &cobra.Command{
	Use:   "view [directory]",
	Short: "Export the knowledge graph to a Mermaid.js diagram",
	Long: `The view command reads the knowledge directory and exports its structure as a Mermaid.js graph.
Prerequisite links are shown as solid arrows, and sub-guide relationships are shown as dotted arrows.
Use --guide (-g) to export a localized forward-only subgraph centered on a specific guide.`,
	Example: `# Output global mermaid to stdout
qst view my_knowledge_graph_dir/

# Output localized subgraph for a specific guide
qst view . --guide "Printf Function"

# Output localized subgraph specifying file name or path
qst view . --guide "Printf Function.md"
qst view -g examples/C_programming/Printf\ Function.md

# Save to a markdown file
qst view . > graph.md`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		} else if viewGuide != "" {
			guideDir := filepath.Dir(viewGuide)
			if guideDir != "." && guideDir != "" {
				dir = guideDir
			}
		}
		if viewGuide != "" {
			viewGuide = cleanGuideName(viewGuide)
		}

		guides, err := graph.ParseGraph(dir)
		if err != nil {
			cmd.PrintErrf("Error parsing graph: %v\n", err)
			exitFunc(1)
			return
		}

		if viewGuide != "" {
			if _, exists := guides[viewGuide]; !exists {
				cmd.PrintErrf("Error: guide '%s' not found in graph\n", viewGuide)
				exitFunc(1)
				return
			}
			config, err := graph.ParseConfig(dir)
			if err != nil {
				cmd.PrintErrf("Error parsing manifest: %v\n", err)
				exitFunc(1)
				return
			}
			diag, err := ssg.GenerateGuideMermaidText(viewGuide, guides, config)
			if err != nil {
				cmd.PrintErrf("Error generating localized diagram: %v\n", err)
				exitFunc(1)
				return
			}
			cmd.Println(diag)
			return
		}

		cmd.Println(generateMermaid(guides))
	},
}

func generateMermaid(guides map[string]graph.Guide) string {
	var sb strings.Builder
	sb.WriteString("```mermaid\ngraph TD\n")

	// Sort guides for deterministic output
	var ids []string
	for id := range guides {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	r := strings.NewReplacer("-", "_", " ", "_")

	// Declare all nodes first (ensures nodes with no connections are still visible)
	for _, id := range ids {
		cleanID := strings.ToLower(r.Replace(id))
		fmt.Fprintf(&sb, "    %s[\"%s\"]\n", cleanID, id)
	}

	// Declare all edges
	for _, id := range ids {
		guide := guides[id]
		cleanID := strings.ToLower(r.Replace(id))

		// Prerequisites (Solid links: Prereq -> Current)
		for _, prereq := range guide.Metadata.Prerequisites {
			cleanPrereq := strings.ToLower(r.Replace(prereq))
			fmt.Fprintf(&sb, "    %s --> %s\n", cleanPrereq, cleanID)
		}

		// Subguides (Dotted links: Current -> Sub)
		for _, sub := range guide.Metadata.SubGuides {
			cleanSub := strings.ToLower(r.Replace(sub.Guide))
			fmt.Fprintf(&sb, "    %s -.-> %s\n", cleanID, cleanSub)
		}
	}

	sb.WriteString("```")
	return sb.String()
}

func init() {
	viewCmd.Flags().StringVarP(&viewGuide, "guide", "g", "", "Export localized subgraph centered on a specific guide")
	rootCmd.AddCommand(viewCmd)
}
