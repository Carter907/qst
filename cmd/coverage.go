package cmd

import (
	"github.com/spf13/cobra"
)

var coverageDir string

var coverageCmd = &cobra.Command{
	Use:     "coverage [guide]",
	Aliases: []string{"segment"},
	Short:   "Visualize line-by-line subguide coverage of a guide's body",
	Long: `Coverage displays the markdown body of a guide with 1-based line numbers prefixed.
Each line is annotated with the subguide that covers it, highlighting unmapped and overlapping lines.
This is a direct shortcut for 'qst inspect <guide> --coverage'.`,
	Example: `# Inspect body line coverage
qst coverage "Variables and Assignment"

# Inspect in specific directory
qst coverage "The Compilation Process" -d examples/C_programming`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runInspect(cmd, args[0], coverageDir, true)
	},
}

func init() {
	coverageCmd.Flags().StringVarP(&coverageDir, "dir", "d", ".", "Knowledge graph directory")
	rootCmd.AddCommand(coverageCmd)
}
