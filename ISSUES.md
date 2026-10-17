# Project Enhancement Roadmap & Issues

This document tracks completed enhancements and open high-impact tasks across the static site generator, graph core, packaging, and CLI tooling.

---

## Completed Enhancements

### 1. Static Site Generator (`internal/ssg/`)
- [x] **Populate Sidebar Navigation**: Implemented dynamic sidebar grouping guides hierarchically by tours and scopes with collapsible `<details>` sections and active route highlights (`layout.html`, `engine.go`, `style.css`).
- [x] **Sidebar Session Persistence**: Added `localStorage` session state persistence to keep collapsible tour and scope sections synchronized across page reloads and browsing sessions (`layout.html`, `style.css`).
- [x] **Integrate Graph Metadata into Guides**: Guides display metadata bylines (scope, tags) and an academic appendix listing prerequisites and subguides with adherence ratings and segment line numbers (`guide.html`, `engine.go`).
- [x] **Nested Directory & Relative Asset Resolution**: Introduced `RelativeRoot` and `relURL` to compute relative paths for CSS, JS vendor bundles, and cross-guide links, allowing guides at arbitrary directory depths to render without 404s (`engine.go`, `layout.html`).
- [x] **Knowledge Graph Overview on Index**: Automated generation of an interactive Mermaid diagram on `index.html` displaying prerequisites (solid arrows), subguides (dotted arrows), and clickable links to guide pages (`mermaid.go`, `index.html`).
- [x] **Interactive Pan, Zoom & Fullscreen Controls**: Integrated SVG pan/zoom toolbar controls (`panzoom`, zoom in/out, fit-to-view, reset, fullscreen toggle) on the interactive Mermaid knowledge diagram (`layout.html`, `style.css`).
- [x] **Scope-Layered Subgraphs with Horizontal Flow**: Formatted index Mermaid diagram with top-down scope subgraphs containing horizontal (`direction LR`) intra-scope node flow (`mermaid.go`).
- [x] **Academic Styling & Gutter Line Numbering**: Integrated dedicated line markers with hair-line gutter numbering, booktabs-style tables, and serif typography (`style.css`, `engine.go`).
- [x] **Zip Slip / Path Traversal Protection in KaTeX Downloader**: Refactored tar extraction into `extractTar` with directory traversal validation (`strings.HasPrefix(cleanTarget, cleanDest+string(os.PathSeparator))`) to prevent malicious archives from extracting files outside `destDir` (`download.go`, `download_test.go`).
- [x] **Network Robustness & Offline Builds (`qst build --offline`)**: Hardened HTTP requests with 30-second timeouts, added `--offline` flag to compile without network calls, integrated shared cross-platform user caching (`~/.cache/qst/vendor`), and added `--vendor-dir` / `QST_VENDOR_DIR` custom asset sourcing for air-gapped systems (`cmd/build.go`, `internal/ssg/download.go`, `internal/ssg/engine.go`).
- [x] **Guide-Specific Localized Mermaid Subgraphs (`guide.html`, `qst view --guide`)**: Generates forward-only 1-hop localized subgraphs in the guide appendix as clean static academic figure plates with standardized `<Scope>: <Title>` node labels, active guide highlighting (`classDef currentGuide`), relative click navigation, and added CLI localized filtering via `qst view --guide <name>` (`mermaid.go`, `guide.html`, `engine.go`, `cmd/view.go`).
- [x] **Structured Academic Syllabus Homepage (`index.html`)**: Replaced generic home page Mermaid flowchart with an organized, booktabs-style curriculum catalog grouped top-down by scope hierarchy (Lessons → Explanations → Definitions) with guide links, prerequisite badges, and tag chips (`index.html`, `engine.go`, `style.css`).

### 2. Graph Core & Relationship Logic (`internal/graph/`)
- [x] **Flexible Configuration File Naming**: `ParseConfig` now supports falling back from `manifest.yaml` to `quest.yaml` seamlessly (`parser.go`).
- [x] **Relationship Mutation Helpers**: Implemented `AddPrerequisite`, `RemovePrerequisite`, `UpsertSubGuide`, and `RemoveSubGuide` with frontmatter persistence (`parser.go`).
- [x] **Link Validation & Cycle Prevention**: Added `ValidatePrerequisiteLink` and `ValidateSubGuideLink` with scope hierarchy enforcement and acyclicity verification (`validator.go`).
- [x] **Line Coverage Analysis**: Added `AnalyzeGuideCoverage`, `LineCoverage`, and `GuideCoverageReport` to measure subguide coverage across guide markdown bodies (`validator.go`).
- [x] **Single-Line Segment Shorthand & Whitespace Tolerance**: Added support for standalone single line numbers (e.g., `4` equivalent to `4-4`) in subguide segments, and tolerated blank lines and trailing newlines in coverage verification (`validator.go`).
- [x] **Recursive Subdirectory Support in `ParseGraph`**: `ParseGraph` traverses directory trees recursively via `filepath.Walk`, allowing guides to be organized into nested subfolders while ignoring `public/` and hidden folders, and detecting duplicate guide IDs (`parser.go`).
- [x] **Detailed Cycle Path Reporting in `CheckAcyclic`**: `CheckAcyclic` tracks the active DFS traversal stack and outputs the exact loop chain (e.g., `cycle detected in graph: Guide A -> Guide B -> Guide A`) with deterministic ordering (`validator.go`).
- [x] **Taxonomy Tag & Tour Mutations**: Added programmatic API helpers (`AddTag`, `RemoveTag`, `CreateTour`, `DeleteTour`, `AddGuideToTour`, `RemoveGuideFromTour`) with manifest and frontmatter persistence (`parser.go`).
- [x] **Cascading Guide Deletion (`DeleteGuide`)**: Prunes incoming references across prerequisites, subguides, and manifest tours, removing orphaned links safely (`parser.go`).
- [x] **Dot Path Parsing Fix**: Resolved root directory traversal issue when parsing graphs from `.` or relative dot paths (`parser.go`).

