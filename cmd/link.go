package cmd

import (
	"path/filepath"
	"sort"
	"strings"

	huh "charm.land/huh/v2"
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

# Link multiple subguides at once
qst link "Chapter 13 Strings" --subguide "String Fundamentals, String IO" --adherence introductory

# Link a prerequisite of identical scope
qst link "Scanf Function" --prereq "Printf Function"

# Preview body lines before linking
qst link "Variables and Assignment" --preview

# Interactive link selection
qst link "Variables and Assignment" -i`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		guideName := args[0]
		filename := filepath.Join(linkDir, guideName+".md")

		guides, err := graph.ParseGraph(linkDir)
		if err != nil {
			cmd.PrintErrf("Error parsing graph: %v\n", err)
			exitFunc(1)
			return
		}

		currentGuide, exists := guides[guideName]
		if !exists {
			cmd.PrintErrf("Error: guide '%s' not found in graph\n", guideName)
			exitFunc(1)
			return
		}

		config, err := graph.ParseConfig(linkDir)
		if err != nil {
			cmd.PrintErrf("Failed to load manifest.yaml config: %v\n", err)
			exitFunc(1)
			return
		}

		defer func() {
			linkPrereqs = nil
			linkSubguides = nil
		}()

		normalizeSlice := func(items []string) []string {
			var res []string
			for _, item := range items {
				item = strings.TrimPrefix(item, "[")
				item = strings.TrimSuffix(item, "]")
				for _, part := range strings.Split(item, ",") {
					if trimmed := strings.Trim(part, "[] \t"); trimmed != "" {
						res = append(res, trimmed)
					}
				}
			}
			return res
		}

		isPrereqSet := cmd.Flags().Changed("prereq")
		isSubguideSet := cmd.Flags().Changed("subguide") || cmd.Flags().Changed("sub")

		// Show preview if requested
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
				return
			}
		}

		// Interactive mode
		if linkInteractive || (!isPrereqSet && !isSubguideSet) {
			linkType := "subguide"
			if !isPrereqSet && !isSubguideSet {
				err := huh.NewSelect[string]().
					Title("Choose relationship type to link").
					Options(
						huh.NewOption("Subguide (hierarchical child)", "subguide"),
						huh.NewOption("Prerequisite (horizontal dependency)", "prereq"),
					).
					Value(&linkType).
					Run()
				if err != nil {
					cmd.PrintErrf("Link aborted: %v\n", err)
					exitFunc(1)
					return
				}
			} else if isPrereqSet {
				linkType = "prereq"
			}

			if linkType == "prereq" {
				if !isPrereqSet {
					var candidateOptions []huh.Option[string]
					var sortedCandidates []string
					currentScope := currentGuide.Metadata.Scope

					for id, g := range guides {
						if id == guideName || g.Metadata.Scope != currentScope {
							continue
						}
						alreadyPrereq := false
						for _, p := range currentGuide.Metadata.Prerequisites {
							if p == id {
								alreadyPrereq = true
								break
							}
						}
						if !alreadyPrereq {
							sortedCandidates = append(sortedCandidates, id)
						}
					}
					sort.Strings(sortedCandidates)
					for _, c := range sortedCandidates {
						candidateOptions = append(candidateOptions, huh.NewOption(c, c))
					}

					if len(candidateOptions) == 0 {
						cmd.PrintErrf("No available guides with matching scope '%s' to link as prerequisite.\n", currentScope)
						exitFunc(1)
						return
					}

					selectedPrereq := ""
					err = huh.NewSelect[string]().
						Title("Select prerequisite guide").
						Options(candidateOptions...).
						Value(&selectedPrereq).
						Run()
					if err != nil {
						cmd.PrintErrf("Link aborted: %v\n", err)
						exitFunc(1)
						return
					}
					linkPrereqs = []string{selectedPrereq}
					isPrereqSet = true
				}
			} else {
				// Subguide
				if !isSubguideSet {
					parentScopeVal := graph.GetScopeValue(currentGuide.Metadata.Scope, config)
					var candidateOptions []huh.Option[string]
					var sortedCandidates []string

					for id, g := range guides {
						if id == guideName {
							continue
						}
						subScopeVal := graph.GetScopeValue(g.Metadata.Scope, config)
						if subScopeVal >= parentScopeVal {
							continue
						}
						if !config.RelaxedSubguides && parentScopeVal-subScopeVal != 1 {
							continue
						}
						sortedCandidates = append(sortedCandidates, id)
					}
					sort.Strings(sortedCandidates)
					for _, c := range sortedCandidates {
						candidateOptions = append(candidateOptions, huh.NewOption(cmdFormatOption(c, guides[c].Metadata.Scope), c))
					}

					if len(candidateOptions) == 0 {
						cmd.PrintErrf("No available guides with valid lower scope to link as subguide under '%s'.\n", guideName)
						exitFunc(1)
						return
					}

					selectedSubguide := ""
					err = huh.NewSelect[string]().
						Title("Select subguide to link").
						Options(candidateOptions...).
						Value(&selectedSubguide).
						Run()
					if err != nil {
						cmd.PrintErrf("Link aborted: %v\n", err)
						exitFunc(1)
						return
					}
					linkSubguides = []string{selectedSubguide}
					isSubguideSet = true
				}

				if !cmd.Flags().Changed("adherence") {
					var adherenceOptions []huh.Option[string]
					for _, c := range config.Adherences {
						adherenceOptions = append(adherenceOptions, huh.NewOption(c, c))
					}
					err = huh.NewSelect[string]().
						Title("Choose adherence").
						Options(adherenceOptions...).
						Value(&linkAdherence).
						Run()
					if err != nil {
						cmd.PrintErrf("Link aborted: %v\n", err)
						exitFunc(1)
						return
					}
				}

				if !cmd.Flags().Changed("segment") {
					// Display body preview to assist with segment input
					cmd.Println("\nBody Lines Preview:")
					bodyLines := strings.Split(currentGuide.Body, "\n")
					for i, line := range bodyLines {
						cmd.Printf("[L%3d] %s\n", i+1, line)
					}
					cmd.Println()

					err = huh.NewInput().
						Title("Enter segment range in body (optional)").
						Placeholder("e.g. 4-4 or 10-20").
						Value(&linkSegment).
						Run()
					if err != nil {
						cmd.PrintErrf("Link aborted: %v\n", err)
						exitFunc(1)
						return
					}
				}
			}
		}

		// Handle Prerequisites (batch support)
		if isPrereqSet {
			prereqs := normalizeSlice(linkPrereqs)
			if len(prereqs) == 0 {
				cmd.PrintErrln("Error: no prerequisite specified")
				exitFunc(1)
				return
			}
			for _, prereq := range prereqs {
				if err := graph.ValidatePrerequisiteLink(guides, config, guideName, prereq); err != nil {
					cmd.PrintErrf("Error: %v\n", err)
					exitFunc(1)
					return
				}

				if err := graph.AddPrerequisite(filename, prereq); err != nil {
					cmd.PrintErrf("Failed to update guide: %v\n", err)
					exitFunc(1)
					return
				}

				// Update local guide state for subsequent validations
				g := guides[guideName]
				g.Metadata.Prerequisites = append(g.Metadata.Prerequisites, prereq)
				guides[guideName] = g

				cmd.Printf("Successfully linked prerequisite '%s' to '%s'\n", prereq, guideName)
			}
			return
		}

		// Handle Subguides (batch support)
		if isSubguideSet {
			subguides := normalizeSlice(linkSubguides)
			if len(subguides) == 0 {
				cmd.PrintErrln("Error: no subguide specified")
				exitFunc(1)
				return
			}
			if linkAdherence == "" && len(config.Adherences) > 0 {
				linkAdherence = config.Adherences[0]
			}

			for _, sub := range subguides {
				if err := graph.ValidateSubGuideLink(guides, config, guideName, sub, linkAdherence, linkSegment); err != nil {
					cmd.PrintErrf("Error: %v\n", err)
					exitFunc(1)
					return
				}

				newSubguide := graph.SubGuideRelation{
					Guide:     sub,
					Adherence: linkAdherence,
					Segment:   linkSegment,
				}

				if err := graph.UpsertSubGuide(filename, newSubguide); err != nil {
					cmd.PrintErrf("Failed to update guide: %v\n", err)
					exitFunc(1)
					return
				}

				// Update local guide state for subsequent validations
				g := guides[guideName]
				g.Metadata.SubGuides = append(g.Metadata.SubGuides, newSubguide)
				guides[guideName] = g

				cmd.Printf("Successfully linked subguide '%s' to '%s'\n", sub, guideName)
			}
			return
		}

		cmd.PrintErrln("Error: must specify either --subguide or --prereq, or use -i for interactive mode.")
		exitFunc(1)
	},
}

func cmdFormatOption(id, scope string) string {
	return id + " (" + scope + ")"
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
