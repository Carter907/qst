---
scope: component
prerequisites: []
sub_guides:
  - guide: Guide and Manifest Data Models
    adherence: strict
    segment: 6-8
  - guide: Line-Preserving Frontmatter Parser
    adherence: detailed
    segment: 10-16
tags: [graph, parser, mutations, refactoring]
---

# Graph Parsing and Cascading Mutations

The graph persistence layer in `internal/graph/parser.go` and `internal/graph/mutations.go` bridges in-memory domain representations with atomic, reference-safe filesystem operations.

## Directory Discovery and Model Hydration

When `ParseGraph(dirPath)` walks a knowledge graph directory, it skips hidden directories (`.git`) and compiled output (`public/`), hydrating each `.md` file into a `Guide` struct keyed by its filename stem (`Guide.ID`). If two guides in nested subdirectories share the same filename stem, `ParseGraph` immediately aborts with a duplicate guide ID error (`duplicate guide ID '%s': found in %s and %s`) so graph references remain globally unambiguous.

## Reference-Safe Cascading Graph Mutations

Because guide relationships are stored distributedly across individual Markdown frontmatter blocks and `manifest.yaml` tours, `internal/graph/mutations.go` provides high-level transactional operations that maintain referential integrity across the entire directory:

* **Edge & Tag Mutations (`AddPrerequisite`, `UpsertSubGuide`, `AddTag`)**: Idempotently insert or update frontmatter slices in-place using `UpdateGuideMetadata`, preserving original CRLF/LF line endings and body line numbers.
* **Cascading Guide Deletion (`DeleteGuide`)**: Before removing `targetGuide.Path` from disk, `DeleteGuide` scans every other guide's `Prerequisites` and `SubGuides` as well as every tour in `manifest.yaml`, pruning all incoming references and returning the exact `prunedCount`.
* **Cascading Guide Rename (`RenameGuide`)**: Verifies the old ID exists and the new ID does not collide, rewrites every incoming prerequisite, sub-guide, and tour reference across the repository to `newID`, and renames the underlying `.md` file via `os.Rename`.
