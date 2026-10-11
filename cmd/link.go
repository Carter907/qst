package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	huh "charm.land/huh/v2"
	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var (
	linkInteractive bool
	linkDir         string
	linkGuide       string
	linkPrereq      string
	linkAdherence   string
	linkSegment     string
)

var linkCmd = &cobra.Command{
	Use:   "link [guide]",
	Short: "Link a subguide or prerequisite to an existing guide.",
	Long: `Link allows you to connect guides together in the knowledge graph.
- Use --guide to link a hierarchical subguide (with --adherence and optional --segment).
  If the subguide is already linked, its adherence and segment will be updated in place.
- Use --prereq to link a horizontal prerequisite of the exact same scope.
- Use --interactive (-i) for interactive guided selection with constraint checking.`,
	Example: `# Link a subguide with adherence and line segment
qst link "The Compilation Process" --guide "Compiler" --adherence introductory --segment 6-6

# Link a prerequisite of identical scope
qst link "Scanf Function" --prereq "Printf Function"

# Interactive link selection
qst link "Variables and Assignment" -i`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		guideName := args[0]
		filename := filepath.Join(linkDir, guideName+".md")

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			cmd.Printf("Error: guide '%s' does not exist at %s\n", guideName, filename)
			os.Exit(1)
		}

		guides, err := graph.ParseGraph(linkDir)
		if err != nil {
			cmd.Printf("Error parsing graph: %v\n", err)
			os.Exit(1)
		}

		config, err := graph.ParseConfig(linkDir)
		if err != nil {
			cmd.Printf("Failed to load manifest.yaml config: %v\n", err)
			os.Exit(1)
		}

		currentGuide, exists := guides[guideName]
		if !exists {
			cmd.Printf("Error: guide '%s' not found in graph\n", guideName)
			os.Exit(1)
		}

		isPrereqSet := cmd.Flags().Changed("prereq")
		isGuideSet := cmd.Flags().Changed("guide")

		// Interactive mode
		if linkInteractive || (!isPrereqSet && !isGuideSet) {
			linkType := "subguide"
			if !isPrereqSet && !isGuideSet {
				err := huh.NewSelect[string]().
					Title("Choose relationship type to link").
					Options(
						huh.NewOption("Subguide (hierarchical child)", "subguide"),
						huh.NewOption("Prerequisite (horizontal dependency)", "prereq"),
					).
					Value(&linkType).
					Run()
				if err != nil {
					cmd.Printf("Link aborted: %v\n", err)
					os.Exit(1)
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
						cmd.Printf("No available guides with matching scope '%s' to link as prerequisite.\n", currentScope)
						os.Exit(1)
					}

					err = huh.NewSelect[string]().
						Title("Select prerequisite guide").
						Options(candidateOptions...).
						Value(&linkPrereq).
						Run()
					if err != nil {
						cmd.Printf("Link aborted: %v\n", err)
						os.Exit(1)
					}
					isPrereqSet = true
				}
			} else {
				// Subguide
				if !isGuideSet {
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
						candidateOptions = append(candidateOptions, huh.NewOption(fmt.Sprintf("%s (%s)", c, guides[c].Metadata.Scope), c))
					}

					if len(candidateOptions) == 0 {
						cmd.Printf("No available guides with valid lower scope to link as subguide under '%s'.\n", guideName)
						os.Exit(1)
					}

					err = huh.NewSelect[string]().
						Title("Select subguide to link").
						Options(candidateOptions...).
						Value(&linkGuide).
						Run()
					if err != nil {
						cmd.Printf("Link aborted: %v\n", err)
						os.Exit(1)
					}
					isGuideSet = true
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
						cmd.Printf("Link aborted: %v\n", err)
						os.Exit(1)
					}
				}

				if !cmd.Flags().Changed("segment") {
					err = huh.NewInput().
						Title("Enter segment range in body (optional)").
						Placeholder("e.g. 4-4 or 10-20").
						Value(&linkSegment).
						Run()
					if err != nil {
						cmd.Printf("Link aborted: %v\n", err)
						os.Exit(1)
					}
				}
			}
		}

		// Handle Prerequisite
		if isPrereqSet && linkPrereq != "" {
			if err := graph.ValidatePrerequisiteLink(guides, config, guideName, linkPrereq); err != nil {
				cmd.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			if err := graph.AddPrerequisite(filename, linkPrereq); err != nil {
				cmd.Printf("Failed to update guide: %v\n", err)
				os.Exit(1)
			}

			cmd.Printf("Successfully linked prerequisite '%s' to '%s'\n", linkPrereq, guideName)
			return
		}

		// Handle Subguide
		if isGuideSet && linkGuide != "" {
			if linkAdherence == "" && len(config.Adherences) > 0 {
				linkAdherence = config.Adherences[0]
			}

			if err := graph.ValidateSubGuideLink(guides, config, guideName, linkGuide, linkAdherence, linkSegment); err != nil {
				cmd.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			newSubguide := graph.SubGuideRelation{
				Guide:     linkGuide,
				Adherence: linkAdherence,
				Segment:   linkSegment,
			}

			if err := graph.UpsertSubGuide(filename, newSubguide); err != nil {
				cmd.Printf("Failed to update guide: %v\n", err)
				os.Exit(1)
			}

			cmd.Printf("Successfully linked subguide '%s' to '%s'\n", linkGuide, guideName)
			return
		}

		cmd.Println("Error: must specify either --guide or --prereq, or use -i for interactive mode.")
		os.Exit(1)
	},
}

func init() {
	linkCmd.Flags().BoolVarP(&linkInteractive, "interactive", "i", false, "Interactive mode with guided selection")
	linkCmd.Flags().StringVarP(&linkDir, "dir", "d", ".", "Knowledge graph directory")
	linkCmd.Flags().StringVar(&linkGuide, "guide", "", "Name of the subguide to link")
	linkCmd.Flags().StringVar(&linkPrereq, "prereq", "", "Name of the prerequisite guide to link")
	linkCmd.Flags().StringVar(&linkAdherence, "adherence", "", "Adherence of the subguide relation")
	linkCmd.Flags().StringVar(&linkSegment, "segment", "", "Segment line range in body (e.g. 4-4)")

	rootCmd.AddCommand(linkCmd)
}
