package cmd

import (
	"path/filepath"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var tourDir string

var tourCmd = &cobra.Command{
	Use:   "tour",
	Short: "Manage curriculum tours in the knowledge graph",
	Long: `Tour provides commands to create, list, modify, and delete curriculum tours
configured in manifest.yaml. Tours guide readers through sequential learning paths.`,
}

var tourListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tours and their guide sequences",
	Example: `# List all tours in current directory
qst tour list

# List tours in specific directory
qst tour list -d examples/Conditionals`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := graph.ParseConfig(tourDir)
		if err != nil {
			cmd.PrintErrf("Failed to load manifest.yaml config: %v\n", err)
			exitFunc(1)
			return
		}

		if len(config.Tours) == 0 {
			cmd.Println("No tours defined in manifest.yaml.")
			return
		}

		cmd.Printf("Tours (%d):\n", len(config.Tours))
		for _, t := range config.Tours {
			cmd.Printf("\n• %s (%d guides):\n", t.Name, len(t.Guides))
			if len(t.Guides) == 0 {
				cmd.Println("    (no guides in this tour)")
			} else {
				for i, g := range t.Guides {
					cmd.Printf("    %2d. %s\n", i+1, g)
				}
			}
		}
	},
}

var tourCreateCmd = &cobra.Command{
	Use:   "create [tour-name]",
	Short: "Create a new tour in manifest.yaml",
	Example: `# Create a new tour
qst tour create "Core Fundamentals" -d my_graph/`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tourName := args[0]
		err := graph.CreateTour(tourDir, graph.TourConfig{
			Name:   tourName,
			Guides: []string{},
		})
		if err != nil {
			cmd.PrintErrf("Error creating tour: %v\n", err)
			exitFunc(1)
			return
		}

		cmd.Printf("Successfully created tour '%s'\n", tourName)
	},
}

var tourAddCmd = &cobra.Command{
	Use:   "add [tour-name] [guide...]",
	Short: "Add one or more guides to a tour in manifest.yaml",
	Example: `# Add guides to a tour
qst tour add "Core Fundamentals" "Concept A" "Concept B" -d my_graph/

# Add guides specifying file names or paths
qst tour add "Core Fundamentals" "Concept A.md" "Concept B.md" -d my_graph/
qst tour add "Core Fundamentals" my_graph/Concept\ A.md`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		tourName := args[0]
		rawGuides := args[1:]

		dirFlag := cmd.Flags().Lookup("dir")
		if dirFlag == nil || !dirFlag.Changed {
			for _, g := range rawGuides {
				dir := filepath.Dir(g)
				if dir != "." && dir != "" {
					tourDir = dir
					break
				}
			}
		}
		guideNames := cleanGuideList(rawGuides)

		guides, err := graph.ParseGraph(tourDir)
		if err != nil {
			cmd.PrintErrf("Error parsing graph: %v\n", err)
			exitFunc(1)
			return
		}

		for _, gName := range guideNames {
			if _, exists := guides[gName]; !exists {
				cmd.PrintErrf("Error: guide '%s' not found in graph\n", gName)
				exitFunc(1)
				return
			}
		}

		addedCount := 0
		for _, gName := range guideNames {
			err := graph.AddGuideToTour(tourDir, tourName, gName)
			if err != nil {
				cmd.PrintErrf("Failed to add '%s' to tour '%s': %v\n", gName, tourName, err)
				exitFunc(1)
				return
			}
			addedCount++
		}

		cmd.Printf("Successfully added %d guide(s) to tour '%s'\n", addedCount, tourName)
	},
}

var tourRemoveCmd = &cobra.Command{
	Use:   "remove [tour-name] [guide]",
	Short: "Remove a guide from a tour in manifest.yaml",
	Example: `# Remove a guide from a tour
qst tour remove "Core Fundamentals" "Concept A" -d my_graph/

# Remove a guide specifying file name or path
qst tour remove "Core Fundamentals" "Concept A.md" -d my_graph/
qst tour remove "Core Fundamentals" my_graph/Concept\ A.md`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		tourName := args[0]
		guideName := resolveGuideAndDir(cmd, args[1], &tourDir)

		err := graph.RemoveGuideFromTour(tourDir, tourName, guideName)
		if err != nil {
			cmd.PrintErrf("Error: %v\n", err)
			exitFunc(1)
			return
		}

		cmd.Printf("Successfully removed guide '%s' from tour '%s'\n", guideName, tourName)
	},
}

var tourDeleteCmd = &cobra.Command{
	Use:   "delete [tour-name]",
	Short: "Delete an entire tour from manifest.yaml",
	Example: `# Delete a tour
qst tour delete "Core Fundamentals" -d my_graph/`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tourName := args[0]

		err := graph.DeleteTour(tourDir, tourName)
		if err != nil {
			cmd.PrintErrf("Error: %v\n", err)
			exitFunc(1)
			return
		}

		cmd.Printf("Successfully deleted tour '%s'\n", tourName)
	},
}

func init() {
	tourCmd.PersistentFlags().StringVarP(&tourDir, "dir", "d", ".", "Knowledge graph directory")

	tourCmd.AddCommand(tourListCmd)
	tourCmd.AddCommand(tourCreateCmd)
	tourCmd.AddCommand(tourAddCmd)
	tourCmd.AddCommand(tourRemoveCmd)
	tourCmd.AddCommand(tourDeleteCmd)

	rootCmd.AddCommand(tourCmd)
}
