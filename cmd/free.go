package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var freeCmd = &cobra.Command{
	Use:   "free",
	Short: "Unzip a .kng file so you can read or edit the knowledge graph",
	Long:  "Free is how you start learning new knowledge. Unzipped archives are placed in a directory with the same name as the file. All markdown guides are losslessly decompressed.",
	Example: `# Free the Knowledge
qst free my-graph.kng`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please specify a .kng file to unpack")
			os.Exit(1)
		}
		inputKng := args[0]
		ext := filepath.Ext(inputKng)
		if ext != ".kng" {
			fmt.Println("Missing required file extension (.kng)")
			os.Exit(1)
		}

		filename := filepath.Base(inputKng)
		destDir := strings.TrimSuffix(filename, ext)

		err := os.MkdirAll(destDir, 0o755)
		if err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
			os.Exit(1)
		}

		err = graph.UnarchiveGraph(inputKng, destDir)
		if err != nil {
			fmt.Printf("Error unarchiving graph: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully unarchived to %s\n", destDir)
	},
}

func init() {
	rootCmd.AddCommand(freeCmd)
}
