package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Carter907/qst/internal/graph"
)

func TestDeleteCommand(t *testing.T) {
	dir := setupTestGraph(t)

	// Setup: link Def 1 as subguide to Desc 1, add Def 1 to a tour
	_, err := executeCommand("link", "Desc 1", "-d", dir, "--subguide", "Def 1", "--adherence", "introductory", "--segment", "3-3")
	if err != nil {
		t.Fatalf("link subguide failed: %v", err)
	}
	_, err = executeCommand("link", "Def 2", "-d", dir, "--prereq", "Def 1")
	if err != nil {
		t.Fatalf("link prereq failed: %v", err)
	}

	// Create and add to tour
	_, err = executeCommand("tour", "create", "Sample Tour", "-d", dir)
	if err != nil {
		t.Fatalf("tour create failed: %v", err)
	}
	_, err = executeCommand("tour", "add", "Sample Tour", "Def 1", "Def 2", "-d", dir)
	if err != nil {
		t.Fatalf("tour add failed: %v", err)
	}

	// 1. Dry run preview
	outDry, exitCode, _ := executeCommandWithExit("delete", "Def 1", "-d", dir, "--dry-run")
	if exitCode != 0 {
		t.Fatalf("delete dry-run failed with code %d: %s", exitCode, outDry)
	}
	if !strings.Contains(outDry, "Dry run: Would delete 'Def 1'") || !strings.Contains(outDry, "References to prune (3 total)") {
		t.Errorf("unexpected dry-run output:\n%s", outDry)
	}
	if _, err := os.Stat(filepath.Join(dir, "Def 1.md")); os.IsNotExist(err) {
		t.Fatalf("dry run deleted the file unexpectedly!")
	}

	// 2. Perform actual deletion
	outDel, exitCode, err := executeCommandWithExit("delete", "Def 1", "-d", dir)
	if err != nil || exitCode != 0 {
		t.Fatalf("delete failed: %v (code %d, out: %s)", err, exitCode, outDel)
	}
	if !strings.Contains(outDel, "Successfully deleted guide 'Def 1' and pruned 3 references.") {
		t.Errorf("unexpected delete output:\n%s", outDel)
	}

	// Verify file is removed
	if _, err := os.Stat(filepath.Join(dir, "Def 1.md")); !os.IsNotExist(err) {
		t.Errorf("expected Def 1.md to be deleted, but it still exists")
	}

	// Verify references pruned in Def 2
	def2, _ := graph.ParseGuide(filepath.Join(dir, "Def 2.md"))
	if len(def2.Metadata.Prerequisites) != 0 {
		t.Errorf("expected Def 1 pruned from Def 2 prereqs, got: %v", def2.Metadata.Prerequisites)
	}

	// Verify references pruned in Desc 1
	desc1, _ := graph.ParseGuide(filepath.Join(dir, "Desc 1.md"))
	if len(desc1.Metadata.SubGuides) != 0 {
		t.Errorf("expected Def 1 pruned from Desc 1 subguides, got: %v", desc1.Metadata.SubGuides)
	}

	// Verify references pruned in manifest tour
	cfg, _ := graph.ParseConfig(dir)
	if len(cfg.Tours[0].Guides) != 1 || cfg.Tours[0].Guides[0] != "Def 2" {
		t.Errorf("expected Def 1 pruned from tour, got: %v", cfg.Tours[0].Guides)
	}

	// Verify graph still validates
	outVal, exitCode, err := executeCommandWithExit("validate", dir)
	if err != nil || exitCode != 0 {
		t.Fatalf("graph failed validation after delete: %v (code %d, out: %s)", err, exitCode, outVal)
	}
}
