package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	huh "charm.land/huh/v2"
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

# Unlink a prerequisite
qst unlink "Scanf Function" --prereq "Printf Function"

# Interactive link removal
qst unlink "Variables and Assignment" -i`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		guideName := args[0]
		filename := filepath.Join(unlinkDir, guideName+".md")

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			cmd.PrintErrf("Error: guide '%s' does not exist at %s\n", guideName, filename)
			exitFunc(1)
			return
		}

		currentGuide, err := graph.ParseGuide(filename)
		if err != nil {
			cmd.PrintErrf("Failed to parse guide '%s': %v\n", guideName, err)
			exitFunc(1)
			return
		}

		isPrereqSet := cmd.Flags().Changed("prereq")
		isSubguideSet := cmd.Flags().Changed("subguide") || cmd.Flags().Changed("sub")

		// Interactive mode
		if unlinkInteractive || (!isPrereqSet && !isSubguideSet) {
			var linkOptions []huh.Option[string]

			for _, p := range currentGuide.Metadata.Prerequisites {
				label := fmt.Sprintf("[Prerequisite] %s", p)
				val := "prereq:" + p
				linkOptions = append(linkOptions, huh.NewOption(label, val))
			}

			for _, s := range currentGuide.Metadata.SubGuides {
				label := fmt.Sprintf("[Subguide] %s (adherence: %s, segment: %s)", s.Guide, s.Adherence, s.Segment)
				val := "subguide:" + s.Guide
				linkOptions = append(linkOptions, huh.NewOption(label, val))
			}

			if len(linkOptions) == 0 {
				cmd.Printf("Guide '%s' has no prerequisites or subguides to unlink.\n", guideName)
				return
			}

			selected := ""
			err := huh.NewSelect[string]().
				Title("Select link to remove").
				Options(linkOptions...).
				Value(&selected).
				Run()
			if err != nil {
				cmd.PrintErrf("Unlink aborted: %v\n", err)
				exitFunc(1)
				return
			}

			if strings.HasPrefix(selected, "prereq:") {
				unlinkPrereq = strings.TrimPrefix(selected, "prereq:")
				isPrereqSet = true
			} else if strings.HasPrefix(selected, "subguide:") {
				unlinkSubguide = strings.TrimPrefix(selected, "subguide:")
				isSubguideSet = true
			}
		}

		if isPrereqSet && unlinkPrereq != "" {
			if err := graph.RemovePrerequisite(filename, unlinkPrereq); err != nil {
				cmd.PrintErrf("Error: %v\n", err)
				exitFunc(1)
				return
			}
			cmd.Printf("Successfully unlinked prerequisite '%s' from '%s'\n", unlinkPrereq, guideName)
			return
		}

		if isSubguideSet && unlinkSubguide != "" {
			if err := graph.RemoveSubGuide(filename, unlinkSubguide); err != nil {
				cmd.PrintErrf("Error: %v\n", err)
				exitFunc(1)
				return
			}
			cmd.Printf("Successfully unlinked subguide '%s' from '%s'\n", unlinkSubguide, guideName)
			return
		}

		cmd.PrintErrln("Error: must specify either --subguide or --prereq, or use -i for interactive mode.")
		exitFunc(1)
	},
}

func init() {
	unlinkCmd.Flags().BoolVarP(&unlinkInteractive, "interactive", "i", false, "Interactive mode with guided selection")
	unlinkCmd.Flags().StringVarP(&unlinkDir, "dir", "d", ".", "Knowledge graph directory")
	unlinkCmd.Flags().StringVarP(&unlinkSubguide, "subguide", "s", "", "Name of the subguide to unlink")
	unlinkCmd.Flags().StringVar(&unlinkSubguide, "sub", "", "Alias for --subguide")
	unlinkCmd.Flags().StringVar(&unlinkPrereq, "prereq", "", "Name of the prerequisite guide to unlink")

	rootCmd.AddCommand(unlinkCmd)
}
