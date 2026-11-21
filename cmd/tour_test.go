package cmd

import (
	"strings"
	"testing"

	"github.com/Carter907/qst/internal/graph"
)

func TestTourCommands(t *testing.T) {
	dir := setupTestGraph(t)

	// 1. Tour list empty
	outList, err := executeCommand("tour", "list", "-d", dir)
	if err != nil {
		t.Fatalf("tour list failed: %v", err)
	}
	if !strings.Contains(outList, "No tours defined") {
		t.Errorf("expected 'No tours defined', got: %s", outList)
	}

	// 2. Tour create
	outCreate, err := executeCommand("tour", "create", "Track 1", "-d", dir)
	if err != nil {
		t.Fatalf("tour create failed: %v, out: %s", err, outCreate)
	}

	// Duplicate create should fail
	_, exitCode, _ := executeCommandWithExit("tour", "create", "Track 1", "-d", dir)
	if exitCode != 1 {
		t.Errorf("expected duplicate tour create to fail with code 1")
	}

	// 3. Tour add unknown guide should fail
	_, exitCode, _ = executeCommandWithExit("tour", "add", "Track 1", "NonExistent", "-d", dir)
	if exitCode != 1 {
		t.Errorf("expected tour add unknown guide to fail with code 1")
	}

	// Tour add valid guides
	outAdd, err := executeCommand("tour", "add", "Track 1", "Def 1", "Def 2", "-d", dir)
	if err != nil {
		t.Fatalf("tour add failed: %v, out: %s", err, outAdd)
	}

	// Tour list populated
	outListPop, err := executeCommand("tour", "list", "-d", dir)
	if err != nil {
		t.Fatalf("tour list failed: %v", err)
	}
	if !strings.Contains(outListPop, "Track 1 (2 guides)") || !strings.Contains(outListPop, "Def 1") {
		t.Errorf("unexpected populated tour list:\n%s", outListPop)
	}

	// 4. Tour remove
	outRem, err := executeCommand("tour", "remove", "Track 1", "Def 1", "-d", dir)
	if err != nil {
		t.Fatalf("tour remove failed: %v, out: %s", err, outRem)
	}
	cfg, _ := graph.ParseConfig(dir)
	if len(cfg.Tours[0].Guides) != 1 || cfg.Tours[0].Guides[0] != "Def 2" {
		t.Errorf("expected only Def 2 in tour after remove, got: %v", cfg.Tours[0].Guides)
	}

	// 5. Tour delete
	outDel, err := executeCommand("tour", "delete", "Track 1", "-d", dir)
	if err != nil {
		t.Fatalf("tour delete failed: %v, out: %s", err, outDel)
	}
	cfgAfterDel, _ := graph.ParseConfig(dir)
	if len(cfgAfterDel.Tours) != 0 {
		t.Errorf("expected 0 tours after delete, got: %d", len(cfgAfterDel.Tours))
	}
}
