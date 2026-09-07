package graph

import (
	"testing"
)

func TestValidateGraph(t *testing.T) {
	mockConfig := Manifest{
		Scopes: []string{"definition", "description", "explanation", "lesson"},
		Adherences: []string{"vague", "introductory", "detailed", "strict"},
	}
	tests := []struct {
		name    string
		guides  map[string]Guide
		wantErr bool
	}{
		{
			name: "valid single guide",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, Metadata: GuideMetadata{
						Scope: "definition",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid scope",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, Metadata: GuideMetadata{
						Scope: "invalid_scope",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid prerequisite",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, Metadata: GuideMetadata{
						Scope:         "lesson",
						Prerequisites: []string{"guide2"},
					},
				},
				"guide2": {
					ID:         "guide2",
					HasContent: true, Metadata: GuideMetadata{
						Scope: "lesson",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unknown prerequisite",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, Metadata: GuideMetadata{
						Scope:         "lesson",
						Prerequisites: []string{"unknown_guide"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "mismatched prerequisite scope",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, Metadata: GuideMetadata{
						Scope:         "lesson",
						Prerequisites: []string{"guide2"},
					},
				},
				"guide2": {
					ID:         "guide2",
					HasContent: true, Metadata: GuideMetadata{
						Scope: "definition",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid sub-guide",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, Metadata: GuideMetadata{
						Scope:     "lesson",
						SubGuides: []SubGuideRelation{{Guide: "guide2", Adherence: "detailed", Segment: "1-10"}},
					},
				},
				"guide2": {
					ID:         "guide2",
					HasContent: true, Metadata: GuideMetadata{
						Scope: "explanation",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unknown sub-guide",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, Metadata: GuideMetadata{
						Scope:     "lesson",
						SubGuides: []SubGuideRelation{{Guide: "unknown_guide"}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid sub-guide scope (jumping level without relaxed_subguides)",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, Metadata: GuideMetadata{
						Scope:     "lesson", // 4
						SubGuides: []SubGuideRelation{{Guide: "guide2", Adherence: "detailed", Segment: "1-10"}},
					},
				},
				"guide2": {
					ID:         "guide2",
					HasContent: true, Metadata: GuideMetadata{
						Scope: "description", // 2 (differs by 2)
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid sub-guide scope (larger scope)",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, Metadata: GuideMetadata{
						Scope:     "explanation",
						SubGuides: []SubGuideRelation{{Guide: "guide2", Adherence: "detailed", Segment: "1-10"}},
					},
				},
				"guide2": {
					ID:         "guide2",
					HasContent: true, Metadata: GuideMetadata{
						Scope: "lesson", // larger scope than Explanation
					},
				},
			},
			wantErr: true,
		},
		{
			name: "cyclic prerequisites",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, Metadata: GuideMetadata{
						Scope:         "lesson",
						Prerequisites: []string{"guide2"},
					},
				},
				"guide2": {
					ID:         "guide2",
					HasContent: true, Metadata: GuideMetadata{
						Scope:         "lesson",
						Prerequisites: []string{"guide1"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "transitive cycle in prerequisites",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, Metadata: GuideMetadata{
						Scope:         "lesson",
						Prerequisites: []string{"guide2"},
					},
				},
				"guide2": {
					ID:         "guide2",
					HasContent: true, Metadata: GuideMetadata{
						Scope:         "lesson",
						Prerequisites: []string{"guide3"},
					},
				},
				"guide3": {
					ID:         "guide3",
					HasContent: true, Metadata: GuideMetadata{
						Scope:         "lesson",
						Prerequisites: []string{"guide1"},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGraph(tt.guides, mockConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGraph() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckAcyclic(t *testing.T) {
	tests := []struct {
		name    string
		guides  map[string]Guide
		wantErr bool
	}{
		{
			name:    "empty graph",
			guides:  map[string]Guide{},
			wantErr: false,
		},
		{
			name: "single node without dependencies",
			guides: map[string]Guide{
				"a": {ID: "a"},
			},
			wantErr: false,
		},
		{
			name: "linear acyclic chain",
			guides: map[string]Guide{
				"a": {ID: "a", HasContent: true, Metadata: GuideMetadata{Prerequisites: []string{"b"}}},
				"b": {ID: "b", HasContent: true, Metadata: GuideMetadata{Prerequisites: []string{"c"}}},
				"c": {ID: "c"},
			},
			wantErr: false,
		},
		{
			name: "diamond acyclic graph",
			guides: map[string]Guide{
				"a": {ID: "a", HasContent: true, Metadata: GuideMetadata{Prerequisites: []string{"b", "c"}}},
				"b": {ID: "b", HasContent: true, Metadata: GuideMetadata{Prerequisites: []string{"d"}}},
				"c": {ID: "c", HasContent: true, Metadata: GuideMetadata{Prerequisites: []string{"d"}}},
				"d": {ID: "d"},
			},
			wantErr: false,
		},
		{
			name: "self-referential cycle",
			guides: map[string]Guide{
				"a": {ID: "a", HasContent: true, Metadata: GuideMetadata{Prerequisites: []string{"a"}}},
			},
			wantErr: true,
		},
		{
			name: "sub_guide cycle",
			guides: map[string]Guide{
				"a": {ID: "a", HasContent: true, Metadata: GuideMetadata{SubGuides: []SubGuideRelation{{Guide: "b"}}}},
				"b": {ID: "b", HasContent: true, Metadata: GuideMetadata{SubGuides: []SubGuideRelation{{Guide: "a"}}}},
			},
			wantErr: true,
		},
		{
			name: "mixed prerequisites and sub_guides cycle",
			guides: map[string]Guide{
				"a": {ID: "a", HasContent: true, Metadata: GuideMetadata{Prerequisites: []string{"b"}}},
				"b": {ID: "b", HasContent: true, Metadata: GuideMetadata{SubGuides: []SubGuideRelation{{Guide: "c"}}}},
				"c": {ID: "c", HasContent: true, Metadata: GuideMetadata{Prerequisites: []string{"a"}}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckAcyclic(tt.guides)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAcyclic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGraphRelaxedSubguides(t *testing.T) {
	mockConfig := Manifest{
		Scopes: []string{"definition", "description", "explanation", "lesson"},
		Adherences: []string{"vague", "introductory", "detailed", "strict"},

		RelaxedSubguides: true,
	}

	guides := map[string]Guide{
		"guide1": {
			ID:         "guide1",
			HasContent: true, Metadata: GuideMetadata{
				Scope:     "lesson", // 4
				SubGuides: []SubGuideRelation{{Guide: "guide2", Adherence: "detailed", Segment: "1-10"}},
			},
		},
		"guide2": {
			ID:         "guide2",
			HasContent: true, Metadata: GuideMetadata{
				Scope: "description", // 2 (differs by 2)
			},
		},
	}

	err := ValidateGraph(guides, mockConfig)
	if err != nil {
		t.Errorf("ValidateGraph() with relaxed_subguides=true returned error = %v, want no error", err)
	}
}

func TestValidateGraphStrictCoverage(t *testing.T) {
	mockConfig := Manifest{
		Scopes: []string{"definition", "description", "explanation", "lesson"},
		Adherences: []string{"vague", "introductory", "detailed", "strict"},
		StrictCoverage: true,
	}

	tests := []struct {
		name    string
		guides  map[string]Guide
		wantErr bool
	}{
		{
			name: "full coverage valid",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, 
					LineCount: 10,
					Metadata: GuideMetadata{
						Scope:     "lesson",
						SubGuides: []SubGuideRelation{
							{Guide: "guide2", Adherence: "detailed", Segment: "1-5"},
							{Guide: "guide3", Adherence: "detailed", Segment: "6-10"},
						},
					},
				},
				"guide2": {ID: "guide2", HasContent: true, Metadata: GuideMetadata{Scope: "explanation"}},
				"guide3": {ID: "guide3", HasContent: true, Metadata: GuideMetadata{Scope: "explanation"}},
			},
			wantErr: false,
		},
		{
			name: "missing coverage at start",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, 
					LineCount: 10,
					Metadata: GuideMetadata{
						Scope:     "lesson",
						SubGuides: []SubGuideRelation{
							{Guide: "guide2", Adherence: "detailed", Segment: "2-10"},
						},
					},
				},
				"guide2": {ID: "guide2", HasContent: true, Metadata: GuideMetadata{Scope: "explanation"}},
			},
			wantErr: true, // missing line 1
		},
		{
			name: "missing coverage in middle",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, 
					LineCount: 10,
					Metadata: GuideMetadata{
						Scope:     "lesson",
						SubGuides: []SubGuideRelation{
							{Guide: "guide2", Adherence: "detailed", Segment: "1-4"},
							{Guide: "guide3", Adherence: "detailed", Segment: "6-10"},
						},
					},
				},
				"guide2": {ID: "guide2", HasContent: true, Metadata: GuideMetadata{Scope: "explanation"}},
				"guide3": {ID: "guide3", HasContent: true, Metadata: GuideMetadata{Scope: "explanation"}},
			},
			wantErr: true, // missing line 5
		},
		{
			name: "missing coverage at end",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, 
					LineCount: 10,
					Metadata: GuideMetadata{
						Scope:     "lesson",
						SubGuides: []SubGuideRelation{
							{Guide: "guide2", Adherence: "detailed", Segment: "1-9"},
						},
					},
				},
				"guide2": {ID: "guide2", HasContent: true, Metadata: GuideMetadata{Scope: "explanation"}},
			},
			wantErr: true, // missing line 10
		},
		{
			name: "overlapping segments",
			guides: map[string]Guide{
				"guide1": {
					ID:         "guide1",
					HasContent: true, 
					LineCount: 10,
					Metadata: GuideMetadata{
						Scope:     "lesson",
						SubGuides: []SubGuideRelation{
							{Guide: "guide2", Adherence: "detailed", Segment: "1-5"},
							{Guide: "guide3", Adherence: "detailed", Segment: "5-10"},
						},
					},
				},
				"guide2": {ID: "guide2", HasContent: true, Metadata: GuideMetadata{Scope: "explanation"}},
				"guide3": {ID: "guide3", HasContent: true, Metadata: GuideMetadata{Scope: "explanation"}},
			},
			wantErr: true, // overlaps at 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGraph(tt.guides, mockConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGraph() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
