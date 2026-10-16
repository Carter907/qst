package ssg

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// extractTar unpacks tar entries into destDir with Zip Slip protection.
func extractTar(tr *tar.Reader, destDir string) error {
	cleanDest := filepath.Clean(destDir)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(cleanDest, header.Name)
		cleanTarget := filepath.Clean(target)

		// Zip Slip / Path Traversal protection
		if cleanTarget != cleanDest && !strings.HasPrefix(cleanTarget, cleanDest+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", header.Name)
		}

		if header.FileInfo().IsDir() {
			if mkdirErr := os.MkdirAll(cleanTarget, 0o755); mkdirErr != nil {
				return mkdirErr
			}
			continue
		}

		if mkdirErr := os.MkdirAll(filepath.Dir(cleanTarget), 0o755); mkdirErr != nil {
			return mkdirErr
		}
		f, err := os.OpenFile(cleanTarget, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			_ = f.Close()
			return err
		}
		_ = f.Close()
	}
	return nil
}

// downloadKaTeX downloads and extracts the KaTeX tarball to outDir/assets/vendor.
// It creates the katex/ folder inside vendor/.
func downloadKaTeX(destDir string) error {
	targetBase := filepath.Join(destDir, "katex")
	// If already downloaded, skip
	if _, err := os.Stat(filepath.Join(targetBase, "katex.min.css")); err == nil {
		return nil
	}

	resp, err := http.Get("https://github.com/KaTeX/KaTeX/releases/download/v0.16.9/katex.tar.gz")
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer func() { _ = gr.Close() }()

	tr := tar.NewReader(gr)
	return extractTar(tr, destDir)
}

// downloadMermaid downloads the standalone mermaid bundle to outDir/assets/vendor/mermaid/mermaid.min.js.
func downloadMermaid(destDir string) error {
	target := filepath.Join(destDir, "mermaid", "mermaid.min.js")
	if _, err := os.Stat(target); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	resp, err := http.Get("https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js")
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_, err = io.Copy(f, resp.Body)
	return err
}

// DownloadAssets orchestrates the downloading of KaTeX and Mermaid to outDir/assets/vendor.
func DownloadAssets(outDir string) error {
	vendorDir := filepath.Join(outDir, "assets", "vendor")
	if err := os.MkdirAll(vendorDir, 0o755); err != nil {
		return err
	}

	if err := downloadKaTeX(vendorDir); err != nil {
		return err
	}

	if err := downloadMermaid(vendorDir); err != nil {
		return err
	}

	return nil
}
