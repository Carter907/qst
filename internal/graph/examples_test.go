package graph

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExamplesDirectoryValid ensures all example knowledge graphs in the repository
// are continuously well-formed, parse cleanly, and satisfy all graph validation rules.
func TestExamplesDirectoryValid(t *testing.T) {
	examplesDir := filepath.Join("..", "..", "examples")
	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("failed to read examples directory: %v", err)
	}

	foundExamples := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		exampleName := entry.Name()
		dirPath := filepath.Join(examplesDir, exampleName)

		t.Run(exampleName, func(t *testing.T) {
			foundExamples++

			// 1. Manifest must exist and parse
			manifest, err := ParseConfig(dirPath)
			if err != nil {
				t.Fatalf("failed to parse manifest in %s: %v", dirPath, err)
			}
			if manifest.Title == "" {
				t.Errorf("manifest in %s is missing title", dirPath)
			}
			if len(manifest.Scopes) == 0 {
				t.Errorf("manifest in %s has no scopes defined", dirPath)
			}
			if len(manifest.Adherences) == 0 {
				t.Errorf("manifest in %s has no adherences defined", dirPath)
			}

			// 2. Guides must parse and not be empty
			guides, err := ParseGraph(dirPath)
			if err != nil {
				t.Fatalf("failed to parse guides in %s: %v", dirPath, err)
			}
			if len(guides) == 0 {
				t.Fatalf("example %s contains 0 guides (empty knowledge graph)", exampleName)
			}

			// 3. Graph must validate against manifest constraints
			if err := ValidateGraph(guides, manifest); err != nil {
				t.Fatalf("validation failed for example %s: %v", exampleName, err)
			}

			// 4. Graph must be acyclic
			if err := CheckAcyclic(guides); err != nil {
				t.Fatalf("acyclic check failed for example %s: %v", exampleName, err)
			}
		})
	}

	if foundExamples == 0 {
		t.Fatalf("no example directories found in %s", examplesDir)
	}
}
