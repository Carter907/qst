package ssg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSite(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	outDir := filepath.Join(tmpDir, "public")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Create a mock markdown file with MathJax, Mermaid, and GFM

	manifestContent := "name: Test\nversion: 1.0.0\n"
	if err := os.WriteFile(filepath.Join(sourceDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to write mock manifest: %v", err)
	}

	markdownContent := `
# Test Knowledge Graph

## LaTeX Math
Block Math:
$$
E=mc^2
$$

Inline Math:
The formula is $x = y^2$.

## Mermaid
` + "```" + `mermaid
graph TD;
    A-->B;
` + "```" + `

## Table
| Header 1 | Header 2 |
|----------|----------|
| Row 1    | Value 1  |
`
	if err := os.WriteFile(filepath.Join(sourceDir, "test.md"), []byte(markdownContent), 0644); err != nil {
		t.Fatalf("Failed to write mock md: %v", err)
	}

	// Build the site
	if err := BuildSite(sourceDir, outDir); err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// Check if output files exist
	expectedHTML := filepath.Join(outDir, "test.html")
	content, err := os.ReadFile(expectedHTML)
	if err != nil {
		t.Fatalf("Failed to read generated HTML: %v", err)
	}
	htmlStr := string(content)

	// Verify math rendering wrappers
	if !strings.Contains(htmlStr, `<span class="math display">`) {
		t.Errorf("Expected block math in HTML, got: %s", htmlStr)
	}

	// Verify mermaid code block (Goldmark should render it as <code class="language-mermaid">)
	if !strings.Contains(htmlStr, `class="language-mermaid"`) {
		t.Errorf("Expected mermaid class in HTML")
	}

	// Verify table
	if !strings.Contains(htmlStr, `<table>`) {
		t.Errorf("Expected table in HTML")
	}

	// Verify assets downloaded
	if _, err := os.Stat(filepath.Join(outDir, "assets", "vendor", "katex", "katex.min.css")); err != nil {
		t.Errorf("KaTeX CSS not found in outDir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "assets", "vendor", "mermaid", "mermaid.min.js")); err != nil {
		t.Errorf("Mermaid bundle not found in outDir: %v", err)
	}
}

func TestBuildSiteGraphMetadataAndRelativePaths(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	outDir := filepath.Join(tmpDir, "public")

	if err := os.MkdirAll(filepath.Join(sourceDir, "sub"), 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	manifestContent := `title: Test Graph
description: A knowledge graph for testing
scopes:
  - definition
  - explanation
tours:
  - name: Core Tour
    guides:
      - GuideA
      - GuideB
`
	if err := os.WriteFile(filepath.Join(sourceDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	guideAContent := `---
prerequisites: []
sub_guides:
  - guide: GuideB
    adherence: strict
    segment: 10-20
scope: explanation
tags:
  - go
---
# Guide A Heading

Body text for Guide A.
`
	if err := os.WriteFile(filepath.Join(sourceDir, "GuideA.md"), []byte(guideAContent), 0644); err != nil {
		t.Fatalf("Failed to write GuideA: %v", err)
	}

	guideBContent := `---
prerequisites:
  - GuideA
sub_guides: []
scope: definition
tags:
  - core
---
# Guide B Heading

Body text for Guide B.
`
	if err := os.WriteFile(filepath.Join(sourceDir, "sub", "GuideB.md"), []byte(guideBContent), 0644); err != nil {
		t.Fatalf("Failed to write GuideB: %v", err)
	}

	if err := BuildSite(sourceDir, outDir); err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// 1. Verify GuideA (root-level guide)
	guideAHtmlBytes, err := os.ReadFile(filepath.Join(outDir, "GuideA.html"))
	if err != nil {
		t.Fatalf("Failed to read GuideA.html: %v", err)
	}
	guideAHtml := string(guideAHtmlBytes)

	// Issue 1: No YAML frontmatter leaked into output
	if strings.Contains(guideAHtml, "prerequisites: []") || strings.Contains(guideAHtml, "sub_guides:") {
		t.Errorf("GuideA.html contains raw frontmatter: %s", guideAHtml)
	}
	if !strings.Contains(guideAHtml, "<h1 id=\"guide-a-heading\">Guide A Heading</h1>") &&
		!strings.Contains(guideAHtml, "Guide A Heading") {
		t.Errorf("GuideA.html missing body heading")
	}

	// Issue 2: Graph relations rendered
	if !strings.Contains(guideAHtml, "guide-byline") || !strings.Contains(guideAHtml, "explanation") {
		t.Errorf("GuideA.html missing explanation scope byline")
	}
	if !strings.Contains(guideAHtml, "href=\"sub/GuideB.html\"") {
		t.Errorf("GuideA.html should link to sub/GuideB.html")
	}
	if !strings.Contains(guideAHtml, "Sub-Guides") {
		t.Errorf("GuideA.html should show Sub-Guides for GuideB")
	}
	if !strings.Contains(guideAHtml, "strict") {
		t.Errorf("GuideA.html should display adherence 'strict'")
	}
	if !strings.Contains(guideAHtml, "lines 10-20") {
		t.Errorf("GuideA.html should display segment 'lines 10-20'")
	}

	// Issue 3: Root paths
	if !strings.Contains(guideAHtml, "href=\"assets/style.css\"") {
		t.Errorf("GuideA.html should reference assets/style.css")
	}

	// Issue 4: Dynamic Sidebar
	if !strings.Contains(guideAHtml, "Core Tour") {
		t.Errorf("GuideA.html sidebar missing Core Tour")
	}
	if !strings.Contains(guideAHtml, "<details class=\"sidebar-tour-group\"") {
		t.Errorf("GuideA.html sidebar missing collapsible tour groups")
	}
	if !strings.Contains(guideAHtml, "<details class=\"sidebar-scope-group\"") {
		t.Errorf("GuideA.html sidebar missing scope groups")
	}
	if !strings.Contains(guideAHtml, "class=\"sidebar-item active\"") {
		t.Errorf("GuideA.html sidebar missing active class for GuideA")
	}

	// Tour Navigation in Sidebar (in-body tour banners/steppers removed for clean academic feel)
	if strings.Contains(guideAHtml, "tour-banner") || strings.Contains(guideAHtml, "tour-stepper") {
		t.Errorf("GuideA.html should not contain in-body tour banners or steppers")
	}
	if !strings.Contains(guideAHtml, "href=\"sub/GuideB.html\"") {
		t.Errorf("GuideA.html sidebar missing tour link to sub/GuideB.html")
	}

	// 2. Verify GuideB (nested guide in subfolder)
	guideBHtmlBytes, err := os.ReadFile(filepath.Join(outDir, "sub", "GuideB.html"))
	if err != nil {
		t.Fatalf("Failed to read sub/GuideB.html: %v", err)
	}
	guideBHtml := string(guideBHtmlBytes)

	// Issue 1: No YAML frontmatter
	if strings.Contains(guideBHtml, "scope: definition") {
		t.Errorf("sub/GuideB.html contains raw frontmatter")
	}

	// Issue 2: Prerequisite link pointing to GuideA
	if !strings.Contains(guideBHtml, "href=\"../GuideA.html\"") {
		t.Errorf("sub/GuideB.html should link back to ../GuideA.html")
	}
	if !strings.Contains(guideBHtml, "Prerequisites") {
		t.Errorf("sub/GuideB.html should show Prerequisites section for guide GuideA")
	}

	// Issue 3: Relative asset path in nested directory
	if !strings.Contains(guideBHtml, "href=\"../assets/style.css\"") {
		t.Errorf("sub/GuideB.html should reference ../assets/style.css, got: %s", guideBHtml)
	}
	if !strings.Contains(guideBHtml, "href=\"../index.html\"") {
		t.Errorf("sub/GuideB.html should link to ../index.html")
	}
	if !strings.Contains(guideBHtml, "src=\"../assets/vendor/katex/katex.min.js\"") {
		t.Errorf("sub/GuideB.html should reference ../assets/vendor/katex/katex.min.js")
	}

	// Issue 4: Nested Tour in Sidebar & clean body
	if !strings.Contains(guideBHtml, "<details class=\"sidebar-tour-group\"") {
		t.Errorf("sub/GuideB.html sidebar missing collapsible tour groups")
	}
	if !strings.Contains(guideBHtml, "href=\"../GuideA.html\"") {
		t.Errorf("sub/GuideB.html sidebar should link to ../GuideA.html")
	}
	if strings.Contains(guideBHtml, "tour-banner") || strings.Contains(guideBHtml, "tour-stepper") {
		t.Errorf("sub/GuideB.html should not contain in-body tour banners or steppers")
	}

	// 3. Verify index.html contains Knowledge Graph Mermaid diagram
	idxHtmlBytes, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("Failed to read index.html: %v", err)
	}
	idxHtml := string(idxHtmlBytes)

	if !strings.Contains(idxHtml, "Knowledge Graph Overview") {
		t.Errorf("index.html missing Knowledge Graph Overview heading")
	}
	if !strings.Contains(idxHtml, `class="language-mermaid"`) {
		t.Errorf("index.html missing mermaid code block")
	}
	if !strings.Contains(idxHtml, `subgraph scope_explanation ["Scope: explanation"]`) {
		t.Errorf("index.html mermaid diagram missing subgraph for explanation")
	}
	if !strings.Contains(idxHtml, `subgraph scope_definition ["Scope: definition"]`) {
		t.Errorf("index.html mermaid diagram missing subgraph for definition")
	}
	if !strings.Contains(idxHtml, "direction LR") {
		t.Errorf("index.html mermaid diagram missing direction LR")
	}
	if !strings.Contains(idxHtml, "guidea --> guideb") {
		t.Errorf("index.html mermaid diagram missing prerequisite edge 'guidea --> guideb'")
	}
	if !strings.Contains(idxHtml, "guidea -.-> guideb") {
		t.Errorf("index.html mermaid diagram missing subguide edge 'guidea -.-> guideb'")
	}
	if !strings.Contains(idxHtml, `click guidea href "GuideA.html"`) {
		t.Errorf("index.html mermaid diagram missing click href for GuideA")
	}
	if !strings.Contains(idxHtml, `click guideb href "sub/GuideB.html"`) {
		t.Errorf("index.html mermaid diagram missing click href for sub/GuideB")
	}
}
