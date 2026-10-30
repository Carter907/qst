package graph

import (
	"fmt"
	"slices"
	"sort"
)

// PathOptions configures graph traversal and filtering for learning paths.
type PathOptions struct {
	IncludeSubguides bool // Include vertical subguides as dependencies (default: true)
	PrereqsOnly      bool // Only follow horizontal prerequisites (overrides IncludeSubguides)
	DirectOnly       bool // Only include direct 1-hop prerequisites (not transitive closure)
}

// PathStep represents an individual guide within an ordered learning sequence.
type PathStep struct {
	Index        int      `json:"index"`                   // 1-based sequential step order
	GuideID      string   `json:"guide_id"`                // Guide name / identifier
	Scope        string   `json:"scope"`                   // Scope taxonomy (e.g. definition, explanation)
	ScopeLevel   int      `json:"scope_level"`             // 1-indexed hierarchy level from manifest (1 = base)
	SegmentStart int      `json:"segment_start,omitempty"` // Start line in encompassing parent, if applicable
	Dependencies []string `json:"dependencies,omitempty"` // Guides in path that must be completed prior
	RequiredBy   []string `json:"required_by,omitempty"`   // Guides in path that directly require this guide
}

// ScopeGroup groups learning path steps by their manifest scope tier.
type ScopeGroup struct {
	Scope      string     `json:"scope"`
	ScopeLevel int        `json:"scope_level"`
	Guides     []PathStep `json:"guides"`
}

// LearningPath holds the ordered result of a path discovery or topological sort.
type LearningPath struct {
	Target string       `json:"target,omitempty"` // Target guide (empty for global graph)
	Steps  []PathStep   `json:"steps"`            // Linear sequential learning path
	Scopes []ScopeGroup `json:"scopes,omitempty"` // Steps partitioned by scope tier
}

// TopologicalSort performs a deterministic topological sort across all guides in the knowledge graph.
func TopologicalSort(guides map[string]Guide, config Manifest, opts PathOptions) (*LearningPath, error) {
	if len(guides) == 0 {
		return &LearningPath{}, nil
	}

	return solveTopological(guides, guides, config, opts, "")
}

// PrerequisitePath computes the minimal ordered learning path required to master targetID.
func PrerequisitePath(targetID string, guides map[string]Guide, config Manifest, opts PathOptions) (*LearningPath, error) {
	targetGuide, exists := guides[targetID]
	if !exists {
		return nil, fmt.Errorf("guide '%s' not found in knowledge graph", targetID)
	}

	// 1. Discover the upstream dependency closure
	closure := make(map[string]Guide)
	closure[targetID] = targetGuide

	queue := []string{targetID}
	visited := map[string]bool{targetID: true}

	for len(queue) > 0 {
		currID := queue[0]
		queue = queue[1:]
		currGuide := guides[currID]

		var upstream []string
		upstream = append(upstream, currGuide.Metadata.Prerequisites...)

		if !opts.PrereqsOnly {
			for _, sub := range currGuide.Metadata.SubGuides {
				upstream = append(upstream, sub.Guide)
			}
		}

		for _, depID := range upstream {
			depGuide, exists := guides[depID]
			if !exists {
				continue
			}
			if !visited[depID] {
				visited[depID] = true
				closure[depID] = depGuide
				if !opts.DirectOnly {
					queue = append(queue, depID)
				}
			}
		}
	}

	return solveTopological(closure, guides, config, opts, targetID)
}

// UnlocksPath discovers all downstream guides that targetID directly or transitively unlocks.
func UnlocksPath(targetID string, guides map[string]Guide, config Manifest, opts PathOptions) (*LearningPath, error) {
	targetGuide, exists := guides[targetID]
	if !exists {
		return nil, fmt.Errorf("guide '%s' not found in knowledge graph", targetID)
	}

	idx := BuildIndex(guides, config)

	// Discover forward downstream closure
	closure := make(map[string]Guide)
	closure[targetID] = targetGuide

	queue := []string{targetID}
	visited := map[string]bool{targetID: true}

	for len(queue) > 0 {
		currID := queue[0]
		queue = queue[1:]

		var downstream []string
		downstream = append(downstream, idx.GetDependents(currID)...)
		if !opts.PrereqsOnly {
			downstream = append(downstream, idx.GetParentGuides(currID)...)
		}

		for _, downID := range downstream {
			downGuide, exists := guides[downID]
			if !exists {
				continue
			}
			if !visited[downID] {
				visited[downID] = true
				closure[downID] = downGuide
				if !opts.DirectOnly {
					queue = append(queue, downID)
				}
			}
		}
	}

	return solveTopological(closure, guides, config, opts, "")
}

