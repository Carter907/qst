package cmd

import (
	"fmt"
	"os"
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
			fmt.Printf("Error parsing graph: %v\n", err)
			os.Exit(0)
		}

		config, err := graph.ParseConfig(dir)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(0)
		}

		err = graph.ValidateGraph(guides, config)
		if err != nil {
			fmt.Printf("Validation failed:\n%v\n", err)
			os.Exit(0)
		}

		if _, err := filepath.Abs(dir); err != nil {
			fmt.Printf("Error resolving path: %v\n", err)
			os.Exit(0)
		}

		fmt.Printf("Knowledge graph parses and is valid with %d guides in %s\n", len(guides), dir)
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
