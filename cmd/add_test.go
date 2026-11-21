package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Carter907/qst/internal/graph"
)

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

	// 7. --in-place without --file should fail
	out, exitCode, _ = executeCommandWithExit("add", "InPlaceFail", "-d", dir, "--scope", "definition", "--in-place")
	if exitCode != 1 || !strings.Contains(out, "--in-place requires --file") {
		t.Errorf("expected exit code 1 and '--in-place requires --file', got %d, out: %s", exitCode, out)
	}

	// 8. Valid --in-place addition transforming a raw markdown file
	draftPath := filepath.Join(dir, "RawDraft.md")
	rawBody := "## Raw Draft\n\nThis is existing body text from draft.\n"
	if err := os.WriteFile(draftPath, []byte(rawBody), 0644); err != nil {
		t.Fatalf("failed to write raw draft: %v", err)
	}

	out, exitCode, err = executeCommandWithExit("add", "RawDraft", "-d", dir, "--file", draftPath, "--in-place", "--scope", "definition", "--tags", "draft,imported")
	if err != nil || exitCode != 0 {
		t.Fatalf("failed to add guide in place: %v (code %d, out: %s)", err, exitCode, out)
	}
	if !strings.Contains(out, "Successfully added new guide: "+draftPath) {
		t.Errorf("unexpected output for in-place add: %s", out)
	}

	transformedGuide, err := graph.ParseGuide(draftPath)
	if err != nil {
		t.Fatalf("failed to parse transformed guide: %v", err)
	}
	if transformedGuide.Metadata.Scope != "definition" {
		t.Errorf("expected scope definition, got %s", transformedGuide.Metadata.Scope)
	}
	if len(transformedGuide.Metadata.Tags) != 2 || transformedGuide.Metadata.Tags[0] != "draft" || transformedGuide.Metadata.Tags[1] != "imported" {
		t.Errorf("expected tags [draft, imported], got %v", transformedGuide.Metadata.Tags)
	}
	if !strings.Contains(transformedGuide.Body, "This is existing body text from draft.") {
		t.Errorf("expected original draft content preserved in body, got:\n%s", transformedGuide.Body)
	}
}
