package graph

// LongestPrerequisiteChain returns the longest sequential chain of prerequisites in the knowledge graph.
func LongestPrerequisiteChain(guides map[string]Guide) []string {
	memo := make(map[string][]string)
	visiting := make(map[string]bool)

	var getChain func(id string) []string
	getChain = func(id string) []string {
		if chain, ok := memo[id]; ok {
			return chain
		}
		if visiting[id] {
			return []string{id}
		}
		visiting[id] = true
		defer func() { visiting[id] = false }()

		guide, ok := guides[id]
		if !ok {
			return nil
		}

		var bestParentChain []string
		for _, p := range guide.Metadata.Prerequisites {
			if _, exists := guides[p]; !exists {
				continue
			}
			pChain := getChain(p)
			if len(pChain) > len(bestParentChain) {
				bestParentChain = pChain
			}
		}

		res := append(append([]string(nil), bestParentChain...), id)
		memo[id] = res
		return res
	}

	var longest []string
	for id := range guides {
		chain := getChain(id)
		if len(chain) > len(longest) {
			longest = chain
		}
	}
	return longest
}

// MaxSubguideDepth returns the maximum hierarchical nesting depth across all subguides in the graph.
func MaxSubguideDepth(guides map[string]Guide) int {
	hasSubguides := false
	for _, g := range guides {
		if len(g.Metadata.SubGuides) > 0 {
			hasSubguides = true
			break
		}
	}
	if !hasSubguides {
		return 0
	}

	memo := make(map[string]int)
	visiting := make(map[string]bool)

	var getDepth func(id string) int
	getDepth = func(id string) int {
		if d, ok := memo[id]; ok {
			return d
		}
		if visiting[id] {
			return 1
		}
		visiting[id] = true
		defer func() { visiting[id] = false }()

		guide, ok := guides[id]
		if !ok {
			return 0
		}

		maxChildDepth := 0
		for _, s := range guide.Metadata.SubGuides {
			if _, exists := guides[s.Guide]; !exists {
				continue
			}
			childDepth := getDepth(s.Guide)
			if childDepth > maxChildDepth {
				maxChildDepth = childDepth
			}
		}

		res := 1 + maxChildDepth
		memo[id] = res
		return res
	}

	maxDepth := 0
	for id := range guides {
		d := getDepth(id)
		if d > maxDepth {
			maxDepth = d
		}
	}
	return maxDepth
}
