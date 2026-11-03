---
scope: primitive
prerequisites: []
sub_guides: []
tags: [graph, models, yaml, frontmatter]
---

# Guide and Manifest Data Models

The foundational domain structures of `qst` reside in `internal/graph/models.go` and model a knowledge graph as a two-dimensional Directed Acyclic Graph (DAG) with **horizontal prerequisite edges** and **vertical sub-guide edges**.

## Core Guide Structures

Each Markdown file in a knowledge graph is parsed into a `Guide` struct containing both filesystem identity and YAML frontmatter metadata:

```go
type SubGuideRelation struct {
	Guide     string `yaml:"guide"`
	Adherence string `yaml:"adherence"`
	Segment   string `yaml:"segment"`
}

type GuideMetadata struct {
	Prerequisites []string           `yaml:"prerequisites"`
	SubGuides     []SubGuideRelation `yaml:"sub_guides"`
	Scope         string             `yaml:"scope"`
	Tags          []string           `yaml:"tags"`
}

type Guide struct {
	ID         string
	Path       string
	Metadata   GuideMetadata
	LineCount  int
	HasContent bool
	Body       string
}
```

### Low-Level Design Choices
* **Filename-Stem Identity (`Guide.ID`)**: Rather than duplicating an `id` field inside YAML frontmatter (which risks desynchronization when files are moved or renamed), `Guide.ID` is derived directly from `strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))`.
* **Post-Frontmatter `LineCount`**: `LineCount` records only the lines *after* the closing `---` frontmatter delimiter (`len(parts.lines) - (parts.frontMatterEnd + 1)`). This ensures that editing YAML metadata (such as adding tags or prerequisites) never shifts the line numbers referenced by `SubGuideRelation.Segment` ranges or Goldmark `#L<n>` HTML anchors.
* **`HasContent` Guard**: Tracks whether at least one non-whitespace line exists in `Body`, allowing the validator to reject empty stub guides early.

## Manifest Configuration (`manifest.yaml`)

Graph-wide invariants and curriculum tours are governed by `Manifest`:

```go
type Manifest struct {
	Title            string       `yaml:"title"`
	Description      string       `yaml:"description"`
	Scopes           []string     `yaml:"scopes"`
	Adherences       []string     `yaml:"adherences"`
	RelaxedSubguides bool         `yaml:"relaxed_subguides"`
	StrictCoverage   bool         `yaml:"strict_coverage"`
	RequireSubguides bool         `yaml:"require_subguides"`
	Tours            []TourConfig `yaml:"tours"`
}
```

The `Scopes` slice is **order-significant**: index `0` represents the smallest, most granular scope (e.g., `definition` or `primitive`), while higher indices represent progressively broader synthesis tiers.
