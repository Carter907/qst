---
scope: primitive
prerequisites:
  - Guide and Manifest Data Models
sub_guides: []
tags: [graph, parser, yaml, serialization]
---

# Line-Preserving Frontmatter Parser

Standard Markdown frontmatter libraries typically parse the document into an abstract buffer, stripping original line endings or reformatting the markdown body on write. In `internal/graph/parser.go`, `qst` implements a custom line-preserving parser (`extractFrontmatter`) so that CLI mutations (`qst link`, `qst tag`, `qst rename`) never alter user-authored Markdown bodies or line numbers.

## Frontmatter Extraction (`extractFrontmatter`)

```go
type frontmatterParts struct {
	lines            []string
	frontmatter      string
	body             string
	newline          string
	frontMatterStart int
	frontMatterEnd   int
}
```

1. **Line-Ending Detection**: Before splitting lines, `extractFrontmatter` inspects the raw byte slice for `"\r\n"`. If present, it records `newline = "\r\n"` (otherwise `"\n"`), normalizes lines for scanning, and restores the exact original line ending upon serialization in `UpdateGuideMetadata`.
2. **Exact Boundary Indexing**: It scans for the opening and closing `---` delimiters (`strings.TrimSpace(line) == "---"`), recording `frontMatterStart` and `frontMatterEnd` line indices. Everything after `frontMatterEnd + 1` is joined into `body`.

## Surgical In-Place Metadata Updates (`UpdateGuideMetadata`)

When `UpdateGuideMetadata(path, meta)` mutates a guide's frontmatter:

```go
// Ensure nil slices serialize to empty arrays [] in yaml instead of null
if meta.Prerequisites == nil {
	meta.Prerequisites = []string{}
}
if meta.SubGuides == nil {
	meta.SubGuides = []SubGuideRelation{}
}
if meta.Tags == nil {
	meta.Tags = []string{}
}
```

* **Preventing YAML `null` Pollution**: In `gopkg.in/yaml.v3`, a `nil` slice marshals as `null` (`prerequisites: null`), whereas an initialized empty slice marshals as `prerequisites: []`. Explicitly replacing `nil` slices before `yaml.Marshal` keeps guide frontmatter clean and uniform.
* **Splice Reconstruction**: Only the lines between `frontMatterStart+1` and `frontMatterEnd` are replaced; `parts.lines[frontMatterEnd:]` (the closing `---` and entire Markdown body) is spliced back untouched.
* **Manifest Fallback**: `ParseConfig` reads `manifest.yaml` first and transparently falls back to `quest.yaml` for backwards compatibility with early graph repositories.
