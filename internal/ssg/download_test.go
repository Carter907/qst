package ssg

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractTarZipSlipProtection(t *testing.T) {
	destDir := filepath.Join(t.TempDir(), "safe_vendor")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatalf("failed to create safe destDir: %v", err)
	}

	traversalPaths := []string{
		"../outside.txt",
		"../../evil.txt",
		"katex/../../outside2.txt",
	}

	for _, badPath := range traversalPaths {
		buf := new(bytes.Buffer)
		tw := tar.NewWriter(buf)

		data := []byte("malicious payload")
		hdr := &tar.Header{
			Name: badPath,
			Mode: 0o644,
			Size: int64(len(data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("failed to write tar header for %s: %v", badPath, err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatalf("failed to write tar body for %s: %v", badPath, err)
		}
		if err := tw.Close(); err != nil {
			t.Fatalf("failed to close tar writer: %v", err)
		}

		tr := tar.NewReader(buf)
		err := extractTar(tr, destDir)
		if err == nil {
			t.Fatalf("expected Zip Slip error for %s, got nil", badPath)
		}
		if !strings.Contains(err.Error(), "illegal file path") {
			t.Errorf("expected 'illegal file path' error for %s, got: %v", badPath, err)
		}

		// Ensure no file escaped destDir
		parentDir := filepath.Dir(destDir)
		if _, statErr := os.Stat(filepath.Join(parentDir, "outside.txt")); statErr == nil {
			t.Errorf("file escaped to parent directory!")
		}
		if _, statErr := os.Stat(filepath.Join(parentDir, "evil.txt")); statErr == nil {
			t.Errorf("file escaped to parent directory!")
		}
	}
}

func TestExtractTarValidEntries(t *testing.T) {
	destDir := filepath.Join(t.TempDir(), "vendor")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatalf("failed to create destDir: %v", err)
	}

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// Directory entry
	dirHdr := &tar.Header{
		Name:     "katex/",
		Mode:     0o755,
		Typeflag: tar.TypeDir,
	}
	if err := tw.WriteHeader(dirHdr); err != nil {
		t.Fatalf("failed to write dir header: %v", err)
	}

	// File entry
	cssContent := []byte("/* katex min css */")
	fileHdr := &tar.Header{
		Name: "katex/katex.min.css",
		Mode: 0o644,
		Size: int64(len(cssContent)),
	}
	if err := tw.WriteHeader(fileHdr); err != nil {
		t.Fatalf("failed to write file header: %v", err)
	}
	if _, err := tw.Write(cssContent); err != nil {
		t.Fatalf("failed to write file body: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}

	tr := tar.NewReader(buf)
	if err := extractTar(tr, destDir); err != nil {
		t.Fatalf("extractTar failed on valid entries: %v", err)
	}

	targetFile := filepath.Join(destDir, "katex", "katex.min.css")
	readData, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(readData) != string(cssContent) {
		t.Errorf("extracted content mismatch, got %q, want %q", string(readData), string(cssContent))
	}
}
