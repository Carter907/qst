package ssg

import (
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"unicode"

	"github.com/Carter907/qst/internal/graph"
)

func cleanMermaidID(id string) string {
	var sb strings.Builder
	for _, r := range strings.ToLower(id) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	s := sb.String()
	s = strings.Trim(s, "_")
	if s == "" || (s[0] >= '0' && s[0] <= '9') {
		s = "node_" + s
	}
	return s
}

// wrapNodeTitle wraps a guide title into multiple lines at word boundaries
// targeting approximately targetLen characters per line, separated by <br/>.
func wrapNodeTitle(title string, targetLen int) string {
	words := strings.Fields(title)
	if len(words) <= 1 {
		return title
	}

	var lines []string
	var current strings.Builder

	for _, word := range words {
		if current.Len() == 0 {
			current.WriteString(word)
		} else if current.Len()+1+len(word) <= targetLen {
			current.WriteByte(' ')
			current.WriteString(word)
		} else {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
		}
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return strings.Join(lines, "<br/>")
}

// formatGuideNodeLabel formats a guide node label with a capitalized scope header
// and wrapped title separated by <br/> according to the minimal academic format:
// "<Scope><br/><Wrapped Title>". If no scope is present, it returns the wrapped title.
func formatGuideNodeLabel(id string, sc string) string {
	wrappedID := wrapNodeTitle(id, 24)
	escapedID := strings.ReplaceAll(wrappedID, "\"", "#quot;")
	if sc == "" {
		return escapedID
	}
	escapedScope := strings.ReplaceAll(sc, "\"", "#quot;")
	r := []rune(escapedScope)
	capScope := string(unicode.ToUpper(r[0])) + string(r[1:])
	return fmt.Sprintf("%s<br/>%s", capScope, escapedID)
}

// MermaidOptions specifies formatting options when rendering localized Mermaid diagrams.
type MermaidOptions struct {
	MarkdownWrap     bool
	RequireNeighbors bool
	ClickURLFunc     func(guideID string) string // if non-nil, emits click directives
	Dependents       []string
	ParentGuides     []string
}

// buildMermaidNodeIDs maps a sorted slice of guide IDs to unique, collision-free Mermaid node identifiers.
func buildMermaidNodeIDs(ids []string) map[string]string {
	nodeIDs := make(map[string]string, len(ids))
	usedIDs := make(map[string]bool, len(ids))

	for _, id := range ids {
		clean := cleanMermaidID(id)
		candidate := clean
		counter := 1
		for usedIDs[candidate] {
			counter++
			candidate = fmt.Sprintf("%s_%d", clean, counter)
		}
		usedIDs[candidate] = true
		nodeIDs[id] = candidate
	}
	return nodeIDs
}

// GenerateGlobalMermaid generates a full-graph Mermaid.js TD flowchart for all guides in the graph.
func GenerateGlobalMermaid(guides map[string]graph.Guide) string {
	var sb strings.Builder
	sb.WriteString("```mermaid\ngraph TD\n")

	var ids []string
	seen := make(map[string]bool, len(guides))
	for id := range guides {
		ids = append(ids, id)
		seen[id] = true
	}
	sort.Strings(ids)

	allIDs := append([]string(nil), ids...)
	for _, id := range ids {
		g := guides[id]
		for _, p := range g.Metadata.Prerequisites {
			if !seen[p] {
				seen[p] = true
				allIDs = append(allIDs, p)
			}
		}
		for _, s := range g.Metadata.SubGuides {
			if !seen[s.Guide] {
				seen[s.Guide] = true
				allIDs = append(allIDs, s.Guide)
			}
		}
	}
	sort.Strings(allIDs)
	nodeIDs := buildMermaidNodeIDs(allIDs)

	for _, id := range ids {
		escapedID := strings.ReplaceAll(id, "\"", "#quot;")
		fmt.Fprintf(&sb, "    %s[\"%s\"]\n", nodeIDs[id], escapedID)
	}

	for _, id := range ids {
		guide := guides[id]
		cleanID := nodeIDs[id]

		for _, prereq := range guide.Metadata.Prerequisites {
			fmt.Fprintf(&sb, "    %s --> %s\n", nodeIDs[prereq], cleanID)
		}

		for _, sub := range guide.Metadata.SubGuides {
			fmt.Fprintf(&sb, "    %s -.-> %s\n", cleanID, nodeIDs[sub.Guide])
		}
	}

	sb.WriteString("```")
	return sb.String()
}

// sortSubGuidesByNarrative orders sub-guides so that segmented guides appear first
// sorted by their start line (top-to-bottom reading order in the document), followed
// by unsegmented sub-guides in their original frontmatter declaration order.
func sortSubGuidesByNarrative(subguides []graph.SubGuideRelation) []graph.SubGuideRelation {
	type subItem struct {
		rel       graph.SubGuideRelation
		startLine int
		index     int
	}

	items := make([]subItem, len(subguides))
	for i, s := range subguides {
		startLine := math.MaxInt
		if s.Segment != "" {
			if ranges, err := graph.ParseSegments(s.Segment); err == nil && len(ranges) > 0 {
				minStart := ranges[0][0]
				for _, r := range ranges[1:] {
					if r[0] < minStart {
						minStart = r[0]
					}
				}
				startLine = minStart
			}
		}
		items[i] = subItem{
			rel:       s,
			startLine: startLine,
			index:     i,
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].startLine != items[j].startLine {
			return items[i].startLine < items[j].startLine
		}
		return items[i].index < items[j].index
	})

	sorted := make([]graph.SubGuideRelation, len(items))
	for i, item := range items {
		sorted[i] = item.rel
	}
	return sorted
}

