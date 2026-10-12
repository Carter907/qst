package ssg

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/Carter907/qst/internal/graph"
)

// generateIndexMermaid produces a Mermaid diagram definition representing the
// complete knowledge graph for embedding on the home page (index.html).
// Guides are partitioned into scope subgraphs with horizontal flow (direction LR)
// ordered top-down according to the manifest scope hierarchy.
func generateIndexMermaid(config graph.Manifest, guideFiles []*guideFile, filesByID map[string]*guideFile) string {
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

	// Helper to generate unique subgraph ID
	getSubgraphID := func(sc string) string {
		clean := cleanMermaidID(sc)
		if sc == "" || clean == "node_" {
			clean = "other"
		}
		base := "scope_" + clean
		candidate := base
		counter := 1
		for usedIDs[candidate] {
			counter++
			candidate = fmt.Sprintf("%s_%d", base, counter)
		}
		usedIDs[candidate] = true
		return candidate
	}

	// 3. Group guide IDs by scope (preserves sorted guide ID order)
	guidesByScope := make(map[string][]string)
	for _, id := range ids {
		gf := filesByID[id]
		if gf == nil {
			continue
		}
		sc := gf.guide.Metadata.Scope
		guidesByScope[sc] = append(guidesByScope[sc], id)
	}

	// 4. Order scopes top-down (highest scope level at the top, lowest at bottom)
	var orderedScopes []string
	seenScopes := make(map[string]bool)

	for i := len(config.Scopes) - 1; i >= 0; i-- {
		sc := config.Scopes[i]
		if len(guidesByScope[sc]) > 0 && !seenScopes[sc] {
			orderedScopes = append(orderedScopes, sc)
			seenScopes[sc] = true
		}
	}

	var otherScopes []string
	for sc := range guidesByScope {
		if !seenScopes[sc] {
			otherScopes = append(otherScopes, sc)
		}
	}
	sort.Strings(otherScopes)
	orderedScopes = append(orderedScopes, otherScopes...)

	// 5. Render scope subgraphs with direction LR
	seenEdges := make(map[string]bool)

	for _, sc := range orderedScopes {
		scopeGuideIDs := guidesByScope[sc]
		if len(scopeGuideIDs) == 0 {
			continue
		}

		subgraphID := getSubgraphID(sc)
		scopeLabel := sc
		if scopeLabel == "" {
			scopeLabel = "other"
		}
		escapedScopeLabel := strings.ReplaceAll(scopeLabel, "\"", "#quot;")

		fmt.Fprintf(&sb, "    subgraph %s [\"Scope: %s\"]\n", subgraphID, escapedScopeLabel)
		sb.WriteString("        direction LR\n")

		// Declare nodes in this scope
		for _, id := range scopeGuideIDs {
			nodeID := nodeIDs[id]
			label := id
			escapedLabel := strings.ReplaceAll(label, "\"", "#quot;")
			fmt.Fprintf(&sb, "        %s[\"%s\"]\n", nodeID, escapedLabel)
		}

		// Intra-scope edges: prerequisites between guides in the same scope
		var intraEdges []string
		for _, id := range scopeGuideIDs {
			gf := filesByID[id]
			if gf == nil {
				continue
			}
			nodeID := nodeIDs[id]
			for _, prereq := range gf.guide.Metadata.Prerequisites {
				prereqGF := filesByID[prereq]
				if prereqGF != nil && prereqGF.guide.Metadata.Scope == sc {
					prereqID, exists := nodeIDs[prereq]
					if exists {
						edgeKey := fmt.Sprintf("%s-->%s", prereqID, nodeID)
						if !seenEdges[edgeKey] {
							seenEdges[edgeKey] = true
							intraEdges = append(intraEdges, fmt.Sprintf("        %s --> %s", prereqID, nodeID))
						}
					}
				}
			}
		}
		sort.Strings(intraEdges)
		for _, edge := range intraEdges {
			sb.WriteString(edge)
			sb.WriteString("\n")
		}

		sb.WriteString("    end\n")
	}

	// 6. Cross-scope and hierarchical edges
	var crossEdges []string
	for _, id := range ids {
		gf := filesByID[id]
		if gf == nil {
			continue
		}
		nodeID := nodeIDs[id]
		currentScope := gf.guide.Metadata.Scope

		// Solid links: cross-scope prerequisites (if any)
		for _, prereq := range gf.guide.Metadata.Prerequisites {
			prereqGF := filesByID[prereq]
			if prereqGF == nil || prereqGF.guide.Metadata.Scope != currentScope {
				prereqID, exists := nodeIDs[prereq]
				if exists {
					edgeKey := fmt.Sprintf("%s-->%s", prereqID, nodeID)
					if !seenEdges[edgeKey] {
						seenEdges[edgeKey] = true
						crossEdges = append(crossEdges, fmt.Sprintf("    %s --> %s", prereqID, nodeID))
					}
				}
			}
		}

		// Dotted links: Current -> Subguide (cross-scope / hierarchical)
		for _, sub := range gf.guide.Metadata.SubGuides {
			subID, exists := nodeIDs[sub.Guide]
			if exists {
				edgeKey := fmt.Sprintf("%s-.->%s", nodeID, subID)
				if !seenEdges[edgeKey] {
					seenEdges[edgeKey] = true
					crossEdges = append(crossEdges, fmt.Sprintf("    %s -.-> %s", nodeID, subID))
				}
			}
		}
	}
	sort.Strings(crossEdges)
	if len(crossEdges) > 0 {
		sb.WriteString("\n")
		for _, edge := range crossEdges {
			sb.WriteString(edge)
			sb.WriteString("\n")
		}
	}

	// 7. Click directives for interactive navigation
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
	if len(clickLines) > 0 {
		sb.WriteString("\n")
		for _, click := range clickLines {
			sb.WriteString(click)
			sb.WriteString("\n")
		}
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
