package graph

import (
	"path/filepath"
	"testing"
)

func TestTopologicalSort_Empty(t *testing.T) {
	guides := make(map[string]Guide)
	config := Manifest{}
	path, err := TopologicalSort(guides, config, PathOptions{IncludeSubguides: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path.Steps) != 0 {
		t.Fatalf("expected 0 steps, got %d", len(path.Steps))
	}
}

func TestTopologicalSort_LinearChain(t *testing.T) {
	guides := map[string]Guide{
		"A": {ID: "A", Metadata: GuideMetadata{Scope: "definition"}},
		"B": {ID: "B", Metadata: GuideMetadata{Scope: "definition", Prerequisites: []string{"A"}}},
		"C": {ID: "C", Metadata: GuideMetadata{Scope: "definition", Prerequisites: []string{"B"}}},
	}
	config := Manifest{Scopes: []string{"definition"}}

	path, err := TopologicalSort(guides, config, PathOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"A", "B", "C"}
	if len(path.Steps) != len(expected) {
		t.Fatalf("expected %d steps, got %d", len(expected), len(path.Steps))
	}
	for i, step := range path.Steps {
		if step.GuideID != expected[i] {
			t.Errorf("step %d: expected %s, got %s", i+1, expected[i], step.GuideID)
		}
	}
}

func TestTopologicalSort_Diamond(t *testing.T) {
	guides := map[string]Guide{
		"A": {ID: "A", Metadata: GuideMetadata{Scope: "definition"}},
		"B": {ID: "B", Metadata: GuideMetadata{Scope: "definition", Prerequisites: []string{"A"}}},
		"C": {ID: "C", Metadata: GuideMetadata{Scope: "definition", Prerequisites: []string{"A"}}},
		"D": {ID: "D", Metadata: GuideMetadata{Scope: "definition", Prerequisites: []string{"B", "C"}}},
	}
	config := Manifest{Scopes: []string{"definition"}}

	path, err := TopologicalSort(guides, config, PathOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path.Steps[0].GuideID != "A" {
		t.Errorf("step 1: expected A, got %s", path.Steps[0].GuideID)
	}
	if path.Steps[3].GuideID != "D" {
		t.Errorf("step 4: expected D, got %s", path.Steps[3].GuideID)
	}
}

func TestTopologicalSort_SegmentNarrativeTieBreaking(t *testing.T) {
	// Parent guide P has two independent subguides S1 and S2.
	// S2 appears at line 5, S1 appears at line 20.
	// Even though S1 comes first alphabetically, S2 appears earlier in parent P text!
	guides := map[string]Guide{
		"S1": {ID: "S1", Metadata: GuideMetadata{Scope: "definition"}},
		"S2": {ID: "S2", Metadata: GuideMetadata{Scope: "definition"}},
		"P": {
			ID: "P",
			Metadata: GuideMetadata{
				Scope: "description",
				SubGuides: []SubGuideRelation{
					{Guide: "S2", Segment: "5-10"},
					{Guide: "S1", Segment: "20-30"},
				},
			},
		},
	}
	config := Manifest{Scopes: []string{"definition", "description"}}

	path, err := PrerequisitePath("P", guides, config, PathOptions{IncludeSubguides: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(path.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(path.Steps))
	}
	// S2 must come before S1 due to segment narrative line ordering!
	if path.Steps[0].GuideID != "S2" {
		t.Errorf("step 1: expected S2 (segment line 5), got %s", path.Steps[0].GuideID)
	}
	if path.Steps[1].GuideID != "S1" {
		t.Errorf("step 2: expected S1 (segment line 20), got %s", path.Steps[1].GuideID)
	}
	if path.Steps[2].GuideID != "P" {
		t.Errorf("step 3: expected P, got %s", path.Steps[2].GuideID)
	}
}

func TestPrerequisitePath_Isolation(t *testing.T) {
	guides := map[string]Guide{
		"Unrelated1": {ID: "Unrelated1", Metadata: GuideMetadata{Scope: "definition"}},
		"Unrelated2": {ID: "Unrelated2", Metadata: GuideMetadata{Scope: "definition", Prerequisites: []string{"Unrelated1"}}},
		"Root":       {ID: "Root", Metadata: GuideMetadata{Scope: "definition"}},
		"Middle":     {ID: "Middle", Metadata: GuideMetadata{Scope: "definition", Prerequisites: []string{"Root"}}},
		"Target":     {ID: "Target", Metadata: GuideMetadata{Scope: "description", SubGuides: []SubGuideRelation{{Guide: "Middle"}}}},
	}
	config := Manifest{Scopes: []string{"definition", "description"}}

	path, err := PrerequisitePath("Target", guides, config, PathOptions{IncludeSubguides: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only contain Root, Middle, and Target (3 guides)
	if len(path.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(path.Steps))
	}
	expected := []string{"Root", "Middle", "Target"}
	for i, step := range path.Steps {
		if step.GuideID != expected[i] {
			t.Errorf("step %d: expected %s, got %s", i+1, expected[i], step.GuideID)
		}
	}
}

func TestUnlocksPath(t *testing.T) {
	guides := map[string]Guide{
		"Base":  {ID: "Base", Metadata: GuideMetadata{Scope: "definition"}},
		"Next1": {ID: "Next1", Metadata: GuideMetadata{Scope: "definition", Prerequisites: []string{"Base"}}},
		"Next2": {ID: "Next2", Metadata: GuideMetadata{Scope: "definition", Prerequisites: []string{"Next1"}}},
		"Other": {ID: "Other", Metadata: GuideMetadata{Scope: "definition"}},
	}
	config := Manifest{Scopes: []string{"definition"}}

	path, err := UnlocksPath("Base", guides, config, PathOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"Base", "Next1", "Next2"}
	if len(path.Steps) != len(expected) {
		t.Fatalf("expected %d steps, got %d", len(expected), len(path.Steps))
	}
	for i, step := range path.Steps {
		if step.GuideID != expected[i] {
			t.Errorf("step %d: expected %s, got %s", i+1, expected[i], step.GuideID)
		}
	}
}

func TestTopologicalSort_Cycle(t *testing.T) {
	guides := map[string]Guide{
		"A": {ID: "A", Metadata: GuideMetadata{Scope: "definition", Prerequisites: []string{"B"}}},
		"B": {ID: "B", Metadata: GuideMetadata{Scope: "definition", Prerequisites: []string{"A"}}},
	}
	config := Manifest{Scopes: []string{"definition"}}

	_, err := TopologicalSort(guides, config, PathOptions{})
	if err == nil {
		t.Fatalf("expected cycle error, got nil")
	}
}

func TestPrerequisitePath_RealExamples(t *testing.T) {
	// Test against examples/C_programming
	cDir := filepath.Join("..", "..", "examples", "C_programming")
	guides, err := ParseGraph(cDir)
	if err != nil {
		t.Fatalf("failed to parse C_programming guides: %v", err)
	}
	config, err := ParseConfig(cDir)
	if err != nil {
		t.Fatalf("failed to parse C_programming config: %v", err)
	}

	// 1. Path for Variables and Assignment
	path, err := PrerequisitePath("Variables and Assignment", guides, config, PathOptions{IncludeSubguides: true})
	if err != nil {
		t.Fatalf("failed to compute path for Variables and Assignment: %v", err)
	}
	if len(path.Steps) != 3 {
		t.Fatalf("expected 3 steps for Variables and Assignment, got %d", len(path.Steps))
	}
	if path.Steps[0].GuideID != "Variable" {
		t.Errorf("step 1: expected Variable, got %s", path.Steps[0].GuideID)
	}
	if path.Steps[1].GuideID != "Assignment" {
		t.Errorf("step 2: expected Assignment, got %s", path.Steps[1].GuideID)
	}
	if path.Steps[2].GuideID != "Variables and Assignment" {
		t.Errorf("step 3: expected Variables and Assignment, got %s", path.Steps[2].GuideID)
	}

	// 2. Global TopologicalSort across all 21 guides
	allPath, err := TopologicalSort(guides, config, PathOptions{IncludeSubguides: true})
	if err != nil {
		t.Fatalf("failed global topological sort: %v", err)
	}
	if len(allPath.Steps) != len(guides) {
		t.Fatalf("expected %d steps in global sort, got %d", len(guides), len(allPath.Steps))
	}
	// Verify that in allPath, every prerequisite/subguide appears before its dependent
	seen := make(map[string]bool)
	for _, step := range allPath.Steps {
		g := guides[step.GuideID]
		for _, p := range g.Metadata.Prerequisites {
			if !seen[p] {
				t.Errorf("guide %s appeared before its prerequisite %s", step.GuideID, p)
			}
		}
		for _, s := range g.Metadata.SubGuides {
			if !seen[s.Guide] {
				t.Errorf("guide %s appeared before its subguide %s", step.GuideID, s.Guide)
			}
		}
		seen[step.GuideID] = true
	}
}
