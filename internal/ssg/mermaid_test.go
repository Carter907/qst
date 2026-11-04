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
	if !strings.Contains(diag, `current_lesson["Lesson<br/>Current Lesson"]`) {
		t.Errorf("expected current lesson node in diagram, got:\n%s", diag)
	}
	if !strings.Contains(diag, `prereq_lesson["Lesson<br/>Prereq Lesson"]`) {
		t.Errorf("expected prereq lesson node in diagram, got:\n%s", diag)
	}
	if !strings.Contains(diag, `sub_definition["Definition<br/>Sub Definition"]`) {
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

	// 3. Foundational guide with downstream connections
	foundationalDiag := generateGuideMermaid(guideSub, manifest, filesByID)
	if !strings.HasPrefix(foundationalDiag, "graph LR\n") {
		t.Fatalf("expected graph LR prefix for foundational guide with dependents, got:\n%s", foundationalDiag)
	}
	if !strings.Contains(foundationalDiag, `sub_definition["Definition<br/>Sub Definition"]`) {
		t.Errorf("expected sub definition node in foundational diagram, got:\n%s", foundationalDiag)
	}
	if !strings.Contains(foundationalDiag, "current_lesson -.-> sub_definition") {
		t.Errorf("expected parent decomposition edge in foundational diagram, got:\n%s", foundationalDiag)
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
	if !strings.Contains(diag, `guide_1["Lesson<br/>Guide 1"]`) {
		t.Errorf("expected guide_1 node with scope prefix, got:\n%s", diag)
	}
	if !strings.Contains(diag, `guide_2["Lesson<br/>Guide 2"]`) {
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

func TestWrapNodeTitle(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		targetLen int
		expected  string
	}{
		{
			name:      "single word",
			input:     "Variables",
			targetLen: 24,
			expected:  "Variables",
		},
		{
			name:      "short multi-word within limit",
			input:     "Writing a Simple C",
			targetLen: 24,
			expected:  "Writing a Simple C",
		},
		{
			name:      "balanced multi-word exceeding limit",
			input:     "Automated CLI Markdown Documentation Generator",
			targetLen: 24,
			expected:  "Automated CLI Markdown<br/>Documentation Generator",
		},
		{
			name:      "three line wrap",
			input:     "Testable Cobra Output Channels and Exit Interception",
			targetLen: 24,
			expected:  "Testable Cobra Output<br/>Channels and Exit<br/>Interception",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapNodeTitle(tt.input, tt.targetLen)
			if got != tt.expected {
				t.Errorf("wrapNodeTitle(%q, %d) = %q, expected %q", tt.input, tt.targetLen, got, tt.expected)
			}
		})
	}
}

func TestFormatGuideNodeLabel(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		scope    string
		expected string
	}{
		{
			name:     "with scope and short title",
			id:       "Variable",
			scope:    "definition",
			expected: "Definition<br/>Variable",
		},
		{
			name:     "with scope and wrapped long title",
			id:       "Automated CLI Markdown Documentation Generator",
			scope:    "explanation",
			expected: "Explanation<br/>Automated CLI Markdown<br/>Documentation Generator",
		},
		{
			name:     "empty scope",
			id:       "Automated CLI Markdown Documentation Generator",
			scope:    "",
			expected: "Automated CLI Markdown<br/>Documentation Generator",
		},
		{
			name:     "escapes quotes",
			id:       `Guide with "quotes"`,
			scope:    "definition",
			expected: "Definition<br/>Guide with #quot;quotes#quot;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatGuideNodeLabel(tt.id, tt.scope)
			if got != tt.expected {
				t.Errorf("formatGuideNodeLabel(%q, %q) = %q, expected %q", tt.id, tt.scope, got, tt.expected)
			}
		})
	}
}

func TestBuildGuideMermaid_NarrativeOrdering(t *testing.T) {
	guide := graph.Guide{
		ID: "Parent Guide",
		Metadata: graph.GuideMetadata{
			Scope: "explanation",
			Prerequisites: []string{
				"Prereq Zebra",
				"Prereq Apple",
			},
			SubGuides: []graph.SubGuideRelation{
				{Guide: "Sub Zebra", Segment: "5-10"},
				{Guide: "Sub Apple", Segment: "20-30"},
				{Guide: "Sub Middle", Segment: "12-15"},
				{Guide: "Sub Unsegmented Two"},
				{Guide: "Sub Unsegmented One"},
			},
		},
	}

	getScope := func(id string) string {
		return "definition"
	}

	diag := BuildGuideMermaid(guide, getScope, MermaidOptions{
		MarkdownWrap:     false,
		RequireNeighbors: false,
	})

	// 1. Prerequisites should appear in frontmatter order: Zebra before Apple
	posPrereqZ := strings.Index(diag, "prereq_zebra --> parent_guide")
	posPrereqA := strings.Index(diag, "prereq_apple --> parent_guide")
	if posPrereqZ == -1 || posPrereqA == -1 {
		t.Fatalf("missing prerequisite edges in diagram:\n%s", diag)
	}
	if posPrereqZ >= posPrereqA {
		t.Errorf("expected Prereq Zebra (pos %d) before Prereq Apple (pos %d) in frontmatter order", posPrereqZ, posPrereqA)
	}

	// 2. Sub-guides should appear in narrative start-line order:
	// "Sub Zebra" (5-10) -> "Sub Middle" (12-15) -> "Sub Apple" (20-30)
	// followed by unsegmented: "Sub Unsegmented Two" -> "Sub Unsegmented One"
	edgeSection := diag[strings.Index(diag, "parent_guide -.->"):]
	posSubZ := strings.Index(edgeSection, "sub_zebra")
	posSubM := strings.Index(edgeSection, "sub_middle")
	posSubA := strings.Index(edgeSection, "sub_apple")
	posSubU2 := strings.Index(edgeSection, "sub_unsegmented_two")
	posSubU1 := strings.Index(edgeSection, "sub_unsegmented_one")

	if posSubZ == -1 || posSubM == -1 || posSubA == -1 || posSubU2 == -1 || posSubU1 == -1 {
		t.Fatalf("missing subguide edges in diagram:\n%s", diag)
	}

	if !(posSubZ < posSubM && posSubM < posSubA && posSubA < posSubU2 && posSubU2 < posSubU1) {
		t.Errorf("expected narrative order Zebra (5) < Middle (12) < Apple (20) < UnsegTwo < UnsegOne, got pos: Z=%d, M=%d, A=%d, U2=%d, U1=%d",
			posSubZ, posSubM, posSubA, posSubU2, posSubU1)
	}
}

