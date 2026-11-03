---
scope: component
prerequisites:
  - CLI Command Architecture and Dual-Mode UX
  - Graph Indexing Analytics and KNG Packaging
sub_guides:
  - guide: Automated CLI Markdown Documentation Generator
    adherence: strict
    segment: 12-18
tags: [cli, inspect, stats, docgen, ci]
---

# Graph Inspection Diagnostics and CI Documentation Parity

Beyond authoring mutations, `cmd/` and `internal/tools/docgen` provide read-only diagnostic commands (`inspect`, `coverage`, `stats`, `list`, `view`, `validate`) and automated CI quality gates.

## Read-Only Graph Introspection Commands

* **`qst inspect` & `qst coverage` (`cmd/inspect.go`, `cmd/coverage.go`)**: Parses the target guide's sub-guide `segment` ranges (`graph.ParseSegments`) and renders either a compact Unicode progress bar (`[████████░░] 80.0%`) or a line-by-line annotated source listing (`-c` / `--coverage`) showing exactly which child guide covers each line of the parent guide.
* **`qst stats` & `qst list` (`cmd/stats.go`, `cmd/list.go`)**: Uses `graph.BuildIndex`, `LongestPrerequisiteChain`, and `MaxSubguideDepth` to report total guide counts, scope/adherence breakdowns, orphan guides, and filterable guide tables (`--scope`, `--tag`, `--tour`, `--orphans`).
* **`qst view` (`cmd/view.go`)**: Exports either the global DAG (`generateMermaid`) or a localized 1-hop guide subgraph (`--guide` via `ssg.GenerateGuideMermaidText`) as raw Mermaid Markdown to stdout.

## Automated CLI Documentation and Repository Verification

To keep contributor documentation and example graphs continuously verified, the repository enforces a five-stage verification pipeline in `.github/workflows/go.yml`:
1. `go build -v ./...` (compilation check across all packages and tools).
2. `errcheck ./...` and `go vet ./...` (static analysis ensuring zero unhandled `io.Writer` or `fmt.Fprintf` errors).
3. `go run ./internal/tools/docgen -out ./docs/cli -format markdown` followed by `git diff --exit-code docs/cli/` (guaranteeing `docs/cli/*.md` matches the Cobra command definitions in `cmd/`).
4. `go test -count=1 ./...` (executing unit tests alongside `TestExamplesDirectoryValid`, which parses and validates every knowledge graph inside `examples/`).
