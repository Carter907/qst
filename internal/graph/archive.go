package graph

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ArchiveGraph packages all the markdown guides into a .kng zip archive.
func ArchiveGraph(dirPath string, outputPath string) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = outFile.Close() }()

	w := zip.NewWriter(outFile)
	defer func() { _ = w.Close() }()

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Only archive markdown files
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		path := filepath.Join(dirPath, entry.Name())

		f, err := w.Create(entry.Name())
		if err != nil {
			return err
		}

		fileContent, err := os.Open(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, fileContent)
		_ = fileContent.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// UnarchiveGraph unpacks all the markdown guides from a .kng zip archive.
func UnarchiveGraph(inputPath string, destDir string) error {
	r, err := zip.OpenReader(inputPath)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {

		if f.FileInfo().IsDir() || filepath.Ext(f.Name) != ".md" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outPath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(outPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", outPath)
		}

		dst, err := os.Create(outPath)
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
