package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [directory]",
	Short: "Validate the knowledge directory",
	Long:  `Validate is useful for checking if your knowledge directory correctly adheres to the formatting and structural constraints.`,
	Example: `# Validate Current Dir
qst validate

# Specify a Path
qst validate my_knowledge_graph_dir/`,
	Args: cobra.MaximumNArgs(1),
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		guides, err := graph.ParseGraph(dir)
		if err != nil {
			return fmt.Errorf("Error parsing graph: %w", err)
		}

		config, err := graph.ParseConfig(dir)
		if err != nil {
			return fmt.Errorf("Error loading config: %w", err)
		}

		if err := graph.ValidateGraph(guides, config); err != nil {
			return fmt.Errorf("Validation failed:\n%w", err)
		}

		if _, err := filepath.Abs(dir); err != nil {
			return fmt.Errorf("Error resolving path: %w", err)
		}

		cmd.Printf("Knowledge graph parses and is valid with %d guides in %s\n", len(guides), dir)
		return nil
	}),
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