// solveTopological runs Kahn's algorithm on the induced subgraph with deterministic tie-breaking.
func solveTopological(subgraph map[string]Guide, allGuides map[string]Guide, config Manifest, opts PathOptions, targetID string) (*LearningPath, error) {
	// Construct the directed progression graph within the subgraph:
	// If A is required by B (A -> B), A must be studied before B.
	dependents := make(map[string][]string, len(subgraph)) // dependency -> slice of downstream guides
	inDegree := make(map[string]int, len(subgraph))        // guide -> number of unmet dependencies in subgraph
	dependencies := make(map[string][]string, len(subgraph))

	for id := range subgraph {
		inDegree[id] = 0
	}

	for id, g := range subgraph {
		var prereqs []string
		prereqs = append(prereqs, g.Metadata.Prerequisites...)
		if !opts.PrereqsOnly {
			for _, sub := range g.Metadata.SubGuides {
				prereqs = append(prereqs, sub.Guide)
			}
		}

		for _, p := range prereqs {
			if _, inSub := subgraph[p]; inSub {
				dependents[p] = append(dependents[p], id)
				dependencies[id] = append(dependencies[id], p)
				inDegree[id]++
			}
		}
	}

	// Pre-calculate narrative ordering metrics (segment start lines and list index)
	type parentNarrative struct {
		startLine int
		listIndex int
	}
	subguideNarrative := make(map[string]parentNarrative)
	for _, parentGuide := range allGuides {
		for idx, s := range parentGuide.Metadata.SubGuides {
			startLine := 0
			if ranges, err := ParseSegments(s.Segment); err == nil && len(ranges) > 0 {
				startLine = ranges[0][0]
			}
			if existing, ok := subguideNarrative[s.Guide]; !ok || (startLine > 0 && (existing.startLine == 0 || startLine < existing.startLine)) {
				subguideNarrative[s.Guide] = parentNarrative{
					startLine: startLine,
					listIndex: idx,
				}
			}
		}
	}

	// Comparator for deterministic tie-breaking:
	// 1. Lower scope level (foundations first)
	// 2. Segment start line in encompassing parent (narrative flow)
	// 3. Subguide declaration index in parent
	// 4. Alphabetical by Guide ID
	less := func(a, b string) bool {
		scopeA := GetScopeValue(subgraph[a].Metadata.Scope, config)
		scopeB := GetScopeValue(subgraph[b].Metadata.Scope, config)
		if scopeA != scopeB {
			return scopeA < scopeB
		}

		nA, hasA := subguideNarrative[a]
		nB, hasB := subguideNarrative[b]
		if hasA && hasB {
			if nA.startLine > 0 && nB.startLine > 0 && nA.startLine != nB.startLine {
				return nA.startLine < nB.startLine
			}
			if nA.listIndex != nB.listIndex {
				return nA.listIndex < nB.listIndex
			}
		}

		return a < b
	}

	// Initial ready queue: all nodes with inDegree == 0
	var ready []string
	for id, deg := range inDegree {
		if deg == 0 {
			ready = append(ready, id)
		}
	}

	var orderedIDs []string
	for len(ready) > 0 {
		sort.Slice(ready, func(i, j int) bool {
			return less(ready[i], ready[j])
		})

		curr := ready[0]
		ready = ready[1:]
		orderedIDs = append(orderedIDs, curr)

		for _, downstream := range dependents[curr] {
			inDegree[downstream]--
			if inDegree[downstream] == 0 {
				ready = append(ready, downstream)
			}
		}
	}

	if len(orderedIDs) < len(subgraph) {
		return nil, fmt.Errorf("cycle detected in graph: unable to topologically order %d guides", len(subgraph)-len(orderedIDs))
	}

	// Build PathSteps with detailed metadata
	steps := make([]PathStep, 0, len(orderedIDs))
	for i, id := range orderedIDs {
		g := subgraph[id]
		scopeLevel := GetScopeValue(g.Metadata.Scope, config)

		deps := dependencies[id]
		slices.Sort(deps)

		reqBy := dependents[id]
		slices.Sort(reqBy)

		step := PathStep{
			Index:        i + 1,
			GuideID:      id,
			Scope:        g.Metadata.Scope,
			ScopeLevel:   scopeLevel,
			SegmentStart: subguideNarrative[id].startLine,
			Dependencies: deps,
			RequiredBy:   reqBy,
		}
		steps = append(steps, step)
	}

	// Build ScopeGroups partitioned by manifest scopes
	scopeMap := make(map[string][]PathStep)
	for _, step := range steps {
		scopeMap[step.Scope] = append(scopeMap[step.Scope], step)
	}

	var scopeGroups []ScopeGroup
	for _, s := range config.Scopes {
		if gSteps, ok := scopeMap[s]; ok && len(gSteps) > 0 {
			scopeGroups = append(scopeGroups, ScopeGroup{
				Scope:      s,
				ScopeLevel: GetScopeValue(s, config),
				Guides:     gSteps,
			})
		}
	}
	// Catch any guides with unlisted or empty scopes
	for scopeName, gSteps := range scopeMap {
		if !slices.Contains(config.Scopes, scopeName) && len(gSteps) > 0 {
			scopeGroups = append(scopeGroups, ScopeGroup{
				Scope:      scopeName,
				ScopeLevel: 0,
				Guides:     gSteps,
			})
		}
	}

	return &LearningPath{
		Target: targetID,
		Steps:  steps,
		Scopes: scopeGroups,
	}, nil
}
