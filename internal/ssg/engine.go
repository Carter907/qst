package ssg

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

//go:embed templates/*
var templatesFS embed.FS

type PageData struct {
	Title   string
	Content template.HTML
}

type linkASTTransformer struct{}

func (l *linkASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if link, ok := n.(*ast.Link); ok {
			dest := string(link.Destination)
			// Check if it's a relative link pointing to a .md file
			if strings.HasSuffix(dest, ".md") && !strings.HasPrefix(dest, "http") {
				baseDest := filepath.Base(dest)
				if strings.ToLower(baseDest) == "index.md" {
					dirDest := filepath.Dir(dest)
					if dirDest == "." {
						link.Destination = []byte("index-guide.html")
					} else {
						link.Destination = []byte(dirDest + "/index-guide.html")
					}
				} else {
					newDest := strings.TrimSuffix(dest, ".md") + ".html"
					link.Destination = []byte(newDest)
				}
			}
		}
		return ast.WalkContinue, nil
	})
}

func BuildSite(dirPath string, outDir string) error {
	// Ensure outDir exists
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create outDir: %w", err)
	}

	// 1. Download vendor assets to outDir/assets/vendor
	if err := DownloadAssets(outDir); err != nil {
		return fmt.Errorf("failed to download vendor assets: %w", err)
	}

	// 2. Copy style.css to outDir/assets
	styleData, err := templatesFS.ReadFile("templates/style.css")
	if err != nil {
		return fmt.Errorf("failed to read embedded style.css: %w", err)
	}
	if writeErr := os.WriteFile(filepath.Join(outDir, "assets", "style.css"), styleData, 0644); writeErr != nil {
		return fmt.Errorf("failed to write style.css: %w", writeErr)
	}

	// 3. Initialize Goldmark
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			mathjax.MathJax,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithASTTransformers(
				util.Prioritized(&linkASTTransformer{}, 100),
			),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	// 4. Parse Layout
	layoutStr, err := templatesFS.ReadFile("templates/layout.html")
	if err != nil {
		return fmt.Errorf("failed to read layout.html: %w", err)
	}
	tmpl, err := template.New("layout").Parse(string(layoutStr))
	if err != nil {
		return fmt.Errorf("failed to parse layout template: %w", err)
	}

	// 5. Process Markdown Files
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Skip output directory itself to avoid infinite loops if outDir is inside dirPath
		if info.IsDir() {
			absPath, _ := filepath.Abs(path)
			absOutDir, _ := filepath.Abs(outDir)
			if absPath == absOutDir {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(info.Name(), ".md") {
			return processMarkdownFile(path, dirPath, outDir, info, md, tmpl)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 6. Generate Index Page
	indexContent, err := templatesFS.ReadFile("templates/index.html")
	if err != nil {
		return fmt.Errorf("failed to read embedded index.html: %w", err)
	}

	indexData := PageData{
		Title:   "Knowledge Graph",
		Content: template.HTML(indexContent),
	}

	idxOutPath := filepath.Join(outDir, "index.html")
	idxF, err := os.Create(idxOutPath)
	if err != nil {
		return fmt.Errorf("failed to create index.html: %w", err)
	}
	defer idxF.Close()

	if tmplErr := tmpl.Execute(idxF, indexData); tmplErr != nil {
		return fmt.Errorf("failed to execute template for index.html: %w", tmplErr)
	}

	return nil
}

func processMarkdownFile(path, dirPath, outDir string, info os.FileInfo, md goldmark.Markdown, tmpl *template.Template) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if convErr := md.Convert(src, &buf); convErr != nil {
		return convErr
	}

	relPath, err := filepath.Rel(dirPath, path)
	if err != nil {
		return err
	}

	baseName := filepath.Base(relPath)
	if strings.ToLower(baseName) == "index.md" {
		dirName := filepath.Dir(relPath)
		if dirName == "." {
			relPath = "index-guide.md"
		} else {
			relPath = filepath.Join(dirName, "index-guide.md")
		}
	}

	outPath := filepath.Join(outDir, strings.TrimSuffix(relPath, ".md")+".html")

	if mkdirErr := os.MkdirAll(filepath.Dir(outPath), 0755); mkdirErr != nil {
		return mkdirErr
	}

	data := PageData{
		Title:   strings.TrimSuffix(info.Name(), ".md"),
		Content: template.HTML(buf.String()),
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if tmplErr := tmpl.Execute(f, data); tmplErr != nil {
		return tmplErr
	}

	return nil
}