### 3. CLI Commands & Tooling (`cmd/`)
- [x] **Reciprocal Relationship Management (`qst unlink`)**: Added `qst unlink` to remove prerequisites or subguides interactively (`-i`) or via flags (`--subguide`/`--sub`, `--prereq`) (`cmd/unlink.go`).
- [x] **Enhanced Linking (`qst link`)**: Extended `qst link` to support horizontal prerequisites (`--prereq`), interactive prompt workflows, batch linking, validation, and in-place upserts (`cmd/link.go`).
- [x] **Guide Body Preview in `qst link`**: Added `--preview` (`-p`) flag to `qst link` displaying numbered body lines directly in terminal output to streamline segment range selection (`cmd/link.go`).
- [x] **Guide Inspector & Coverage Commands (`qst inspect`, `qst coverage`)**: Added `qst inspect` for comprehensive guide metadata and body statistics, with `--coverage` (`-c`) and `qst coverage` providing line-by-line coverage breakdowns (`cmd/inspect.go`, `cmd/coverage.go`).
- [x] **Fix Exit Codes in `qst validate`**: Validation, configuration, and parsing errors now terminate with exit code 1 to ensure CI reliability (`cmd/validate.go`).
- [x] **Deterministic Sorting & Rich Table View in `qst list`**: Guides are sorted alphabetically, with `-l, --long` rendering an aligned tabular view showing scope, tags, and link counts (`cmd/list.go`).
- [x] **Enrich `qst stats` Dashboard**: Displays sectioned statistics including overview metrics, scope distribution, graph depth / longest prerequisite chain, tour coverage, and aggregate body line coverage (`cmd/stats.go`).
- [x] **Safe Guide Deletion (`qst delete`, `qst rm`)**: Added `qst delete` (with `rm` alias) supporting dry-run previews (`--dry-run`, `-n`) and cascading removal of incoming references across prerequisites, subguides, and manifest tours (`cmd/delete.go`).
- [x] **Subguide Flag Replacement & Batch Linking (`qst link`, `qst unlink`)**: Replaced deprecated `--guide` flag with `--subguide` / `--sub`, added batch linking for multiple prerequisites and subguides simultaneously (`cmd/link.go`, `cmd/unlink.go`).
- [x] **Robust Creation in `qst add`**: Enforced valid `--scope`, validated prerequisite existence and scope parity against the graph, added subguide linking flags (`--subguide`, `--adherence`, `--segment`), added body/file import flags (`--body`, `--file`), and standardized error channels (`cmd/add.go`).
- [x] **Tour Management Commands (`qst tour`)**: Added `qst tour` subcommand group (`list`, `create`, `add`, `remove`, `delete`) with manifest persistence and guide existence verification (`cmd/tour.go`).
- [x] **Taxonomy Tag Tooling (`qst tag`)**: Added `qst tag` subcommand group (`list`, `add`, `remove`) with frontmatter persistence (`cmd/tag.go`).
- [x] **Opt-in Strict Coverage Flag in `qst new`**: Added `--strict-coverage` flag to `qst new` to configure `strict_coverage: true` in `manifest.yaml` (defaulting to `false`) (`cmd/new.go`).
- [x] **Diagnostics & Standard Channel Compliance**: Fixed false-positive strict coverage failure on valid leaf guides, harmonized line count reporting, initialized default stdout and stderr on `rootCmd` for test harness interceptability, and converted legacy `fmt.Printf`/`os.Exit` calls across `new`, `rename`, and `add` (`cmd/inspect.go`, `cmd/coverage.go`, `cmd/new.go`, `cmd/rename.go`, `cmd/root.go`, `internal/graph/validator.go`).
- [x] **Standardize Channel & Exit Compliance in `build.go`, `view.go`, and `inspect.go`**: Replaced all remaining legacy `fmt.Printf`, `fmt.Fprintf(os.Stderr, ...)`, and raw `os.Exit(1)` calls across `build.go`, `view.go`, and `inspect.go` with `cmd.Printf`, `cmd.PrintErrf`, and `exitFunc(1)` to satisfy `AGENTS.md` testing standards, and added comprehensive failure interception unit tests (`cmd/build.go`, `cmd/view.go`, `cmd/inspect.go`, `cmd/commands_test.go`).
- [x] **Full CLI Documentation Parity**: Synchronized all newly added commands and options (`qst delete`, `qst tag`, `qst tour`, etc.) into `docs/cli/` via `internal/tools/docgen`.

