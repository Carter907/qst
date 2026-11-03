---
scope: subsystem
prerequisites:
  - Knowledge Graph Core Engine
sub_guides:
  - guide: Markdown AST Pipeline and Diagram Synthesis
    adherence: strict
    segment: 6-8
  - guide: Static Site Compilation and Academic Layout
    adherence: strict
    segment: 10-12
  - guide: Live-Reloading Development Server Architecture
    adherence: detailed
    segment: 14-16
tags: [ssg, subsystem, compiler, server]
---

# Static Site Compiler and Live Server

The `internal/ssg` package consumes the validated domain structures of `internal/graph` and compiles them into an interactive, academic-styled static HTML website (`qst build`) or serves them dynamically with live reload and real-time validation diagnostics (`qst serve`).

## 1. Goldmark AST Transformation and Neighborhood Diagrams

In `markdown.go` and `mermaid.go`, each guide's Markdown body is processed through a customized Goldmark AST pipeline that rewrites relative `.md` links to `.html` (`linkASTTransformer`) and binary-searches source byte offsets to inject `#L<n>` line number anchors (`lineASTTransformer`). In parallel, `BuildGuideMermaid` synthesizes a localized 1-hop `graph LR` Mermaid diagram displaying the guide's immediate prerequisites and sub-guides with academic scope prefixes and clickable navigation links.

## 2. Static Site Generation, Navigation, and Vendor Caching

In `engine.go`, `navigation.go`, and `download.go`, `BuildSiteWithOptions` resolves third-party KaTeX and Mermaid assets through a 4-tier offline-capable cache (`EnsureVendorAssets`), copies static media assets, and renders embedded HTML templates (`layout.html`, `index.html`, `guide.html`). It synthesizes a top-down scope syllabus on the landing page, collapsible sidebar tour/scope trees backed by `sessionStorage`, and relative URLs (`relURL`) that work from any base path.

## 3. Ephemeral Live-Reloading Development Server

In `server.go`, `DevServer` provides a zero-pollution authoring workflow by compiling into an ephemeral `os.MkdirTemp` directory, watching recursive directory trees (or `.kng` archives) via debounced `fsnotify` events, guarding rebuilds with a `sync.RWMutex`, and streaming `reload`, `build_error`, and `clear_error` events to connected browsers over Server-Sent Events (`/_qst/events`) with split `Cache-Control` headers and scroll-position preservation.
