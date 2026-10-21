package ssg

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

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

// LineNumNode represents a synthetic line marker anchor injected into the Goldmark AST.
type LineNumNode struct {
	ast.BaseInline
	Line int
}

// KindLineNum is the Goldmark NodeKind for line markers.
var KindLineNum = ast.NewNodeKind("LineNum")

func (n *LineNumNode) Kind() ast.NodeKind {
	return KindLineNum
}

func (n *LineNumNode) Dump(source []byte, level int) {
	_ = level
}

// LineNumRenderer renders LineNumNode elements as HTML span line markers.
type LineNumRenderer struct{}

func (r *LineNumRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindLineNum, r.renderLineNum)
}

func (r *LineNumRenderer) renderLineNum(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		ln := node.(*LineNumNode)
		if _, err := fmt.Fprintf(w, `<span class="line-marker" id="L%d" data-line="%d"><span class="line-num">%d</span></span>`, ln.Line, ln.Line, ln.Line); err != nil {
			return ast.WalkStop, err
		}
	}
	return ast.WalkContinue, nil
}

type lineASTTransformer struct{}

func (lt *lineASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	src := reader.Source()
	if len(src) == 0 {
		return
	}

	lineStarts := []int{0}
	for i, b := range src {
		if b == '\n' {
			lineStarts = append(lineStarts, i+1)
		}
	}

	getLine := func(pos int) int {
		return sort.Search(len(lineStarts), func(i int) bool {
			return lineStarts[i] > pos
		})
	}

	currentLine := 0
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if t, ok := n.(*ast.Text); ok {
			lineNo := getLine(t.Segment.Start)
			if lineNo != currentLine && lineNo > 0 {
				currentLine = lineNo
				lnNode := &LineNumNode{Line: lineNo}
				target := ast.Node(t)
				for target.Parent() != nil && target.Parent().Type() != ast.TypeBlock {
					target = target.Parent()
				}
				parent := target.Parent()
				if parent != nil {
					parent.InsertBefore(parent, target, lnNode)
				}
			}
		} else if cb, ok := n.(*ast.FencedCodeBlock); ok {
			lines := cb.Lines()
			if lines.Len() > 0 {
				lineNo := getLine(lines.At(0).Start) - 1
				if lineNo > 0 && lineNo != currentLine {
					currentLine = lineNo
					parent := cb.Parent()
					if parent != nil {
						lnNode := &LineNumNode{Line: lineNo}
						parent.InsertBefore(parent, cb, lnNode)
					}
				}
			}
		}
		return ast.WalkContinue, nil
	})
}

// newMarkdownParser creates a preconfigured Goldmark Markdown parser with academic styling,
// mathjax formula support, syntax highlighting, link rewrites, and line numbering.
func newMarkdownParser() goldmark.Markdown {
	return goldmark.New(
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
				util.Prioritized(&lineASTTransformer{}, 200),
			),
		),
		goldmark.WithRendererOptions(
			renderer.WithNodeRenderers(
				util.Prioritized(&LineNumRenderer{}, 500),
			),
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)
}
