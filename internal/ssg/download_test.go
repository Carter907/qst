package ssg

import (
	"archive/tar"
	"bytes"
	"fmt"
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

func TestResolveUserCacheDir(t *testing.T) {
	cacheDir, err := ResolveUserCacheDir()
	if err != nil {
		t.Skipf("skipping ResolveUserCacheDir test: %v", err)
	}
	if !strings.Contains(cacheDir, "qst") || !strings.Contains(cacheDir, "vendor") {
		t.Errorf("expected cacheDir to contain 'qst' and 'vendor', got: %s", cacheDir)
	}
}

func TestEnsureVendorAssetsOfflineExisting(t *testing.T) {
	outDir := t.TempDir()
	vendorDir := filepath.Join(outDir, "assets", "vendor")
	if err := os.MkdirAll(filepath.Join(vendorDir, "katex"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(vendorDir, "mermaid"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vendorDir, "katex", "katex.min.css"), []byte("/* katex */"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vendorDir, "mermaid", "mermaid.min.js"), []byte("/* mermaid */"), 0o644); err != nil {
		t.Fatal(err)
	}

	warned := false
	warnFn := func(format string, args ...any) {
		warned = true
	}

	err := EnsureVendorAssets(outDir, VendorOptions{
		Offline:  true,
		WarnFunc: warnFn,
	})
	if err != nil {
		t.Fatalf("unexpected error in EnsureVendorAssets: %v", err)
	}
	if warned {
		t.Errorf("did not expect warning when vendor assets are already present")
	}
}

func TestEnsureVendorAssetsOfflineFromCustomVendorDir(t *testing.T) {
	outDir := t.TempDir()
	customVendor := t.TempDir()

	if err := os.MkdirAll(filepath.Join(customVendor, "katex"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(customVendor, "mermaid"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(customVendor, "katex", "katex.min.css"), []byte("/* custom katex */"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(customVendor, "mermaid", "mermaid.min.js"), []byte("/* custom mermaid */"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := EnsureVendorAssets(outDir, VendorOptions{
		Offline:   true,
		VendorDir: customVendor,
	})
	if err != nil {
		t.Fatalf("EnsureVendorAssets failed with custom VendorDir: %v", err)
	}

	targetCSS := filepath.Join(outDir, "assets", "vendor", "katex", "katex.min.css")
	if _, err := os.Stat(targetCSS); err != nil {
		t.Errorf("expected copied katex.min.css in outDir: %v", err)
	}

	targetJS := filepath.Join(outDir, "assets", "vendor", "mermaid", "mermaid.min.js")
	if _, err := os.Stat(targetJS); err != nil {
		t.Errorf("expected copied mermaid.min.js in outDir: %v", err)
	}
}

func TestEnsureVendorAssetsOfflineFromCache(t *testing.T) {
	outDir := t.TempDir()
	mockCache := t.TempDir()

	if err := os.MkdirAll(filepath.Join(mockCache, "katex"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(mockCache, "mermaid"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mockCache, "katex", "katex.min.css"), []byte("/* cache katex */"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mockCache, "mermaid", "mermaid.min.js"), []byte("/* cache mermaid */"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := EnsureVendorAssets(outDir, VendorOptions{
		Offline:  true,
		CacheDir: mockCache,
	})
	if err != nil {
		t.Fatalf("EnsureVendorAssets failed with mock CacheDir: %v", err)
	}

	targetCSS := filepath.Join(outDir, "assets", "vendor", "katex", "katex.min.css")
	if _, err := os.Stat(targetCSS); err != nil {
		t.Errorf("expected copied katex.min.css from cache into outDir: %v", err)
	}
}

func TestEnsureVendorAssetsOfflineMissing(t *testing.T) {
	outDir := t.TempDir()
	emptyCache := t.TempDir()

	var warningMsg string
	warnFn := func(format string, args ...any) {
		warningMsg = fmt.Sprintf(format, args...)
	}

	err := EnsureVendorAssets(outDir, VendorOptions{
		Offline:  true,
		CacheDir: emptyCache,
		WarnFunc: warnFn,
	})
	if err != nil {
		t.Fatalf("EnsureVendorAssets returned unexpected error in offline mode: %v", err)
	}
	if !strings.Contains(warningMsg, "offline mode enabled and vendor assets (KaTeX / Mermaid) were not found") {
		t.Errorf("expected warning message about missing vendor assets, got: %q", warningMsg)
	}
}

