package ssg

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// generateIndexMermaid produces a Mermaid diagram definition representing the
// complete knowledge graph for embedding on the home page (index.html).
func generateIndexMermaid(guideFiles []*guideFile, filesByID map[string]*guideFile) string {
	if len(guideFiles) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("graph TD\n")

	// 1. Collect and sort all guide IDs for deterministic output
	var ids []string
	for _, gf := range guideFiles {
		ids = append(ids, gf.guide.ID)
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

	// 3. Declare nodes
	for _, id := range ids {
		nodeID := nodeIDs[id]
		label := id
		escapedLabel := strings.ReplaceAll(label, "\"", "#quot;")
		fmt.Fprintf(&sb, "    %s[\"%s\"]\n", nodeID, escapedLabel)
	}

	// 4. Edges (Prerequisites: solid arrows; Subguides: dotted arrows)
	seenEdges := make(map[string]bool)
	var edgeLines []string
	for _, id := range ids {
		gf := filesByID[id]
		if gf == nil {
			continue
		}
		nodeID := nodeIDs[id]

		// Solid links: Prereq -> Current
		for _, prereq := range gf.guide.Metadata.Prerequisites {
			prereqID, exists := nodeIDs[prereq]
			if exists {
				edge := fmt.Sprintf("    %s --> %s", prereqID, nodeID)
				if !seenEdges[edge] {
					seenEdges[edge] = true
					edgeLines = append(edgeLines, edge)
				}
			}
		}

		// Dotted links: Current -> Subguide
		for _, sub := range gf.guide.Metadata.SubGuides {
			subID, exists := nodeIDs[sub.Guide]
			if exists {
				edge := fmt.Sprintf("    %s -.-> %s", nodeID, subID)
				if !seenEdges[edge] {
					seenEdges[edge] = true
					edgeLines = append(edgeLines, edge)
				}
			}
		}
	}
	sort.Strings(edgeLines)
	for _, edge := range edgeLines {
		sb.WriteString(edge)
		sb.WriteString("\n")
	}

	// 5. Click directives for interactive navigation
	var clickLines []string
	for _, id := range ids {
		gf := filesByID[id]
		if gf == nil {
			continue
		}
		nodeID := nodeIDs[id]
		targetURL := relURL("index.html", gf.outRelPath)
		u := (&url.URL{Path: targetURL}).EscapedPath()
		clickLines = append(clickLines, fmt.Sprintf("    click %s href \"%s\"", nodeID, u))
	}
	sort.Strings(clickLines)
	for _, click := range clickLines {
		sb.WriteString(click)
		sb.WriteString("\n")
	}

	return sb.String()
}

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
