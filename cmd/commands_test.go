package cmd

import (
	"bytes"
	"fmt"
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

func executeCommandWithExit(args ...string) (string, int, error) {
	var exitCode int
	oldExit := exitFunc
	exitFunc = func(code int) {
		exitCode = code
	}
	defer func() { exitFunc = oldExit }()

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
	return buf.String(), exitCode, err
}

func executeCommand(args ...string) (string, error) {
	out, exitCode, err := executeCommandWithExit(args...)
	if err == nil && exitCode != 0 {
		err = fmt.Errorf("command exited with code %d: %s", exitCode, out)
	}
	return out, err
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
	out, err := executeCommand("link", "Desc 1", "--dir", dir, "--subguide", "Def 1", "--adherence", "introductory", "--segment", "3-3")
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
	out, err = executeCommand("link", "Desc 1", "--dir", dir, "--subguide", "Def 1", "--adherence", "strict", "--segment", "2-3")
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
	out, err = executeCommand("unlink", "Desc 1", "--dir", dir, "--subguide", "Def 1")
	if err != nil {
		t.Fatalf("unlink subguide error: %v, out: %s", err, out)
	}

	guide, _ = graph.ParseGuide(filepath.Join(dir, "Desc 1.md"))
	if len(guide.Metadata.SubGuides) != 0 {
		t.Fatalf("expected Def 1 removed from subguides, got: %v", guide.Metadata.SubGuides)
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

func TestAddCommand(t *testing.T) {
	dir := setupTestGraph(t)

	// 1. Missing scope in non-interactive mode should fail
	out, exitCode, _ := executeCommandWithExit("add", "MissingScope", "-d", dir)
	if exitCode != 1 || !strings.Contains(out, "--scope is required") {
		t.Errorf("expected exit code 1 and required scope error, got %d, out: %s", exitCode, out)
	}

	// 2. Invalid scope should fail
	out, exitCode, _ = executeCommandWithExit("add", "BadScope", "-d", dir, "--scope", "invalid_scope")
	if exitCode != 1 || !strings.Contains(out, "invalid scope") {
		t.Errorf("expected exit code 1 for invalid scope, got %d, out: %s", exitCode, out)
	}

	// 3. Unknown prerequisite should fail
	out, exitCode, _ = executeCommandWithExit("add", "BadPrereq", "-d", dir, "--scope", "definition", "--prerequisites", "NonExistent")
	if exitCode != 1 || !strings.Contains(out, "does not exist in graph") {
		t.Errorf("expected exit code 1 for unknown prereq, got %d, out: %s", exitCode, out)
	}

	// 4. Mismatched prerequisite scope should fail
	out, exitCode, _ = executeCommandWithExit("add", "MismatchedPrereq", "-d", dir, "--scope", "definition", "--prerequisites", "Desc 1")
	if exitCode != 1 || !strings.Contains(out, "does not match guide scope") {
		t.Errorf("expected exit code 1 for scope mismatch, got %d, out: %s", exitCode, out)
	}

	// 5. Valid addition with prerequisite, tags, and custom body
	out, exitCode, err := executeCommandWithExit("add", "Def 3", "-d", dir, "--scope", "definition", "--prerequisites", "Def 1", "--tags", "test,added", "--body", "## Def 3\n\nCustom content here.")
	if err != nil || exitCode != 0 {
		t.Fatalf("failed to add guide: %v (code %d, out: %s)", err, exitCode, out)
	}

	guide, err := graph.ParseGuide(filepath.Join(dir, "Def 3.md"))
	if err != nil {
		t.Fatalf("failed to parse added guide: %v", err)
	}
	if guide.Metadata.Scope != "definition" {
		t.Errorf("expected scope definition, got %s", guide.Metadata.Scope)
	}
	if len(guide.Metadata.Prerequisites) != 1 || guide.Metadata.Prerequisites[0] != "Def 1" {
		t.Errorf("expected prerequisite Def 1, got %v", guide.Metadata.Prerequisites)
	}
	if len(guide.Metadata.Tags) != 2 || guide.Metadata.Tags[0] != "test" {
		t.Errorf("expected tags [test, added], got %v", guide.Metadata.Tags)
	}
	if !strings.Contains(guide.Body, "Custom content here.") {
		t.Errorf("expected custom body, got:\n%s", guide.Body)
	}

	// 6. Valid addition of higher-tier guide with subguides
	out, exitCode, err = executeCommandWithExit("add", "Desc 2", "-d", dir, "--scope", "description", "--subguide", "Def 1, Def 2", "--adherence", "strict")
	if err != nil || exitCode != 0 {
		t.Fatalf("failed to add guide with subguides: %v (code %d, out: %s)", err, exitCode, out)
	}
	desc2, err := graph.ParseGuide(filepath.Join(dir, "Desc 2.md"))
	if err != nil {
		t.Fatalf("failed to parse Desc 2: %v", err)
	}
	if len(desc2.Metadata.SubGuides) != 2 {
		t.Fatalf("expected 2 subguides in Desc 2, got %d", len(desc2.Metadata.SubGuides))
	}
}

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

func TestBatchLinkAndPreview(t *testing.T) {
	dir := setupTestGraph(t)

	// 1. Preview body lines
	outPrev, err := executeCommand("link", "Desc 1", "-d", dir, "--preview")
	if err != nil {
		t.Fatalf("preview failed: %v", err)
	}
	if !strings.Contains(outPrev, "Body Lines Preview: Desc 1") || !strings.Contains(outPrev, "[L  3]") {
		t.Errorf("unexpected preview output:\n%s", outPrev)
	}

	// 2. Batch link subguides using comma-separated flag
	outLink, err := executeCommand("link", "Desc 1", "-d", dir, "--subguide", "Def 1, Def 2", "--adherence", "introductory")
	if err != nil {
		t.Fatalf("batch link subguide failed: %v (out: %s)", err, outLink)
	}
	desc1, _ := graph.ParseGuide(filepath.Join(dir, "Desc 1.md"))
	if len(desc1.Metadata.SubGuides) != 2 {
		t.Fatalf("expected 2 subguides after batch link, got %d", len(desc1.Metadata.SubGuides))
	}
}

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

func TestBuildCommand(t *testing.T) {
	dir := setupTestGraph(t)
	outDir := filepath.Join(t.TempDir(), "public")

	// Success case
	out, exitCode, err := executeCommandWithExit("build", dir, "--out", outDir)
	if err != nil || exitCode != 0 {
		t.Fatalf("build command failed: exitCode=%d, err=%v, out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "Building static site") || !strings.Contains(out, "Build complete!") {
		t.Errorf("unexpected build output: %s", out)
	}
	if _, err := os.Stat(filepath.Join(outDir, "index.html")); err != nil {
		t.Errorf("expected index.html in outDir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "Def 1.html")); err != nil {
		t.Errorf("expected Def 1.html in outDir: %v", err)
	}

	// Failure case: non-existent directory
	badDir := filepath.Join(t.TempDir(), "nonexistent")
	out, exitCode, _ = executeCommandWithExit("build", badDir)
	if exitCode != 1 {
		t.Errorf("expected exitCode 1 for nonexistent build dir, got %d", exitCode)
	}
	if !strings.Contains(out, "Error building site:") {
		t.Errorf("expected 'Error building site:' in output, got: %s", out)
	}
}

func TestBuildCommand_Offline(t *testing.T) {
	dir := setupTestGraph(t)
	outDir := filepath.Join(t.TempDir(), "public_offline")

	// Offline build
	out, exitCode, err := executeCommandWithExit("build", dir, "--out", outDir, "--offline")
	if err != nil || exitCode != 0 {
		t.Fatalf("build --offline failed: exitCode=%d, err=%v, out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "Building static site") || !strings.Contains(out, "Build complete!") {
		t.Errorf("unexpected build output: %s", out)
	}
	if _, err := os.Stat(filepath.Join(outDir, "index.html")); err != nil {
		t.Errorf("expected index.html in outDir: %v", err)
	}
}

func TestBuildCommand_VendorDir(t *testing.T) {
	dir := setupTestGraph(t)
	outDir := filepath.Join(t.TempDir(), "public_custom_vendor")
	customVendor := t.TempDir()

	if err := os.MkdirAll(filepath.Join(customVendor, "katex"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(customVendor, "mermaid"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(customVendor, "katex", "katex.min.css"), []byte("/* custom */"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(customVendor, "mermaid", "mermaid.min.js"), []byte("/* custom */"), 0o644); err != nil {
		t.Fatal(err)
	}

	// 1. Via flag --vendor-dir
	out, exitCode, err := executeCommandWithExit("build", dir, "--out", outDir, "--offline", "--vendor-dir", customVendor)
	if err != nil || exitCode != 0 {
		t.Fatalf("build --vendor-dir failed: exitCode=%d, err=%v, out=%s", exitCode, err, out)
	}
	if _, err := os.Stat(filepath.Join(outDir, "assets", "vendor", "katex", "katex.min.css")); err != nil {
		t.Errorf("expected copied katex.min.css: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "assets", "vendor", "mermaid", "mermaid.min.js")); err != nil {
		t.Errorf("expected copied mermaid.min.js: %v", err)
	}

	// 2. Via QST_VENDOR_DIR environment variable
	outDirEnv := filepath.Join(t.TempDir(), "public_env_vendor")
	t.Setenv("QST_VENDOR_DIR", customVendor)
	outEnv, exitCodeEnv, errEnv := executeCommandWithExit("build", dir, "--out", outDirEnv, "--offline")
	if errEnv != nil || exitCodeEnv != 0 {
		t.Fatalf("build with QST_VENDOR_DIR failed: exitCode=%d, err=%v, out=%s", exitCodeEnv, errEnv, outEnv)
	}
	if _, err := os.Stat(filepath.Join(outDirEnv, "assets", "vendor", "katex", "katex.min.css")); err != nil {
		t.Errorf("expected copied katex.min.css via QST_VENDOR_DIR: %v", err)
	}
}


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
	if !strings.Contains(out, `def_1["Definition: Def 1"]`) {
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

func TestInspectCommand_Failures(t *testing.T) {
	dir := setupTestGraph(t)

	// Failure case 1: Non-existent guide
	out, exitCode, _ := executeCommandWithExit("inspect", "NonExistentGuide", "--dir", dir)
	if exitCode != 1 {
		t.Errorf("expected exitCode 1 for nonexistent guide inspect, got %d", exitCode)
	}
	if !strings.Contains(out, "Error: guide 'NonExistentGuide' not found") {
		t.Errorf("expected not found error in inspect output, got: %s", out)
	}

	// Failure case 2: Guide exists but manifest is missing
	badDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(badDir, "TestGuide.md"), []byte("---\nscope: foo\n---\nBody\n"), 0644); err != nil {
		t.Fatalf("failed to write test guide: %v", err)
	}
	out, exitCode, _ = executeCommandWithExit("inspect", "TestGuide", "--dir", badDir)
	if exitCode != 1 {
		t.Errorf("expected exitCode 1 for inspect with missing manifest, got %d", exitCode)
	}
	if !strings.Contains(out, "Failed to load manifest.yaml:") {
		t.Errorf("expected 'Failed to load manifest.yaml:' in inspect output, got: %s", out)
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
