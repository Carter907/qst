package ssg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Carter907/qst/internal/graph"
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
	if !strings.Contains(guideAHtml, "<figure class=\"appendix-figure\">") {
		t.Errorf("GuideA.html should contain appendix-figure container")
	}
	if !strings.Contains(guideAHtml, "class=\"language-mermaid\"") {
		t.Errorf("GuideA.html should contain localized mermaid diagram")
	}
	if !strings.Contains(guideAHtml, "Explanation: GuideA") {
		t.Errorf("GuideA.html should contain Explanation: GuideA node label")
	}
	if !strings.Contains(guideAHtml, "guidea -.->") {
		t.Errorf("GuideA.html should show subguide edge for GuideB")
	}
	if !strings.Contains(guideAHtml, "strict") {
		t.Errorf("GuideA.html should display adherence 'strict'")
	}
	if !strings.Contains(guideAHtml, "10-20") {
		t.Errorf("GuideA.html should display segment '10-20'")
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
	if strings.Contains(guideAHtml, "<details class=\"sidebar-tour-group\" open") {
		t.Errorf("GuideA.html sidebar tour groups should be closed by default")
	}
	if !strings.Contains(guideAHtml, "<details class=\"sidebar-scope-group\"") {
		t.Errorf("GuideA.html sidebar missing scope groups")
	}
	if strings.Contains(guideAHtml, "<details class=\"sidebar-scope-group\" open") {
		t.Errorf("GuideA.html sidebar scope groups should be closed by default")
	}
	if !strings.Contains(guideAHtml, "sidebar-section-tours") || !strings.Contains(guideAHtml, "sidebar-section-scopes") {
		t.Errorf("GuideA.html sidebar missing section classes for tours or scopes")
	}
	if !strings.Contains(guideAHtml, `data-group-id="tour-Core Tour"`) {
		t.Errorf("GuideA.html sidebar missing data-group-id for Core Tour")
	}
	if !strings.Contains(guideAHtml, `data-group-id="scope-explanation"`) {
		t.Errorf("GuideA.html sidebar missing data-group-id for explanation scope")
	}
	if !strings.Contains(guideAHtml, "qst_sidebar_open_groups") {
		t.Errorf("GuideA.html missing sidebar sessionStorage persistence script")
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
	if !strings.Contains(guideBHtml, "<figure class=\"appendix-figure\">") {
		t.Errorf("sub/GuideB.html should contain appendix-figure container")
	}
	if !strings.Contains(guideBHtml, "class=\"language-mermaid\"") {
		t.Errorf("sub/GuideB.html should contain localized mermaid diagram")
	}
	if !strings.Contains(guideBHtml, "Definition: GuideB") {
		t.Errorf("sub/GuideB.html should contain Definition: GuideB node label")
	}
	if !strings.Contains(guideBHtml, "guidea --> guideb") {
		t.Errorf("sub/GuideB.html should show prerequisite edge from GuideA")
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
	if strings.Contains(guideBHtml, "<details class=\"sidebar-tour-group\" open") {
		t.Errorf("sub/GuideB.html sidebar tour groups should be closed by default")
	}
	if !strings.Contains(guideBHtml, "href=\"../GuideA.html\"") {
		t.Errorf("sub/GuideB.html sidebar should link to ../GuideA.html")
	}
	if strings.Contains(guideBHtml, "tour-banner") || strings.Contains(guideBHtml, "tour-stepper") {
		t.Errorf("sub/GuideB.html should not contain in-body tour banners or steppers")
	}

	// 3. Verify index.html contains Academic Syllabus Table (and no Mermaid diagram)
	idxHtmlBytes, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("Failed to read index.html: %v", err)
	}
	idxHtml := string(idxHtmlBytes)

	if !strings.Contains(idxHtml, "Curriculum & Syllabus") {
		t.Errorf("index.html missing Curriculum & Syllabus heading")
	}
	if !strings.Contains(idxHtml, "syllabus-table") {
		t.Errorf("index.html missing syllabus-table class")
	}
	if !strings.Contains(idxHtml, "explanation") {
		t.Errorf("index.html missing explanation scope")
	}
	if !strings.Contains(idxHtml, "definition") {
		t.Errorf("index.html missing definition scope")
	}
	if !strings.Contains(idxHtml, "href=\"GuideA.html\"") {
		t.Errorf("index.html missing link to GuideA.html")
	}
	if !strings.Contains(idxHtml, "href=\"sub/GuideB.html\"") {
		t.Errorf("index.html missing link to sub/GuideB.html")
	}
	// Prerequisite plain-text link for GuideB pointing to GuideA
	if !strings.Contains(idxHtml, "syllabus-prereq-link") {
		t.Errorf("index.html missing prerequisite link in syllabus")
	}
	if strings.Contains(idxHtml, "syllabus-badge") {
		t.Errorf("index.html should not contain badge containers")
	}
	// GuideA has zero prerequisites, should render en-dash
	if !strings.Contains(idxHtml, "syllabus-empty") {
		t.Errorf("index.html missing empty prerequisite placeholder (–)")
	}
	// Verify global Mermaid diagram and pan/zoom are NOT in index.html
	if strings.Contains(idxHtml, "Knowledge Graph Overview") {
		t.Errorf("index.html should not contain Knowledge Graph Overview heading")
	}
	if strings.Contains(idxHtml, "setupPanZoom") {
		t.Errorf("index.html should not contain setupPanZoom script")
	}

	// 4. Verify Syllabus styles in style.css
	cssBytes, err := os.ReadFile(filepath.Join(outDir, "assets", "style.css"))
	if err != nil {
		t.Fatalf("Failed to read style.css: %v", err)
	}
	cssStr := string(cssBytes)
	if !strings.Contains(cssStr, ".syllabus-table") {
		t.Errorf("style.css missing .syllabus-table styles")
	}
	if !strings.Contains(cssStr, ".syllabus-prereq-link") {
		t.Errorf("style.css missing .syllabus-prereq-link styles")
	}
	if !strings.Contains(cssStr, ".syllabus-tag") {
		t.Errorf("style.css missing .syllabus-tag styles")
	}
	if strings.Contains(cssStr, ".syllabus-badge") {
		t.Errorf("style.css should not contain .syllabus-badge styles")
	}
}

func TestBuildSyllabusScopes(t *testing.T) {
	manifest := graph.Manifest{
		Scopes: []string{"definition", "explanation", "lesson"},
	}

	gDef := &guideFile{
		guide: graph.Guide{
			ID: "Def 1",
			Metadata: graph.GuideMetadata{
				Scope: "definition",
			},
		},
		outRelPath: "Def 1.html",
	}

	gExp := &guideFile{
		guide: graph.Guide{
			ID: "Exp 1",
			Metadata: graph.GuideMetadata{
				Scope:         "explanation",
				Prerequisites: []string{"Def 1"},
				Tags:          []string{"math"},
			},
		},
		outRelPath: "Exp 1.html",
	}

	gLes := &guideFile{
		guide: graph.Guide{
			ID: "Lesson 1",
			Metadata: graph.GuideMetadata{
				Scope:         "lesson",
				Prerequisites: []string{"Exp 1"},
				Tags:          []string{"intro", "core"},
			},
		},
		outRelPath: "sub/Lesson 1.html",
	}

	gOther := &guideFile{
		guide: graph.Guide{
			ID: "Unscoped Guide",
			Metadata: graph.GuideMetadata{
				Scope: "",
			},
		},
		outRelPath: "Unscoped Guide.html",
	}

	guideFiles := []*guideFile{gDef, gExp, gLes, gOther}
	filesByID := map[string]*guideFile{
		"Def 1":          gDef,
		"Exp 1":          gExp,
		"Lesson 1":       gLes,
		"Unscoped Guide": gOther,
	}

	sections := buildSyllabusScopes(manifest, guideFiles, filesByID)

	// Top-down scope order: lesson -> explanation -> definition -> other
	if len(sections) != 4 {
		t.Fatalf("expected 4 syllabus sections, got %d", len(sections))
	}
	if sections[0].Scope != "lesson" {
		t.Errorf("expected first section to be 'lesson', got %q", sections[0].Scope)
	}
	if sections[1].Scope != "explanation" {
		t.Errorf("expected second section to be 'explanation', got %q", sections[1].Scope)
	}
	if sections[2].Scope != "definition" {
		t.Errorf("expected third section to be 'definition', got %q", sections[2].Scope)
	}
	if sections[3].Scope != "other" {
		t.Errorf("expected fourth section to be 'other', got %q", sections[3].Scope)
	}

	// Verify Lesson 1 guide details
	lesGuide := sections[0].Guides[0]
	if lesGuide.Title != "Lesson 1" || lesGuide.URL != "sub/Lesson 1.html" {
		t.Errorf("unexpected Lesson 1 ref: %+v", lesGuide)
	}
	if len(lesGuide.Prerequisites) != 1 || lesGuide.Prerequisites[0].Title != "Exp 1" || lesGuide.Prerequisites[0].URL != "Exp 1.html" {
		t.Errorf("unexpected Lesson 1 prerequisites: %+v", lesGuide.Prerequisites)
	}
	if len(lesGuide.Tags) != 2 || lesGuide.Tags[0] != "intro" || lesGuide.Tags[1] != "core" {
		t.Errorf("unexpected Lesson 1 tags: %+v", lesGuide.Tags)
	}

	// Verify Def 1 guide details (zero prerequisites)
	defGuide := sections[2].Guides[0]
	if defGuide.Title != "Def 1" || len(defGuide.Prerequisites) != 0 {
		t.Errorf("unexpected Def 1 ref: %+v", defGuide)
	}
}

