package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestViewCommand(t *testing.T) {
	dir := setupTestGraph(t)

	// Success case
	out, exitCode, err := executeCommandWithExit("view", dir)
	if err != nil || exitCode != 0 {
		t.Fatalf("view command failed: exitCode=%d, err=%v, out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "```mermaid") || !strings.Contains(out, "graph TD") {
		t.Errorf("expected mermaid diagram in view output, got: %s", out)
	}
	if !strings.Contains(out, "def_1") || !strings.Contains(out, "def_2") {
		t.Errorf("expected guide nodes in view output, got: %s", out)
	}

	// Failure case: non-existent directory
	badDir := filepath.Join(t.TempDir(), "nonexistent")
	out, exitCode, _ = executeCommandWithExit("view", badDir)
	if exitCode != 1 {
		t.Errorf("expected exitCode 1 for nonexistent view dir, got %d", exitCode)
	}
	if !strings.Contains(out, "Error parsing graph:") {
		t.Errorf("expected 'Error parsing graph:' in output, got: %s", out)
	}
}

func TestViewCommand_LocalizedGuide(t *testing.T) {
	dir := setupTestGraph(t)

	// 1. Success case with --guide
	out, exitCode, err := executeCommandWithExit("view", dir, "--guide", "Def 1")
	if err != nil || exitCode != 0 {
		t.Fatalf("view --guide failed: exitCode=%d, err=%v, out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "```mermaid") || !strings.Contains(out, "graph LR") {
		t.Errorf("expected mermaid block in localized view output, got: %s", out)
	}
	if !strings.Contains(out, `def_1["Definition<br/>Def 1"]`) {
		t.Errorf("expected def_1 node with scope prefix, got: %s", out)
	}
	if !strings.Contains(out, "class def_1 currentGuide;") {
		t.Errorf("expected currentGuide styling for def_1, got: %s", out)
	}

	// 2. Failure case: non-existent guide
	outFail, exitCodeFail, _ := executeCommandWithExit("view", dir, "--guide", "NonExistent")
	if exitCodeFail != 1 {
		t.Errorf("expected exitCode 1 for non-existent guide view, got %d", exitCodeFail)
	}
	if !strings.Contains(outFail, "Error: guide 'NonExistent' not found in graph") {
		t.Errorf("expected not found error in output, got: %s", outFail)
	}
}
