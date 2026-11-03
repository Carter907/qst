---
scope: primitive
prerequisites:
  - Testable Cobra Output Channels and Exit Interception
sub_guides: []
tags: [cli, tui, huh, interactive]
---

# Interactive Terminal Forms with Huh

Authoring commands in `cmd/` (`new`, `add`, `link`, `unlink`, `tour`) support a dual-mode user experience: they can be driven non-interactively via CLI flags in scripts, or interactively via terminal UI (TUI) forms powered by `charm.land/huh/v2` when flags are omitted.

## Dynamic Form Construction from Graph State

Rather than presenting static text prompts, commands first parse the target knowledge graph (`graph.ParseGraph` and `graph.ParseConfig`) and dynamically populate `huh.NewSelect` and `huh.NewMultiSelect` options from live graph entities:

* **`qst add` (`cmd/add.go`)**:
  * Populates the **Scope** selector directly from `config.Scopes`.
  * Dynamically filters the **Prerequisites** multi-select so only existing guides matching the chosen scope are offered.
  * Multi-step **Sub-Guides** loop prompts for target child guide, adherence level (`config.Adherences`), and line segment range (`e.g. 10-20`).
* **`qst link` / `qst unlink` (`cmd/link.go`, `cmd/unlink.go`)**:
  * Lets the user choose between horizontal (`prerequisite`) and vertical (`subguide`) edge operations and pre-populates only valid candidate guides.

## Inline Pre-Submission Validation

Interactive fields attach `Validate(func(string) error)` closures that invoke `graph.ParseSegments` or check for duplicate guide/tour names in real time before the form submits, preventing invalid frontmatter from ever being written to disk.
