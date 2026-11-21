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
