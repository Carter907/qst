package graph

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// CheckAcyclic checks if the knowledge graph is acyclic using DFS.
// If a cycle is detected, it formats the complete cycle path in the error message.
func CheckAcyclic(guides map[string]Guide) error {
	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)

	state := make(map[string]int, len(guides))
	var path []string

	var dfs func(id string) error
	dfs = func(id string) error {
		state[id] = visiting
		path = append(path, id)
		defer func() {
			path = path[:len(path)-1]
		}()

		if guide, exists := guides[id]; exists {
			neighbors := append([]string(nil), guide.Metadata.Prerequisites...)
			for _, sub := range guide.Metadata.SubGuides {
				neighbors = append(neighbors, sub.Guide)
			}

			for _, neighborID := range neighbors {
				if _, exists := guides[neighborID]; !exists {
					continue
				}
				if state[neighborID] == visiting {
					cycleStart := -1
					for i, node := range path {
						if node == neighborID {
							cycleStart = i
							break
						}
					}
					var cycleChain []string
					if cycleStart != -1 {
						cycleChain = append(append([]string(nil), path[cycleStart:]...), neighborID)
					} else {
						cycleChain = []string{neighborID, neighborID}
					}
					return fmt.Errorf("cycle detected in graph: %s", strings.Join(cycleChain, " -> "))
				}
				if state[neighborID] == unvisited {
					if err := dfs(neighborID); err != nil {
						return err
					}
				}
			}
		}

		state[id] = visited
		return nil
	}

	ids := make([]string, 0, len(guides))
	for id := range guides {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		if state[id] == unvisited {
			if err := dfs(id); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseSegments parses a segment string like "10-20, 25-30" or "4, 6" into a slice of [2]int
func parseSegments(segmentStr string) ([][2]int, error) {
	if strings.TrimSpace(segmentStr) == "" {
		return nil, nil
	}
	var ranges [][2]int
	parts := strings.SplitSeq(segmentStr, ",")
	for p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		bounds := strings.Split(p, "-")
		if len(bounds) == 1 {
			line, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid line number: %w", err)
			}
			if line < 1 {
				return nil, fmt.Errorf("line number must be positive: %d", line)
			}
			ranges = append(ranges, [2]int{line, line})
			continue
		}
		if len(bounds) != 2 {
			return nil, fmt.Errorf("invalid range format: %s", p)
		}
		start, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid start line: %w", err)
		}
		end, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid end line: %w", err)
		}
		if start < 1 {
			return nil, fmt.Errorf("start line must be positive: %d", start)
		}
		if start > end {
			return nil, fmt.Errorf("start line %d is greater than end line %d", start, end)
		}
		ranges = append(ranges, [2]int{start, end})
	}
	return ranges, nil
}

// getScopeValue helper returns the hierarchical index of the given scope (1-indexed)
func getScopeValue(scope string, config Manifest) int {
	for i, s := range config.Scopes {
		if s == scope {
			return i + 1
		}
	}
	return 0
}

func isValidAdherence(adherence string, config Manifest) bool {
	return slices.Contains(config.Adherences, adherence)
}

// ValidateGraph checks structural constraints of the knowledge graph
func ValidateGraph(guides map[string]Guide, config Manifest) error {
	if err := validateTours(guides, config); err != nil {
		return err
	}

	for _, guide := range guides {
		if err := validateGuide(guide, guides, config); err != nil {
			return err
		}
	}
	return CheckAcyclic(guides)
}

func validateTours(guides map[string]Guide, config Manifest) error {
	for _, tour := range config.Tours {
		for _, tGuide := range tour.Guides {
			if _, exists := guides[tGuide]; !exists {
				return fmt.Errorf("tour '%s' references unknown guide: '%s'", tour.Name, tGuide)
			}
		}
	}
	return nil
}

func validateGuide(guide Guide, guides map[string]Guide, config Manifest) error {
	if !guide.HasContent {
		return fmt.Errorf("guide '%s' has no markdown content", guide.ID)
	}

	// Validate Scope
	guideScopeVal := getScopeValue(guide.Metadata.Scope, config)
	if guideScopeVal == 0 {
		return fmt.Errorf("guide '%s' has invalid scope: '%s'", guide.ID, guide.Metadata.Scope)
	}

	// Check RequireSubguides for non-leaf scopes
	if config.RequireSubguides && guideScopeVal > 1 && len(guide.Metadata.SubGuides) == 0 {
		return fmt.Errorf("guide '%s' is a non-leaf scope but has no subguides (require_subguides is true)", guide.ID)
	}

	if err := validatePrerequisites(guide, guides); err != nil {
		return err
	}

	allRanges, err := validateSubGuides(guide, guides, config, guideScopeVal)
	if err != nil {
		return err
	}

	return validateCoverage(guide, allRanges, config)
}

