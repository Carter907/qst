package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var (
	interactive      bool
	addDir           string
	addPrerequisites []string
	addScope         string
	addTags          []string
	addSubguides     []string
	addAdherence     string
	addSegment       string
	addBody          string
	addFile          string
)

var addCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new guide to the knowledge graph.",
	Long:  "Add allows you to insert a guide into the knowledge graph by specifying its prerequisites, subguides, and scope. A new markdown file will be inserted into the target directory.",
	Example: `# Add a new definition
qst add Exponent --scope definition

# Add a description specifying file name or path
qst add "String Operations.md" --scope description --subguide "String Length.md, String Copy.md" --adherence introductory

# Add to a specific directory with initial body content
qst add "Memory Layout" --dir my_graph/ --scope explanation --body "## Memory Layout\n\nExplanation text."`,
	Args: cobra.ExactArgs(1),
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		dir := addDir
		name := resolveGuideAndDir(cmd, args[0], &dir)
		filename := filepath.Join(dir, name+".md")

		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			return fmt.Errorf("Error: %s already exists", filename)
		}
		config, err := graph.ParseConfig(dir)
		if err != nil {
			return fmt.Errorf("Failed to load manifest.yaml config: %w", err)
		}

		scope := addScope
		prereqs := parseGuideCSVOrSlice(addPrerequisites)
		subguides := parseGuideCSVOrSlice(addSubguides)
		tags := parseCSVOrSlice(addTags)
		adherence := addAdherence
		segment := addSegment

		if interactive {
			if err := promptAddGuide(cmd, config, &scope, &prereqs, &tags); err != nil {
				return err
			}
		} else if !cmd.Flags().Changed("scope") || scope == "" {
			return fmt.Errorf("Error: --scope is required in non-interactive mode")
		}

		if !slices.Contains(config.Scopes, scope) {
			return fmt.Errorf("Error: invalid scope '%s'. Valid scopes are: %v", scope, config.Scopes)
		}

		var guides map[string]graph.Guide
		if len(prereqs) > 0 || len(subguides) > 0 {
			guides, err = graph.ParseGraph(dir)
			if err != nil {
				return fmt.Errorf("Error parsing graph: %w", err)
			}
		}

		for _, prereq := range prereqs {
			p, exists := guides[prereq]
			if !exists {
				return fmt.Errorf("Error: prerequisite '%s' does not exist in graph", prereq)
			}
			if p.Metadata.Scope != scope {
				return fmt.Errorf("Error: prerequisite '%s' has scope '%s', which does not match guide scope '%s'", prereq, p.Metadata.Scope, scope)
			}
		}

		body := fmt.Sprintf("## %s\n\nThis is a guide for %s.\n", name, name)
		if cmd.Flags().Changed("file") && addFile != "" {
			fileBytes, readErr := os.ReadFile(addFile)
			if readErr != nil {
				return fmt.Errorf("Failed to read body file %s: %w", addFile, readErr)
			}
			body = string(fileBytes)
			if !strings.HasSuffix(body, "\n") {
				body += "\n"
			}
		} else if cmd.Flags().Changed("body") {
			body = addBody
			if !strings.HasSuffix(body, "\n") {
				body += "\n"
			}
		}

		bodyLineCount := len(strings.Split(strings.TrimSuffix(body, "\n"), "\n"))

		var subRelations []graph.SubGuideRelation
		if len(subguides) > 0 {
			if adherence == "" && len(config.Adherences) > 0 {
				adherence = config.Adherences[0]
			}

			guides[name] = graph.Guide{
				ID:        name,
				Metadata:  graph.GuideMetadata{Scope: scope},
				LineCount: bodyLineCount,
			}

			for _, sub := range subguides {
				if valErr := graph.ValidateSubGuideLink(guides, config, name, sub, adherence, segment); valErr != nil {
					return fmt.Errorf("Error: %w", valErr)
				}
				subRelations = append(subRelations, graph.SubGuideRelation{
					Guide:     sub,
					Adherence: adherence,
					Segment:   segment,
				})
			}
		}

		meta := graph.GuideMetadata{
			Prerequisites: prereqs,
			SubGuides:     subRelations,
			Scope:         scope,
			Tags:          tags,
		}

		if err := graph.CreateGuide(filename, meta, body); err != nil {
			return fmt.Errorf("Failed to write %s: %w", filename, err)
		}

		cmd.Printf("Successfully added new guide: %s\n", filename)
		return nil
	}),
}

func init() {
	addCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode")
	addCmd.Flags().StringVarP(&addDir, "dir", "d", ".", "Knowledge graph directory")
	addCmd.Flags().StringSliceVar(&addPrerequisites, "prerequisites", nil, "List of prerequisites")
	addCmd.Flags().StringVar(&addScope, "scope", "", "Scope of the guide (e.g. definition, description)")
	addCmd.Flags().StringSliceVar(&addTags, "tags", nil, "List of tags")
	addCmd.Flags().StringSliceVarP(&addSubguides, "subguide", "s", nil, "List of subguides to link")
	addCmd.Flags().StringSliceVar(&addSubguides, "sub", nil, "Alias for --subguide")
	addCmd.Flags().StringVar(&addAdherence, "adherence", "", "Adherence of the subguide relation")
	addCmd.Flags().StringVar(&addSegment, "segment", "", "Segment line range in body (e.g. 4-4)")
	addCmd.Flags().StringVar(&addBody, "body", "", "Initial markdown body for the guide")
	addCmd.Flags().StringVar(&addFile, "file", "", "Path to file containing initial markdown body")

	rootCmd.AddCommand(addCmd)
}
