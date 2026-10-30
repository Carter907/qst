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

// runWithExit wraps a command handler returning an error so any returned error is printed
// to the command's standard error channel and triggers exitFunc(1).
func runWithExit(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if err := fn(cmd, args); err != nil {
			cmd.PrintErrln(err)
			exitFunc(1)
		}
	}
}

func init() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
}
