package ssg

import (
	"strings"
	"testing"

	"github.com/Carter907/qst/internal/graph"
)

func TestGenerateIndexMermaid(t *testing.T) {
	manifest := graph.Manifest{
		Scopes: []string{"definition", "explanation", "lesson"},
	}

	guideA := &guideFile{
		guide: graph.Guide{
			ID: "Guide A",
			Metadata: graph.GuideMetadata{
				Scope: "lesson",
				SubGuides: []graph.SubGuideRelation{
					{Guide: "Guide B"},
				},
			},
		},
		outRelPath: "Guide A.html",
	}

	guideC := &guideFile{
		guide: graph.Guide{
			ID: "Guide C",
			Metadata: graph.GuideMetadata{
				Scope: "lesson",
				Prerequisites: []string{
					"Guide A",
				},
			},
		},
		outRelPath: "Guide C.html",
	}

	guideB := &guideFile{
		guide: graph.Guide{
			ID: "Guide B",
			Metadata: graph.GuideMetadata{
				Scope: "definition",
			},
		},
		outRelPath: "sub/Guide B.html",
	}

	guideFiles := []*guideFile{guideB, guideA, guideC}
	filesByID := map[string]*guideFile{
		"Guide A": guideA,
		"Guide B": guideB,
		"Guide C": guideC,
	}

	result := generateIndexMermaid(manifest, guideFiles, filesByID)

	if !strings.HasPrefix(result, "graph TD\n") {
		t.Fatalf("Expected graph TD prefix, got:\n%s", result)
	}

	// Verify scope subgraphs and direction LR
	if !strings.Contains(result, `subgraph scope_lesson ["Scope: lesson"]`) {
		t.Errorf("Expected subgraph for scope 'lesson', got:\n%s", result)
	}
	if !strings.Contains(result, `subgraph scope_definition ["Scope: definition"]`) {
		t.Errorf("Expected subgraph for scope 'definition', got:\n%s", result)
	}
	if !strings.Contains(result, "direction LR") {
		t.Errorf("Expected 'direction LR' in subgraphs, got:\n%s", result)
	}

	// Verify top-down hierarchy: higher scope (lesson) must appear before lower scope (definition)
	idxLesson := strings.Index(result, `subgraph scope_lesson`)
	idxDef := strings.Index(result, `subgraph scope_definition`)
	if idxLesson == -1 || idxDef == -1 || idxLesson >= idxDef {
		t.Errorf("Expected scope_lesson to appear before scope_definition in output, got:\n%s", result)
	}

	// Verify node declarations
	if !strings.Contains(result, `guide_a["Guide A"]`) {
		t.Errorf("Expected node declaration for Guide A, got:\n%s", result)
	}
	if !strings.Contains(result, `guide_b["Guide B"]`) {
		t.Errorf("Expected node declaration for Guide B, got:\n%s", result)
	}
	if !strings.Contains(result, `guide_c["Guide C"]`) {
		t.Errorf("Expected node declaration for Guide C, got:\n%s", result)
	}

	// Verify intra-scope horizontal prerequisite edge (Guide A -> Guide C inside scope_lesson)
	if !strings.Contains(result, "guide_a --> guide_c") {
		t.Errorf("Expected intra-scope prerequisite edge 'guide_a --> guide_c', got:\n%s", result)
	}

	// Verify cross-scope hierarchical subguide edge (Guide A -.-> Guide B)
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
	if !strings.Contains(result, `click guide_c href "Guide%20C.html"`) {
		t.Errorf("Expected click directive for Guide C, got:\n%s", result)
	}
}

func TestGenerateIndexMermaid_FallbackScopesAndEmpty(t *testing.T) {
	// Empty guide list
	if got := generateIndexMermaid(graph.Manifest{}, nil, nil); got != "" {
		t.Errorf("Expected empty string for nil guideFiles, got %q", got)
	}

	// Guides without scope
	g1 := &guideFile{
		guide: graph.Guide{
			ID: "NoScope",
		},
		outRelPath: "NoScope.html",
	}
	res := generateIndexMermaid(graph.Manifest{}, []*guideFile{g1}, map[string]*guideFile{"NoScope": g1})
	if !strings.Contains(res, `subgraph scope_other ["Scope: other"]`) {
		t.Errorf("Expected fallback subgraph for unscoped guide, got:\n%s", res)
	}
}

func TestGenerateIndexMermaid_IDCollisionAvoidance(t *testing.T) {
	// Guide with ID matching a scope candidate name
	g := &guideFile{
		guide: graph.Guide{
			ID: "scope_lesson",
			Metadata: graph.GuideMetadata{
				Scope: "lesson",
			},
		},
		outRelPath: "scope_lesson.html",
	}

	manifest := graph.Manifest{
		Scopes: []string{"lesson"},
	}

	res := generateIndexMermaid(manifest, []*guideFile{g}, map[string]*guideFile{"scope_lesson": g})
	// Node ID is scope_lesson, Subgraph ID should be differentiated as scope_lesson_2
	if !strings.Contains(res, `subgraph scope_lesson_2 ["Scope: lesson"]`) {
		t.Errorf("Expected collision avoidance in subgraph ID, got:\n%s", res)
	}
	if !strings.Contains(res, `scope_lesson["scope_lesson"]`) {
		t.Errorf("Expected node declaration scope_lesson, got:\n%s", res)
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