// BuildGuideMermaid generates a localized 1-hop Mermaid diagram for a guide.
func BuildGuideMermaid(guide graph.Guide, getScope func(id string) string, opts MermaidOptions) string {
	prereqs := guide.Metadata.Prerequisites
	subguides := sortSubGuidesByNarrative(guide.Metadata.SubGuides)
	isFoundational := len(prereqs) == 0 && len(subguides) == 0
	if opts.RequireNeighbors && isFoundational && len(opts.Dependents) == 0 && len(opts.ParentGuides) == 0 {
		return ""
	}

	var sb strings.Builder
	if opts.MarkdownWrap {
		sb.WriteString("```mermaid\n")
	}
	sb.WriteString("graph LR\n")

	// 1. Collect all unique guide IDs in this localized 1-hop horizon
	seen := make(map[string]bool)
	seen[guide.ID] = true
	for _, p := range prereqs {
		seen[p] = true
	}
	for _, s := range subguides {
		seen[s.Guide] = true
	}
	if isFoundational {
		for _, d := range opts.Dependents {
			seen[d] = true
		}
		for _, p := range opts.ParentGuides {
			seen[p] = true
		}
	}

	// Order nodes deterministically:
	// 1. Prerequisites (in author frontmatter order)
	// 2. Current guide (center rank)
	// 3. Sub-guides (in narrative reading order)
	// 4. Foundational connections (dependents, parent guides)
	var ids []string
	addID := func(id string) {
		if seen[id] {
			ids = append(ids, id)
			delete(seen, id)
		}
	}

	for _, p := range prereqs {
		addID(p)
	}
	addID(guide.ID)
	for _, s := range subguides {
		addID(s.Guide)
	}
	if isFoundational {
		for _, d := range opts.Dependents {
			addID(d)
		}
		for _, p := range opts.ParentGuides {
			addID(p)
		}
	}
	if len(seen) > 0 {
		var rem []string
		for id := range seen {
			rem = append(rem, id)
		}
		sort.Strings(rem)
		for _, id := range rem {
			addID(id)
		}
	}

	// 2. Map guide IDs to unique safe Mermaid node identifiers
	nodeIDs := buildMermaidNodeIDs(ids)

	// 3. Declare nodes with academic scope prefix
	for _, id := range ids {
		nodeID := nodeIDs[id]
		sc := ""
		if getScope != nil {
			sc = getScope(id)
		}
		label := formatGuideNodeLabel(id, sc)
		fmt.Fprintf(&sb, "    %s[\"%s\"]\n", nodeID, label)
	}
	sb.WriteString("\n")

	// 4. Edges
	seenEdges := make(map[string]bool)
	currentNodeID := nodeIDs[guide.ID]

	var prereqEdges []string
	for _, p := range prereqs {
		pNodeID, exists := nodeIDs[p]
		if exists {
			edgeKey := fmt.Sprintf("%s-->%s", pNodeID, currentNodeID)
			if !seenEdges[edgeKey] {
				seenEdges[edgeKey] = true
				prereqEdges = append(prereqEdges, fmt.Sprintf("    %s --> %s", pNodeID, currentNodeID))
			}
		}
	}
	for _, edge := range prereqEdges {
		sb.WriteString(edge)
		sb.WriteString("\n")
	}

	var subEdges []string
	for _, sub := range subguides {
		subNodeID, exists := nodeIDs[sub.Guide]
		if exists {
			edgeKey := fmt.Sprintf("%s-.->%s", currentNodeID, subNodeID)
			if !seenEdges[edgeKey] {
				seenEdges[edgeKey] = true
				edgeLabel := ""
				if sub.Adherence != "" && sub.Segment != "" {
					edgeLabel = fmt.Sprintf("%s [%s]", sub.Adherence, sub.Segment)
				} else if sub.Adherence != "" {
					edgeLabel = sub.Adherence
				} else if sub.Segment != "" {
					edgeLabel = fmt.Sprintf("[%s]", sub.Segment)
				}
				if edgeLabel != "" {
					escapedEdgeLabel := strings.ReplaceAll(edgeLabel, "\"", "#quot;")
					subEdges = append(subEdges, fmt.Sprintf("    %s -.->|\"%s\"| %s", currentNodeID, escapedEdgeLabel, subNodeID))
				} else {
					subEdges = append(subEdges, fmt.Sprintf("    %s -.-> %s", currentNodeID, subNodeID))
				}
			}
		}
	}
	for _, edge := range subEdges {
		sb.WriteString(edge)
		sb.WriteString("\n")
	}

	if isFoundational {
		var depEdges []string
		for _, dep := range opts.Dependents {
			dNodeID, exists := nodeIDs[dep]
			if exists {
				edgeKey := fmt.Sprintf("%s-->%s", currentNodeID, dNodeID)
				if !seenEdges[edgeKey] {
					seenEdges[edgeKey] = true
					depEdges = append(depEdges, fmt.Sprintf("    %s --> %s", currentNodeID, dNodeID))
				}
			}
		}
		for _, edge := range depEdges {
			sb.WriteString(edge)
			sb.WriteString("\n")
		}

		var parentEdges []string
		for _, p := range opts.ParentGuides {
			pNodeID, exists := nodeIDs[p]
			if exists {
				edgeKey := fmt.Sprintf("%s-.->%s", pNodeID, currentNodeID)
				if !seenEdges[edgeKey] {
					seenEdges[edgeKey] = true
					parentEdges = append(parentEdges, fmt.Sprintf("    %s -.-> %s", pNodeID, currentNodeID))
				}
			}
		}
		for _, edge := range parentEdges {
			sb.WriteString(edge)
			sb.WriteString("\n")
		}
	}

	// 5. Click directives for interactive navigation
	if opts.ClickURLFunc != nil {
		var clickLines []string
		for _, id := range ids {
			if id == guide.ID {
				continue
			}
			targetURL := opts.ClickURLFunc(id)
			if targetURL == "" {
				continue
			}
			nodeID := nodeIDs[id]
			u := (&url.URL{Path: targetURL}).EscapedPath()
			clickLines = append(clickLines, fmt.Sprintf("    click %s href \"%s\"", nodeID, u))
		}
		if len(clickLines) > 0 {
			sb.WriteString("\n")
			for _, click := range clickLines {
				sb.WriteString(click)
				sb.WriteString("\n")
			}
		}
	}

	// 6. Monochrome academic styles
	sb.WriteString("\n")
	sb.WriteString("    classDef default fill:#ffffff,stroke:#1f2937,stroke-width:1.2px,rx:3px,ry:3px;\n")
	sb.WriteString("    classDef currentGuide fill:#f9fafb,stroke:#000000,stroke-width:2.5px,font-weight:bold;\n")
	fmt.Fprintf(&sb, "    class %s currentGuide;\n", currentNodeID)

	if opts.MarkdownWrap {
		sb.WriteString("```")
	}

	return sb.String()
}

