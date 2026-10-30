package cmd

import (
	"fmt"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var (
	unlinkInteractive bool
	unlinkDir         string
	unlinkSubguide    string
	unlinkPrereq      string
)

var unlinkCmd = &cobra.Command{
	Use:   "unlink [guide]",
	Short: "Unlink a subguide or prerequisite from an existing guide.",
	Long: `Unlink allows you to sever relationships between guides in the knowledge graph.
- Use --subguide (or --sub, -s) to remove a subguide relation.
- Use --prereq to remove a prerequisite relation.
- Use --interactive (-i) to choose from existing links to remove.`,
	Example: `# Unlink a subguide
qst unlink "The Compilation Process" --subguide "Compiler"

# Unlink specifying file names or paths
qst unlink "The Compilation Process.md" --subguide "Compiler.md"
qst unlink examples/C_programming/Scanf\ Function.md --prereq examples/C_programming/Printf\ Function.md

# Unlink a prerequisite
qst unlink "Scanf Function" --prereq "Printf Function"

# Interactive link removal
qst unlink "Variables and Assignment" -i`,
	Args: cobra.ExactArgs(1),
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		dir := unlinkDir
		guideName := resolveGuideAndDir(cmd, args[0], &dir)
		prereq := cleanGuideName(unlinkPrereq)
		subguide := cleanGuideName(unlinkSubguide)

		currentGuide, _, err := resolveGuide(dir, guideName)
		if err != nil {
			return fmt.Errorf("Error: %w", err)
		}

		isPrereqSet := cmd.Flags().Changed("prereq")
		isSubguideSet := cmd.Flags().Changed("subguide") || cmd.Flags().Changed("sub")

		if unlinkInteractive || (!isPrereqSet && !isSubguideSet) {
			done, err := promptUnlinkGuide(cmd, guideName, currentGuide, &isPrereqSet, &isSubguideSet, &prereq, &subguide)
			if err != nil || done {
				return err
			}
		}

		if isPrereqSet && prereq != "" {
			if err := graph.RemovePrerequisite(currentGuide.Path, prereq); err != nil {
				return fmt.Errorf("Error: %w", err)
			}
			cmd.Printf("Successfully unlinked prerequisite '%s' from '%s'\n", prereq, guideName)
			return nil
		}

		if isSubguideSet && subguide != "" {
			if err := graph.RemoveSubGuide(currentGuide.Path, subguide); err != nil {
				return fmt.Errorf("Error: %w", err)
			}
			cmd.Printf("Successfully unlinked subguide '%s' from '%s'\n", subguide, guideName)
			return nil
		}

		return fmt.Errorf("Error: must specify either --subguide or --prereq, or use -i for interactive mode.")
	}),
}

func init() {
	unlinkCmd.Flags().BoolVarP(&unlinkInteractive, "interactive", "i", false, "Interactive mode with guided selection")
	unlinkCmd.Flags().StringVarP(&unlinkDir, "dir", "d", ".", "Knowledge graph directory")
	unlinkCmd.Flags().StringVarP(&unlinkSubguide, "subguide", "s", "", "Name of the subguide to unlink")
	unlinkCmd.Flags().StringVar(&unlinkSubguide, "sub", "", "Alias for --subguide")
	unlinkCmd.Flags().StringVar(&unlinkPrereq, "prereq", "", "Name of the prerequisite guide to unlink")

	rootCmd.AddCommand(unlinkCmd)
}
