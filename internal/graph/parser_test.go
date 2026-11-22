package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseGraph(t *testing.T) {
	tempDir := t.TempDir()

	validGuide := `---
scope: definition
prerequisites:
  - other
sub_guides:
  - guide: sub
tags:
  - a
  - b
---
# Content here
`
	err := os.WriteFile(filepath.Join(tempDir, "valid.md"), []byte(validGuide), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	invalidYamlGuide := `---
scope: [invalid yaml
---
# Content here
`
	err = os.WriteFile(filepath.Join(tempDir, "invalid_yaml.md"), []byte(invalidYamlGuide), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	noFrontmatterGuide := `# Just some content
No frontmatter at all.
`
	err = os.WriteFile(filepath.Join(tempDir, "no_frontmatter.md"), []byte(noFrontmatterGuide), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create a sub directory to ensure it is ignored
	err = os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
	if err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Test parsing the whole dir (will fail because of invalid files)
	_, err = ParseGraph(tempDir)
	if err == nil {
		t.Error("ParseGraph() expected error due to invalid files, got nil")
	}

	// Test ParseGuide directly for individual behaviors
	t.Run("valid guide", func(t *testing.T) {
		g, err := ParseGuide(filepath.Join(tempDir, "valid.md"))
		if err != nil {
			t.Fatalf("ParseGuide() unexpected error: %v", err)
		}
		if g.ID != "valid" {
			t.Errorf("expected ID 'valid', got '%s'", g.ID)
		}
		if g.Metadata.Scope != "definition" {
			t.Errorf("expected Scope 'definition', got '%s'", g.Metadata.Scope)
		}
		if len(g.Metadata.Prerequisites) != 1 || g.Metadata.Prerequisites[0] != "other" {
			t.Errorf("expected 1 prerequisite 'other', got %v", g.Metadata.Prerequisites)
		}
		if g.Body != "# Content here\n" {
			t.Errorf("expected Body '# Content here\\n', got %q", g.Body)
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		_, err := ParseGuide(filepath.Join(tempDir, "invalid_yaml.md"))
		if err == nil {
			t.Error("ParseGuide() expected error for invalid yaml, got nil")
		}
	})

	t.Run("no frontmatter", func(t *testing.T) {
		_, err := ParseGuide(filepath.Join(tempDir, "no_frontmatter.md"))
		if err == nil {
			t.Error("ParseGuide() expected error for no frontmatter, got nil")
		}
	})

	t.Run("parse config quest.yaml fallback", func(t *testing.T) {
		configDir := t.TempDir()
		questYaml := "title: Fallback Test\ndescription: Testing quest.yaml\n"
		if err := os.WriteFile(filepath.Join(configDir, "quest.yaml"), []byte(questYaml), 0644); err != nil {
			t.Fatalf("failed to write quest.yaml: %v", err)
		}
		cfg, err := ParseConfig(configDir)
		if err != nil {
			t.Fatalf("ParseConfig() failed with quest.yaml: %v", err)
		}
		if cfg.Title != "Fallback Test" {
			t.Errorf("expected Title 'Fallback Test', got %q", cfg.Title)
		}
	})

	t.Run("recursive subdirectory parsing", func(t *testing.T) {
		graphDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(graphDir, "ch1"), 0755); err != nil {
			t.Fatalf("failed to create ch1: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(graphDir, "ch2"), 0755); err != nil {
			t.Fatalf("failed to create ch2: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(graphDir, "public"), 0755); err != nil {
			t.Fatalf("failed to create public: %v", err)
		}

		g1 := "---\nscope: definition\nprerequisites: []\nsub_guides: []\ntags: []\n---\n# G1\n"
		g2 := "---\nscope: description\nprerequisites: []\nsub_guides: []\ntags: []\n---\n# G2\n"
		pubG := "---\nscope: definition\nprerequisites: []\nsub_guides: []\ntags: []\n---\n# Pub\n"

		_ = os.WriteFile(filepath.Join(graphDir, "ch1", "Guide1.md"), []byte(g1), 0644)
		_ = os.WriteFile(filepath.Join(graphDir, "ch2", "Guide2.md"), []byte(g2), 0644)
		_ = os.WriteFile(filepath.Join(graphDir, "public", "Ignored.md"), []byte(pubG), 0644)
		_ = os.WriteFile(filepath.Join(graphDir, ".hidden.md"), []byte(pubG), 0644)

		guides, err := ParseGraph(graphDir)
		if err != nil {
			t.Fatalf("ParseGraph() unexpected error: %v", err)
		}
		if len(guides) != 2 {
			t.Errorf("expected 2 guides parsed from subdirectories, got %d", len(guides))
		}
		if _, ok := guides["Guide1"]; !ok {
			t.Errorf("expected Guide1 in guides map")
		}
		if _, ok := guides["Guide2"]; !ok {
			t.Errorf("expected Guide2 in guides map")
		}
		if _, ok := guides["Ignored"]; ok {
			t.Errorf("expected public/Ignored.md to be excluded")
		}
	})

	t.Run("duplicate guide id error", func(t *testing.T) {
		graphDir := t.TempDir()
		_ = os.MkdirAll(filepath.Join(graphDir, "folderA"), 0755)
		_ = os.MkdirAll(filepath.Join(graphDir, "folderB"), 0755)

		g := "---\nscope: definition\nprerequisites: []\nsub_guides: []\ntags: []\n---\n# Dup\n"
		_ = os.WriteFile(filepath.Join(graphDir, "folderA", "SameID.md"), []byte(g), 0644)
		_ = os.WriteFile(filepath.Join(graphDir, "folderB", "SameID.md"), []byte(g), 0644)

		_, err := ParseGraph(graphDir)
		if err == nil {
			t.Fatal("expected error for duplicate guide IDs across folders, got nil")
		}
		if !strings.Contains(err.Error(), "duplicate guide ID 'SameID'") {
			t.Errorf("expected duplicate guide ID error message, got: %v", err)
		}
	})

	t.Run("parse graph current directory dot", func(t *testing.T) {
		graphDir := t.TempDir()
		g := "---\nscope: definition\nprerequisites: []\nsub_guides: []\ntags: []\n---\n# Current Dir\n"
		_ = os.WriteFile(filepath.Join(graphDir, "CurrentGuide.md"), []byte(g), 0644)

		t.Chdir(graphDir)
		guides, err := ParseGraph(".")
		if err != nil {
			t.Fatalf("ParseGraph('.') unexpected error: %v", err)
		}
		if len(guides) != 1 {
			t.Fatalf("expected 1 guide when parsing '.', got %d", len(guides))
		}
		if _, ok := guides["CurrentGuide"]; !ok {
			t.Errorf("expected 'CurrentGuide' in guides map")
		}
	})
}

func TestParseGraphLenient(t *testing.T) {
	graphDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(graphDir, "dirA"), 0755)
	_ = os.MkdirAll(filepath.Join(graphDir, "dirB"), 0755)

	valid := "---\nscope: definition\nprerequisites: []\nsub_guides: []\ntags: []\n---\n# Valid\n"
	badYAML := "---\nscope: [unterminated\n---\n# Bad YAML\n"
	noFM := "# Plain draft without frontmatter\n"

	_ = os.WriteFile(filepath.Join(graphDir, "ValidGuide.md"), []byte(valid), 0644)
	_ = os.WriteFile(filepath.Join(graphDir, "BadYAML.md"), []byte(badYAML), 0644)
	_ = os.WriteFile(filepath.Join(graphDir, "DraftNoFM.md"), []byte(noFM), 0644)
	_ = os.WriteFile(filepath.Join(graphDir, "dirA", "DupGuide.md"), []byte(valid), 0644)
	_ = os.WriteFile(filepath.Join(graphDir, "dirB", "DupGuide.md"), []byte(valid), 0644)

	res, err := ParseGraphLenient(graphDir)
	if err != nil {
		t.Fatalf("ParseGraphLenient() unexpected fatal error: %v", err)
	}

	if len(res.Guides) != 1 {
		t.Fatalf("expected 1 valid non-duplicate guide, got %d: %v", len(res.Guides), res.Guides)
	}
	if _, ok := res.Guides["ValidGuide"]; !ok {
		t.Errorf("expected 'ValidGuide' in res.Guides")
	}
	if _, ok := res.Guides["DupGuide"]; ok {
		t.Errorf("expected duplicated 'DupGuide' to be excluded from res.Guides")
	}

	if len(res.Errors) != 3 {
		t.Fatalf("expected 3 collected parse errors, got %d: %v", len(res.Errors), res.Errors)
	}

	if res.ErrorForID("ValidGuide") != nil {
		t.Errorf("expected nil ErrorForID for ValidGuide")
	}
	if pe := res.ErrorForID("BadYAML"); pe == nil || !strings.Contains(pe.Error(), "failed to parse yaml") {
		t.Errorf("expected YAML parse error for BadYAML, got: %v", pe)
	}
	if pe := res.ErrorForID("DraftNoFM"); pe == nil || !strings.Contains(pe.Error(), "invalid or missing YAML frontmatter") {
		t.Errorf("expected frontmatter error for DraftNoFM, got: %v", pe)
	}
	if pe := res.ErrorForID("DupGuide"); pe == nil || !strings.Contains(pe.Error(), "duplicate guide ID 'DupGuide'") {
		t.Errorf("expected duplicate ID error for DupGuide, got: %v", pe)
	}

	// Strict ParseGraph should fail with all 3 errors joined
	_, strictErr := ParseGraph(graphDir)
	if strictErr == nil {
		t.Fatal("expected strict ParseGraph to fail")
	}
	for _, want := range []string{"BadYAML.md", "DraftNoFM.md", "duplicate guide ID 'DupGuide'"} {
		if !strings.Contains(strictErr.Error(), want) {
			t.Errorf("expected strict ParseGraph error to contain %q, got:\n%v", want, strictErr)
		}
	}
}
