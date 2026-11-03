---
scope: component
prerequisites:
  - Markdown AST Pipeline and Diagram Synthesis
  - Graph Indexing Analytics and KNG Packaging
sub_guides:
  - guide: Vendor Asset Acquisition and Caching
    adherence: detailed
    segment: 6-10
  - guide: Academic Sidebar and Syllabus Builder
    adherence: strict
    segment: 12-16
tags: [ssg, engine, templates, html]
---

# Static Site Compilation and Academic Layout

In `internal/ssg/engine.go`, `BuildSiteWithOptions` orchestrates the end-to-end compilation of a knowledge graph directory into a self-contained static HTML website requiring no backend server when deployed.

## Asset Provisioning and Embedded Template Compilation

At the start of a build, `BuildSiteWithOptions` ensures `outDir` exists, provisions KaTeX and Mermaid inside `outDir/assets/vendor` via `EnsureVendorAssets` (unless `SkipVendor` is set by `qst serve`), and writes the embedded `templates/style.css` stylesheet via Go's `//go:embed templates/*` filesystem (`templatesFS`). It then compiles two template sets from `templatesFS`:
* `indexTmpl`: `layout.html` + `index.html` (the course syllabus landing page).
* `guideTmpl`: `layout.html` + `guide.html` (individual guide pages with metadata cards, line-numbered content, and 1-hop Mermaid diagrams).

## Relative Path Resolution and Page Rendering

During its walk of `dirPath`, `BuildSiteWithOptions` copies static media files (`.png`, `.jpg`, `.jpeg`, `.svg`, `.gif`, `.webp`, `.mmd`, `.mermaid`) directly to their relative paths in `outDir` and computes each guide's `relRoot` (`filepath.Rel(outRelDir, ".")`).

Using `relURL(fromOutRelPath, toOutRelPath)`, every link in the sidebar (`buildSidebar`), tour step bar (`buildTourSteps`), syllabus (`buildSyllabusScopes`), and relationship cards (`Prerequisites`, `Dependents`, `SubGuides`, `ParentGuides`) is emitted as a relative URL. This guarantees that compiled sites work seamlessly whether hosted at a domain root (`https://example.com/`), inside a subpath (`https://example.com/courses/c/`), or opened directly from the local filesystem.
