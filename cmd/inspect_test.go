package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectCommand(t *testing.T) {
	dir := setupTestGraph(t)

	// Default inspect (compact coverage row)
	out, err := executeCommand("inspect", "Desc 1", "--dir", dir)
	if err != nil {
		t.Fatalf("inspect error: %v, out: %s", err, out)
	}
	if !strings.Contains(out, "Guide: Desc 1") || !strings.Contains(out, "Scope: description") {
		t.Errorf("unexpected inspect output: %s", out)
	}
	if !strings.Contains(out, "Coverage:") || !strings.Contains(out, "use -c for line breakdown") {
		t.Errorf("expected compact coverage hint in default inspect, got: %s", out)
	}

	// Inspect with --coverage / -c
	outWithCov, err := executeCommand("inspect", "Desc 1", "--dir", dir, "-c")
	if err != nil {
		t.Fatalf("inspect -c error: %v, out: %s", err, outWithCov)
	}
	if !strings.Contains(outWithCov, "Line-by-Line Body Coverage: Desc 1") || !strings.Contains(outWithCov, "[UNMAPPED]") {
		t.Errorf("expected expanded line-by-line coverage in inspect -c, got: %s", outWithCov)
	}
}

func TestLeafInspectCoverage(t *testing.T) {
	tmp := t.TempDir()
	manifest := `title: Strict Graph
scopes:
  - definition
  - description
adherences:
  - strict
strict_coverage: true
tours: []
`
	if err := os.WriteFile(filepath.Join(tmp, "manifest.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}
	def := `---
scope: definition
prerequisites: []
sub_guides: []
tags: []
---

## Leaf Guide

Just some body text.
`
	if err := os.WriteFile(filepath.Join(tmp, "Leaf Guide.md"), []byte(def), 0644); err != nil {
		t.Fatalf("failed to write leaf guide: %v", err)
	}

	out, err := executeCommand("inspect", "Leaf Guide", "-d", tmp, "-c")
	if err != nil {
		t.Fatalf("inspect failed: %v, out: %s", err, out)
	}
	if strings.Contains(out, "[strict: FAILED]") {
		t.Errorf("expected leaf guide not to fail strict coverage, but got:\n%s", out)
	}
	if !strings.Contains(out, "[strict: N/A (leaf)]") && !strings.Contains(out, "Status: N/A (leaf guide - no subguides defined)") {
		t.Errorf("expected leaf guide N/A status in inspect output, got:\n%s", out)
	}
}

func TestInspectCommand_Failures(t *testing.T) {
	dir := setupTestGraph(t)

	// Failure case 1: Non-existent guide
	res := executeCommandContext(context.Background(), "inspect", "NonExistentGuide", "--dir", dir)
	if res.ExitCode != 1 {
		t.Errorf("expected exitCode 1 for nonexistent guide inspect, got %d", res.ExitCode)
	}
	if res.Stdout != "" {
		t.Errorf("expected empty stdout on failure, got: %s", res.Stdout)
	}
	if !strings.Contains(res.Stderr, "Error: guide 'NonExistentGuide' not found") {
		t.Errorf("expected not found error in stderr, got: %s", res.Stderr)
	}

	// Failure case 2: Guide exists but manifest is missing
	badDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(badDir, "TestGuide.md"), []byte("---\nscope: foo\n---\nBody\n"), 0644); err != nil {
		t.Fatalf("failed to write test guide: %v", err)
	}
	res = executeCommandContext(context.Background(), "inspect", "TestGuide", "--dir", badDir)
	if res.ExitCode != 1 {
		t.Errorf("expected exitCode 1 for inspect with missing manifest, got %d", res.ExitCode)
	}
	if !strings.Contains(res.Stderr, "Failed to load manifest.yaml:") {
		t.Errorf("expected 'Failed to load manifest.yaml:' in stderr, got: %s", res.Stderr)
	}
}
