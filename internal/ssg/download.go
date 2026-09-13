package ssg

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

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
	defer resp.Body.Close()

	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)
		if header.FileInfo().IsDir() {
			if mkdirErr := os.MkdirAll(target, 0755); mkdirErr != nil {
				return mkdirErr
			}
			continue
		}

		if mkdirErr := os.MkdirAll(filepath.Dir(target), 0755); mkdirErr != nil {
			return mkdirErr
		}
		f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
}

// downloadMermaid downloads the mermaid ESM module to outDir/assets/vendor/mermaid/mermaid.esm.min.mjs.
func downloadMermaid(destDir string) error {
	target := filepath.Join(destDir, "mermaid", "mermaid.esm.min.mjs")
	if _, err := os.Stat(target); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	resp, err := http.Get("https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

// DownloadAssets orchestrates the downloading of KaTeX and Mermaid to outDir/assets/vendor.
func DownloadAssets(outDir string) error {
	vendorDir := filepath.Join(outDir, "assets", "vendor")
	if err := os.MkdirAll(vendorDir, 0755); err != nil {
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
