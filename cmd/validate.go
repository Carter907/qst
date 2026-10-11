package cmd

import (
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
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		guides, err := graph.ParseGraph(dir)
		if err != nil {
			cmd.PrintErrf("Error parsing graph: %v\n", err)
			exitFunc(1)
			return
		}

		config, err := graph.ParseConfig(dir)
		if err != nil {
			cmd.PrintErrf("Error loading config: %v\n", err)
			exitFunc(1)
			return
		}

		err = graph.ValidateGraph(guides, config)
		if err != nil {
			cmd.PrintErrf("Validation failed:\n%v\n", err)
			exitFunc(1)
			return
		}

		if _, err := filepath.Abs(dir); err != nil {
			cmd.PrintErrf("Error resolving path: %v\n", err)
			exitFunc(1)
			return
		}

		cmd.Printf("Knowledge graph parses and is valid with %d guides in %s\n", len(guides), dir)
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
