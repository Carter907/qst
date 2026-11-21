package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Carter907/qst/internal/graph"
)

func TestCoverageCommand(t *testing.T) {
	dir := setupTestGraph(t)
	out, err := executeCommand("coverage", "Desc 1", "--dir", dir)
	if err != nil {
		t.Fatalf("coverage error: %v, out: %s", err, out)
	}
	// Standalone coverage displays the unified view (metadata header + line-by-line coverage)
	if !strings.Contains(out, "Guide: Desc 1") || !strings.Contains(out, "Line-by-Line Body Coverage: Desc 1") || !strings.Contains(out, "[UNMAPPED]") {
		t.Errorf("unexpected coverage output: %s", out)
	}
}

func TestValidateCommand(t *testing.T) {
	dir := setupTestGraph(t)

	// Valid graph
	out, exitCode, err := executeCommandWithExit("validate", dir)
	if err != nil || exitCode != 0 {
		t.Fatalf("validate expected success (0), got code: %d, err: %v, out: %s", exitCode, err, out)
	}
	if !strings.Contains(out, "valid with 3 guides") {
		t.Errorf("expected success message in validate output, got: %s", out)
	}

	// Invalid graph: missing manifest
	emptyDir := t.TempDir()
	out, exitCode, _ = executeCommandWithExit("validate", emptyDir)
	if exitCode != 1 {
		t.Errorf("validate expected failure code 1 for missing manifest, got %d (out: %s)", exitCode, out)
	}
	if !strings.Contains(out, "Error loading config") {
		t.Errorf("expected 'Error loading config' in error output, got: %s", out)
	}
}

