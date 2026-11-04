package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var (
	pathDir         string
	pathPrereqsOnly bool
	pathDirect      bool
	pathSubsequent  bool
	pathFormat      string
	pathSaveTour    string
	pathVerbose     bool
)

var pathCmd = &cobra.Command{
	Use:     "path [guide]",
	Aliases: []string{"learn", "roadmap"},
	Short:   "Generate a topological learning path for a guide or the entire graph",
	Long: `Path computes an optimal, prerequisite-satisfying sequence to learn a specific guide
or the entire knowledge graph using topological sorting.

By default, learning paths resolve both horizontal prerequisites (same scope) and
vertical subguide components (lower scopes), ensuring foundational building blocks
are mastered before high-level topics.

When no guide argument is supplied, a complete curriculum for the entire knowledge
graph is generated and grouped by manifest scopes.`,
	Example: `# Discover learning path for a specific guide in current directory
qst path "Variables and Assignment"

# Discover learning path in a specific knowledge graph directory
qst path "Chapter 2 C Fundamentals" -d examples/C_programming

# Output learning path as a Mermaid diagram
qst path "Variables and Assignment" -d examples/C_programming --format mermaid

# Discover subsequent guides dependent on a foundational guide
qst path "Variable" -d examples/C_programming --subsequent

# Generate full syllabus for the entire knowledge graph
qst path -d examples/C_programming

# Save the resulting sequence directly as an official tour in manifest.yaml
qst path "Chapter 2 C Fundamentals" -d examples/C_programming --save-tour "Chapter 2 Learning Path"`,
	Args: cobra.MaximumNArgs(1),
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		targetGuideName := ""
		if len(args) > 0 {
			targetGuideName = resolveGuideAndDir(cmd, args[0], &pathDir)
		}

		guides, err := graph.ParseGraph(pathDir)
		if err != nil {
			return fmt.Errorf("failed to parse guides: %w", err)
		}
		if len(guides) == 0 {
			return fmt.Errorf("no guides found in directory: %s", pathDir)
		}

		config, err := graph.ParseConfig(pathDir)
		if err != nil {
			return fmt.Errorf("failed to parse manifest.yaml: %w", err)
		}

		opts := graph.PathOptions{
			IncludeSubguides: !pathPrereqsOnly,
			PrereqsOnly:      pathPrereqsOnly,
			DirectOnly:       pathDirect,
		}

		var learningPath *graph.LearningPath
		if pathSubsequent {
			if targetGuideName == "" {
				return fmt.Errorf("a target guide name is required when using --subsequent")
			}
			resolved, _, err := resolveGuide(pathDir, targetGuideName)
			if err != nil {
				return fmt.Errorf("target guide '%s' not found: %w", targetGuideName, err)
			}
			learningPath, err = graph.SubsequentPath(resolved.ID, guides, config, opts)
			if err != nil {
				return fmt.Errorf("failed to compute subsequent path: %w", err)
			}
		} else if targetGuideName != "" {
			resolved, _, err := resolveGuide(pathDir, targetGuideName)
			if err != nil {
				return fmt.Errorf("target guide '%s' not found: %w", targetGuideName, err)
			}
			learningPath, err = graph.PrerequisitePath(resolved.ID, guides, config, opts)
			if err != nil {
				return fmt.Errorf("failed to compute learning path: %w", err)
			}
		} else {
			learningPath, err = graph.TopologicalSort(guides, config, opts)
			if err != nil {
				return fmt.Errorf("failed to topologically sort knowledge graph: %w", err)
			}
		}

		// Save tour if requested
		if pathSaveTour != "" {
			var tourGuides []string
			for _, step := range learningPath.Steps {
				tourGuides = append(tourGuides, step.GuideID)
			}
			err := graph.CreateTour(pathDir, graph.TourConfig{
				Name:   pathSaveTour,
				Guides: tourGuides,
			})
			if err != nil {
				return fmt.Errorf("failed to save tour '%s': %w", pathSaveTour, err)
			}
			cmd.Printf("Successfully created tour '%s' with %d guides in manifest.yaml\n\n", pathSaveTour, len(tourGuides))
		}

		switch strings.ToLower(pathFormat) {
		case "json":
			data, err := json.MarshalIndent(learningPath, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to serialize JSON output: %w", err)
			}
			cmd.Println(string(data))
			return nil

		case "mermaid":
			cmd.Println(generatePathMermaid(learningPath, guides))
			return nil

		case "text", "":
			printPathText(cmd, learningPath, config, targetGuideName, pathSubsequent, pathVerbose)
			return nil

		default:
			return fmt.Errorf("unsupported format '%s' (allowed: text, json, mermaid)", pathFormat)
		}
	}),
}

