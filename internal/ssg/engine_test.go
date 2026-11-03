package ssg

import (
	"fmt"
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

func TestBuildSiteLearningPath(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	outDir := filepath.Join(tmpDir, "public")

	if err := os.MkdirAll(filepath.Join(sourceDir, "intermediate"), 0o755); err != nil {
		t.Fatalf("Failed to create source dirs: %v", err)
	}

	manifestContent := `title: Learning Path Test
scopes:
  - definition
  - explanation
  - application
`
	if err := os.WriteFile(filepath.Join(sourceDir, "manifest.yaml"), []byte(manifestContent), 0o644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	baseContent := `---
scope: definition
tags:
  - base
---
# Base Concept
Foundational text.
`
	if err := os.WriteFile(filepath.Join(sourceDir, "BaseConcept.md"), []byte(baseContent), 0o644); err != nil {
		t.Fatalf("Failed to write BaseConcept.md: %v", err)
	}

	interContent := `---
sub_guides:
  - guide: BaseConcept
scope: explanation
tags:
  - mid
---
# Intermediate Concept
Middle text.
`
	if err := os.WriteFile(filepath.Join(sourceDir, "intermediate", "IntermediateConcept.md"), []byte(interContent), 0o644); err != nil {
		t.Fatalf("Failed to write IntermediateConcept.md: %v", err)
	}

	advContent := `---
sub_guides:
  - guide: IntermediateConcept
scope: application
tags:
  - adv
---
# Advanced Concept
Advanced text.
`
	if err := os.WriteFile(filepath.Join(sourceDir, "AdvancedConcept.md"), []byte(advContent), 0o644); err != nil {
		t.Fatalf("Failed to write AdvancedConcept.md: %v", err)
	}

	if err := BuildSite(sourceDir, outDir); err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// 1. Verify foundational guide has unified trajectory table without boilerplate notice
	baseBytes, err := os.ReadFile(filepath.Join(outDir, "BaseConcept.html"))
	if err != nil {
		t.Fatalf("Failed to read BaseConcept.html: %v", err)
	}
	baseHTML := string(baseBytes)
	if strings.Contains(baseHTML, "appendix-path-notice") {
		t.Errorf("BaseConcept.html should not contain appendix-path-notice")
	}
	if !strings.Contains(baseHTML, "Learning Trajectory") {
		t.Errorf("BaseConcept.html missing Learning Trajectory section")
	}
	if !strings.Contains(baseHTML, "appendix-path-table") {
		t.Errorf("BaseConcept.html missing appendix-path-table")
	}
	if !strings.Contains(baseHTML, ">Current<") {
		t.Errorf("BaseConcept.html missing Current offset cell")
	}
	if !strings.Contains(baseHTML, ">+1<") {
		t.Errorf("BaseConcept.html missing +1 offset cell")
	}
	if !strings.Contains(baseHTML, "href=\"intermediate/IntermediateConcept.html\"") {
		t.Errorf("BaseConcept.html missing unlocked link to intermediate/IntermediateConcept.html")
	}
	// Has distant unlock (hop +2), so details dropdown should be present with complete trajectory
	if !strings.Contains(baseHTML, "appendix-trajectory-details") {
		t.Errorf("BaseConcept.html missing appendix-trajectory-details")
	}
	if !strings.Contains(baseHTML, "Show complete trajectory (3 steps)") {
		t.Errorf("BaseConcept.html missing trajectory summary label")
	}
	if !strings.Contains(baseHTML, ">+2<") {
		t.Errorf("BaseConcept.html missing +2 offset in full trajectory")
	}
	if !strings.Contains(baseHTML, "href=\"AdvancedConcept.html\"") {
		t.Errorf("BaseConcept.html missing unlocked link to AdvancedConcept.html")
	}
	if !strings.Contains(baseHTML, "<figure class=\"appendix-figure\">") {
		t.Errorf("BaseConcept.html missing forward localized concept horizon figure")
	}

	// 2. Verify intermediate guide has immediate horizon and no distant details toggle
	midBytes, err := os.ReadFile(filepath.Join(outDir, "intermediate", "IntermediateConcept.html"))
	if err != nil {
		t.Fatalf("Failed to read IntermediateConcept.html: %v", err)
	}
	midHTML := string(midBytes)
	if strings.Contains(midHTML, "appendix-trajectory-details") {
		t.Errorf("IntermediateConcept.html should not have details toggle when all steps are in horizon")
	}
	if !strings.Contains(midHTML, ">-1<") || !strings.Contains(midHTML, ">+1<") {
		t.Errorf("IntermediateConcept.html missing -1 or +1 offset cells")
	}

	// 3. Verify terminal guide has ordered topological trajectory with details for distant prereq
	advBytes, err := os.ReadFile(filepath.Join(outDir, "AdvancedConcept.html"))
	if err != nil {
		t.Fatalf("Failed to read AdvancedConcept.html: %v", err)
	}
	advHTML := string(advBytes)

	if !strings.Contains(advHTML, "Learning Trajectory") {
		t.Errorf("AdvancedConcept.html missing Learning Trajectory section")
	}
	if !strings.Contains(advHTML, "appendix-path-table") {
		t.Errorf("AdvancedConcept.html missing appendix-path-table")
	}
	if !strings.Contains(advHTML, "href=\"BaseConcept.html\"") {
		t.Errorf("AdvancedConcept.html missing link to BaseConcept.html")
	}
	if !strings.Contains(advHTML, "href=\"intermediate/IntermediateConcept.html\"") {
		t.Errorf("AdvancedConcept.html missing link to intermediate/IntermediateConcept.html")
	}
	if !strings.Contains(advHTML, "path-current-step") {
		t.Errorf("AdvancedConcept.html missing path-current-step class")
	}
	if !strings.Contains(advHTML, "(current guide)") {
		t.Errorf("AdvancedConcept.html missing (current guide) label")
	}
	if strings.Contains(advHTML, "Terminal guide") {
		t.Errorf("AdvancedConcept.html should not have terminal guide notice")
	}
	if !strings.Contains(advHTML, ">-2<") {
		t.Errorf("AdvancedConcept.html missing -2 offset cell for distant prerequisite")
	}
	if !strings.Contains(advHTML, ">-1<") {
		t.Errorf("AdvancedConcept.html missing -1 offset cell for immediate prerequisite")
	}
	if !strings.Contains(advHTML, "appendix-trajectory-details") {
		t.Errorf("AdvancedConcept.html missing trajectory details dropdown for -2 prereq")
	}
	if !strings.Contains(advHTML, "definition") || !strings.Contains(advHTML, "explanation") {
		t.Errorf("AdvancedConcept.html missing scope badges in path table")
	}

	// Verify step ordering: BaseConcept -> IntermediateConcept -> AdvancedConcept
	posBase := strings.Index(advHTML, "BaseConcept.html")
	posMid := strings.Index(advHTML, "intermediate/IntermediateConcept.html")
	posAdv := strings.Index(advHTML, "AdvancedConcept</strong>")
	if posBase == -1 || posMid == -1 || posAdv == -1 {
		t.Fatalf("Failed to locate step links in AdvancedConcept.html (posBase=%d, posMid=%d, posAdv=%d)", posBase, posMid, posAdv)
	}
	if !(posBase < posMid && posMid < posAdv) {
		t.Errorf("Expected topological sequence Base < Mid < Adv, but got posBase=%d, posMid=%d, posAdv=%d", posBase, posMid, posAdv)
	}

	// Localized Mermaid diagram should also be present below the table
	if !strings.Contains(advHTML, "<figure class=\"appendix-figure\">") {
		t.Errorf("AdvancedConcept.html missing localized concept horizon figure")
	}

	// 3. Verify resilient warning on graph cycle
	cycleDir := filepath.Join(tmpDir, "cycle_src")
	cycleOut := filepath.Join(tmpDir, "cycle_out")
	if err := os.MkdirAll(cycleDir, 0o755); err != nil {
		t.Fatalf("Failed to create cycle dir: %v", err)
	}
	_ = os.WriteFile(filepath.Join(cycleDir, "manifest.yaml"), []byte("title: Cycle\nscopes: [def]\n"), 0o644)
	_ = os.WriteFile(filepath.Join(cycleDir, "CycleA.md"), []byte("---\nscope: def\nprerequisites: [CycleB]\n---\n# CycleA\n"), 0o644)
	_ = os.WriteFile(filepath.Join(cycleDir, "CycleB.md"), []byte("---\nscope: def\nprerequisites: [CycleA]\n---\n# CycleB\n"), 0o644)

	var warnings []string
	err = BuildSiteWithOptions(BuildOptions{
		DirPath: cycleDir,
		OutDir:  cycleOut,
		Offline: true,
		WarnFunc: func(format string, args ...any) {
			warnings = append(warnings, fmt.Sprintf(format, args...))
		},
	})
	if err != nil {
		t.Fatalf("BuildSiteWithOptions should not fail on cyclic graph: %v", err)
	}
	foundWarning := false
	for _, w := range warnings {
		if strings.Contains(w, "skipping prerequisite path for") && strings.Contains(w, "cycle detected") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Errorf("Expected cycle warning in WarnFunc, got: %v", warnings)
	}
}

func TestBuildSiteSubguideSegments(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	outDir := filepath.Join(tmpDir, "public")

	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("Failed to create source dirs: %v", err)
	}

	manifestContent := `title: Segment Test
scopes:
  - definition
  - lesson
adherences:
  - strict
`
	if err := os.WriteFile(filepath.Join(sourceDir, "manifest.yaml"), []byte(manifestContent), 0o644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	childContent := `---
scope: definition
---
# Child Concept
Detailed child text.
`
	if err := os.WriteFile(filepath.Join(sourceDir, "ChildConcept.md"), []byte(childContent), 0o644); err != nil {
		t.Fatalf("Failed to write ChildConcept.md: %v", err)
	}

	parentContent := `---
scope: lesson
sub_guides:
  - guide: ChildConcept
    adherence: strict
    segment: 2-3
---
# Parent Lesson

## Introduction to Child
Here is the text relating to the child concept.
`
	if err := os.WriteFile(filepath.Join(sourceDir, "ParentLesson.md"), []byte(parentContent), 0o644); err != nil {
		t.Fatalf("Failed to write ParentLesson.md: %v", err)
	}

	if err := BuildSite(sourceDir, outDir); err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	parentBytes, err := os.ReadFile(filepath.Join(outDir, "ParentLesson.html"))
	if err != nil {
		t.Fatalf("Failed to read ParentLesson.html: %v", err)
	}
	parentHTML := string(parentBytes)

	// Verify subtle left margin indicator and segment body wrapper
	if !strings.Contains(parentHTML, "segment-section") {
		t.Errorf("ParentLesson.html missing segment-section container")
	}
	if !strings.Contains(parentHTML, "segment-indicator") {
		t.Errorf("ParentLesson.html missing segment-indicator")
	}
	if !strings.Contains(parentHTML, "ChildConcept.html") {
		t.Errorf("ParentLesson.html missing link to ChildConcept.html in segment indicator")
	}
	if !strings.Contains(parentHTML, `title="Sub-guide: ChildConcept (strict, lines 2–3)"`) {
		t.Errorf("ParentLesson.html missing native tooltip title on segment indicator")
	}
	if !strings.Contains(parentHTML, ">↳</a>") {
		t.Errorf("ParentLesson.html missing ↳ arrow icon in segment indicator")
	}
	if !strings.Contains(parentHTML, "segment-body") {
		t.Errorf("ParentLesson.html missing segment-body")
	}

	// Verify bulky chip elements are removed
	if strings.Contains(parentHTML, "segment-chip") {
		t.Errorf("ParentLesson.html should not contain segment-chip (bulky chip removed)")
	}
	if strings.Contains(parentHTML, "segment-callout") {
		t.Errorf("ParentLesson.html should not contain segment-callout")
	}

	// Verify raw line numbers are removed, while anchor spans are preserved
	if strings.Contains(parentHTML, "line-num") {
		t.Errorf("ParentLesson.html should not contain line-num elements (raw gutter line numbers)")
	}
	if !strings.Contains(parentHTML, `class="line-marker" id="L`) {
		t.Errorf("ParentLesson.html missing anchor line-marker spans for deep linking")
	}
}

