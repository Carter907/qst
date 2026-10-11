package cmd

import (
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

		absDir, err := filepath.Abs(dir)
		if err != nil {
			cmd.PrintErrf("Error resolving path: %v\n", err)
			exitFunc(1)
			return
		}

		outPath := formOutput
		if outPath == "" {
			baseName := filepath.Base(absDir)
			outPath = baseName + ".kng"
		}

		err = graph.ArchiveGraph(dir, outPath)
		if err != nil {
			cmd.PrintErrf("Error archiving graph: %v\n", err)
			exitFunc(1)
			return
		}

		cmd.Printf("Successfully packaged %d guides into %s\n", len(guides), outPath)
	},
}

func init() {
	formCmd.Flags().StringVarP(&formOutput, "output", "o", "", "Output path for the .kng archive (default: <directory-name>.kng)")
	rootCmd.AddCommand(formCmd)
}
