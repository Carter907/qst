---
scope: subsystem
prerequisites: []
sub_guides:
  - guide: Graph Parsing and Cascading Mutations
    adherence: strict
    segment: 6-8
  - guide: Structural Hierarchy and Acyclic Validation
    adherence: strict
    segment: 10-12
  - guide: Graph Indexing Analytics and KNG Packaging
    adherence: detailed
    segment: 14-16
tags: [graph, subsystem, engine, dag]
---

# Knowledge Graph Core Engine

The `internal/graph` package forms the foundational domain subsystem of `qst`. It is deliberately decoupled from both terminal UI dependencies and HTML rendering templates so it can serve as a pure, fast, and deterministic graph library.

## 1. Line-Preserving Parsing and Cascading Persistence

At the persistence boundary (`models.go`, `parser.go`, `mutations.go`), `internal/graph` treats Markdown files on disk as the single source of truth. By deriving guide IDs from filenames, measuring `LineCount` strictly after the closing `---` frontmatter delimiter, and performing surgical YAML line splicing in `UpdateGuideMetadata`, mutations (`AddPrerequisite`, `UpsertSubGuide`, `RenameGuide`, `DeleteGuide`) cascade reference updates across all guides and `manifest.yaml` tours without ever altering Markdown prose or shifting body line numbers.

## 2. Two-Dimensional DAG Validation and Pre-Flight Simulation

In `validator.go`, `ValidateGraph` enforces the structural grammar of a `qst` knowledge graph: horizontal `prerequisites` must share identical scopes (`prereq.Scope == guide.Scope`), vertical `sub_guides` must step down the scope hierarchy (`subScope < guideScope` with `relaxed_subguides` and `require_subguides` controls), `segment` line ranges must not overlap (`validateCoverage`), and the combined horizontal+vertical graph must pass three-color DFS cycle detection (`CheckAcyclic`). Pre-flight simulation helpers (`ValidatePrerequisiteLink`, `ValidateSubGuideLink`) verify these invariants in memory before CLI mutations touch the disk.

## 3. Inverted Indexing, DAG Analytics, and `.kng` Archiving

To support fast reverse queries and portable distribution, `BuildIndex` (`index.go`) constructs an in-memory `GraphIndex` of incoming `Dependents`, `ParentGuides`, `Tours`, and vertex degrees alongside memoized path analytics (`LongestPrerequisiteChain`, `MaxSubguideDepth` in `analytics.go`). Finally, `ArchiveGraph` and `UnarchiveGraph` (`archive.go`) losslessly package and extract `.kng` Deflate zip bundles with cross-platform path normalization and Zip Slip directory traversal protection.
