package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var formOutput string

var formCmd = &cobra.Command{
	Use:   "form [directory]",
	Short: "zip a knowledge graph directory into the .kng archive file format",
	Long:  `Validates the formatting constraints of the knowledge graph and packages it losslessly into a compressed .kng archive.`,
	Example: `# From a directory
qst form my_knowledge

# From Current Directory
qst form

# Specify custom output path
qst form my_knowledge -o dist/bundle.kng`,
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

		absDir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("Error resolving path: %w", err)
		}

		outPath := formOutput
		if outPath == "" {
			baseName := filepath.Base(absDir)
			outPath = baseName + ".kng"
		}

		if err := graph.ArchiveGraph(dir, outPath); err != nil {
			return fmt.Errorf("Error archiving graph: %w", err)
		}

		cmd.Printf("Successfully packaged %d guides into %s\n", len(guides), outPath)
		return nil
	}),
}

func init() {
	formCmd.Flags().StringVarP(&formOutput, "output", "o", "", "Output path for the .kng archive (default: <directory-name>.kng)")
	rootCmd.AddCommand(formCmd)
}
