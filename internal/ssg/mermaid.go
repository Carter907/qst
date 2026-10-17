package ssg

import (
	"fmt"
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

// formatGuideNodeLabel formats a guide node label with a capitalized scope prefix
// according to the academic format: "<Scope>: <Title>" (e.g. "Explanation: De Morgan's Laws").
// If no scope is present, it returns the escaped title.
func formatGuideNodeLabel(id string, sc string) string {
	escapedID := strings.ReplaceAll(id, "\"", "#quot;")
	if sc == "" {
		return escapedID
	}
	escapedScope := strings.ReplaceAll(sc, "\"", "#quot;")
	r := []rune(escapedScope)
	capScope := string(unicode.ToUpper(r[0])) + string(r[1:])
	return fmt.Sprintf("%s: %s", capScope, escapedID)
}

// generateGuideMermaid produces a Mermaid diagram representing the localized
// forward-only 1-hop neighborhood for a specific guide (the guide itself,
// its prerequisites, and its subguides) in an open monochrome academic LR layout.
// Neighbor nodes include clickable hyperlinks to their respective guide pages.
func generateGuideMermaid(currentGF *guideFile, config graph.Manifest, filesByID map[string]*guideFile) string {
	if currentGF == nil {
		return ""
	}

	prereqs := currentGF.guide.Metadata.Prerequisites
	subguides := currentGF.guide.Metadata.SubGuides
	if len(prereqs) == 0 && len(subguides) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("graph LR\n")

	// 1. Collect all unique guide IDs in this forward-only 1-hop horizon
	seen := make(map[string]bool)
	seen[currentGF.guide.ID] = true
	for _, p := range prereqs {
		seen[p] = true
	}
	for _, s := range subguides {
		seen[s.Guide] = true
	}

	var ids []string
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// 2. Map guide IDs to unique safe Mermaid node identifiers
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

	// 3. Declare nodes with academic scope prefix
	for _, id := range ids {
		nodeID := nodeIDs[id]
		sc := ""
		if gf := filesByID[id]; gf != nil {
			sc = gf.guide.Metadata.Scope
		}
		label := formatGuideNodeLabel(id, sc)
		fmt.Fprintf(&sb, "    %s[\"%s\"]\n", nodeID, label)
	}
	sb.WriteString("\n")

	// 4. Edges
	seenEdges := make(map[string]bool)
	currentNodeID := nodeIDs[currentGF.guide.ID]

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
	sort.Strings(prereqEdges)
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
	sort.Strings(subEdges)
	for _, edge := range subEdges {
		sb.WriteString(edge)
		sb.WriteString("\n")
	}

	// 5. Click directives for interactive navigation
	var clickLines []string
	for _, id := range ids {
		if id == currentGF.guide.ID {
			continue
		}
		gf := filesByID[id]
		if gf == nil {
			continue
		}
		nodeID := nodeIDs[id]
		targetURL := relURL(currentGF.outRelPath, gf.outRelPath)
		u := (&url.URL{Path: targetURL}).EscapedPath()
		clickLines = append(clickLines, fmt.Sprintf("    click %s href \"%s\"", nodeID, u))
	}
	sort.Strings(clickLines)
	if len(clickLines) > 0 {
		sb.WriteString("\n")
		for _, click := range clickLines {
			sb.WriteString(click)
			sb.WriteString("\n")
		}
	}

	// 6. Monochrome academic styles
	sb.WriteString("\n")
	sb.WriteString("    classDef default fill:#ffffff,stroke:#1f2937,stroke-width:1.2px,rx:3px,ry:3px;\n")
	sb.WriteString("    classDef currentGuide fill:#f9fafb,stroke:#000000,stroke-width:2.5px,font-weight:bold;\n")
	fmt.Fprintf(&sb, "    class %s currentGuide;\n", currentNodeID)

	return sb.String()
}

// GenerateGuideMermaidText generates a localized Mermaid diagram markdown code block
// for a specific guide in a knowledge graph (without click URLs) in open monochrome academic LR layout.
func GenerateGuideMermaidText(guideID string, guides map[string]graph.Guide, config graph.Manifest) (string, error) {
	guide, exists := guides[guideID]
	if !exists {
		return "", fmt.Errorf("guide '%s' not found in graph", guideID)
	}

	prereqs := guide.Metadata.Prerequisites
	subguides := guide.Metadata.SubGuides

	var sb strings.Builder
	sb.WriteString("```mermaid\ngraph LR\n")

	// 1. Collect all unique guide IDs
	seen := make(map[string]bool)
	seen[guideID] = true
	for _, p := range prereqs {
		seen[p] = true
	}
	for _, s := range subguides {
		seen[s.Guide] = true
	}

	var ids []string
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// 2. Map guide IDs to unique safe Mermaid node identifiers
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

	// 3. Declare nodes with academic scope prefix
	for _, id := range ids {
		nodeID := nodeIDs[id]
		sc := ""
		if g, ok := guides[id]; ok {
			sc = g.Metadata.Scope
		}
		label := formatGuideNodeLabel(id, sc)
		fmt.Fprintf(&sb, "    %s[\"%s\"]\n", nodeID, label)
	}
	sb.WriteString("\n")

	// 4. Edges
	seenEdges := make(map[string]bool)
	currentNodeID := nodeIDs[guideID]

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
	sort.Strings(prereqEdges)
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
	sort.Strings(subEdges)
	for _, edge := range subEdges {
		sb.WriteString(edge)
		sb.WriteString("\n")
	}

	// 5. Monochrome academic styles
	sb.WriteString("\n")
	sb.WriteString("    classDef default fill:#ffffff,stroke:#1f2937,stroke-width:1.2px,rx:3px,ry:3px;\n")
	sb.WriteString("    classDef currentGuide fill:#f9fafb,stroke:#000000,stroke-width:2.5px,font-weight:bold;\n")
	fmt.Fprintf(&sb, "    class %s currentGuide;\n", currentNodeID)
	sb.WriteString("```")

	return sb.String(), nil
}