func validatePrerequisites(guide Guide, guides map[string]Guide) error {
	for _, prereqID := range guide.Metadata.Prerequisites {
		prereq, exists := guides[prereqID]
		if !exists {
			return fmt.Errorf("guide '%s' references unknown prerequisite: '%s'", guide.ID, prereqID)
		}
		if prereq.Metadata.Scope != guide.Metadata.Scope {
			return fmt.Errorf("guide '%s' (scope: '%s') has prerequisite '%s' with mismatched scope: '%s'. Horizontal edges must have exactly identical scope",
				guide.ID, guide.Metadata.Scope, prereqID, prereq.Metadata.Scope)
		}
	}
	return nil
}

func validateSubGuides(guide Guide, guides map[string]Guide, config Manifest, guideScopeVal int) ([][2]int, error) {
	var allRanges [][2]int

	for _, subRelation := range guide.Metadata.SubGuides {
		if subRelation.Guide == "" {
			return nil, fmt.Errorf("guide '%s' has an empty guide reference in its subguides", guide.ID)
		}
		if subRelation.Adherence == "" {
			return nil, fmt.Errorf("guide '%s' is missing adherence for subguide '%s'", guide.ID, subRelation.Guide)
		}
		if subRelation.Segment == "" {
			return nil, fmt.Errorf("guide '%s' is missing segment for subguide '%s'", guide.ID, subRelation.Guide)
		}

		subID := subRelation.Guide
		sub, exists := guides[subID]
		if !exists {
			return nil, fmt.Errorf("guide '%s' references unknown sub_guide: '%s'", guide.ID, subID)
		}

		if !isValidAdherence(subRelation.Adherence, config) {
			return nil, fmt.Errorf("guide '%s' has invalid subguide adherence: '%s' for subguide '%s'", guide.ID, subRelation.Adherence, subID)
		}

		ranges, err := parseSegments(subRelation.Segment)
		if err != nil {
			return nil, fmt.Errorf("guide '%s' has invalid segment for subguide '%s': %w", guide.ID, subID, err)
		}
		allRanges = append(allRanges, ranges...)

		subScopeVal := getScopeValue(sub.Metadata.Scope, config)

		if subScopeVal >= guideScopeVal {
			return nil, fmt.Errorf("guide '%s' (scope: '%s') has sub_guide '%s' with invalid scope: '%s'. Sub-guides must have a strictly smaller scope",
				guide.ID, guide.Metadata.Scope, subID, sub.Metadata.Scope)
		}

		if !config.RelaxedSubguides && guideScopeVal-subScopeVal != 1 {
			return nil, fmt.Errorf("guide '%s' (scope: '%s') has sub_guide '%s' with scope: '%s', but relaxed_subguides is false. Sub-guides must be exactly one scope level below their parent",
				guide.ID, guide.Metadata.Scope, subID, sub.Metadata.Scope)
		}
	}
	return allRanges, nil
}

func validateCoverage(guide Guide, allRanges [][2]int, config Manifest) error {
	// Sort and check overlaps
	slices.SortFunc(allRanges, func(a, b [2]int) int {
		if a[0] == b[0] {
			return a[1] - b[1]
		}
		return a[0] - b[0]
	})

	for i := 1; i < len(allRanges); i++ {
		if allRanges[i][0] <= allRanges[i-1][1] {
			return fmt.Errorf("guide '%s' has overlapping subguide segments: %v and %v", guide.ID, allRanges[i-1], allRanges[i])
		}
	}

	if config.StrictCoverage && len(guide.Metadata.SubGuides) > 0 {
		if len(allRanges) == 0 {
			return fmt.Errorf("guide '%s' has strict_coverage enabled but no segments defined", guide.ID)
		}
		if allRanges[0][0] > 1 {
			return fmt.Errorf("guide '%s' is missing coverage for lines 1 to %d", guide.ID, allRanges[0][0]-1)
		}
		for i := 1; i < len(allRanges); i++ {
			if allRanges[i][0] > allRanges[i-1][1]+1 {
				return fmt.Errorf("guide '%s' is missing coverage for lines %d to %d", guide.ID, allRanges[i-1][1]+1, allRanges[i][0]-1)
			}
		}
		effectiveLineCount := guide.LineCount
		// If the file ends with a trailing newline (POSIX style EOF), the final body line is empty.
		// Allow coverage to terminate at either LineCount or LineCount-1 if the trailing line is empty.
		if effectiveLineCount > 0 && strings.HasSuffix(guide.Body, "\n") {
			if allRanges[len(allRanges)-1][1] == effectiveLineCount-1 {
				effectiveLineCount--
			}
		}
		if allRanges[len(allRanges)-1][1] < effectiveLineCount {
			return fmt.Errorf("guide '%s' is missing coverage for lines %d to %d", guide.ID, allRanges[len(allRanges)-1][1]+1, effectiveLineCount)
		}
	}
	return nil
}

