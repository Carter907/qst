package ssg

import (
	"strings"
	"testing"

	"github.com/Carter907/qst/internal/graph"
)

func TestGenerateIndexMermaid(t *testing.T) {
	guideA := &guideFile{
		guide: graph.Guide{
			ID: "Guide A",
			Metadata: graph.GuideMetadata{
				SubGuides: []graph.SubGuideRelation{
					{Guide: "Guide B"},
				},
			},
		},
		outRelPath: "Guide A.html",
	}

	guideB := &guideFile{
		guide: graph.Guide{
			ID: "Guide B",
			Metadata: graph.GuideMetadata{
				Prerequisites: []string{
					"Guide A",
				},
			},
		},
		outRelPath: "sub/Guide B.html",
	}

	guideFiles := []*guideFile{guideB, guideA}
	filesByID := map[string]*guideFile{
		"Guide A": guideA,
		"Guide B": guideB,
	}

	result := generateIndexMermaid(guideFiles, filesByID)

	if !strings.HasPrefix(result, "graph TD\n") {
		t.Fatalf("Expected graph TD prefix, got:\n%s", result)
	}

	// Verify node declarations
	if !strings.Contains(result, `guide_a["Guide A"]`) {
		t.Errorf("Expected node declaration for Guide A, got:\n%s", result)
	}
	if !strings.Contains(result, `guide_b["Guide B"]`) {
		t.Errorf("Expected node declaration for Guide B, got:\n%s", result)
	}

	// Verify solid arrow for prerequisite (Guide A -> Guide B)
	if !strings.Contains(result, "guide_a --> guide_b") {
		t.Errorf("Expected prerequisite edge 'guide_a --> guide_b', got:\n%s", result)
	}

	// Verify dotted arrow for subguide (Guide A -.-> Guide B)
	if !strings.Contains(result, "guide_a -.-> guide_b") {
		t.Errorf("Expected subguide edge 'guide_a -.-> guide_b', got:\n%s", result)
	}

	// Verify clickable href links
	if !strings.Contains(result, `click guide_a href "Guide%20A.html"`) {
		t.Errorf("Expected click directive for Guide A, got:\n%s", result)
	}
	if !strings.Contains(result, `click guide_b href "sub/Guide%20B.html"`) {
		t.Errorf("Expected click directive for Guide B, got:\n%s", result)
	}
}

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