// generateGuideMermaid produces a Mermaid diagram representing the localized
// 1-hop neighborhood for a specific guide in an open monochrome academic LR layout.
// For foundational guides without incoming prerequisites or subguides, outgoing
// dependents and parent guides are rendered to provide forward navigation.
// Neighbor nodes include clickable hyperlinks to their respective guide pages.
func generateGuideMermaid(currentGF *guideFile, config graph.Manifest, filesByID map[string]*guideFile, graphIdx ...*graph.GraphIndex) string {
	if currentGF == nil {
		return ""
	}
	getScope := func(id string) string {
		if gf := filesByID[id]; gf != nil {
			return gf.guide.Metadata.Scope
		}
		return ""
	}
	getUrl := func(id string) string {
		if gf := filesByID[id]; gf != nil {
			return relURL(currentGF.outRelPath, gf.outRelPath)
		}
		return ""
	}

	var depIDs, parentIDs []string
	if len(graphIdx) > 0 && graphIdx[0] != nil {
		depIDs = graphIdx[0].GetDependents(currentGF.guide.ID)
		parentIDs = graphIdx[0].GetParentGuides(currentGF.guide.ID)
	} else if len(filesByID) > 0 {
		guidesMap := make(map[string]graph.Guide, len(filesByID))
		for id, gf := range filesByID {
			if gf != nil {
				guidesMap[id] = gf.guide
			}
		}
		idx := graph.BuildIndex(guidesMap, config)
		depIDs = idx.GetDependents(currentGF.guide.ID)
		parentIDs = idx.GetParentGuides(currentGF.guide.ID)
	}

	return BuildGuideMermaid(currentGF.guide, getScope, MermaidOptions{
		MarkdownWrap:     false,
		RequireNeighbors: true,
		ClickURLFunc:     getUrl,
		Dependents:       depIDs,
		ParentGuides:     parentIDs,
	})
}

// GenerateGuideMermaidText generates a localized Mermaid diagram markdown code block
// for a specific guide in a knowledge graph (without click URLs) in open monochrome academic LR layout.
func GenerateGuideMermaidText(guideID string, guides map[string]graph.Guide, config graph.Manifest) (string, error) {
	guide, exists := guides[guideID]
	if !exists {
		return "", fmt.Errorf("guide '%s' not found in graph", guideID)
	}
	getScope := func(id string) string {
		if g, ok := guides[id]; ok {
			return g.Metadata.Scope
		}
		return ""
	}
	idx := graph.BuildIndex(guides, config)
	return BuildGuideMermaid(guide, getScope, MermaidOptions{
		MarkdownWrap:     true,
		RequireNeighbors: false,
		Dependents:       idx.GetDependents(guideID),
		ParentGuides:     idx.GetParentGuides(guideID),
	}), nil
}
