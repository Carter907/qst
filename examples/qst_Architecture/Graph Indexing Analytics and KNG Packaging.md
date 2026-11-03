---
scope: component
prerequisites:
  - Structural Hierarchy and Acyclic Validation
sub_guides:
  - guide: Inverted Graph Index and Degree Metrics
    adherence: strict
    segment: 6-10
  - guide: Deflate Zip Packaging and Zip Slip Guard
    adherence: detailed
    segment: 12-14
tags: [graph, index, analytics, archive, kng]
---

# Graph Indexing Analytics and KNG Packaging

Once a knowledge graph is parsed and validated, `internal/graph` provides two higher-level capabilities: structural graph indexing/analytics (`index.go`, `analytics.go`) and lossless `.kng` archive packaging (`archive.go`).

## Bidirectional Graph Indexing and Topology Analytics

Because frontmatter only encodes forward outgoing links, `BuildIndex` constructs an in-memory `GraphIndex` mapping every guide to its incoming `Dependents` (reverse prerequisites), `ParentGuides` (reverse sub-guides), containing `Tours`, and vertex degrees (`InDegree`, `OutDegree`, `IsOrphan`). This inverted index is shared across:
* **`qst stats` & `qst inspect`**: Combines `GraphIndex` degree metrics with `LongestPrerequisiteChain` and `MaxSubguideDepth` to report orphan nodes, scope distributions, and critical prerequisite paths.
* **Static Site Compiler (`internal/ssg`)**: Populates bidirectional "Dependents" and "Parent Guides" cards on every compiled HTML page in $O(1)$ lookup time.

## Lossless `.kng` Distribution Lifecycle

To distribute a knowledge graph as a single portable artifact, `qst form` first runs `ParseGraph`, `ParseConfig`, and `ValidateGraph` as a mandatory gate before calling `ArchiveGraph` to compress all manifests, guides, diagrams, and images into a `.kng` Deflate zip bundle. When a learner runs `qst free` (or `qst serve bundle.kng`), `UnarchiveGraph` verifies every entry against Zip Slip traversal before reconstructing the directory tree losslessly.
