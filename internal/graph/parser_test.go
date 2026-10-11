package graph

import (
	"os"
	"path/filepath"
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
}
