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
}
