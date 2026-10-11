package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/pflag"
)

func setupTestGraph(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	manifest := `title: Test Graph
scopes:
  - definition
  - description
adherences:
  - strict
  - introductory
relaxed_subguides: false
strict_coverage: false
tours: []
`
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	def1 := `---
scope: definition
prerequisites: []
sub_guides: []
tags: [basic]
---

## Def 1
Definition content here.
`
	if err := os.WriteFile(filepath.Join(dir, "Def 1.md"), []byte(def1), 0644); err != nil {
		t.Fatalf("failed to write def1: %v", err)
	}

	def2 := `---
scope: definition
prerequisites: []
sub_guides: []
tags: [basic]
---

## Def 2
Second definition.
`
	if err := os.WriteFile(filepath.Join(dir, "Def 2.md"), []byte(def2), 0644); err != nil {
		t.Fatalf("failed to write def2: %v", err)
	}

	desc1 := `---
scope: description
prerequisites: []
sub_guides: []
tags: [concept]
---

## Desc 1

Line 3 has Def 1 reference.
`
	if err := os.WriteFile(filepath.Join(dir, "Desc 1.md"), []byte(desc1), 0644); err != nil {
		t.Fatalf("failed to write desc1: %v", err)
	}

	return dir
}

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	// Reset all command flags so state does not leak between test invocations
	for _, c := range rootCmd.Commands() {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			f.Changed = false
			_ = f.Value.Set(f.DefValue)
		})
	}
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		_ = f.Value.Set(f.DefValue)
	})

	err := rootCmd.Execute()
	return buf.String(), err
}

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

func TestLinkAndUnlinkPrerequisite(t *testing.T) {
	dir := setupTestGraph(t)

	// 1. Link prerequisite Def 2 -> Def 1
	out, err := executeCommand("link", "Def 1", "--dir", dir, "--prereq", "Def 2")
	if err != nil {
		t.Fatalf("link prerequisite error: %v, out: %s", err, out)
	}

	// Verify file content has prerequisite
	content, err := os.ReadFile(filepath.Join(dir, "Def 1.md"))
	if err != nil {
		t.Fatalf("failed to read Def 1: %v", err)
	}
	if !strings.Contains(string(content), "Def 2") {
		t.Fatalf("expected Def 2 in prerequisites, got:\n%s", string(content))
	}

	// 2. Unlink prerequisite
	out, err = executeCommand("unlink", "Def 1", "--dir", dir, "--prereq", "Def 2")
	if err != nil {
		t.Fatalf("unlink error: %v, out: %s", err, out)
	}

	content, _ = os.ReadFile(filepath.Join(dir, "Def 1.md"))
	if strings.Contains(string(content), "Def 2") {
		t.Fatalf("expected Def 2 removed from prerequisites, got:\n%s", string(content))
	}
}

func TestLinkAndUnlinkSubguide(t *testing.T) {
	dir := setupTestGraph(t)

	// 1. Link subguide Def 1 to Desc 1
	out, err := executeCommand("link", "Desc 1", "--dir", dir, "--guide", "Def 1", "--adherence", "introductory", "--segment", "3-3")
	if err != nil {
		t.Fatalf("link subguide error: %v, out: %s", err, out)
	}

	content, err := os.ReadFile(filepath.Join(dir, "Desc 1.md"))
	if err != nil {
		t.Fatalf("failed to read Desc 1: %v", err)
	}
	if !strings.Contains(string(content), "Def 1") || !strings.Contains(string(content), "3-3") {
		t.Fatalf("expected subguide Def 1 with segment 3-3, got:\n%s", string(content))
	}

	// 2. Update subguide in place with new segment
	out, err = executeCommand("link", "Desc 1", "--dir", dir, "--guide", "Def 1", "--adherence", "strict", "--segment", "2-3")
	if err != nil {
		t.Fatalf("link update error: %v, out: %s", err, out)
	}

	guide, err := graph.ParseGuide(filepath.Join(dir, "Desc 1.md"))
	if err != nil {
		t.Fatalf("failed to parse Desc 1: %v", err)
	}
	if len(guide.Metadata.SubGuides) != 1 {
		t.Fatalf("expected exactly 1 subguide after in-place update, got: %d", len(guide.Metadata.SubGuides))
	}
	if guide.Metadata.SubGuides[0].Segment != "2-3" || guide.Metadata.SubGuides[0].Adherence != "strict" {
		t.Fatalf("expected updated segment 2-3 and adherence strict, got: %v", guide.Metadata.SubGuides[0])
	}

	// 3. Unlink subguide
	out, err = executeCommand("unlink", "Desc 1", "--dir", dir, "--guide", "Def 1")
	if err != nil {
		t.Fatalf("unlink subguide error: %v, out: %s", err, out)
	}

	guide, _ = graph.ParseGuide(filepath.Join(dir, "Desc 1.md"))
	if len(guide.Metadata.SubGuides) != 0 {
		t.Fatalf("expected Def 1 removed from subguides, got: %v", guide.Metadata.SubGuides)
	}
}

func executeCommandWithExit(args ...string) (string, int, error) {
	var exitCode int
	oldExit := exitFunc
	exitFunc = func(code int) {
		exitCode = code
	}
	defer func() { exitFunc = oldExit }()

	out, err := executeCommand(args...)
	return out, exitCode, err
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


