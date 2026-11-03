---
scope: component
prerequisites:
  - Graph Parsing and Cascading Mutations
sub_guides:
  - guide: Goldmark AST Link and Line Transformers
    adherence: strict
    segment: 6-10
  - guide: Localized 1-Hop Mermaid Graph Generator
    adherence: detailed
    segment: 12-14
tags: [ssg, markdown, goldmark, mermaid]
---

# Markdown AST Pipeline and Diagram Synthesis

The content transformation pipeline in `internal/ssg/markdown.go` and `internal/ssg/mermaid.go` converts raw guide Markdown and frontmatter relationships into academic HTML annotated with line markers, math formulas, syntax-highlighted code blocks, and interactive neighborhood graphs.

## Goldmark Extension Pipeline and Line Anchoring

`newMarkdownParser()` constructs a configured Goldmark compiler combining GitHub Flavored Markdown (`extension.GFM`), footnotes (`extension.Footnote`), MathJax (`goldmark-mathjax`), and Chroma syntax highlighting (`github` style) with `qst`'s custom AST transformers:
1. **`linkASTTransformer` (Priority 100)**: Rewrites relative `.md` links to `.html` and maps `index.md` to `index-guide.html`.
2. **`lineASTTransformer` (Priority 200) & `LineNumRenderer`**: Binary-searches source newline offsets to inject `<span class="line-marker" id="L<n>">` anchors before every block element, enabling direct deep-linking to the line segments cited by parent guides.

## Contextual 1-Hop Mermaid Synthesis

Alongside the rendered Markdown body, `generateGuideMermaid` invokes `BuildGuideMermaid` for every guide that has at least one prerequisite or sub-guide (`RequireNeighbors: true`). The resulting `graph LR` diagram styles the active guide with a bold border (`classDef currentGuide`), labels every node with its capitalized scope prefix (`<Scope>: <Title>`), distinguishes horizontal prerequisites (`-->`) from vertical sub-guides (`-.->|"adherence [segment]"|`), and attaches clickable relative URLs to every neighbor.