func TestListCommand(t *testing.T) {
	dir := setupTestGraph(t)

	// Default: alphabetically sorted IDs
	out, err := executeCommand("list", "-d", dir)
	if err != nil {
		t.Fatalf("list error: %v, out: %s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	expected := []string{"Def 1", "Def 2", "Desc 1"}
	if len(lines) != len(expected) {
		t.Fatalf("expected %d guides in list, got %d: %v", len(expected), len(lines), lines)
	}
	for i, exp := range expected {
		if lines[i] != exp {
			t.Errorf("expected line %d to be %q, got %q", i, exp, lines[i])
		}
	}

	// Scope filter
	outScope, err := executeCommand("list", "-d", dir, "--scope", "definition")
	if err != nil {
		t.Fatalf("list --scope error: %v, out: %s", err, outScope)
	}
	if strings.Contains(outScope, "Desc 1") || !strings.Contains(outScope, "Def 1") || !strings.Contains(outScope, "Def 2") {
		t.Errorf("unexpected filtered list output: %s", outScope)
	}

	// Long list (-l): aligned table view
	outLong, err := executeCommand("list", "-d", dir, "-l")
	if err != nil {
		t.Fatalf("list -l error: %v, out: %s", err, outLong)
	}
	if !strings.Contains(outLong, "GUIDE") || !strings.Contains(outLong, "SCOPE") || !strings.Contains(outLong, "PREREQS") {
		t.Errorf("expected table header in list -l, got: %s", outLong)
	}
	if !strings.Contains(outLong, "Def 1") || !strings.Contains(outLong, "definition") {
		t.Errorf("expected Def 1 row in list -l, got: %s", outLong)
	}
}

func TestStatsCommand(t *testing.T) {
	dir := setupTestGraph(t)

	// Valid stats
	out, exitCode, err := executeCommandWithExit("stats", dir)
	if err != nil || exitCode != 0 {
		t.Fatalf("stats unexpected exit: %d, err: %v, out: %s", exitCode, err, out)
	}

	sections := []string{
		"Graph Statistics: Test Graph",
		"Overview:",
		"Total Guides:          3",
		"Scope Distribution:",
		"definition:",
		"description:",
		"Graph Depth & Learning Paths:",
		"Prerequisite Depth:",
		"Tour Coverage:",
		"Total Tours:",
		"Content & Subguide Coverage:",
		"Total Body Lines:",
	}
	for _, s := range sections {
		if !strings.Contains(out, s) {
			t.Errorf("expected stats output to contain %q, but got:\n%s", s, out)
		}
	}

	// Missing manifest should exit with code 1
	emptyDir := t.TempDir()
	out, exitCode, _ = executeCommandWithExit("stats", emptyDir)
	if exitCode != 1 {
		t.Errorf("stats expected exit code 1 for missing manifest, got %d (out: %s)", exitCode, out)
	}
	if !strings.Contains(out, "Error loading config") {
		t.Errorf("expected 'Error loading config' in stats output, got: %s", out)
	}
}

func TestFormAndFreeCommands(t *testing.T) {
	srcDir := setupTestGraph(t)
	kngPath := filepath.Join(t.TempDir(), "packaged.kng")

	// 1. qst form with -o
	outForm, exitCode, err := executeCommandWithExit("form", srcDir, "-o", kngPath)
	if err != nil || exitCode != 0 {
		t.Fatalf("form failed (code: %d, err: %v, out: %s)", exitCode, err, outForm)
	}
	if !strings.Contains(outForm, "Successfully packaged 3 guides") {
		t.Errorf("expected 'Successfully packaged 3 guides', got: %s", outForm)
	}
	if _, err := os.Stat(kngPath); err != nil {
		t.Fatalf(".kng archive was not created at %s: %v", kngPath, err)
	}

	// 2. qst free with custom destination
	unpackedDir := filepath.Join(t.TempDir(), "unpacked")
	outFree, exitCode, err := executeCommandWithExit("free", kngPath, unpackedDir)
	if err != nil || exitCode != 0 {
		t.Fatalf("free failed (code: %d, err: %v, out: %s)", exitCode, err, outFree)
	}
	if !strings.Contains(outFree, "Successfully unarchived to") {
		t.Errorf("expected 'Successfully unarchived to', got: %s", outFree)
	}

	// Verify unarchived manifest and guides
	cfg, err := graph.ParseConfig(unpackedDir)
	if err != nil {
		t.Fatalf("unarchived graph failed ParseConfig: %v", err)
	}
	if cfg.Title != "Test Graph" {
		t.Errorf("expected Title 'Test Graph', got %q", cfg.Title)
	}

	guides, err := graph.ParseGraph(unpackedDir)
	if err != nil {
		t.Fatalf("unarchived graph failed ParseGraph: %v", err)
	}
	if len(guides) != 3 {
		t.Errorf("expected 3 guides in unarchived graph, got %d", len(guides))
	}

	// 3. qst free with invalid extension
	outBad, exitCode, _ := executeCommandWithExit("free", "invalid.zip")
	if exitCode != 1 {
		t.Errorf("free expected exit code 1 for invalid extension, got %d", exitCode)
	}
	if !strings.Contains(outBad, "Missing required file extension (.kng)") {
		t.Errorf("expected error message for missing extension, got: %s", outBad)
	}
}

func TestTagCommands(t *testing.T) {
	dir := setupTestGraph(t)

	// 1. Tag list
	outList, err := executeCommand("tag", "list", "-d", dir)
	if err != nil {
		t.Fatalf("tag list failed: %v", err)
	}
	if !strings.Contains(outList, "basic") || !strings.Contains(outList, "concept") {
		t.Errorf("expected basic and concept tags in list, got: %s", outList)
	}

	// 2. Tag add
	outAdd, err := executeCommand("tag", "add", "Def 1", "fresh, tested", "-d", dir)
	if err != nil {
		t.Fatalf("tag add failed: %v, out: %s", err, outAdd)
	}
	def1, _ := graph.ParseGuide(filepath.Join(dir, "Def 1.md"))
	if len(def1.Metadata.Tags) != 3 || def1.Metadata.Tags[1] != "fresh" || def1.Metadata.Tags[2] != "tested" {
		t.Errorf("expected tags to include fresh and tested, got: %v", def1.Metadata.Tags)
	}

	// 3. Tag remove
	outRem, err := executeCommand("tag", "remove", "Def 1", "fresh", "-d", dir)
	if err != nil {
		t.Fatalf("tag remove failed: %v, out: %s", err, outRem)
	}
	def1After, _ := graph.ParseGuide(filepath.Join(dir, "Def 1.md"))
	if len(def1After.Metadata.Tags) != 2 || strings.Contains(strings.Join(def1After.Metadata.Tags, ","), "fresh") {
		t.Errorf("expected tag fresh to be removed, got: %v", def1After.Metadata.Tags)
	}
}

func TestNewCommand(t *testing.T) {
	tmp := t.TempDir()
	graphDir := filepath.Join(tmp, "my_graph")

	// 1. Default new should set strict_coverage: false
	out, exitCode, err := executeCommandWithExit("new", graphDir)
	if err != nil || exitCode != 0 {
		t.Fatalf("new failed: %v (code %d, out: %s)", err, exitCode, out)
	}
	cfg, err := graph.ParseConfig(graphDir)
	if err != nil {
		t.Fatalf("failed to parse manifest from new: %v", err)
	}
	if cfg.StrictCoverage {
		t.Errorf("expected strict_coverage: false by default in qst new")
	}

	// 2. Opt-in with --strict-coverage
	strictDir := filepath.Join(tmp, "strict_graph")
	outStrict, exitCode, err := executeCommandWithExit("new", strictDir, "--strict-coverage")
	if err != nil || exitCode != 0 {
		t.Fatalf("new --strict-coverage failed: %v (code %d, out: %s)", err, exitCode, outStrict)
	}
	cfgStrict, err := graph.ParseConfig(strictDir)
	if err != nil {
		t.Fatalf("failed to parse manifest from new --strict-coverage: %v", err)
	}
	if !cfgStrict.StrictCoverage {
		t.Errorf("expected strict_coverage: true with --strict-coverage flag")
	}
}

func TestRenameCommand(t *testing.T) {
	dir := setupTestGraph(t)

	// Non-existent guide rename should fail
	out, exitCode, _ := executeCommandWithExit("rename", "DoesNotExist", "NewName", "-d", dir)
	if exitCode != 1 || !strings.Contains(out, "does not exist") {
		t.Errorf("expected code 1 for renaming non-existent guide, got %d, out: %s", exitCode, out)
	}

	// Valid rename
	outRen, exitCode, err := executeCommandWithExit("rename", "Def 1", "Renamed Def 1", "-d", dir)
	if err != nil || exitCode != 0 {
		t.Fatalf("rename failed: %v (code %d, out: %s)", err, exitCode, outRen)
	}
	if !strings.Contains(outRen, "Successfully renamed 'Def 1' to 'Renamed Def 1'") {
		t.Errorf("unexpected rename output:\n%s", outRen)
	}
	if _, err := os.Stat(filepath.Join(dir, "Renamed Def 1.md")); os.IsNotExist(err) {
		t.Errorf("renamed file Renamed Def 1.md does not exist")
	}
}

func TestGuideFileAndPathSpecification(t *testing.T) {
	// Unit tests for cleanGuideName
	tests := []struct {
		input    string
		expected string
	}{
		{"Guide.md", "Guide"},
		{"Guide.MD", "Guide"},
		{"Guide", "Guide"},
		{"path/to/Guide.md", "Guide"},
		{"path/to/Guide", "Guide"},
		{"./Guide.md", "Guide"},
		{"", ""},
		{"   ", ""},
	}
	for _, tc := range tests {
		if actual := cleanGuideName(tc.input); actual != tc.expected {
			t.Errorf("cleanGuideName(%q) = %q; expected %q", tc.input, actual, tc.expected)
		}
	}

	dir := setupTestGraph(t)

	// 1. inspect with .md and with full path auto-inferring --dir
	out, exitCode, err := executeCommandWithExit("inspect", "Def 1.md", "--dir", dir)
	if err != nil || exitCode != 0 {
		t.Fatalf("inspect Def 1.md failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "Guide: Def 1") {
		t.Errorf("expected inspect output to contain 'Guide: Def 1', got: %s", out)
	}

	guidePath := filepath.Join(dir, "Def 1.md")
	out, exitCode, err = executeCommandWithExit("inspect", guidePath)
	if err != nil || exitCode != 0 {
		t.Fatalf("inspect with full path failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "Guide: Def 1") {
		t.Errorf("expected inspect output to contain 'Guide: Def 1', got: %s", out)
	}

	// 2. coverage with full path auto-inferring --dir
	out, exitCode, err = executeCommandWithExit("coverage", guidePath)
	if err != nil || exitCode != 0 {
		t.Fatalf("coverage with full path failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "Guide: Def 1") {
		t.Errorf("expected coverage output to contain 'Guide: Def 1', got: %s", out)
	}

	// 3. add guide with .md and path, auto-inferring --dir
	newGuidePath := filepath.Join(dir, "Def 3.md")
	out, exitCode, err = executeCommandWithExit("add", newGuidePath, "--scope", "definition", "--prerequisites", "Def 1.md")
	if err != nil || exitCode != 0 {
		t.Fatalf("add with path and .md prereq failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "Def 3.md")); os.IsNotExist(statErr) {
		t.Errorf("expected Def 3.md to be created in %s", dir)
	}

	// 4. link with path and .md flags
	out, exitCode, err = executeCommandWithExit("link", filepath.Join(dir, "Def 1.md"), "--prereq", filepath.Join(dir, "Def 2.md"))
	if err != nil || exitCode != 0 {
		t.Fatalf("link with path and .md prereq failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "Successfully linked prerequisite 'Def 2' to 'Def 1'") {
		t.Errorf("unexpected link output: %s", out)
	}

	// 5. unlink with path and .md flags
	out, exitCode, err = executeCommandWithExit("unlink", filepath.Join(dir, "Def 1.md"), "--prereq", "Def 2.md")
	if err != nil || exitCode != 0 {
		t.Fatalf("unlink with path and .md prereq failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "Successfully unlinked prerequisite 'Def 2' from 'Def 1'") {
		t.Errorf("unexpected unlink output: %s", out)
	}

	// 6. tag add & remove with path and .md
	out, exitCode, err = executeCommandWithExit("tag", "add", filepath.Join(dir, "Def 1.md"), "tagged")
	if err != nil || exitCode != 0 {
		t.Fatalf("tag add failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	out, exitCode, err = executeCommandWithExit("tag", "remove", filepath.Join(dir, "Def 1.md"), "tagged")
	if err != nil || exitCode != 0 {
		t.Fatalf("tag remove failed: code=%d err=%v out=%s", exitCode, err, out)
	}

	// 7. tour add & remove with path and .md
	out, exitCode, err = executeCommandWithExit("tour", "create", "Basics", "--dir", dir)
	if err != nil || exitCode != 0 {
		t.Fatalf("tour create failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	out, exitCode, err = executeCommandWithExit("tour", "add", "Basics", filepath.Join(dir, "Def 1.md"), filepath.Join(dir, "Def 2.md"))
	if err != nil || exitCode != 0 {
		t.Fatalf("tour add failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	out, exitCode, err = executeCommandWithExit("tour", "remove", "Basics", filepath.Join(dir, "Def 1.md"))
	if err != nil || exitCode != 0 {
		t.Fatalf("tour remove failed: code=%d err=%v out=%s", exitCode, err, out)
	}

	// 8. view with -g and path
	out, exitCode, err = executeCommandWithExit("view", "-g", filepath.Join(dir, "Def 1.md"))
	if err != nil || exitCode != 0 {
		t.Fatalf("view -g failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "Def 1") {
		t.Errorf("expected view output to contain 'Def 1', got: %s", out)
	}

	// 9. rename with old and new .md and path
	out, exitCode, err = executeCommandWithExit("rename", filepath.Join(dir, "Def 2.md"), "Def 2 Renamed.md")
	if err != nil || exitCode != 0 {
		t.Fatalf("rename failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "Def 2 Renamed.md")); os.IsNotExist(statErr) {
		t.Errorf("expected Def 2 Renamed.md to exist after rename")
	}

	// 10. delete with path and .md
	out, exitCode, err = executeCommandWithExit("delete", filepath.Join(dir, "Def 3.md"))
	if err != nil || exitCode != 0 {
		t.Fatalf("delete failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "Def 3.md")); !os.IsNotExist(statErr) {
		t.Errorf("expected Def 3.md to be deleted")
	}
}

func TestVersionFlag(t *testing.T) {
	defer SetVersionInfo("dev", "none", "unknown")

	SetVersionInfo("v1.2.3", "abc1234", "2026-11-13T00:00:00Z")
	out, exitCode, err := executeCommandWithExit("--version")
	if err != nil || exitCode != 0 {
		t.Fatalf("--version failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	want := "qst v1.2.3 (commit abc1234, built 2026-11-13T00:00:00Z)"
	if !strings.Contains(out, want) {
		t.Errorf("expected version output %q, got %q", want, out)
	}
}
