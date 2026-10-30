package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var freeOutput string

var freeCmd = &cobra.Command{
	Use:   "free <archive.kng> [destination]",
	Short: "Unzip a .kng file so you can read or edit the knowledge graph",
	Long:  "Free is how you start learning new knowledge. Unzipped archives are placed in a directory with the same name as the file (or a specified destination). All markdown guides, manifests, and assets are losslessly decompressed.",
	Example: `# Free the Knowledge to default directory
qst free my-graph.kng

# Free to a custom destination directory
qst free my-graph.kng custom_dir
qst free my-graph.kng -o custom_dir`,
	Args: cobra.RangeArgs(1, 2),
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		inputKng := args[0]
		ext := filepath.Ext(inputKng)
		if ext != ".kng" {
			return fmt.Errorf("Missing required file extension (.kng)")
		}

		destDir := freeOutput
		if destDir == "" && len(args) > 1 {
			destDir = args[1]
		}
		if destDir == "" {
			filename := filepath.Base(inputKng)
			destDir = strings.TrimSuffix(filename, ext)
		}

		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return fmt.Errorf("Error creating directory: %w", err)
		}

		if err := graph.UnarchiveGraph(inputKng, destDir); err != nil {
			return fmt.Errorf("Error unarchiving graph: %w", err)
		}

		cmd.Printf("Successfully unarchived to %s\n", destDir)
		return nil
	}),
}

func init() {
	freeCmd.Flags().StringVarP(&freeOutput, "output", "o", "", "Destination directory for unpacked graph (default: <archive-name>)")
	rootCmd.AddCommand(freeCmd)
}
