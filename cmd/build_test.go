package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
