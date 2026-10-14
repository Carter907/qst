# Project Enhancement Roadmap & Issues

This document tracks completed enhancements and open high-impact tasks across the static site generator, graph core, packaging, and CLI tooling.

---

## Completed Enhancements

### 1. Static Site Generator (`internal/ssg/`)
- [x] **Populate Sidebar Navigation**: Implemented dynamic sidebar grouping guides hierarchically by tours and scopes with collapsible `<details>` sections and active route highlights (`layout.html`, `engine.go`, `style.css`).
- [x] **Integrate Graph Metadata into Guides**: Guides display metadata bylines (scope, tags) and an academic appendix listing prerequisites and subguides with adherence ratings and segment line numbers (`guide.html`, `engine.go`).
- [x] **Nested Directory & Relative Asset Resolution**: Introduced `RelativeRoot` and `relURL` to compute relative paths for CSS, JS vendor bundles, and cross-guide links, allowing guides at arbitrary directory depths to render without 404s (`engine.go`, `layout.html`).
- [x] **Knowledge Graph Overview on Index**: Added automated generation of an interactive Mermaid diagram on `index.html` displaying prerequisites (solid arrows), subguides (dotted arrows), and clickable links to guide pages (`mermaid.go`, `index.html`).
- [x] **Academic Styling & Gutter Line Numbering**: Integrated dedicated line markers with hair-line gutter numbering and booktabs-style tables (`style.css`, `engine.go`).

### 2. Graph Core & Relationship Logic (`internal/graph/`)
- [x] **Flexible Configuration File Naming**: `ParseConfig` now supports falling back from `manifest.yaml` to `quest.yaml` seamlessly (`parser.go`).
- [x] **Relationship Mutation Helpers**: Implemented `AddPrerequisite`, `RemovePrerequisite`, `UpsertSubGuide`, and `RemoveSubGuide` with frontmatter persistence (`parser.go`).
- [x] **Link Validation & Cycle Prevention**: Added `ValidatePrerequisiteLink` and `ValidateSubGuideLink` with scope hierarchy enforcement and acyclicity verification (`validator.go`).
- [x] **Line Coverage Analysis**: Added `AnalyzeGuideCoverage`, `LineCoverage`, and `GuideCoverageReport` to measure subguide coverage across guide markdown bodies (`validator.go`).
- [x] **Recursive Subdirectory Support in `ParseGraph`**: `ParseGraph` traverses directory trees recursively via `filepath.Walk`, allowing guides to be organized into nested subfolders while ignoring `public/` and hidden folders, and detecting duplicate guide IDs (`parser.go`).
- [x] **Detailed Cycle Path Reporting in `CheckAcyclic`**: `CheckAcyclic` tracks the active DFS traversal stack and outputs the exact loop chain (e.g. `cycle detected in graph: Guide A -> Guide B -> Guide A`) with deterministic ordering (`validator.go`).

### 3. CLI Commands & Tooling (`cmd/`)
- [x] **Reciprocal Relationship Management (`qst unlink`)**: Added `qst unlink` to remove prerequisites or subguides interactively (`-i`) or via flags (`--guide`, `--prereq`) (`cmd/unlink.go`).
- [x] **Enhanced Linking (`qst link`)**: Extended `qst link` to support horizontal prerequisites (`--prereq`), interactive prompt workflows, validation, and in-place upserts (`cmd/link.go`).
- [x] **Guide Inspector & Coverage Commands (`qst inspect`, `qst coverage`)**: Added `qst inspect` for comprehensive guide metadata and body statistics, with `--coverage` (`-c`) and `qst coverage` providing line-by-line coverage breakdowns (`cmd/inspect.go`, `cmd/coverage.go`).
- [x] **Fix Exit Codes in `qst validate`**: Validation, configuration, and parsing errors now terminate with exit code 1 to ensure CI reliability (`cmd/validate.go`).
- [x] **Deterministic Sorting & Rich Table View in `qst list`**: Guides are sorted alphabetically, with `-l, --long` rendering an aligned tabular view showing scope, tags, and link counts (`cmd/list.go`).
- [x] **Enrich `qst stats` Dashboard**: Displays sectioned statistics including overview metrics, scope distribution, graph depth / longest prerequisite chain, tour coverage, and aggregate body line coverage (`cmd/stats.go`).

### 4. Archiving & Packaging (`internal/graph/archive.go`, `cmd/`)
- [x] **Preserve Manifests and Assets in `.kng` Packages**: `ArchiveGraph` recursively walks knowledge graphs using deflate compression, packaging `manifest.yaml` / `quest.yaml`, `.md` guides, diagrams, and media assets while skipping build directories (`public/`), hidden files, and archives (`archive.go`).
- [x] **Lossless Extraction & Nested Directory Support**: `UnarchiveGraph` automatically creates parent directories for nested paths, extracts all packaged assets, and enforces zip slip path validation (`archive.go`).
- [x] **CLI Packaging Customization**: Added `-o, --output` flags and positional destination arguments to `qst form` and `qst free` with standardized error handling and exit codes (`cmd/form.go`, `cmd/free.go`).

---

## Open Tasks & Next Enhancements

### 1. Static Site Generator (`internal/ssg/`)
- [ ] **Guide-Specific Localized Mermaid Subgraphs**:
  - *Location*: `internal/ssg/engine.go`, `internal/ssg/templates/guide.html`
  - *Details*: While `index.html` embeds a global diagram, individual guide pages currently lack localized context. Render a focused 1-hop or 2-hop neighborhood Mermaid subgraph in the guide footer/appendix showing immediate prerequisites and subguides.
- [ ] **Zip Slip / Path Traversal Protection in KaTeX Downloader**:
  - *Location*: `internal/ssg/download.go:43`
  - *Details*: In `downloadKaTeX`, `target := filepath.Join(destDir, header.Name)` does not verify that `target` remains prefixed with `destDir` before extracting files. Add a `strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir))` check similar to `UnarchiveGraph`.
- [ ] **Network Robustness & Offline Builds**:
  - *Location*: `internal/ssg/download.go:21,77`, `cmd/build.go`
  - *Details*: `http.Get` uses the default HTTP client without a timeout. Add a configured `http.Client` with timeout, and introduce an `--offline` / `--skip-vendor` flag to `qst build` to allow building in air-gapped environments when vendor assets are already present or external assets are pre-cached.


