package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	huh "charm.land/huh/v2"
	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var (
	linkInteractive bool
	linkDir         string
	linkGuide       string
	linkAdherence   string
	linkSegment     string
)

var linkCmd = &cobra.Command{
	Use:   "link [parent-guide]",
	Short: "Link a subguide to an existing guide.",
	Long:  "Modifies an existing parent guide's frontmatter to add a new subguide relation with adherence and segment.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		parentName := args[0]
		filename := filepath.Join(linkDir, parentName+".md")

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			fmt.Printf("Error: parent guide %s does not exist\n", filename)
			os.Exit(1)
		}

		config, err := graph.ParseConfig(linkDir)
		if err != nil {
			fmt.Printf("Failed to load manifest.yaml config: %v\n", err)
			os.Exit(1)
		}

		if linkInteractive {
			var fields []huh.Field

			if !cmd.Flags().Changed("guide") {
				fields = append(fields, huh.NewInput().
					Title("Enter subguide name").
					Placeholder("e.g. Subtopic").
					Value(&linkGuide))
			}

			if !cmd.Flags().Changed("adherence") {
				var adherenceOptions []huh.Option[string]
				for _, c := range config.Adherences {
					adherenceOptions = append(adherenceOptions, huh.NewOption(c, c))
				}
				fields = append(fields, huh.NewSelect[string]().
					Title("Choose a adherence").
					Options(adherenceOptions...).
					Value(&linkAdherence))
			}

			if !cmd.Flags().Changed("segment") {
				fields = append(fields, huh.NewInput().
					Title("Enter segment (optional)").
					Placeholder("e.g. 1-10").
					Value(&linkSegment))
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
		} else {
			if cmd.Flags().Changed("adherence") {
				validAdherence := false
				for _, c := range config.Adherences {
					if linkAdherence == c {
						validAdherence = true
						break
					}
				}
				if !validAdherence {
					fmt.Printf("Error: invalid adherence '%s'. Valid clarities are: %v\n", linkAdherence, config.Adherences)
					os.Exit(1)
				}
			}
		}

		if linkGuide == "" {
			fmt.Println("Error: subguide name cannot be empty")
			os.Exit(1)
		}

		parentGuide, err := graph.ParseGuide(filename)
		if err != nil {
			fmt.Printf("Failed to parse parent guide: %v\n", err)
			os.Exit(1)
		}

		// Append the new subguide
		newSubguide := graph.SubGuideRelation{
			Guide:     linkGuide,
			Adherence: linkAdherence,
			Segment:   linkSegment,
		}
		parentGuide.Metadata.SubGuides = append(parentGuide.Metadata.SubGuides, newSubguide)

		err = graph.UpdateGuideMetadata(filename, parentGuide.Metadata)
		if err != nil {
			fmt.Printf("Failed to update parent guide: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully linked subguide '%s' to '%s'\n", linkGuide, filename)
	},
}

func init() {
	linkCmd.Flags().BoolVarP(&linkInteractive, "interactive", "i", false, "Interactive mode")
	linkCmd.Flags().StringVarP(&linkDir, "dir", "d", ".", "Knowledge graph directory")
	linkCmd.Flags().StringVar(&linkGuide, "guide", "", "Name of the subguide to link")
	linkCmd.Flags().StringVar(&linkAdherence, "adherence", "", "Adherence of the subguide relation")
	linkCmd.Flags().StringVar(&linkSegment, "segment", "", "Segment of the subguide relation")

	rootCmd.AddCommand(linkCmd)
}
