package cmd

import (
	"context"
	"strings"
	"testing"
)

func TestPathCommand(t *testing.T) {
	dir := setupTestGraph(t)

	// Link Def 1 as subguide to Desc 1 so there's an active dependency
	out, exitCode, err := executeCommandWithExit("link", "Desc 1", "--subguide", "Def 1", "--segment", "3-3", "--dir", dir)
	if err != nil || exitCode != 0 {
		t.Fatalf("setup link failed: code=%d err=%v out=%s", exitCode, err, out)
	}

	// 1. Target learning path (text)
	out, exitCode, err = executeCommandWithExit("path", "Desc 1", "--dir", dir)
	if err != nil || exitCode != 0 {
		t.Fatalf("path Desc 1 failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "Learning path for 'Desc 1'") {
		t.Errorf("expected target header, got: %s", out)
	}
	if !strings.Contains(out, "Def 1") || !strings.Contains(out, "Desc 1") {
		t.Errorf("expected Def 1 and Desc 1 in path, got: %s", out)
	}

	// 2. Global curriculum (all guides)
	out, exitCode, err = executeCommandWithExit("path", "--dir", dir)
	if err != nil || exitCode != 0 {
		t.Fatalf("path --all failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "Scope: definition") || !strings.Contains(out, "Scope: description") {
		t.Errorf("expected scope groupings in global path, got: %s", out)
	}

	// 3. Format JSON
	jsonRes := executeCommandContext(context.Background(), "path", "Desc 1", "--dir", dir, "--format", "json")
	if jsonRes.Err != nil || jsonRes.ExitCode != 0 {
		t.Fatalf("path json failed: code=%d err=%v out=%s", jsonRes.ExitCode, jsonRes.Err, jsonRes.Combined)
	}
	if jsonRes.Stderr != "" {
		t.Errorf("expected empty stderr for path json, got: %s", jsonRes.Stderr)
	}
	if !strings.Contains(jsonRes.Stdout, `"guide_id": "Desc 1"`) || !strings.Contains(jsonRes.Stdout, `"guide_id": "Def 1"`) {
		t.Errorf("expected JSON output containing guide IDs, got: %s", jsonRes.Stdout)
	}

	// 4. Format Mermaid
	out, exitCode, err = executeCommandWithExit("path", "Desc 1", "--dir", dir, "--format", "mermaid")
	if err != nil || exitCode != 0 {
		t.Fatalf("path mermaid failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "flowchart TD") {
		t.Errorf("expected Mermaid flowchart output, got: %s", out)
	}

	// 5. Subsequent mode
	out, exitCode, err = executeCommandWithExit("path", "Def 1", "--dir", dir, "--subsequent")
	if err != nil || exitCode != 0 {
		t.Fatalf("path subsequent failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "Subsequent guides dependent on 'Def 1'") {
		t.Errorf("expected subsequent header, got: %s", out)
	}
	if !strings.Contains(out, "Desc 1") {
		t.Errorf("expected subsequent guide Desc 1 in output, got: %s", out)
	}

	// 6. Save Tour
	out, exitCode, err = executeCommandWithExit("path", "Desc 1", "--dir", dir, "--save-tour", "Auto Desc Tour")
	if err != nil || exitCode != 0 {
		t.Fatalf("path --save-tour failed: code=%d err=%v out=%s", exitCode, err, out)
	}
	if !strings.Contains(out, "Successfully created tour 'Auto Desc Tour'") {
		t.Errorf("expected success message for tour creation, got: %s", out)
	}
	// Verify tour appears in tour list
	tourOut, tourCode, tourErr := executeCommandWithExit("tour", "list", "--dir", dir)
	if tourErr != nil || tourCode != 0 || !strings.Contains(tourOut, "Auto Desc Tour") {
		t.Errorf("expected tour list to show 'Auto Desc Tour', got code=%d out=%s", tourCode, tourOut)
	}

	// 7. Non-existent guide should exit with code 1
	out, exitCode, _ = executeCommandWithExit("path", "NonexistentGuide", "--dir", dir)
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for non-existent guide, got %d (out: %s)", exitCode, out)
	}
}
