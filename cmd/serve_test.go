package cmd

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestServeCommand(t *testing.T) {
	dir := setupTestGraph(t)

	// 1. Non-existent directory should exit with code 1
	out, exitCode, _ := executeCommandWithExit("serve", filepath.Join(dir, "nonexistent"), "--offline")
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for non-existent path, got %d (out: %s)", exitCode, out)
	}

	// 2. Non-.kng file should exit with code 1
	out, exitCode, _ = executeCommandWithExit("serve", filepath.Join(dir, "manifest.yaml"), "--offline")
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for non-.kng file, got %d (out: %s)", exitCode, out)
	}

	// 3. Valid graph directory should start, log readiness, and exit cleanly on cancel
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res := executeCommandWithHook(ctx, func(stdout string) {
		if strings.Contains(stdout, "Serving knowledge graph") {
			cancel()
		}
	}, "serve", dir, "--host", "127.0.0.1", "--port", "0", "--offline")
	if res.Err != nil || res.ExitCode != 0 {
		t.Fatalf("expected clean exit for serve command, got code=%d err=%v out=%s errOut=%s", res.ExitCode, res.Err, res.Stdout, res.Stderr)
	}
	if !strings.Contains(res.Stdout, "Serving knowledge graph") {
		t.Errorf("expected 'Serving knowledge graph' output, got: %s", res.Stdout)
	}
}
