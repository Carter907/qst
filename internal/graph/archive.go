package graph

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// shouldArchiveEntry determines whether a file should be included in a .kng package.
func shouldArchiveEntry(relPath string, info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}

	base := filepath.Base(relPath)
	if strings.HasPrefix(base, ".") {
		return false
	}

	// Always preserve the knowledge graph manifest
	if base == "manifest.yaml" || base == "manifest.yml" || base == "quest.yaml" || base == "quest.yml" {
		return true
	}

	ext := strings.ToLower(filepath.Ext(base))
	if ext == ".kng" {
		return false
	}

	switch ext {
	case ".md", ".mmd", ".mermaid", ".png", ".jpg", ".jpeg", ".svg", ".gif", ".webp":
		return true
	default:
		return false
	}
}

// ArchiveGraph packages all markdown guides, manifests, and supporting assets into a compressed .kng zip archive.
func ArchiveGraph(dirPath string, outputPath string) error {
	outDir := filepath.Dir(outputPath)
	if outDir != "." && outDir != "" {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return err
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = outFile.Close() }()

	w := zip.NewWriter(outFile)
	defer func() { _ = w.Close() }()

	cleanDir := filepath.Clean(dirPath)

	err = filepath.Walk(cleanDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(cleanDir, path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		// Skip hidden directories and build outputs like public/
		if info.IsDir() {
			base := info.Name()
			if strings.HasPrefix(base, ".") || base == "public" {
				return filepath.SkipDir
			}
			return nil
		}

		if !shouldArchiveEntry(rel, info) {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Store with forward slashes for cross-platform zip compatibility
		header.Name = filepath.ToSlash(rel)
		header.Method = zip.Deflate

		f, err := w.CreateHeader(header)
		if err != nil {
			return err
		}

		fileContent, err := os.Open(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, fileContent)
		_ = fileContent.Close()
		return err
	})

	return err
}

// UnarchiveGraph unpacks all guides, manifests, and assets from a .kng zip archive into destDir.
func UnarchiveGraph(inputPath string, destDir string) error {
	r, err := zip.OpenReader(inputPath)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	cleanDest := filepath.Clean(destDir)

	for _, f := range r.File {
		outPath := filepath.Join(cleanDest, f.Name)
		cleanOut := filepath.Clean(outPath)

		// Zip Slip protection
		if cleanOut != cleanDest && !strings.HasPrefix(cleanOut, cleanDest+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(cleanOut, 0o755); err != nil {
				return err
			}
			continue
		}

		// Ensure parent directory exists for nested assets/guides
		if err := os.MkdirAll(filepath.Dir(cleanOut), 0o755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		dst, err := os.OpenFile(cleanOut, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, f.Mode())
		if err != nil {
			_ = rc.Close()
			return err
		}

		if _, err := io.Copy(dst, rc); err != nil {
			_ = dst.Close()
			_ = rc.Close()
			return err
		}

		_ = dst.Close()
		_ = rc.Close()
	}

	return nil
}
