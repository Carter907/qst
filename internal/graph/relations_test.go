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
