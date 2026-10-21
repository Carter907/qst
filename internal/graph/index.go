package graph

// GraphIndex provides an inverted reference index and edge metrics for a knowledge graph.
type GraphIndex struct {
	Guides       map[string]Guide
	Dependents   map[string][]string // Guide ID -> incoming prereq guides (dependents)
	ParentGuides map[string][]string // Guide ID -> incoming subguides (parents)
	Tours        map[string][]string // Guide ID -> tour names containing this guide
	inDegree     map[string]int
	outDegree    map[string]int
}

// BuildIndex constructs a GraphIndex from a map of guides and an optional Manifest.
func BuildIndex(guides map[string]Guide, config Manifest) *GraphIndex {
	idx := &GraphIndex{
		Guides:       guides,
		Dependents:   make(map[string][]string, len(guides)),
		ParentGuides: make(map[string][]string, len(guides)),
		Tours:        make(map[string][]string),
		inDegree:     make(map[string]int, len(guides)),
		outDegree:    make(map[string]int, len(guides)),
	}

	for id, guide := range guides {
		for _, p := range guide.Metadata.Prerequisites {
			if _, exists := guides[p]; exists {
				idx.Dependents[p] = append(idx.Dependents[p], id)
				idx.inDegree[id]++
				idx.outDegree[p]++
			}
		}
		for _, s := range guide.Metadata.SubGuides {
			if _, exists := guides[s.Guide]; exists {
				idx.ParentGuides[s.Guide] = append(idx.ParentGuides[s.Guide], id)
				idx.inDegree[s.Guide]++
				idx.outDegree[id]++
			}
		}
	}

	for _, tour := range config.Tours {
		for _, g := range tour.Guides {
			if _, exists := guides[g]; exists {
				idx.Tours[g] = append(idx.Tours[g], tour.Name)
			}
		}
	}

	return idx
}

// GetDependents returns guides that have targetID as a prerequisite.
func (idx *GraphIndex) GetDependents(targetID string) []string {
	if idx == nil || idx.Dependents == nil {
		return nil
	}
	return idx.Dependents[targetID]
}

// GetParentGuides returns guides that have targetID as a subguide.
func (idx *GraphIndex) GetParentGuides(targetID string) []string {
	if idx == nil || idx.ParentGuides == nil {
		return nil
	}
	return idx.ParentGuides[targetID]
}

// GetTours returns tour names containing targetID.
func (idx *GraphIndex) GetTours(targetID string) []string {
	if idx == nil || idx.Tours == nil {
		return nil
	}
	return idx.Tours[targetID]
}

// InDegree returns the incoming edge count for a guide.
func (idx *GraphIndex) InDegree(targetID string) int {
	if idx == nil || idx.inDegree == nil {
		return 0
	}
	return idx.inDegree[targetID]
}

// OutDegree returns the outgoing edge count for a guide.
func (idx *GraphIndex) OutDegree(targetID string) int {
	if idx == nil || idx.outDegree == nil {
		return 0
	}
	return idx.outDegree[targetID]
}

// IsOrphan returns true if a guide has no incoming or outgoing prerequisite or subguide edges.
func (idx *GraphIndex) IsOrphan(targetID string) bool {
	return idx.InDegree(targetID) == 0 && idx.OutDegree(targetID) == 0
}
