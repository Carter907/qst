package cmd

import (
	"fmt"
	"os"

	"github.com/Carter907/qst/internal/ssg"
	"github.com/spf13/cobra"
)

var buildOut string

var buildCmd = &cobra.Command{
	Use:   "build [directory]",
	Short: "Compile the knowledge graph into a static HTML website",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		fmt.Printf("Building static site from '%s' to '%s'...\n", dir, buildOut)
		err := ssg.BuildSite(dir, buildOut)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building site: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Build complete!")
	},
}

func init() {
	buildCmd.Flags().StringVarP(&buildOut, "out", "o", "public", "Output directory for the compiled site")
	rootCmd.AddCommand(buildCmd)
}