// GetScopeValue returns the hierarchical index of the given scope (1-indexed)
func GetScopeValue(scope string, config Manifest) int {
	return getScopeValue(scope, config)
}

// IsValidAdherence checks if an adherence value is listed in the manifest
func IsValidAdherence(adherence string, config Manifest) bool {
	return isValidAdherence(adherence, config)
}

// ParseSegments parses a segment string like "10-20, 25-30" into a slice of [2]int
func ParseSegments(segmentStr string) ([][2]int, error) {
	return parseSegments(segmentStr)
}

// ValidatePrerequisiteLink checks if a prerequisite relationship can be legally added
func ValidatePrerequisiteLink(guides map[string]Guide, config Manifest, sourceID, prereqID string) error {
	sourceGuide, exists := guides[sourceID]
	if !exists {
		return fmt.Errorf("guide '%s' does not exist", sourceID)
	}

	prereqGuide, exists := guides[prereqID]
	if !exists {
		return fmt.Errorf("prerequisite guide '%s' does not exist", prereqID)
	}

	if sourceID == prereqID {
		return fmt.Errorf("guide cannot be a prerequisite of itself")
	}

	if sourceGuide.Metadata.Scope != prereqGuide.Metadata.Scope {
		return fmt.Errorf("cannot link prerequisite: '%s' has scope '%s' but '%s' has scope '%s'. Horizontal edges must have identical scope",
			prereqID, prereqGuide.Metadata.Scope, sourceID, sourceGuide.Metadata.Scope)
	}

	if slices.Contains(sourceGuide.Metadata.Prerequisites, prereqID) {
		return fmt.Errorf("prerequisite '%s' is already linked to '%s'", prereqID, sourceID)
	}

	// Cycle simulation
	cloned := make(map[string]Guide, len(guides))
	for k, v := range guides {
		cloned[k] = v
	}
	simGuide := sourceGuide
	simGuide.Metadata.Prerequisites = append([]string(nil), sourceGuide.Metadata.Prerequisites...)
	simGuide.Metadata.Prerequisites = append(simGuide.Metadata.Prerequisites, prereqID)
	cloned[sourceID] = simGuide

	if err := CheckAcyclic(cloned); err != nil {
		return fmt.Errorf("linking prerequisite '%s' to '%s' creates a cycle: %w", prereqID, sourceID, err)
	}

	return nil
}

// ValidateSubGuideLink checks if a subguide relationship can be legally added or updated
func ValidateSubGuideLink(guides map[string]Guide, config Manifest, parentID, subID, adherence, segment string) error {
	parentGuide, exists := guides[parentID]
	if !exists {
		return fmt.Errorf("parent guide '%s' does not exist", parentID)
	}

	subGuide, exists := guides[subID]
	if !exists {
		return fmt.Errorf("subguide '%s' does not exist", subID)
	}

	if parentID == subID {
		return fmt.Errorf("guide cannot be a subguide of itself")
	}

	parentScopeVal := getScopeValue(parentGuide.Metadata.Scope, config)
	if parentScopeVal == 0 {
		return fmt.Errorf("parent guide '%s' has invalid scope '%s'", parentID, parentGuide.Metadata.Scope)
	}

	subScopeVal := getScopeValue(subGuide.Metadata.Scope, config)
	if subScopeVal == 0 {
		return fmt.Errorf("subguide '%s' has invalid scope '%s'", subID, subGuide.Metadata.Scope)
	}

	if subScopeVal >= parentScopeVal {
		return fmt.Errorf("subguide '%s' (scope: '%s') must have a strictly smaller scope than parent '%s' (scope: '%s')",
			subID, subGuide.Metadata.Scope, parentID, parentGuide.Metadata.Scope)
	}

	if !config.RelaxedSubguides && parentScopeVal-subScopeVal != 1 {
		return fmt.Errorf("subguide '%s' (scope: '%s') must be exactly one scope level below parent '%s' (scope: '%s') when relaxed_subguides is false",
			subID, subGuide.Metadata.Scope, parentID, parentGuide.Metadata.Scope)
	}

	if adherence != "" && !isValidAdherence(adherence, config) {
		return fmt.Errorf("invalid adherence '%s'. Valid adherences are: %v", adherence, config.Adherences)
	}

	if segment != "" {
		ranges, err := parseSegments(segment)
		if err != nil {
			return fmt.Errorf("invalid segment '%s': %w", segment, err)
		}
		if parentGuide.LineCount > 0 {
			for _, r := range ranges {
				if r[1] > parentGuide.LineCount {
					return fmt.Errorf("segment end line %d exceeds parent guide body line count (%d)", r[1], parentGuide.LineCount)
				}
			}
		}
	}

	// Cycle simulation
	cloned := make(map[string]Guide, len(guides))
	for k, v := range guides {
		cloned[k] = v
	}
	simGuide := parentGuide
	simGuide.Metadata.SubGuides = append([]SubGuideRelation(nil), parentGuide.Metadata.SubGuides...)
	newRel := SubGuideRelation{Guide: subID, Adherence: adherence, Segment: segment}
	found := false
	for i, s := range simGuide.Metadata.SubGuides {
		if s.Guide == subID {
			simGuide.Metadata.SubGuides[i] = newRel
			found = true
			break
		}
	}
	if !found {
		simGuide.Metadata.SubGuides = append(simGuide.Metadata.SubGuides, newRel)
	}
	cloned[parentID] = simGuide

	if err := CheckAcyclic(cloned); err != nil {
		return fmt.Errorf("linking subguide '%s' to '%s' creates a cycle: %w", subID, parentID, err)
	}

	return nil
}

