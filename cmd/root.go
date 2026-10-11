/*
Package cmd

specifies base commands and flags for quest cli
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var exitFunc = os.Exit

var rootCmd = &cobra.Command{
	Use:   "qst",
	Short: "A lossless compression and formatting tool for knowledge",
	Long:  `qst can be used to initialize knowledge graph directories, update unformatted markdown files, and zip/unzip .kng files.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func Root() *cobra.Command {
	return rootCmd
}

func init() {
}
