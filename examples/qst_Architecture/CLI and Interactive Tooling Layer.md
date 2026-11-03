---
scope: subsystem
prerequisites:
  - Knowledge Graph Core Engine
  - Static Site Compiler and Live Server
sub_guides:
  - guide: CLI Command Architecture and Dual-Mode UX
    adherence: strict
    segment: 6-8
  - guide: Graph Inspection Diagnostics and CI Documentation Parity
    adherence: detailed
    segment: 10-12
tags: [cli, subsystem, cobra, huh, docgen]
---

# CLI and Interactive Tooling Layer

The `cmd/` and `internal/tools/docgen` packages form the user-facing presentation and automation layer of `qst`, uniting scriptable CLI subcommands, interactive terminal forms, diagnostic inspectors, and CI documentation verification.

## 1. Testable Cobra Hierarchy and Dual-Mode Interactive UX

Every command in `cmd/` is registered on `rootCmd` and adheres to a strict testability architecture: all output passes through `cmd.Printf` and `cmd.PrintErrf`, and fatal exits call an interceptable `exitFunc(1)` hook so `cmd/commands_test.go` can test every command in-process with zero state leakage. Authoring commands (`new`, `add`, `link`, `unlink`, `tour`) operate in dual mode—executing immediately when CLI flags are supplied or launching dynamic `charm.land/huh/v2` TUI forms populated from live graph state when flags are omitted.

## 2. Graph Introspection, Visual Export, and CI Documentation Parity

Read-only diagnostic commands (`inspect`, `coverage`, `stats`, `list`, `view`, `validate`) leverage `internal/graph` indexing and `internal/ssg` Mermaid generation to render line-by-line sub-guide coverage bars, orphan and depth metrics, and Mermaid graph exports. Finally, `internal/tools/docgen` compiles deterministic Markdown reference docs (`docs/cli/`) from the Cobra command tree, enforced in GitHub Actions CI alongside `errcheck`, `go vet`, and `TestExamplesDirectoryValid`.