// LineCoverage describes the coverage state of a single body line in a guide
type LineCoverage struct {
	LineNumber int
	Text       string
	SubGuides  []string
	Adherences []string
	IsOverlap  bool
}

// GuideCoverageReport summarizes coverage information for a guide's body
type GuideCoverageReport struct {
	GuideID          string
	Scope            string
	StrictCoverage   bool
	HasSubGuides     bool
	TotalBodyLines   int
	CoveredLines     int
	UnmappedLines    int
	OverlappingLines int
	Lines            []LineCoverage
}

// AnalyzeGuideCoverage evaluates subguide segment coverage against body lines
func AnalyzeGuideCoverage(guide Guide, config Manifest) (GuideCoverageReport, error) {
	report := GuideCoverageReport{
		GuideID:        guide.ID,
		Scope:          guide.Metadata.Scope,
		StrictCoverage: config.StrictCoverage,
		HasSubGuides:   len(guide.Metadata.SubGuides) > 0,
	}

	bodyLines := []string{}
	if guide.Body != "" {
		bodyLines = strings.Split(guide.Body, "\n")
	}
	report.TotalBodyLines = len(bodyLines)

	type subHit struct {
		guide     string
		adherence string
	}
	lineHits := make(map[int][]subHit)

	for _, sub := range guide.Metadata.SubGuides {
		if sub.Segment == "" {
			continue
		}
		ranges, err := parseSegments(sub.Segment)
		if err != nil {
			return report, fmt.Errorf("invalid segment '%s' for subguide '%s': %w", sub.Segment, sub.Guide, err)
		}
		for _, r := range ranges {
			for l := r[0]; l <= r[1]; l++ {
				lineHits[l] = append(lineHits[l], subHit{guide: sub.Guide, adherence: sub.Adherence})
			}
		}
	}

	report.Lines = make([]LineCoverage, report.TotalBodyLines)
	for i := 0; i < report.TotalBodyLines; i++ {
		lineNum := i + 1
		text := bodyLines[i]
		hits := lineHits[lineNum]

		cov := LineCoverage{
			LineNumber: lineNum,
			Text:       text,
		}

		if len(hits) == 0 {
			report.UnmappedLines++
		} else {
			report.CoveredLines++
			if len(hits) > 1 {
				cov.IsOverlap = true
				report.OverlappingLines++
			}
			for _, h := range hits {
				cov.SubGuides = append(cov.SubGuides, h.guide)
				cov.Adherences = append(cov.Adherences, h.adherence)
			}
		}
		report.Lines[i] = cov
	}

	// If the file ends with a trailing newline, the final body line is empty.
	// If that trailing line is unmapped, exclude it from TotalBodyLines and UnmappedLines
	// so POSIX newlines at EOF do not register as unmapped errors.
	if report.TotalBodyLines > 0 && bodyLines[report.TotalBodyLines-1] == "" && len(lineHits[report.TotalBodyLines]) == 0 {
		report.TotalBodyLines--
		report.UnmappedLines--
		report.Lines = report.Lines[:report.TotalBodyLines]
	}

	return report, nil
}
