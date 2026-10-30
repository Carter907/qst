package cmd

import (
	"fmt"
	"sort"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var tagDir string

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage taxonomy tags on guides in the knowledge graph",
	Long:  "Tag provides commands to list, add, and remove tags from guides' frontmatter.",
}

var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags used across the knowledge graph",
	Example: `# List all tags in the graph
qst tag list

# List tags in a specific directory
qst tag list -d examples/C_programming`,
	Args: cobra.NoArgs,
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		guides, err := graph.ParseGraph(tagDir)
		if err != nil {
			return fmt.Errorf("Error parsing graph: %w", err)
		}

		tagCounts := make(map[string]int)
		for _, g := range guides {
			for _, t := range g.Metadata.Tags {
				tagCounts[t]++
			}
		}

		if len(tagCounts) == 0 {
			cmd.Println("No tags found across the knowledge graph.")
			return nil
		}

		var tags []string
		for t := range tagCounts {
			tags = append(tags, t)
		}
		sort.Strings(tags)

		cmd.Printf("Tags (%d):\n", len(tags))
		for _, t := range tags {
			cmd.Printf("  • %-20s (%d guides)\n", t, tagCounts[t])
		}
		return nil
	}),
}

var tagAddCmd = &cobra.Command{
	Use:   "add [guide] [tag...]",
	Short: "Add one or more tags to a guide",
	Example: `# Add tags to a guide
qst tag add "Printf Function" "io" "c-standard" -d examples/C_programming

# Add tags specifying file name or path
qst tag add "Printf Function.md" "io" -d examples/C_programming
qst tag add examples/C_programming/Printf\ Function.md "io"`,
	Args: cobra.MinimumNArgs(2),
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		dir := tagDir
		guideName := resolveGuideAndDir(cmd, args[0], &dir)
		tags := args[1:]

		guides, err := graph.ParseGraph(dir)
		if err != nil {
			return fmt.Errorf("Error parsing graph: %w", err)
		}

		guide, exists := guides[guideName]
		if !exists {
			return fmt.Errorf("Error: guide '%s' does not exist in graph", guideName)
		}
		targetPath := guide.Path

		tagList := parseCSVOrSlice(tags)
		addedCount := 0
		for _, t := range tagList {
			if err := graph.AddTag(targetPath, t); err != nil {
				return fmt.Errorf("Failed to add tag '%s': %w", t, err)
			}
			addedCount++
		}

		cmd.Printf("Successfully added %d tag(s) to '%s'\n", addedCount, guideName)
		return nil
	}),
}

var tagRemoveCmd = &cobra.Command{
	Use:   "remove [guide] [tag...]",
	Short: "Remove one or more tags from a guide",
	Example: `# Remove a tag from a guide
qst tag remove "Printf Function" "io" -d examples/C_programming

# Remove a tag specifying file name or path
qst tag remove "Printf Function.md" "io" -d examples/C_programming
qst tag remove examples/C_programming/Printf\ Function.md "io"`,
	Args: cobra.MinimumNArgs(2),
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		dir := tagDir
		guideName := resolveGuideAndDir(cmd, args[0], &dir)
		tags := args[1:]

		guides, err := graph.ParseGraph(dir)
		if err != nil {
			return fmt.Errorf("Error parsing graph: %w", err)
		}

		guide, exists := guides[guideName]
		if !exists {
			return fmt.Errorf("Error: guide '%s' does not exist in graph", guideName)
		}

		tagList := parseCSVOrSlice(tags)
		removedCount := 0
		for _, t := range tagList {
			if err := graph.RemoveTag(guide.Path, t); err != nil {
				return fmt.Errorf("Failed to remove tag '%s': %w", t, err)
			}
			removedCount++
		}

		cmd.Printf("Successfully removed %d tag(s) from '%s'\n", removedCount, guideName)
		return nil
	}),
}

func init() {
	tagCmd.PersistentFlags().StringVarP(&tagDir, "dir", "d", ".", "Knowledge graph directory")

	tagCmd.AddCommand(tagListCmd)
	tagCmd.AddCommand(tagAddCmd)
	tagCmd.AddCommand(tagRemoveCmd)

	rootCmd.AddCommand(tagCmd)
}
