package cmd

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var (
	listDir   string
	listScope string
	listTags  []string
	listLong  bool
)

var listCmd = &cobra.Command{
	Use:     "list [directory]",
	Aliases: []string{"ls"},
	Short:   "List guides in the knowledge graph",
	Run: runWithExit(func(cmd *cobra.Command, args []string) error {
		targetDir := listDir
		if len(args) > 0 {
			targetDir = args[0]
		}
		guides, err := graph.ParseGraph(targetDir)
		if err != nil {
			return fmt.Errorf("Error parsing graph: %w", err)
		}

		hasScopeFilter := cmd.Flags().Changed("scope") && listScope != ""
		hasTagsFilter := cmd.Flags().Changed("tags") && len(listTags) > 0

		var matchedIDs []string
		for id, guide := range guides {
			if hasScopeFilter && guide.Metadata.Scope != listScope {
				continue
			}
			if hasTagsFilter {
				hasAllTags := true
				guideTags := make(map[string]bool)
				for _, t := range guide.Metadata.Tags {
					guideTags[t] = true
				}
				for _, t := range listTags {
					if !guideTags[t] {
						hasAllTags = false
						break
					}
				}
				if !hasAllTags {
					continue
				}
			}
			matchedIDs = append(matchedIDs, id)
		}

		sort.Strings(matchedIDs)

		if !listLong {
			for _, id := range matchedIDs {
				cmd.Println(id)
			}
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
		if _, err := fmt.Fprintln(w, "GUIDE\tSCOPE\tTAGS\tPREREQS\tSUBGUIDES"); err != nil {
			return fmt.Errorf("Error writing output: %w", err)
		}

		for _, id := range matchedIDs {
			g := guides[id]
			tagsStr := "-"
			if len(g.Metadata.Tags) > 0 {
				tagsStr = strings.Join(g.Metadata.Tags, ",")
			}
			if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\n",
				id,
				g.Metadata.Scope,
				tagsStr,
				len(g.Metadata.Prerequisites),
				len(g.Metadata.SubGuides),
			); err != nil {
				return fmt.Errorf("Error writing output: %w", err)
			}
		}

		if err := w.Flush(); err != nil {
			return fmt.Errorf("Error flushing output: %w", err)
		}
		return nil
	}),
}

func init() {
	listCmd.Flags().StringVarP(&listDir, "dir", "d", ".", "Knowledge graph directory")
	listCmd.Flags().StringVar(&listScope, "scope", "", "Filter by scope")
	listCmd.Flags().StringSliceVar(&listTags, "tags", nil, "Filter by tags (comma separated)")
	listCmd.Flags().BoolVarP(&listLong, "long", "l", false, "Display detailed list with scope, tags, and relationship counts")

	rootCmd.AddCommand(listCmd)
}
