package cmd

import (
	"fmt"
	"sort"
	"strings"

	huh "charm.land/huh/v2"
	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

func cmdFormatOption(id, scope string) string {
	return id + " (" + scope + ")"
}

// promptAddGuide runs the interactive huh form for `qst add -i`.
func promptAddGuide(cmd *cobra.Command, config graph.Manifest, scope *string, prereqs *[]string, tags *[]string) error {
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
			Value(scope))
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
		form := huh.NewForm(huh.NewGroup(fields...))
		if err := form.Run(); err != nil {
			return fmt.Errorf("Failed to run form: %w", err)
		}
	}

	if !cmd.Flags().Changed("prerequisites") {
		*prereqs = parseGuideCSVOrSlice([]string{prereqString})
	}
	if !cmd.Flags().Changed("tags") {
		*tags = parseCSVOrSlice([]string{tagsString})
	}
	return nil
}

// promptLinkGuide runs the interactive huh workflow for `qst link -i`.
func promptLinkGuide(
	cmd *cobra.Command,
	guideName string,
	currentGuide graph.Guide,
	guides map[string]graph.Guide,
	config graph.Manifest,
	isPrereqSet *bool,
	isSubguideSet *bool,
	prereqs *[]string,
	subguides *[]string,
	adherence *string,
	segment *string,
) error {
	linkType := "subguide"
	if !*isPrereqSet && !*isSubguideSet {
		err := huh.NewSelect[string]().
			Title("Choose relationship type to link").
			Options(
				huh.NewOption("Subguide (hierarchical child)", "subguide"),
				huh.NewOption("Prerequisite (horizontal dependency)", "prereq"),
			).
			Value(&linkType).
			Run()
		if err != nil {
			return fmt.Errorf("Link aborted: %w", err)
		}
	} else if *isPrereqSet {
		linkType = "prereq"
	}

	if linkType == "prereq" {
		if !*isPrereqSet {
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
				return fmt.Errorf("No available guides with matching scope '%s' to link as prerequisite.", currentScope)
			}

			selectedPrereq := ""
			err := huh.NewSelect[string]().
				Title("Select prerequisite guide").
				Options(candidateOptions...).
				Value(&selectedPrereq).
				Run()
			if err != nil {
				return fmt.Errorf("Link aborted: %w", err)
			}
			*prereqs = []string{selectedPrereq}
			*isPrereqSet = true
		}
		return nil
	}

	// Subguide flow
	if !*isSubguideSet {
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
			candidateOptions = append(candidateOptions, huh.NewOption(cmdFormatOption(c, guides[c].Metadata.Scope), c))
		}

		if len(candidateOptions) == 0 {
			return fmt.Errorf("No available guides with valid lower scope to link as subguide under '%s'.", guideName)
		}

		selectedSubguide := ""
		err := huh.NewSelect[string]().
			Title("Select subguide to link").
			Options(candidateOptions...).
			Value(&selectedSubguide).
			Run()
		if err != nil {
			return fmt.Errorf("Link aborted: %w", err)
		}
		*subguides = []string{selectedSubguide}
		*isSubguideSet = true
	}

	if !cmd.Flags().Changed("adherence") {
		var adherenceOptions []huh.Option[string]
		for _, c := range config.Adherences {
			adherenceOptions = append(adherenceOptions, huh.NewOption(c, c))
		}
		err := huh.NewSelect[string]().
			Title("Choose adherence").
			Options(adherenceOptions...).
			Value(adherence).
			Run()
		if err != nil {
			return fmt.Errorf("Link aborted: %w", err)
		}
	}

	if !cmd.Flags().Changed("segment") {
		cmd.Println("\nBody Lines Preview:")
		bodyLines := strings.Split(currentGuide.Body, "\n")
		for i, line := range bodyLines {
			cmd.Printf("[L%3d] %s\n", i+1, line)
		}
		cmd.Println()

		err := huh.NewInput().
			Title("Enter segment range in body (optional)").
			Placeholder("e.g. 4-4 or 10-20").
			Value(segment).
			Run()
		if err != nil {
			return fmt.Errorf("Link aborted: %w", err)
		}
	}

	return nil
}

// promptUnlinkGuide runs the interactive huh selection for `qst unlink -i`.
// Returns (true, nil) if the guide had no links to remove and printed a notice.
func promptUnlinkGuide(
	cmd *cobra.Command,
	guideName string,
	currentGuide graph.Guide,
	isPrereqSet *bool,
	isSubguideSet *bool,
	prereq *string,
	subguide *string,
) (bool, error) {
	var linkOptions []huh.Option[string]

	for _, p := range currentGuide.Metadata.Prerequisites {
		label := fmt.Sprintf("[Prerequisite] %s", p)
		val := "prereq:" + p
		linkOptions = append(linkOptions, huh.NewOption(label, val))
	}

	for _, s := range currentGuide.Metadata.SubGuides {
		label := fmt.Sprintf("[Subguide] %s (adherence: %s, segment: %s)", s.Guide, s.Adherence, s.Segment)
		val := "subguide:" + s.Guide
		linkOptions = append(linkOptions, huh.NewOption(label, val))
	}

	if len(linkOptions) == 0 {
		cmd.Printf("Guide '%s' has no prerequisites or subguides to unlink.\n", guideName)
		return true, nil
	}

	selected := ""
	err := huh.NewSelect[string]().
		Title("Select link to remove").
		Options(linkOptions...).
		Value(&selected).
		Run()
	if err != nil {
		return false, fmt.Errorf("Unlink aborted: %w", err)
	}

	if strings.HasPrefix(selected, "prereq:") {
		*prereq = strings.TrimPrefix(selected, "prereq:")
		*isPrereqSet = true
	} else if strings.HasPrefix(selected, "subguide:") {
		*subguide = strings.TrimPrefix(selected, "subguide:")
		*isSubguideSet = true
	}
	return false, nil
}
