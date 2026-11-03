---
scope: component
prerequisites:
  - Graph Parsing and Cascading Mutations
sub_guides:
  - guide: Three-Color DFS Cycle Detection
    adherence: strict
    segment: 6-12
  - guide: Segment Range Parser and Overlap Detector
    adherence: detailed
    segment: 14-18
tags: [graph, validator, hierarchy, dag]
---

# Structural Hierarchy and Acyclic Validation

In `internal/graph/validator.go`, `ValidateGraph(guides, config)` enforces the mathematical and pedagogical invariants that distinguish a `qst` knowledge graph from an unstructured folder of Markdown notes.

## Two-Dimensional Scope Invariants and Cycle Prevention

Every edge in a `qst` knowledge graph is validated against its scope index (`getScopeValue`, 1-indexed from `config.Scopes`) before `CheckAcyclic` verifies global DAG acyclicity:

1. **Horizontal Scope Equality (`validatePrerequisites`)**: A prerequisite must belong to the *exact same scope* as the dependent guide (`prereq.Metadata.Scope == guide.Metadata.Scope`). Horizontal links represent sequential progression at the same level of abstraction.
2. **Vertical Scope Hierarchy (`validateSubGuides`)**: A sub-guide must have a strictly smaller scope index than its parent (`subScopeVal < guideScopeVal`). When `config.RelaxedSubguides` is `false`, the child scope must be *exactly one tier below* the parent (`guideScopeVal - subScopeVal == 1`).
3. **Non-Leaf Decomposition (`RequireSubguides`)**: When `config.RequireSubguides` is `true`, any guide above the base scope tier (`guideScopeVal > 1`) is required to decompose into at least one sub-guide.

## Segment Coverage and Pre-Flight Link Simulation

For every parent guide, `validateSubGuides` parses each child's `segment` specification via `parseSegments` and passes the aggregated intervals to `validateCoverage` to reject overlapping line ranges or uncovered gaps (when `strict_coverage: true`).

To prevent CLI commands (`qst link`) from writing invalid edges to disk, `ValidatePrerequisiteLink` and `ValidateSubGuideLink` clone the guide map in memory, apply the candidate link, and run `CheckAcyclic` and `validateCoverage` as a pre-flight simulation before mutating any file.
