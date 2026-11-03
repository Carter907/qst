---
scope: component
prerequisites: []
sub_guides:
  - guide: Testable Cobra Output Channels and Exit Interception
    adherence: strict
    segment: 6-11
  - guide: Interactive Terminal Forms with Huh
    adherence: detailed
    segment: 13-17
tags: [cli, cobra, huh, authoring]
---

# CLI Command Architecture and Dual-Mode UX

The `cmd/` package exposes the complete graph authoring and mutation surface (`new`, `add`, `link`, `unlink`, `tag`, `tour`, `rename`, `delete`, `form`, `free`) through a unified Cobra command hierarchy engineered for both scriptability and interactive ergonomics.

## Testable Command Execution Contract

Every command in `cmd/` adheres to a strict execution contract:
* All standard output is written through `cmd.Printf` / `cmd.Println` and all errors through `cmd.PrintErrf` / `cmd.PrintErrln`.
* Fatal errors invoke `exitFunc(1)` followed by `return`, allowing `cmd/commands_test.go` to intercept exit codes in-memory via `executeCommandWithExit` and verify all error paths without spawning external subprocesses.
* Guide target arguments pass through `cleanGuideName` (`cmd/add.go`), allowing users to specify guides interchangeably by bare ID (`"Printf Function"`), filename (`"Printf Function.md"`), or relative file path (`examples/C/Printf Function.md`).

## Dual-Mode Scriptable and Interactive Authoring

Commands inspect `cmd.Flags().Changed(...)` and positional arguments to determine whether to run in **non-interactive batch mode** or launch an **interactive `charm.land/huh/v2` TUI form**:
* When flags are provided (e.g., `qst link GuideA --prereq GuideB` or `qst add "Intro" -s primitive -t core`), the command validates and applies the mutation immediately for CI and shell scripts.
* When flags are omitted (e.g., running `qst new`, `qst add`, `qst link`, or `qst tour` with no arguments), the command parses the live knowledge graph and presents multi-step interactive selectors and inline segment validators.