func printPathText(cmd *cobra.Command, path *graph.LearningPath, config graph.Manifest, target string, subsequent bool, verbose bool) {
	if len(path.Steps) == 0 {
		cmd.Println("No guides in learning path.")
		return
	}

	title := config.Title
	if title == "" {
		title = "Knowledge Graph"
	}

	if subsequent {
		cmd.Printf("Subsequent guides dependent on '%s' (%d guides):\n", target, len(path.Steps))
		cmd.Println(strings.Repeat("-", 60))
	} else if target != "" {
		cmd.Printf("Learning path for '%s' (%d guides):\n", target, len(path.Steps))
		cmd.Println(strings.Repeat("-", 60))
	} else {
		cmd.Printf("Topological learning path for '%s' (%d guides):\n", title, len(path.Steps))
		cmd.Println(strings.Repeat("=", 60))
	}

	// For global graph without target, display partitioned by scope groups
	if target == "" && len(path.Scopes) > 0 {
		for i, group := range path.Scopes {
			if i > 0 {
				cmd.Println()
			}
			cmd.Printf("Scope: %s (%d guides)\n", group.Scope, len(group.Guides))
			for _, step := range group.Guides {
				suffix := ""
				if verbose {
					var notes []string
					if len(step.Dependencies) > 0 {
						notes = append(notes, "requires: "+strings.Join(step.Dependencies, ", "))
					}
					if len(notes) > 0 {
						suffix = "  (" + strings.Join(notes, "; ") + ")"
					}
				}
				cmd.Printf("  %2d. %s%s\n", step.Index, step.GuideID, suffix)
			}
		}
		return
	}

	// Single target or subsequent path: print numbered table
	cmd.Printf("%-5s %-15s %s\n", "Step", "Scope", "Guide")
	for _, step := range path.Steps {
		scopeStr := step.Scope
		if scopeStr == "" {
			scopeStr = "-"
		}
		suffix := ""
		if verbose {
			var notes []string
			if len(step.Dependencies) > 0 {
				notes = append(notes, "requires: "+strings.Join(step.Dependencies, ", "))
			}
			if len(notes) > 0 {
				suffix = "  (" + strings.Join(notes, "; ") + ")"
			}
		}
		cmd.Printf("  %2d. %-15s %s%s\n", step.Index, scopeStr, step.GuideID, suffix)
	}
}

func generatePathMermaid(path *graph.LearningPath, guides map[string]graph.Guide) string {
	var sb strings.Builder
	sb.WriteString("flowchart TD\n")

	cleanID := func(raw string) string {
		var out strings.Builder
		for _, r := range raw {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				out.WriteRune(r)
			} else {
				out.WriteRune('_')
			}
		}
		return out.String()
	}

	pathSet := make(map[string]bool, len(path.Steps))
	for _, step := range path.Steps {
		pathSet[step.GuideID] = true
	}

	// Declare nodes
	for _, step := range path.Steps {
		nodeID := cleanID(step.GuideID)
		scopeLabel := step.Scope
		if scopeLabel != "" {
			scopeLabel = fmt.Sprintf("<br/><i>(%s)</i>", scopeLabel)
		}
		sb.WriteString(fmt.Sprintf("    %s[\"%s%s\"]\n", nodeID, step.GuideID, scopeLabel))
	}
	sb.WriteString("\n")

	// Declare edges
	for _, step := range path.Steps {
		currNodeID := cleanID(step.GuideID)
		g := guides[step.GuideID]

		for _, p := range g.Metadata.Prerequisites {
			if pathSet[p] {
				prereqNodeID := cleanID(p)
				sb.WriteString(fmt.Sprintf("    %s --> %s\n", prereqNodeID, currNodeID))
			}
		}
		for _, s := range g.Metadata.SubGuides {
			if pathSet[s.Guide] {
				subNodeID := cleanID(s.Guide)
				sb.WriteString(fmt.Sprintf("    %s -.-> %s\n", subNodeID, currNodeID))
			}
		}
	}

	return sb.String()
}

func init() {
	pathCmd.Flags().StringVarP(&pathDir, "dir", "d", ".", "Knowledge graph directory")
	pathCmd.Flags().BoolVar(&pathPrereqsOnly, "prereqs-only", false, "Only follow horizontal prerequisites, ignoring subguides")
	pathCmd.Flags().BoolVar(&pathDirect, "direct", false, "Only include immediate 1-hop prerequisites")
	pathCmd.Flags().BoolVar(&pathSubsequent, "subsequent", false, "Discover downstream guides dependent on the target")
	pathCmd.Flags().StringVar(&pathFormat, "format", "text", "Output format: text, json, mermaid")
	pathCmd.Flags().StringVar(&pathSaveTour, "save-tour", "", "Save the resulting path as a tour in manifest.yaml")
	pathCmd.Flags().BoolVarP(&pathVerbose, "verbose", "v", false, "Show dependency reasons and requirements per step")
	rootCmd.AddCommand(pathCmd)
}
