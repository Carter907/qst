package graph

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRelationshipMutations(t *testing.T) {
	tempDir := t.TempDir()
	guidePath := filepath.Join(tempDir, "GuideA.md")
	initialContent := `---
scope: description
prerequisites: []
sub_guides: []
tags: [test]
---

## Guide A

This is line 1.
This is line 2.
This is line 3.
`
	if err := os.WriteFile(guidePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// 1. Add prerequisite
	if err := AddPrerequisite(guidePath, "Prereq1"); err != nil {
		t.Fatalf("AddPrerequisite() unexpected error: %v", err)
	}
	g, err := ParseGuide(guidePath)
	if err != nil {
		t.Fatalf("ParseGuide() error: %v", err)
	}
	if len(g.Metadata.Prerequisites) != 1 || g.Metadata.Prerequisites[0] != "Prereq1" {
		t.Fatalf("expected Prereq1 in prerequisites, got %v", g.Metadata.Prerequisites)
	}

	// Adding same prerequisite should be idempotent
	if err := AddPrerequisite(guidePath, "Prereq1"); err != nil {
		t.Fatalf("AddPrerequisite() duplicate unexpected error: %v", err)
	}
	g, _ = ParseGuide(guidePath)
	if len(g.Metadata.Prerequisites) != 1 {
		t.Fatalf("expected still 1 prerequisite, got %v", g.Metadata.Prerequisites)
	}

	// 2. Remove prerequisite
	if err := RemovePrerequisite(guidePath, "Prereq1"); err != nil {
		t.Fatalf("RemovePrerequisite() unexpected error: %v", err)
	}
	g, _ = ParseGuide(guidePath)
	if len(g.Metadata.Prerequisites) != 0 {
		t.Fatalf("expected empty prerequisites, got %v", g.Metadata.Prerequisites)
	}

	// Removing non-existent prerequisite should fail
	if err := RemovePrerequisite(guidePath, "NonExistent"); err == nil {
		t.Fatalf("expected error removing non-existent prerequisite")
	}

	// 3. Upsert subguide (insert)
	rel := SubGuideRelation{
		Guide:     "Sub1",
		Adherence: "introductory",
		Segment:   "1-2",
	}
	if err := UpsertSubGuide(guidePath, rel); err != nil {
		t.Fatalf("UpsertSubGuide() insert unexpected error: %v", err)
	}
	g, _ = ParseGuide(guidePath)
	if len(g.Metadata.SubGuides) != 1 || g.Metadata.SubGuides[0].Segment != "1-2" {
		t.Fatalf("expected subguide Sub1 with segment 1-2, got %v", g.Metadata.SubGuides)
	}

	// Upsert subguide (update in-place)
	relUpdate := SubGuideRelation{
		Guide:     "Sub1",
		Adherence: "detailed",
		Segment:   "2-3",
	}
	if err := UpsertSubGuide(guidePath, relUpdate); err != nil {
		t.Fatalf("UpsertSubGuide() update unexpected error: %v", err)
	}
	g, _ = ParseGuide(guidePath)
	if len(g.Metadata.SubGuides) != 1 {
		t.Fatalf("expected exactly 1 subguide after update, got %d", len(g.Metadata.SubGuides))
	}
	if g.Metadata.SubGuides[0].Adherence != "detailed" || g.Metadata.SubGuides[0].Segment != "2-3" {
		t.Fatalf("expected updated adherence detailed and segment 2-3, got %v", g.Metadata.SubGuides[0])
	}

	// 4. Remove subguide
	if err := RemoveSubGuide(guidePath, "Sub1"); err != nil {
		t.Fatalf("RemoveSubGuide() unexpected error: %v", err)
	}
	g, _ = ParseGuide(guidePath)
	if len(g.Metadata.SubGuides) != 0 {
		t.Fatalf("expected empty subguides, got %v", g.Metadata.SubGuides)
	}
}

func TestValidateLinks(t *testing.T) {
	manifest := Manifest{
		Scopes:           []string{"definition", "description", "explanation"},
		Adherences:       []string{"strict", "detailed", "introductory"},
		RelaxedSubguides: false,
	}

	guides := map[string]Guide{
		"Def1": {
			ID:        "Def1",
			Metadata:  GuideMetadata{Scope: "definition"},
			LineCount: 5,
		},
		"Def2": {
			ID:        "Def2",
			Metadata:  GuideMetadata{Scope: "definition"},
			LineCount: 5,
		},
		"Desc1": {
			ID:        "Desc1",
			Metadata:  GuideMetadata{Scope: "description"},
			LineCount: 10,
		},
		"Desc2": {
			ID:        "Desc2",
			Metadata:  GuideMetadata{Scope: "description"},
			LineCount: 10,
		},
		"Exp1": {
			ID:        "Exp1",
			Metadata:  GuideMetadata{Scope: "explanation"},
			LineCount: 15,
		},
	}

	// 1. Prerequisite validation
	// Valid horizontal link (both are definition)
	if err := ValidatePrerequisiteLink(guides, manifest, "Def1", "Def2"); err != nil {
		t.Errorf("expected valid prereq link, got: %v", err)
	}
	// Invalid: self reference
	if err := ValidatePrerequisiteLink(guides, manifest, "Def1", "Def1"); err == nil {
		t.Errorf("expected error for self prereq link")
	}
	// Invalid: scope mismatch
	if err := ValidatePrerequisiteLink(guides, manifest, "Def1", "Desc1"); err == nil {
		t.Errorf("expected error for prereq link with mismatched scope")
	}

	// 2. Subguide validation
	// Valid vertical link (description -> definition)
	if err := ValidateSubGuideLink(guides, manifest, "Desc1", "Def1", "introductory", "1-2"); err != nil {
		t.Errorf("expected valid subguide link, got: %v", err)
	}
	// Invalid: subguide has higher scope
	if err := ValidateSubGuideLink(guides, manifest, "Def1", "Desc1", "introductory", "1-2"); err == nil {
		t.Errorf("expected error when subguide has higher scope than parent")
	}
	// Invalid: skipped rung when relaxed_subguides is false (explanation -> definition)
	if err := ValidateSubGuideLink(guides, manifest, "Exp1", "Def1", "introductory", "1-2"); err == nil {
		t.Errorf("expected error when relaxed_subguides is false and skipping rung")
	}
	// Invalid adherence
	if err := ValidateSubGuideLink(guides, manifest, "Desc1", "Def1", "unknown_adherence", "1-2"); err == nil {
		t.Errorf("expected error for unknown adherence")
	}
	// Invalid segment format
	if err := ValidateSubGuideLink(guides, manifest, "Desc1", "Def1", "introductory", "not-a-range"); err == nil {
		t.Errorf("expected error for invalid segment format")
	}
	// Segment exceeds line count
	if err := ValidateSubGuideLink(guides, manifest, "Desc1", "Def1", "introductory", "1-20"); err == nil {
		t.Errorf("expected error for segment exceeding line count")
	}
}

func TestAnalyzeGuideCoverage(t *testing.T) {
	manifest := Manifest{
		Scopes:         []string{"definition", "description"},
		StrictCoverage: true,
	}

	guide := Guide{
		ID: "TestGuide",
		Metadata: GuideMetadata{
			Scope: "description",
			SubGuides: []SubGuideRelation{
				{Guide: "Def1", Adherence: "introductory", Segment: "2-3"},
			},
		},
		Body: "\nLine 2\nLine 3\nLine 4", // 4 lines
	}

	report, err := AnalyzeGuideCoverage(guide, manifest)
	if err != nil {
		t.Fatalf("AnalyzeGuideCoverage() unexpected error: %v", err)
	}

	if report.TotalBodyLines != 4 {
		t.Errorf("expected 4 total body lines, got %d", report.TotalBodyLines)
	}
	if report.CoveredLines != 2 {
		t.Errorf("expected 2 covered lines, got %d", report.CoveredLines)
	}
	if report.UnmappedLines != 2 {
		t.Errorf("expected 2 unmapped lines, got %d", report.UnmappedLines)
	}
	if len(report.Lines) != 4 {
		t.Fatalf("expected 4 lines in report, got %d", len(report.Lines))
	}
	if len(report.Lines[1].SubGuides) != 1 || report.Lines[1].SubGuides[0] != "Def1" {
		t.Errorf("expected Def1 on line 2, got %v", report.Lines[1].SubGuides)
	}
}

func TestGraphIndexAndAnalytics(t *testing.T) {
	guides := map[string]Guide{
		"A": {
			ID: "A",
			Metadata: GuideMetadata{
				Scope:         "lesson",
				Prerequisites: []string{},
				SubGuides: []SubGuideRelation{
					{Guide: "B"},
				},
			},
		},
		"B": {
			ID: "B",
			Metadata: GuideMetadata{
				Scope:         "explanation",
				Prerequisites: []string{},
				SubGuides: []SubGuideRelation{
					{Guide: "C"},
				},
			},
		},
		"C": {
			ID: "C",
			Metadata: GuideMetadata{
				Scope:         "definition",
				Prerequisites: []string{"D"},
			},
		},
		"D": {
			ID: "D",
			Metadata: GuideMetadata{
				Scope:         "definition",
				Prerequisites: []string{},
			},
		},
		"Orphan": {
			ID: "Orphan",
			Metadata: GuideMetadata{
				Scope: "definition",
			},
		},
	}

	config := Manifest{
		Tours: []TourConfig{
			{Name: "Tour1", Guides: []string{"A", "B"}},
		},
	}

	idx := BuildIndex(guides, config)

	// Inverted lookup tests
	dependents := idx.GetDependents("D")
	if len(dependents) != 1 || dependents[0] != "C" {
		t.Errorf("expected dependents of D to be [C], got %v", dependents)
	}

	parents := idx.GetParentGuides("B")
	if len(parents) != 1 || parents[0] != "A" {
		t.Errorf("expected parent guides of B to be [A], got %v", parents)
	}

	tours := idx.GetTours("A")
	if len(tours) != 1 || tours[0] != "Tour1" {
		t.Errorf("expected tours for A to be [Tour1], got %v", tours)
	}

	if idx.InDegree("C") != 2 { // Prereq from D + Subguide from B
		t.Errorf("expected inDegree(C) = 2, got %d", idx.InDegree("C"))
	}
	if idx.OutDegree("A") != 1 { // Subguide to B
		t.Errorf("expected outDegree(A) = 1, got %d", idx.OutDegree("A"))
	}
	if !idx.IsOrphan("Orphan") {
		t.Errorf("expected Orphan to be orphan, got false")
	}
	if idx.IsOrphan("A") {
		t.Errorf("expected A not to be orphan, got true")
	}

	// Analytics tests
	chain := LongestPrerequisiteChain(guides)
	if len(chain) != 2 || chain[0] != "D" || chain[1] != "C" {
		t.Errorf("expected longest prereq chain [D, C], got %v", chain)
	}

	depth := MaxSubguideDepth(guides)
	if depth != 3 { // A -> B -> C = 3 levels
		t.Errorf("expected max subguide depth 3, got %d", depth)
	}
}

func TestRenameGuide(t *testing.T) {
	tempDir := t.TempDir()

	manifest := `title: Rename Test
scopes:
  - definition
  - explanation
adherences:
  - strict
tours:
  - name: Core
    guides:
      - OldGuide
`
	if err := os.WriteFile(filepath.Join(tempDir, "manifest.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	oldContent := `---
scope: definition
prerequisites: []
sub_guides: []
tags: [basic]
---
# Old Guide
`
	if err := os.WriteFile(filepath.Join(tempDir, "OldGuide.md"), []byte(oldContent), 0644); err != nil {
		t.Fatalf("failed to write OldGuide: %v", err)
	}

	parentContent := `---
scope: explanation
prerequisites: []
sub_guides:
  - guide: OldGuide
    adherence: strict
    segment: "1-1"
tags: []
---
# Parent
`
	if err := os.WriteFile(filepath.Join(tempDir, "Parent.md"), []byte(parentContent), 0644); err != nil {
		t.Fatalf("failed to write Parent: %v", err)
	}

	prereqContent := `---
scope: definition
prerequisites:
  - OldGuide
sub_guides: []
tags: []
---
# Dependent
`
	if err := os.WriteFile(filepath.Join(tempDir, "Dependent.md"), []byte(prereqContent), 0644); err != nil {
		t.Fatalf("failed to write Dependent: %v", err)
	}

	updated, err := RenameGuide(tempDir, "OldGuide", "NewGuide")
	if err != nil {
		t.Fatalf("RenameGuide() unexpected error: %v", err)
	}
	if updated != 3 { // Parent subguide + Dependent prereq + Tour
		t.Errorf("expected 3 updated references, got %d", updated)
	}

	// Verify old file gone, new file exists
	if _, err := os.Stat(filepath.Join(tempDir, "OldGuide.md")); !os.IsNotExist(err) {
		t.Errorf("expected OldGuide.md to be removed")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "NewGuide.md")); err != nil {
		t.Errorf("expected NewGuide.md to exist: %v", err)
	}

	// Verify references in Parent
	parentG, err := ParseGuide(filepath.Join(tempDir, "Parent.md"))
	if err != nil {
		t.Fatalf("failed to parse Parent: %v", err)
	}
	if len(parentG.Metadata.SubGuides) != 1 || parentG.Metadata.SubGuides[0].Guide != "NewGuide" {
		t.Errorf("expected subguide NewGuide in Parent, got %v", parentG.Metadata.SubGuides)
	}

	// Verify references in Dependent
	depG, err := ParseGuide(filepath.Join(tempDir, "Dependent.md"))
	if err != nil {
		t.Fatalf("failed to parse Dependent: %v", err)
	}
	if len(depG.Metadata.Prerequisites) != 1 || depG.Metadata.Prerequisites[0] != "NewGuide" {
		t.Errorf("expected prerequisite NewGuide in Dependent, got %v", depG.Metadata.Prerequisites)
	}

	// Verify references in Manifest
	cfg, err := ParseConfig(tempDir)
	if err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}
	if len(cfg.Tours[0].Guides) != 1 || cfg.Tours[0].Guides[0] != "NewGuide" {
		t.Errorf("expected tour guide NewGuide, got %v", cfg.Tours[0].Guides)
	}
}

