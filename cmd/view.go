package cmd

import (
	"fmt"
	"path/filepath"

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
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		dir := "."
		targetGuide := viewGuide
		if len(args) > 0 {
			dir = args[0]
		} else if targetGuide != "" {
			guideDir := filepath.Dir(targetGuide)
			if guideDir != "." && guideDir != "" {
				dir = guideDir
			}
		}
		if targetGuide != "" {
			targetGuide = cleanGuideName(targetGuide)
		}

		guides, err := graph.ParseGraph(dir)
		if err != nil {
			return fmt.Errorf("Error parsing graph: %w", err)
		}

		if targetGuide != "" {
			if _, exists := guides[targetGuide]; !exists {
				return fmt.Errorf("Error: guide '%s' not found in graph", targetGuide)
			}
			config, err := graph.ParseConfig(dir)
			if err != nil {
				return fmt.Errorf("Error parsing manifest: %w", err)
			}
			diag, err := ssg.GenerateGuideMermaidText(targetGuide, guides, config)
			if err != nil {
				return fmt.Errorf("Error generating localized diagram: %w", err)
			}
			cmd.Println(diag)
			return nil
		}

		cmd.Println(ssg.GenerateGlobalMermaid(guides))
		return nil
	}),
}

func init() {
	viewCmd.Flags().StringVarP(&viewGuide, "guide", "g", "", "Export localized subgraph centered on a specific guide")
	rootCmd.AddCommand(viewCmd)
}
