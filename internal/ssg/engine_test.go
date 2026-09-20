package ssg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSite(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	outDir := filepath.Join(tmpDir, "public")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Create a mock markdown file with MathJax, Mermaid, and GFM

	manifestContent := "name: Test\nversion: 1.0.0\n"
	if err := os.WriteFile(filepath.Join(sourceDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to write mock manifest: %v", err)
	}

	markdownContent := `
# Test Knowledge Graph

## LaTeX Math
Block Math:
$$
E=mc^2
$$

Inline Math:
The formula is $x = y^2$.

## Mermaid
` + "```" + `mermaid
graph TD;
    A-->B;
` + "```" + `

## Table
| Header 1 | Header 2 |
|----------|----------|
| Row 1    | Value 1  |
`
	if err := os.WriteFile(filepath.Join(sourceDir, "test.md"), []byte(markdownContent), 0644); err != nil {
		t.Fatalf("Failed to write mock md: %v", err)
	}

	// Build the site
	if err := BuildSite(sourceDir, outDir); err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// Check if output files exist
	expectedHTML := filepath.Join(outDir, "test.html")
	content, err := os.ReadFile(expectedHTML)
	if err != nil {
		t.Fatalf("Failed to read generated HTML: %v", err)
	}
	htmlStr := string(content)

	// Verify math rendering wrappers
	if !strings.Contains(htmlStr, `<span class="math display">`) {
		t.Errorf("Expected block math in HTML, got: %s", htmlStr)
	}

	// Verify mermaid code block (Goldmark should render it as <code class="language-mermaid">)
	if !strings.Contains(htmlStr, `class="language-mermaid"`) {
		t.Errorf("Expected mermaid class in HTML")
	}

	// Verify table
	if !strings.Contains(htmlStr, `<table>`) {
		t.Errorf("Expected table in HTML")
	}

	// Verify assets downloaded
	if _, err := os.Stat(filepath.Join(outDir, "assets", "vendor", "katex", "katex.min.css")); err != nil {
		t.Errorf("KaTeX CSS not found in outDir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "assets", "vendor", "mermaid", "mermaid.esm.min.mjs")); err != nil {
		t.Errorf("Mermaid module not found in outDir: %v", err)
	}
}
