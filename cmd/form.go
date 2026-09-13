package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var formCmd = &cobra.Command{
	Use:   "form [directory]",
	Short: "zip a knowledge graph directory into the .kng archive file format",
	Long:  `Validates the formatting constraints of the knowledge graph and packages it losslessly.`,
	Example: `# From a directory
qst form my_knowledge

# From Current Directory
qst form`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		guides, err := graph.ParseGraph(dir)
		if err != nil {
			fmt.Printf("Error parsing graph: %v\n", err)
			os.Exit(1)
		}

		config, err := graph.ParseConfig(dir)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		err = graph.ValidateGraph(guides, config)
		if err != nil {
			fmt.Printf("Validation failed:\n%v\n", err)
			os.Exit(1)
		}

		absDir, err := filepath.Abs(dir)
		if err != nil {
			fmt.Printf("Error resolving path: %v\n", err)
			os.Exit(1)
		}

		baseName := filepath.Base(absDir)
		outPath := baseName + ".kng"

		err = graph.ArchiveGraph(dir, outPath)
		if err != nil {
			fmt.Printf("Error archiving graph: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully packaged %d guides into %s\n", len(guides), outPath)
	},
}

func init() {
	rootCmd.AddCommand(formCmd)
}
