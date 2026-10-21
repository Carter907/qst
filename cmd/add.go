package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	huh "charm.land/huh/v2"
	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v3"
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
	Run: func(cmd *cobra.Command, args []string) {
		name := resolveGuideAndDir(cmd, args[0], &addDir)
		filename := filepath.Join(addDir, name+".md")

		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			cmd.PrintErrf("Error: %s already exists\n", filename)
			exitFunc(1)
			return
		}
		config, err := graph.ParseConfig(addDir)
		if err != nil {
			cmd.PrintErrf("Failed to load manifest.yaml config: %v\n", err)
			exitFunc(1)
			return
		}

		defer func() {
			addPrerequisites = nil
			addSubguides = nil
			addTags = nil
		}()

		if interactive {
			prereqString := ""
			tagsString := ""
			var fields []huh.Field

			if !cmd.Flags().Changed("scope") {
				var scopeOptions []huh.Option[string]
				for _, s := range config.Scopes {
					scopeOptions = append(scopeOptions, huh.NewOption(s, s))
				}
				fields = append(fields, huh.NewSelect[string]().
					Title("Choose your Scope").
					Options(scopeOptions...).
					Value(&addScope))
			}

			if !cmd.Flags().Changed("prerequisites") {
				fields = append(fields, huh.NewInput().
					Title("Enter prerequisites (separated by a comma)").
					Placeholder("e.g. Guide 1, Guide 2, Guide 3").
					Value(&prereqString))
			}

			if !cmd.Flags().Changed("tags") {
				fields = append(fields, huh.NewInput().
					Title("Enter tags (comma separated)").
					Placeholder("e.g. Math, Science, Art").
					Value(&tagsString))
			}

			if len(fields) > 0 {
				form := huh.NewForm(
					huh.NewGroup(fields...),
				)

				err = form.Run()
				if err != nil {
					cmd.PrintErrf("Failed to run form: %v\n", err)
					exitFunc(1)
					return
				}
			}

			if !cmd.Flags().Changed("prerequisites") {
				addPrerequisites = parseGuideCSVOrSlice([]string{prereqString})
			}
			if !cmd.Flags().Changed("tags") {
				addTags = parseCSVOrSlice([]string{tagsString})
			}
		} else {
			if !cmd.Flags().Changed("scope") || addScope == "" {
				cmd.PrintErrln("Error: --scope is required in non-interactive mode")
				exitFunc(1)
				return
			}
		}

		validScope := slices.Contains(config.Scopes, addScope)
		if !validScope {
			cmd.PrintErrf("Error: invalid scope '%s'. Valid scopes are: %v\n", addScope, config.Scopes)
			exitFunc(1)
			return
		}

		addPrerequisites = parseGuideCSVOrSlice(addPrerequisites)
		addSubguides = parseGuideCSVOrSlice(addSubguides)
		addTags = parseCSVOrSlice(addTags)

		// Parse graph to validate prerequisites and subguides if any are specified
		var guides map[string]graph.Guide
		if len(addPrerequisites) > 0 || len(addSubguides) > 0 {
			var parseErr error
			guides, parseErr = graph.ParseGraph(addDir)
			if parseErr != nil {
				cmd.PrintErrf("Error parsing graph: %v\n", parseErr)
				exitFunc(1)
				return
			}
		}

		// Validate prerequisites
		for _, prereq := range addPrerequisites {
			p, exists := guides[prereq]
			if !exists {
				cmd.PrintErrf("Error: prerequisite '%s' does not exist in graph\n", prereq)
				exitFunc(1)
				return
			}
			if p.Metadata.Scope != addScope {
				cmd.PrintErrf("Error: prerequisite '%s' has scope '%s', which does not match guide scope '%s'\n", prereq, p.Metadata.Scope, addScope)
				exitFunc(1)
				return
			}
		}

		// Prepare initial body content
		body := fmt.Sprintf("## %s\n\nThis is a guide for %s.\n", name, name)
		if cmd.Flags().Changed("file") && addFile != "" {
			fileBytes, readErr := os.ReadFile(addFile)
			if readErr != nil {
				cmd.PrintErrf("Failed to read body file %s: %v\n", addFile, readErr)
				exitFunc(1)
				return
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

		// Validate and prepare subguides
		var subRelations []graph.SubGuideRelation
		if len(addSubguides) > 0 {
			if addAdherence == "" && len(config.Adherences) > 0 {
				addAdherence = config.Adherences[0]
			}

			// Insert virtual parent guide into guides map for validation
			guides[name] = graph.Guide{
				ID: name,
				Metadata: graph.GuideMetadata{
					Scope: addScope,
				},
				LineCount: bodyLineCount,
			}

			for _, sub := range addSubguides {
				if valErr := graph.ValidateSubGuideLink(guides, config, name, sub, addAdherence, addSegment); valErr != nil {
					cmd.PrintErrf("Error: %v\n", valErr)
					exitFunc(1)
					return
				}
				subRelations = append(subRelations, graph.SubGuideRelation{
					Guide:     sub,
					Adherence: addAdherence,
					Segment:   addSegment,
				})
			}
		}

		meta := graph.GuideMetadata{
			Prerequisites: addPrerequisites,
			SubGuides:     subRelations,
			Scope:         addScope,
			Tags:          addTags,
		}
		// Ensure nil slices serialize to empty arrays [] in yaml instead of null
		if meta.Prerequisites == nil {
			meta.Prerequisites = []string{}
		}
		if meta.SubGuides == nil {
			meta.SubGuides = []graph.SubGuideRelation{}
		}
		if meta.Tags == nil {
			meta.Tags = []string{}
		}

		metaBytes, err := yaml.Marshal(&meta)
		if err != nil {
			cmd.PrintErrf("Failed to marshal frontmatter: %v\n", err)
			exitFunc(1)
			return
		}

		content := fmt.Sprintf("---\n%s---\n\n%s", string(metaBytes), body)

		err = os.WriteFile(filename, []byte(content), 0o644)
		if err != nil {
			cmd.PrintErrf("Failed to write %s: %v\n", filename, err)
			exitFunc(1)
			return
		}

		cmd.Printf("Successfully added new guide: %s\n", filename)
	},
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
