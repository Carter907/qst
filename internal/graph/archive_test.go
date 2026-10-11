package graph

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchiveAndUnarchiveRoundtrip(t *testing.T) {
	srcDir := t.TempDir()

	// 1. Setup source knowledge graph
	manifestContent := `title: Archive Test Graph
description: Testing archive and unarchive
scopes:
  - definition
  - description
adherences:
  - strict
  - introductory
tours:
  - name: Tour 1
    guides:
      - GuideA
`
	if err := os.WriteFile(filepath.Join(srcDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	guideAContent := `---
scope: definition
prerequisites: []
sub_guides: []
tags: [core]
---
# Guide A
`
	if err := os.WriteFile(filepath.Join(srcDir, "GuideA.md"), []byte(guideAContent), 0644); err != nil {
		t.Fatalf("failed to write GuideA: %v", err)
	}

	// Nested guide
	if err := os.MkdirAll(filepath.Join(srcDir, "sub"), 0755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}
	guideBContent := `---
scope: description
prerequisites: []
sub_guides: []
tags: [nested]
---
# Guide B
`
	if err := os.WriteFile(filepath.Join(srcDir, "sub", "GuideB.md"), []byte(guideBContent), 0644); err != nil {
		t.Fatalf("failed to write GuideB: %v", err)
	}

	// Supporting diagram
	diagramContent := "graph TD\n    a --> b\n"
	if err := os.WriteFile(filepath.Join(srcDir, "diagram.mmd"), []byte(diagramContent), 0644); err != nil {
		t.Fatalf("failed to write diagram: %v", err)
	}

	// Files that MUST be ignored
	if err := os.MkdirAll(filepath.Join(srcDir, "public"), 0755); err != nil {
		t.Fatalf("failed to create public dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "public", "index.html"), []byte("<h1>Built Site</h1>"), 0644); err != nil {
		t.Fatalf("failed to write public index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, ".hidden"), []byte("secret"), 0644); err != nil {
		t.Fatalf("failed to write hidden file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "large.pdf"), []byte("%PDF-1.4"), 0644); err != nil {
		t.Fatalf("failed to write pdf: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "existing.kng"), []byte("dummy zip"), 0644); err != nil {
		t.Fatalf("failed to write dummy kng: %v", err)
	}

	// 2. Archive to .kng
	archivePath := filepath.Join(t.TempDir(), "test_graph.kng")
	if err := ArchiveGraph(srcDir, archivePath); err != nil {
		t.Fatalf("ArchiveGraph failed: %v", err)
	}

	// Inspect zip contents
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("failed to open created zip: %v", err)
	}
	var zippedFiles []string
	for _, f := range r.File {
		zippedFiles = append(zippedFiles, f.Name)
	}
	_ = r.Close()

	expectedIncluded := []string{"manifest.yaml", "GuideA.md", "sub/GuideB.md", "diagram.mmd"}
	for _, exp := range expectedIncluded {
		found := false
		for _, zf := range zippedFiles {
			if zf == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s in archive, but got: %v", exp, zippedFiles)
		}
	}

	excludedFiles := []string{"public/index.html", ".hidden", "large.pdf", "existing.kng"}
	for _, excl := range excludedFiles {
		for _, zf := range zippedFiles {
			if strings.Contains(zf, excl) {
				t.Errorf("file %s should be excluded from archive, but found %s", excl, zf)
			}
		}
	}

	// 3. Unarchive to new destination
	destDir := filepath.Join(t.TempDir(), "unpacked")
	if err := UnarchiveGraph(archivePath, destDir); err != nil {
		t.Fatalf("UnarchiveGraph failed: %v", err)
	}

	// Verify unarchived manifest exists and parses
	cfg, err := ParseConfig(destDir)
	if err != nil {
		t.Fatalf("unarchived graph failed ParseConfig: %v", err)
	}
	if cfg.Title != "Archive Test Graph" {
		t.Errorf("expected Title 'Archive Test Graph', got %q", cfg.Title)
	}

	// Verify GuideA exists
	if _, err := os.Stat(filepath.Join(destDir, "GuideA.md")); err != nil {
		t.Errorf("unarchived GuideA.md missing: %v", err)
	}

	// Verify nested sub/GuideB.md exists
	if _, err := os.Stat(filepath.Join(destDir, "sub", "GuideB.md")); err != nil {
		t.Errorf("unarchived sub/GuideB.md missing: %v", err)
	}

	// Verify diagram.mmd exists
	if _, err := os.Stat(filepath.Join(destDir, "diagram.mmd")); err != nil {
		t.Errorf("unarchived diagram.mmd missing: %v", err)
	}
}

func TestUnarchiveZipSlipProtection(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "malicious.kng")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	zw := zip.NewWriter(f)

	// Add file with directory traversal path
	w, err := zw.Create("../outside.txt")
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}
	_, _ = w.Write([]byte("malicious content"))
	_ = zw.Close()
	_ = f.Close()

	destDir := filepath.Join(t.TempDir(), "safe_dest")
	err = UnarchiveGraph(zipPath, destDir)
	if err == nil {
		t.Fatalf("expected error from zip slip attempt, got nil")
	}
	if !strings.Contains(err.Error(), "illegal file path") {
		t.Errorf("expected 'illegal file path' error, got: %v", err)
	}
}
