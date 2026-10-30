package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var (
	createStarter     bool
	newStrictCoverage bool
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "create a new knowledge graph directory",
	Long:  "You start a new knowledge graph with the new command. You only have to specify the name of the knowledge graph; a directory will be created. If you don't specify a directory, the current directory will be used. You can optionally include a starter guide template.",
	Example: `# Specify New Directory
qst new learn-cpp

# Use Current Directory
qst new

# Create with starter guide
qst new --starter`,
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) != 0 {
			dir = args[0]
		}

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("Failed to create knowledge graph: %w", err)
		}

		if createStarter {
			starterPath := filepath.Join(dir, "starter.md")
			if _, err := os.Stat(starterPath); os.IsNotExist(err) {
				if err := graph.WriteStarterGuide(starterPath); err != nil {
					return fmt.Errorf("Failed to create starter guide: %w", err)
				}
			} else {
				cmd.Println("Sample guide already exists, skipping write...")
			}
		}

		metaFilePath := filepath.Join(dir, "manifest.yaml")
		if _, err := os.Stat(metaFilePath); os.IsNotExist(err) {
			if err := graph.UpdateConfig(dir, graph.DefaultManifest(newStrictCoverage)); err != nil {
				return fmt.Errorf("Failed to write metafile: %w", err)
			}
		} else {
			cmd.Println("Metafile already exists, skipping write...")
		}
		cmd.Printf("Initialized knowledge graph in %s\n", dir)
		return nil
	}),
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().BoolVarP(&createStarter, "starter", "s", false, "Initialize with a starter markdown guide")
	newCmd.Flags().BoolVar(&newStrictCoverage, "strict-coverage", false, "Enable strict coverage mode in manifest.yaml")
}
