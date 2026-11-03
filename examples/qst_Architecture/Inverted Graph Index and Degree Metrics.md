---
scope: primitive
prerequisites:
  - Guide and Manifest Data Models
sub_guides: []
tags: [graph, index, analytics, dag]
---

# Inverted Graph Index and Degree Metrics

While a `Guide`'s frontmatter stores *outgoing* edges (`Prerequisites` and `SubGuides`), rendering back-links in the static site generator and computing graph diagnostics require fast *incoming* edge lookups. `internal/graph/index.go` and `internal/graph/analytics.go` provide an inverted reference index and memoized DAG traversal metrics.

## Inverted Reference Index (`GraphIndex`)

`BuildIndex(guides map[string]Guide, config Manifest) *GraphIndex` performs a single pass over all guides and tours to populate reverse adjacency maps and vertex degree counts:

```go
type GraphIndex struct {
	Guides       map[string]Guide
	Dependents   map[string][]string // Guide ID -> guides that require this guide as a prereq
	ParentGuides map[string][]string // Guide ID -> guides that include this guide as a sub_guide
	Tours        map[string][]string // Guide ID -> tour names containing this guide
	inDegree     map[string]int
	outDegree    map[string]int
}
```

* **Edge Direction Accounting**:
  * For a horizontal prerequisite `p` declared on `id`, `p` gains `id` in `Dependents[p]`, `inDegree[id]` increments, and `outDegree[p]` increments.
  * For a vertical sub-guide `s.Guide` declared on `id`, `s.Guide` gains `id` in `ParentGuides[s.Guide]`, `inDegree[s.Guide]` increments, and `outDegree[id]` increments.
* **Orphan Detection (`IsOrphan`)**: A guide is classified as an orphan when `InDegree(id) == 0 && OutDegree(id) == 0`—meaning it participates in neither horizontal prerequisite chains nor vertical scope decompositions.

## Memoized DAG Path Analytics (`analytics.go`)

For `qst stats`, two dynamic programming traversals compute structural depth across the graph:
1. **`LongestPrerequisiteChain(guides)`**: Uses a `memo map[string][]string` and a `visiting map[string]bool` cycle guard to compute the longest sequential horizontal prerequisite path ending at each node in $O(V + E)$ time.
2. **`MaxSubguideDepth(guides)`**: Recursively evaluates `1 + maxChildDepth` across `SubGuides` with memoization to report the deepest vertical decomposition hierarchy in the graph.
