package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

var (
	inspectDir      string
	inspectCoverage bool
)

var inspectCmd = &cobra.Command{
	Use:     "inspect [guide]",
	Aliases: []string{"show", "info"},
	Short:   "Display metadata, scope, relations, and coverage for a guide",
	Long: `Inspect displays a concise summary of a guide's scope, tags, prerequisites, subguides, coverage, and body statistics.
Use --coverage (-c) to expand the line-by-line body annotation.`,
	Example: `# Inspect a guide in the current directory
qst inspect "Variables and Assignment"

# Inspect specifying file name or path
qst inspect "Variables and Assignment.md" -c
qst inspect examples/C_programming/Scanf\ Function.md -c

# Inspect a guide in a specific directory
qst inspect "Scanf Function" -d examples/C_programming -c`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		guideName := resolveGuideAndDir(cmd, args[0], &inspectDir)
		runInspect(cmd, guideName, inspectDir, inspectCoverage)
	},
}

func runInspect(cmd *cobra.Command, guideName string, dir string, showCoverage bool) {
	filename := filepath.Join(dir, guideName+".md")

	guide, err := graph.ParseGuide(filename)
	if err != nil {
		cmd.PrintErrf("Error: guide '%s' not found or failed to parse: %v\n", guideName, err)
		exitFunc(1)
		return
	}

	config, err := graph.ParseConfig(dir)
	if err != nil {
		cmd.PrintErrf("Failed to load manifest.yaml: %v\n", err)
		exitFunc(1)
		return
	}

	scopeLevel := ""
	val := graph.GetScopeValue(guide.Metadata.Scope, config)
	if val > 0 {
		scopeLevel = fmt.Sprintf(" (Level %d of %d)", val, len(config.Scopes))
	}

	cmd.Printf("Guide: %s\n", guide.ID)
	cmd.Printf("File:  %s\n", guide.Path)
	cmd.Println(strings.Repeat("-", 60))
	cmd.Printf("Scope: %s%s\n", guide.Metadata.Scope, scopeLevel)

	if len(guide.Metadata.Tags) > 0 {
		cmd.Printf("Tags:  [%s]\n", strings.Join(guide.Metadata.Tags, ", "))
	} else {
		cmd.Printf("Tags:  (none)\n")
	}

	cmd.Println("\nPrerequisites:")
	if len(guide.Metadata.Prerequisites) == 0 {
		cmd.Println("  (none)")
	} else {
		for _, p := range guide.Metadata.Prerequisites {
			cmd.Printf("  • %s\n", p)
		}
	}

	cmd.Println("\nSubguides:")
	if len(guide.Metadata.SubGuides) == 0 {
		cmd.Println("  (none)")
	} else {
		for _, s := range guide.Metadata.SubGuides {
			seg := s.Segment
			if seg == "" {
				seg = "all"
			}
			cmd.Printf("  • %-28s adherence: %-12s segment: %s\n", s.Guide, s.Adherence, seg)
		}
	}

	// Coverage analysis
	report, err := graph.AnalyzeGuideCoverage(guide, config)
	if err == nil && report.TotalBodyLines > 0 {
		pct := (float64(report.CoveredLines) / float64(report.TotalBodyLines)) * 100.0
		strictNote := ""
		if report.StrictCoverage {
			if !report.HasSubGuides {
				strictNote = " [strict: N/A (leaf)]"
			} else if report.UnmappedLines > 0 || report.OverlappingLines > 0 {
				strictNote = " [strict: FAILED]"
			} else {
				strictNote = " [strict: PASSED]"
			}
		}

		hint := ""
		if !showCoverage {
			hint = " (use -c for line breakdown)"
		}

		cmd.Printf("\nCoverage: %d/%d body lines (%.1f%%) | %d unmapped, %d overlaps%s%s\n",
			report.CoveredLines, report.TotalBodyLines, pct, report.UnmappedLines, report.OverlappingLines, strictNote, hint)
	}

	// Content stats
	wordCount := len(strings.Fields(guide.Body))
	bodyLines := guide.LineCount
	if report.TotalBodyLines > 0 {
		bodyLines = report.TotalBodyLines
	}
	cmd.Println("\nContent Stats:")
	cmd.Printf("  Body Lines: %d\n", bodyLines)
	cmd.Printf("  Word Count: %d\n", wordCount)

	// Expanded coverage section
	if showCoverage && err == nil {
		cmd.Println("\n" + strings.Repeat("=", 80))
		cmd.Printf("Line-by-Line Body Coverage: %s\n", guide.ID)
		cmd.Println(strings.Repeat("-", 80))

		if report.TotalBodyLines == 0 {
			cmd.Println("  (Guide has no markdown body text)")
		} else {
			for _, line := range report.Lines {
				tag := ""
				if len(line.SubGuides) == 0 {
					tag = "[UNMAPPED]"
				} else if line.IsOverlap {
					var parts []string
					for i := range line.SubGuides {
						parts = append(parts, fmt.Sprintf("%s (%s)", line.SubGuides[i], line.Adherences[i]))
					}
					tag = fmt.Sprintf("[OVERLAP: %s]", strings.Join(parts, ", "))
				} else {
					tag = fmt.Sprintf("[%s (%s)]", line.SubGuides[0], line.Adherences[0])
				}

				cmd.Printf("[L%3d] %-35s %s\n", line.LineNumber, tag, line.Text)
			}
		}

		cmd.Println(strings.Repeat("-", 80))
		pct := 0.0
		if report.TotalBodyLines > 0 {
			pct = (float64(report.CoveredLines) / float64(report.TotalBodyLines)) * 100.0
		}
		cmd.Printf("Coverage Details:\n")
		cmd.Printf("  Total Body Lines:  %d\n", report.TotalBodyLines)
		cmd.Printf("  Covered Lines:     %d (%.1f%%)\n", report.CoveredLines, pct)
		cmd.Printf("  Unmapped Lines:    %d\n", report.UnmappedLines)
		cmd.Printf("  Overlapping Lines: %d\n", report.OverlappingLines)

		if report.StrictCoverage {
			if !report.HasSubGuides {
				cmd.Println("  Status: N/A (leaf guide - no subguides defined)")
			} else if report.UnmappedLines > 0 || report.OverlappingLines > 0 {
				cmd.Println("  Status: FAILED (strict_coverage is enabled but unmapped or overlapping lines exist)")
			} else {
				cmd.Println("  Status: PASSED (100% strict coverage satisfied)")
			}
		}
	}
}

func init() {
	inspectCmd.Flags().StringVarP(&inspectDir, "dir", "d", ".", "Knowledge graph directory")
	inspectCmd.Flags().BoolVarP(&inspectCoverage, "coverage", "c", false, "Show detailed line-by-line body coverage")
	rootCmd.AddCommand(inspectCmd)
}
