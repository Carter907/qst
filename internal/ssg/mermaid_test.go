package ssg

import (
	"strings"
	"testing"

	"github.com/Carter907/qst/internal/graph"
)

func TestCleanMermaidID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple", "simple"},
		{"Guide Name With Spaces", "guide_name_with_spaces"},
		{"11.1 Pointers", "node_11_1_pointers"},
		{"Chapter-12: Arrays", "chapter_12__arrays"},
		{"???", "node_"},
	}

	for _, tt := range tests {
		got := cleanMermaidID(tt.input)
		if got != tt.expected {
			t.Errorf("cleanMermaidID(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestGenerateGuideMermaid(t *testing.T) {
	manifest := graph.Manifest{
		Scopes: []string{"definition", "lesson"},
	}

	guideCurrent := &guideFile{
		guide: graph.Guide{
			ID: "Current Lesson",
			Metadata: graph.GuideMetadata{
				Scope:         "lesson",
				Prerequisites: []string{"Prereq Lesson"},
				SubGuides: []graph.SubGuideRelation{
					{Guide: "Sub Definition"},
				},
			},
		},
		outRelPath: "Current Lesson.html",
	}

	guidePrereq := &guideFile{
		guide: graph.Guide{
			ID: "Prereq Lesson",
			Metadata: graph.GuideMetadata{
				Scope: "lesson",
			},
		},
		outRelPath: "Prereq Lesson.html",
	}

	guideSub := &guideFile{
		guide: graph.Guide{
			ID: "Sub Definition",
			Metadata: graph.GuideMetadata{
				Scope: "definition",
			},
		},
		outRelPath: "sub/Sub Definition.html",
	}

	filesByID := map[string]*guideFile{
		"Current Lesson": guideCurrent,
		"Prereq Lesson":  guidePrereq,
		"Sub Definition": guideSub,
	}

	// 1. Guide with connections
	diag := generateGuideMermaid(guideCurrent, manifest, filesByID)
	if !strings.HasPrefix(diag, "graph LR\n") {
		t.Fatalf("expected graph LR prefix, got:\n%s", diag)
	}
	if !strings.Contains(diag, `current_lesson["Lesson: Current Lesson"]`) {
		t.Errorf("expected current lesson node in diagram, got:\n%s", diag)
	}
	if !strings.Contains(diag, `prereq_lesson["Lesson: Prereq Lesson"]`) {
		t.Errorf("expected prereq lesson node in diagram, got:\n%s", diag)
	}
	if !strings.Contains(diag, `sub_definition["Definition: Sub Definition"]`) {
		t.Errorf("expected sub definition node in diagram, got:\n%s", diag)
	}
	if !strings.Contains(diag, "prereq_lesson --> current_lesson") {
		t.Errorf("expected prerequisite edge, got:\n%s", diag)
	}
	if !strings.Contains(diag, "current_lesson -.->") {
		t.Errorf("expected subguide edge, got:\n%s", diag)
	}
	if !strings.Contains(diag, "class current_lesson currentGuide;") {
		t.Errorf("expected current lesson node styled with currentGuide class, got:\n%s", diag)
	}
	if !strings.Contains(diag, "fill:#f9fafb") {
		t.Errorf("expected monochrome currentGuide fill in diagram, got:\n%s", diag)
	}
	if !strings.Contains(diag, `click prereq_lesson href "Prereq%20Lesson.html"`) {
		t.Errorf("expected click href for prereq lesson, got:\n%s", diag)
	}
	if !strings.Contains(diag, `click sub_definition href "sub/Sub%20Definition.html"`) {
		t.Errorf("expected click href for sub definition, got:\n%s", diag)
	}
	if strings.Contains(diag, `click current_lesson`) {
		t.Errorf("current lesson should not have a self click link, got:\n%s", diag)
	}

	// 2. Leaf guide with zero connections
	leafGuide := &guideFile{
		guide: graph.Guide{
			ID: "Isolated Leaf",
			Metadata: graph.GuideMetadata{
				Scope: "definition",
			},
		},
		outRelPath: "Isolated Leaf.html",
	}
	leafDiag := generateGuideMermaid(leafGuide, manifest, filesByID)
	if leafDiag != "" {
		t.Errorf("expected empty string for leaf guide with no connections, got:\n%s", leafDiag)
	}
}

func TestGenerateGuideMermaidText(t *testing.T) {
	manifest := graph.Manifest{
		Scopes: []string{"definition", "lesson"},
	}

	guides := map[string]graph.Guide{
		"Guide 1": {
			ID: "Guide 1",
			Metadata: graph.GuideMetadata{
				Scope:         "lesson",
				Prerequisites: []string{"Guide 2"},
			},
		},
		"Guide 2": {
			ID: "Guide 2",
			Metadata: graph.GuideMetadata{
				Scope: "lesson",
			},
		},
	}

	diag, err := GenerateGuideMermaidText("Guide 1", guides, manifest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(diag, "```mermaid\ngraph LR\n") {
		t.Errorf("expected mermaid block prefix, got:\n%s", diag)
	}
	if !strings.Contains(diag, `guide_1["Lesson: Guide 1"]`) {
		t.Errorf("expected guide_1 node with scope prefix, got:\n%s", diag)
	}
	if !strings.Contains(diag, `guide_2["Lesson: Guide 2"]`) {
		t.Errorf("expected guide_2 node with scope prefix, got:\n%s", diag)
	}
	if !strings.Contains(diag, "guide_2 --> guide_1") {
		t.Errorf("expected prereq edge in text diagram, got:\n%s", diag)
	}
	if !strings.Contains(diag, "class guide_1 currentGuide;") {
		t.Errorf("expected guide_1 styled with currentGuide class, got:\n%s", diag)
	}
	if !strings.Contains(diag, "fill:#f9fafb") {
		t.Errorf("expected monochrome currentGuide styling, got:\n%s", diag)
	}

	// Non-existent guide
	_, err = GenerateGuideMermaidText("NonExistent", guides, manifest)
	if err == nil {
		t.Errorf("expected error for non-existent guide, got nil")
	}
}

