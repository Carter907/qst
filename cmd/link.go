package cmd

import (
	"fmt"
	"strings"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var (
	linkInteractive bool
	linkDir         string
	linkSubguides   []string
	linkPrereqs     []string
	linkAdherence   string
	linkSegment     string
	linkPreview     bool
)

var linkCmd = &cobra.Command{
	Use:   "link [guide]",
	Short: "Link a subguide or prerequisite to an existing guide.",
	Long: `Link allows you to connect guides together in the knowledge graph.
- Use --subguide (or --sub, -s) to link one or more hierarchical subguides (with --adherence and optional --segment).
  If a subguide is already linked, its adherence and segment will be updated in place.
- Use --prereq to link one or more horizontal prerequisites of the exact same scope.
- Use --preview (-p) to preview numbered body lines for accurate segment range selection.
- Use --interactive (-i) for interactive guided selection with constraint checking.`,
	Example: `# Link a subguide with adherence and line segment
qst link "The Compilation Process" --subguide "Compiler" --adherence introductory --segment 6-6

# Link specifying file names or paths
qst link "The Compilation Process.md" --subguide "Compiler.md" --adherence introductory
qst link examples/C_programming/Scanf\ Function.md --prereq examples/C_programming/Printf\ Function.md

# Link multiple subguides at once
qst link "Chapter 13 Strings" --subguide "String Fundamentals, String IO" --adherence introductory

# Link a prerequisite of identical scope
qst link "Scanf Function" --prereq "Printf Function"

# Preview body lines before linking
qst link "Variables and Assignment" --preview

# Interactive link selection
qst link "Variables and Assignment" -i`,
	Args: cobra.ExactArgs(1),
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		dir := linkDir
		guideName := resolveGuideAndDir(cmd, args[0], &dir)

		guides, err := graph.ParseGraph(dir)
		if err != nil {
			return fmt.Errorf("Error parsing graph: %w", err)
		}

		currentGuide, exists := guides[guideName]
		if !exists {
			return fmt.Errorf("Error: guide '%s' not found in graph", guideName)
		}

		config, err := graph.ParseConfig(dir)
		if err != nil {
			return fmt.Errorf("Failed to load manifest.yaml config: %w", err)
		}

		isPrereqSet := cmd.Flags().Changed("prereq")
		isSubguideSet := cmd.Flags().Changed("subguide") || cmd.Flags().Changed("sub")
		prereqs := append([]string(nil), linkPrereqs...)
		subguides := append([]string(nil), linkSubguides...)
		adherence := linkAdherence
		segment := linkSegment

		if linkPreview {
			cmd.Println("\n" + strings.Repeat("=", 80))
			cmd.Printf("Body Lines Preview: %s\n", guideName)
			cmd.Println(strings.Repeat("-", 80))
			bodyLines := strings.Split(currentGuide.Body, "\n")
			for i, line := range bodyLines {
				cmd.Printf("[L%3d] %s\n", i+1, line)
			}
			cmd.Println(strings.Repeat("-", 80))
			if !isPrereqSet && !isSubguideSet && !linkInteractive {
				return nil
			}
		}

		if linkInteractive || (!isPrereqSet && !isSubguideSet) {
			if err := promptLinkGuide(cmd, guideName, currentGuide, guides, config, &isPrereqSet, &isSubguideSet, &prereqs, &subguides, &adherence, &segment); err != nil {
				return err
			}
		}

		if isPrereqSet {
			parsedPrereqs := parseGuideCSVOrSlice(prereqs)
			if len(parsedPrereqs) == 0 {
				return fmt.Errorf("Error: no prerequisite specified")
			}
			for _, prereq := range parsedPrereqs {
				if err := graph.ValidatePrerequisiteLink(guides, config, guideName, prereq); err != nil {
					return fmt.Errorf("Error: %w", err)
				}
				if err := graph.AddPrerequisite(currentGuide.Path, prereq); err != nil {
					return fmt.Errorf("Failed to update guide: %w", err)
				}
				g := guides[guideName]
				g.Metadata.Prerequisites = append(g.Metadata.Prerequisites, prereq)
				guides[guideName] = g

				cmd.Printf("Successfully linked prerequisite '%s' to '%s'\n", prereq, guideName)
			}
			return nil
		}

		if isSubguideSet {
			parsedSubguides := parseGuideCSVOrSlice(subguides)
			if len(parsedSubguides) == 0 {
				return fmt.Errorf("Error: no subguide specified")
			}
			if adherence == "" && len(config.Adherences) > 0 {
				adherence = config.Adherences[0]
			}

			for _, sub := range parsedSubguides {
				if err := graph.ValidateSubGuideLink(guides, config, guideName, sub, adherence, segment); err != nil {
					return fmt.Errorf("Error: %w", err)
				}
				newSubguide := graph.SubGuideRelation{
					Guide:     sub,
					Adherence: adherence,
					Segment:   segment,
				}
				if err := graph.UpsertSubGuide(currentGuide.Path, newSubguide); err != nil {
					return fmt.Errorf("Failed to update guide: %w", err)
				}
				g := guides[guideName]
				g.Metadata.SubGuides = append(g.Metadata.SubGuides, newSubguide)
				guides[guideName] = g

				cmd.Printf("Successfully linked subguide '%s' to '%s'\n", sub, guideName)
			}
			return nil
		}

		return fmt.Errorf("Error: must specify either --subguide or --prereq, or use -i for interactive mode.")
	}),
}

func init() {
	linkCmd.Flags().BoolVarP(&linkInteractive, "interactive", "i", false, "Interactive mode with guided selection")
	linkCmd.Flags().StringVarP(&linkDir, "dir", "d", ".", "Knowledge graph directory")
	linkCmd.Flags().StringSliceVarP(&linkSubguides, "subguide", "s", nil, "Name(s) of the subguide(s) to link")
	linkCmd.Flags().StringSliceVar(&linkSubguides, "sub", nil, "Alias for --subguide")
	linkCmd.Flags().StringSliceVar(&linkPrereqs, "prereq", nil, "Name(s) of the prerequisite guide(s) to link")
	linkCmd.Flags().StringVar(&linkAdherence, "adherence", "", "Adherence of the subguide relation")
	linkCmd.Flags().StringVar(&linkSegment, "segment", "", "Segment line range in body (e.g. 4-4)")
	linkCmd.Flags().BoolVarP(&linkPreview, "preview", "p", false, "Preview numbered body lines of the guide")

	rootCmd.AddCommand(linkCmd)
}
