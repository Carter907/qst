# Codebase Reformation & Simplification Roadmap

This document tracks active refactoring, deduplication, and architectural simplification tasks across the `qst` codebase. The focus of this roadmap is reducing complexity, eliminating duplicate logic, and lowering the barrier to entry for new contributors without expanding feature scope.

---

## 1. CLI Architecture & Command Simplification (`cmd/`)

- [x] **Eliminate Manual Flag Slice Resets & In-Place Flag Mutation**:
  - *Location*: `cmd/*.go`, `cmd/commands_test.go`
  - *Details*: Avoid mutating package-level flag variables in place inside command handlers and recursively reset `pflag.SliceValue` and subcommand flags in `commands_test.go`, eliminating manual `defer` slice resets.
- [x] **Standardize Error Handling with `runWithExit` Wrapper**:
  - *Location*: `cmd/root.go`, `cmd/*.go`
  - *Details*: Replaced repetitive `cmd.PrintErrf(...)` + `exitFunc(1)` + `return` boilerplate across commands with a centralized `runWithExit` error-returning adapter while preserving test exit-code interception.
- [x] **Consolidate Interactive Terminal Forms (`cmd/forms.go`)**:
  - *Location*: `cmd/forms.go`, `cmd/link.go`, `cmd/unlink.go`, `cmd/add.go`
  - *Details*: Extracted `charm.land/huh/v2` interactive prompt builders out of individual command `Run` functions into `cmd/forms.go` so command handlers remain concise and linear.
- [x] **Unify Global Mermaid Generation in `qst view`**:
  - *Location*: `cmd/view.go`, `internal/ssg/mermaid.go`
  - *Details*: Removed the standalone `generateMermaid` implementation in `cmd/view.go` and delegated global graph diagram generation to `ssg.GenerateGlobalMermaid` so node ID sanitization (`cleanMermaidID`) and collision handling live in a single place.

---

## 2. Graph Core & Validation Consolidation (`internal/graph/`)

- [x] **Use `GraphIndex` in `DeleteGuide` and `RenameGuide`**:
  - *Location*: `internal/graph/mutations.go`, `internal/graph/index.go`
  - *Details*: Refactored `DeleteGuide` and `RenameGuide` to query incoming references via `BuildIndex(guides, config)` (`GetDependents`, `GetParentGuides`, `GetTours`) instead of performing manual $O(N)$ scans across all guides.
- [x] **Deduplicate Relationship Validation Rules**:
  - *Location*: `internal/graph/validator.go`
  - *Details*: Consolidated duplicated scope hierarchy, adherence, and cycle-simulation cloning checks between full-graph validation (`validatePrerequisites`, `validateSubGuides`) and single-link mutation checks (`ValidatePrerequisiteLink`, `ValidateSubGuideLink`).
- [x] **Unify Body Line Counting & EOF Newline Normalization**:
  - *Location*: `internal/graph/validator.go`
  - *Details*: Centralized trailing POSIX newline (`\n` at EOF) detection via `hasTrailingNewline` across `validateCoverage` and `AnalyzeGuideCoverage`.
- [x] **Centralize Ignored Directory & File Rules**:
  - *Location*: `internal/graph/parser.go`, `internal/graph/archive.go`, `internal/ssg/engine.go`, `internal/ssg/watcher.go`
  - *Details*: Unified directory and static asset predicates (`IsIgnoredDir`, `IsStaticAssetExt`) in `internal/graph` so graph parsing, archiving, static site building, and file watching enforce identical traversal boundaries.
- [x] **Move Guide & Manifest Scaffolding into `internal/graph/`**:
  - *Location*: `cmd/add.go`, `cmd/new.go`, `internal/graph/mutations.go`
  - *Details*: Moved YAML frontmatter marshaling and initial guide/manifest creation (`CreateGuide`, `DefaultManifest`, `WriteStarterGuide`) out of `cmd/add.go` and `cmd/new.go` into `internal/graph/mutations.go`.

---

## 3. Static Site Generator Modularization (`internal/ssg/`)

- [x] **Split `internal/ssg/server.go` into Focused Submodules**:
  - *Location*: `internal/ssg/server.go`, `internal/ssg/watcher.go`, `internal/ssg/sse.go`, `internal/ssg/templates/`
  - *Details*: Split the ~825-line `server.go` file into single-responsibility files (`server.go` for HTTP/lifecycle, `watcher.go` for `fsnotify` debounce and directory tracking, `sse.go` for client event broadcasting) and moved inline HTML/JS string literals into `internal/ssg/templates/livereload.html` and `internal/ssg/templates/fallback_index.html`.
- [x] **Reuse `graph.ParseGuidePermissive` & Shared Predicates in SSG Build Pipeline**:
  - *Location*: `internal/ssg/engine.go`, `internal/graph/parser.go`
  - *Details*: Consolidated fallback guide parsing (`ParseGuidePermissive`) and directory/asset predicates (`IsIgnoredDir`, `IsStaticAssetExt`) into `internal/graph` so `BuildSiteWithOptions` delegates guide parsing and file classification to `internal/graph`.

---

## 4. Developer Onboarding & Documentation (`docs/`, `README.md`, `CONTRIBUTING.md`)

- [x] **Surface `examples/qst_Architecture` for Developer Onboarding**:
  - *Location*: `README.md`, `CONTRIBUTING.md`
  - *Details*: Documented the self-hosted `examples/qst_Architecture` knowledge graph in `README.md` and `CONTRIBUTING.md` (`qst serve examples/qst_Architecture --open`) so new contributors can explore `qst`'s subsystem architecture and onboarding tour using `qst` itself.

---

## 5. Topological Path Discovery & Learning Roadmaps (`internal/graph/solver.go`, `cmd/path.go`)

- [x] **Topological Sort & Prerequisite Closure Engine**:
  - *Location*: `internal/graph/solver.go`, `internal/graph/solver_test.go`
  - *Details*: Implemented Kahn's algorithm with deterministic tie-breaking by scope tier, parent narrative order (segment line numbers and declaration indices), and guide IDs. Provides `TopologicalSort`, `PrerequisitePath`, and `UnlocksPath`.
- [x] **Learning Path CLI Command (`qst path`)**:
  - *Location*: `cmd/path.go`, `cmd/commands_test.go`, `docs/cli/qst_path.md`
  - *Details*: Introduced `qst path [guide]` with support for targeted learning paths, global scope-partitioned syllabi, downstream unlock exploration (`--unlocks`), Mermaid diagram output (`--format mermaid`), JSON serialization (`--format json`), and direct tour persistence (`--save-tour`).
