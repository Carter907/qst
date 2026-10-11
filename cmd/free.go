package cmd

import (
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
	Run: func(cmd *cobra.Command, args []string) {
		inputKng := args[0]
		ext := filepath.Ext(inputKng)
		if ext != ".kng" {
			cmd.PrintErrln("Missing required file extension (.kng)")
			exitFunc(1)
			return
		}

		destDir := freeOutput
		if destDir == "" && len(args) > 1 {
			destDir = args[1]
		}
		if destDir == "" {
			filename := filepath.Base(inputKng)
			destDir = strings.TrimSuffix(filename, ext)
		}

		err := os.MkdirAll(destDir, 0o755)
		if err != nil {
			cmd.PrintErrf("Error creating directory: %v\n", err)
			exitFunc(1)
			return
		}

		err = graph.UnarchiveGraph(inputKng, destDir)
		if err != nil {
			cmd.PrintErrf("Error unarchiving graph: %v\n", err)
			exitFunc(1)
			return
		}

		cmd.Printf("Successfully unarchived to %s\n", destDir)
	},
}

func init() {
	freeCmd.Flags().StringVarP(&freeOutput, "output", "o", "", "Destination directory for unpacked graph (default: <archive-name>)")
	rootCmd.AddCommand(freeCmd)
}