### 4. Archiving & Packaging (`internal/graph/archive.go`, `cmd/`)
- [x] **Preserve Manifests and Assets in `.kng` Packages**: `ArchiveGraph` recursively walks knowledge graphs using deflate compression, packaging `manifest.yaml` / `quest.yaml`, `.md` guides, diagrams, and media assets while skipping build directories (`public/`), hidden files, and archives (`archive.go`).
- [x] **Lossless Extraction & Nested Directory Support**: `UnarchiveGraph` automatically creates parent directories for nested paths, extracts all packaged assets, and enforces zip slip path validation (`archive.go`).
- [x] **CLI Packaging Customization**: Added `-o, --output` flags and positional destination arguments to `qst form` and `qst free` with standardized error handling and exit codes (`cmd/form.go`, `cmd/free.go`).

### 5. Curated Example Knowledge Graphs (`examples/`)
- [x] **Pointers & Arrays (`examples/Pointers_and_Arrays/`)**: 44 guides modeling C pointers, address operators, pointer arithmetic, dynamic memory, multi-dimensional arrays, and function pointers.
- [x] **C Strings (`examples/C_Strings/`)**: 28 guides detailing null-terminated byte strings, string literals, `<string.h>` library functions, safety idioms, and bounded operations.
- [x] **Vim Usage (`examples/Vim_Usage/`)**: 48 guides structuring modal text editing, basic/word/line motions, text objects, operators, substitution, registers, macros, and window navigation.
- [x] **Automated Validation Suite (`internal/graph/examples_test.go`)**: Continuous regression test suite ensuring all repository knowledge graphs strictly pass manifest parsing, scope hierarchy rules, segment boundary validations, and acyclicity checks.

---

## Open Tasks & Next Enhancements

### 1. Static Site Generator (`internal/ssg/`)
- [ ] **Built-in Local Development Server (`qst serve`)**:
  - *Location*: `cmd/serve.go`, `internal/ssg/`
  - *Details*: Introduce `qst serve [directory] [--port 8080] [--watch]` to run a lightweight local static HTTP server for browsing compiled sites without external tooling, optionally monitoring guide markdown and manifest changes to auto-rebuild.
- [ ] **Client-Side Search Index Generation**:
  - *Location*: `internal/ssg/engine.go`, `internal/ssg/templates/layout.html`
  - *Details*: Generate a lightweight `search-index.json` containing guide titles, scopes, tags, and summary excerpts. Integrate an instant search modal or filter input in the sidebar to allow quick exploration in large knowledge graphs.

### 2. Graph Core & Relationship Logic (`internal/graph/`)
- [ ] **Reverse Reference Index (Dependents & Inverted Graph)**:
  - *Location*: `internal/graph/parser.go`, `internal/graph/validator.go`
  - *Details*: The `Guide` struct tracks forward prerequisites and subguides. Implement an inverted index helper (e.g., `BuildReverseIndex(guides)`) providing fast queries for incoming dependencies (`GetDependents(guideID)`, `GetParentGuides(guideID)`). This avoids $O(N)$ graph scans during cascading deletions and relationship diagnostics.
- [ ] **Topological Sorting & Prerequisite Path Discovery**:
  - *Location*: `internal/graph/validator.go` or `internal/graph/solver.go`
  - *Details*: Provide `TopologicalSort(guides)` and `PrerequisiteChain(targetGuideID)` returning an ordered linear sequence of guides necessary to learn a topic from foundational definitions to advanced material without violating dependency order.
- [ ] **Standard Graph Export Formats (DOT, Cytoscape / JSON Graph Format)**:
  - *Location*: `internal/graph/`
  - *Details*: Support exporting the knowledge graph structure to Graphviz DOT, Cytoscape JSON, or GraphML for ingestion and network analysis in third-party tools.

### 3. CLI Commands & Tooling (`cmd/`)
- [ ] **Mermaid Generation Parity & Graph Filtering in `qst view`**:
  - *Location*: `cmd/view.go`, `internal/ssg/mermaid.go`
  - *Details*: Refactor `cmd/view.go` to share the robust node ID sanitation, collision handling, and scope clustering implemented in `internal/ssg/mermaid.go`. Add flags for orientation (`--dir TD|LR`), scope grouping (`--scopes`), and localized subgraph extraction (`--guide <name>`).
- [ ] **Interactive Guide Creation (`qst add -i`)**:
  - *Location*: `cmd/add.go`
  - *Details*: Add an interactive mode (`--interactive`, `-i`) using `huh` forms to prompt for guide title, scope, prerequisites, and initial tags interactively, matching the developer ergonomics of `qst link` and `qst unlink`.
- [ ] **Prerequisite Learning Journey Command (`qst path <target>`)**:
  - *Location*: `cmd/path.go`
  - *Details*: Add a command `qst path <guide>` that calculates and displays the shortest learning sequence or topological prerequisite dependencies required to reach a specific guide.
