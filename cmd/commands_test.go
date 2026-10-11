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
