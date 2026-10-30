package cmd

import (
	"fmt"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var renameDir string

var renameCmd = &cobra.Command{
	Use:   "rename [old_name] [new_name]",
	Short: "Rename a guide and update all references to it",
	Example: `# Rename a guide in the graph
qst rename "Old Name" "New Name" -d my_graph/

# Rename specifying file names or paths
qst rename "Old Name.md" "New Name.md"
qst rename my_graph/Old\ Name.md "New Name"`,
	Args: cobra.ExactArgs(2),
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		dir := renameDir
		oldName := resolveGuideAndDir(cmd, args[0], &dir)
		newName := cleanGuideName(args[1])

		updatedCount, err := graph.RenameGuide(dir, oldName, newName)
		if err != nil {
			return fmt.Errorf("Error: %w", err)
		}

		cmd.Printf("Successfully renamed '%s' to '%s' and updated %d references.\n", oldName, newName, updatedCount)
		return nil
	}),
}

func init() {
	renameCmd.Flags().StringVarP(&renameDir, "dir", "d", ".", "Knowledge graph directory")
	rootCmd.AddCommand(renameCmd)
}
