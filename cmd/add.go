package cmd

import (
	"fmt"
	"os"
	"path/filepath"
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
)

var addCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new guide to the knowledge graph.",
	Long:  "Add allows you to insert a guide into the knowledge graph by specifying its prerequisites, and scope. A new markdown file will be inserted into the target directory.",
	Example: `# Add a new definition
qst add Exponent --scope definition

# Add to a specific directory
qst add Exponent --dir my_graph/`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		filename := filepath.Join(addDir, name+".md")

		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			fmt.Printf("Error: %s already exists\n", filename)
			os.Exit(1)
		}
		config, err := graph.ParseConfig(addDir)
		if err != nil {
			fmt.Printf("Failed to load manifest.yaml config: %v\n", err)
			os.Exit(1)
		}

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
					fmt.Printf("Failed to run form: %v\n", err)
					os.Exit(1)
				}
			}

			parseCSV := func(s string) []string {
				s = strings.TrimSpace(s)
				if s == "" {
					return []string{}
				}
				parts := strings.Split(s, ",")
				var res []string
				for _, p := range parts {
					if trimmed := strings.TrimSpace(p); trimmed != "" {
						res = append(res, trimmed)
					}
				}
				return res
			}

			if !cmd.Flags().Changed("prerequisites") {
				addPrerequisites = parseCSV(prereqString)
			}
			if !cmd.Flags().Changed("tags") {
				addTags = parseCSV(tagsString)
			}
		} else {
			if cmd.Flags().Changed("scope") {
				validScope := false
				for _, s := range config.Scopes {
					if addScope == s {
						validScope = true
						break
					}
				}
				if !validScope {
					fmt.Printf("Error: invalid scope '%s'. Valid scopes are: %v\n", addScope, config.Scopes)
					os.Exit(1)
				}
			}

		}

		meta := graph.GuideMetadata{
			Prerequisites: addPrerequisites,
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
			fmt.Printf("Failed to marshal frontmatter: %v\n", err)
			os.Exit(1)
		}

		content := fmt.Sprintf("---\n%s---\n\n## %s\n\nThis is a guide for %s.\n", string(metaBytes), name, name)

		err = os.WriteFile(filename, []byte(content), 0o644)
		if err != nil {
			fmt.Printf("Failed to write %s: %v\n", filename, err)
			os.Exit(1)
		}

		fmt.Printf("Successfully added new guide: %s\n", filename)
	},
}

func init() {
	addCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode")
	addCmd.Flags().StringVarP(&addDir, "dir", "d", ".", "Knowledge graph directory")
	addCmd.Flags().StringSliceVar(&addPrerequisites, "prerequisites", nil, "List of prerequisites")
	addCmd.Flags().StringVar(&addScope, "scope", "", "Scope of the guide (e.g. definition, description)")
	addCmd.Flags().StringSliceVar(&addTags, "tags", nil, "List of tags")

	rootCmd.AddCommand(addCmd)
}
