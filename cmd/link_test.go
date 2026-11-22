package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Carter907/qst/internal/graph"
)

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

func TestLinkLenientWithInvalidGuides(t *testing.T) {
	dir := setupTestGraph(t)

	// Add unfinished/broken files and a duplicate guide ID
	if err := os.WriteFile(filepath.Join(dir, "WIP Draft.md"), []byte("# Unfinished draft\nNo frontmatter yet.\n"), 0644); err != nil {
		t.Fatalf("failed to write WIP Draft: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Broken YAML.md"), []byte("---\nscope: [broken\n---\n# Broken\n"), 0644); err != nil {
		t.Fatalf("failed to write Broken YAML: %v", err)
	}
	_ = os.MkdirAll(filepath.Join(dir, "subA"), 0755)
	_ = os.MkdirAll(filepath.Join(dir, "subB"), 0755)
	dupContent := "---\nscope: definition\nprerequisites: []\nsub_guides: []\ntags: []\n---\n# Dup\n"
	_ = os.WriteFile(filepath.Join(dir, "subA", "DupGuide.md"), []byte(dupContent), 0644)
	_ = os.WriteFile(filepath.Join(dir, "subB", "DupGuide.md"), []byte(dupContent), 0644)

	// 1. Linking two valid guides succeeds with a concise stderr warning
	res := executeCommandContext(nil, "link", "Def 1", "--dir", dir, "--prereq", "Def 2")
	if res.ExitCode != 0 || res.Err != nil {
		t.Fatalf("expected lenient link to succeed, got exit=%d err=%v combined=%s", res.ExitCode, res.Err, res.Combined)
	}
	if !strings.Contains(res.Stderr, "warning: skipped 3 invalid guide(s); run 'qst validate' for details") {
		t.Errorf("expected stderr warning about 3 skipped guides, got stderr=%q", res.Stderr)
	}
	if !strings.Contains(res.Stdout, "Successfully linked prerequisite 'Def 2' to 'Def 1'") {
		t.Errorf("expected success message on stdout, got stdout=%q", res.Stdout)
	}

	// 2. Targeting a broken guide fails with that specific file's parse error
	res = executeCommandContext(nil, "link", "WIP Draft", "--dir", dir, "--prereq", "Def 1")
	if res.ExitCode == 0 {
		t.Fatalf("expected link targeting broken guide to fail, got exit=0")
	}
	if !strings.Contains(res.Combined, "WIP Draft.md") || !strings.Contains(res.Combined, "invalid or missing YAML frontmatter") {
		t.Errorf("expected specific frontmatter parse error for WIP Draft.md, got:\n%s", res.Combined)
	}

	// 3. Referencing a broken guide via --prereq fails with that specific file's parse error
	res = executeCommandContext(nil, "link", "Def 1", "--dir", dir, "--prereq", "Broken YAML")
	if res.ExitCode == 0 {
		t.Fatalf("expected link with broken --prereq to fail, got exit=0")
	}
	if !strings.Contains(res.Combined, "Broken YAML.md") || !strings.Contains(res.Combined, "failed to parse yaml") {
		t.Errorf("expected specific YAML parse error for Broken YAML.md, got:\n%s", res.Combined)
	}

	// 4. Referencing a broken guide via --subguide fails with that specific file's parse error
	res = executeCommandContext(nil, "link", "Desc 1", "--dir", dir, "--subguide", "WIP Draft", "--adherence", "introductory")
	if res.ExitCode == 0 {
		t.Fatalf("expected link with broken --subguide to fail, got exit=0")
	}
	if !strings.Contains(res.Combined, "WIP Draft.md") || !strings.Contains(res.Combined, "invalid or missing YAML frontmatter") {
		t.Errorf("expected specific frontmatter parse error for WIP Draft.md, got:\n%s", res.Combined)
	}

	// 5. Targeting a duplicated guide ID fails with the duplicate conflict error
	res = executeCommandContext(nil, "link", "DupGuide", "--dir", dir, "--prereq", "Def 1")
	if res.ExitCode == 0 {
		t.Fatalf("expected link targeting duplicate guide ID to fail, got exit=0")
	}
	if !strings.Contains(res.Combined, "duplicate guide ID 'DupGuide'") {
		t.Errorf("expected duplicate guide ID error, got:\n%s", res.Combined)
	}

	// 6. Other lenient commands (e.g. list and unlink) succeed with stderr warning
	res = executeCommandContext(nil, "list", "-d", dir)
	if res.ExitCode != 0 || res.Err != nil {
		t.Fatalf("expected lenient list to succeed, got exit=%d err=%v combined=%s", res.ExitCode, res.Err, res.Combined)
	}
	if !strings.Contains(res.Stdout, "Def 1") || !strings.Contains(res.Stderr, "warning: skipped 3 invalid guide(s)") {
		t.Errorf("unexpected list output: stdout=%q stderr=%q", res.Stdout, res.Stderr)
	}

	res = executeCommandContext(nil, "unlink", "Def 1", "-d", dir, "--prereq", "Def 2")
	if res.ExitCode != 0 || res.Err != nil {
		t.Fatalf("expected lenient unlink to succeed, got exit=%d err=%v combined=%s", res.ExitCode, res.Err, res.Combined)
	}

	// 7. qst validate stays strict and reports ALL 3 collected parse errors in one pass
	res = executeCommandContext(nil, "validate", dir)
	if res.ExitCode == 0 {
		t.Fatalf("expected validate to fail on invalid guides, got exit=0")
	}
	for _, want := range []string{"3 error(s)", "WIP Draft.md", "Broken YAML.md", "duplicate guide ID 'DupGuide'"} {
		if !strings.Contains(res.Combined, want) {
			t.Errorf("expected validate output to contain %q, got:\n%s", want, res.Combined)
		}
	}
}
