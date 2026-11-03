package ssg

import (
	"fmt"
	"html"
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
	htmlRenderer "github.com/yuin/goldmark/renderer/html"
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

// LineNumRenderer renders LineNumNode elements as invisible HTML span line anchors.
type LineNumRenderer struct{}

func (r *LineNumRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindLineNum, r.renderLineNum)
}

func (r *LineNumRenderer) renderLineNum(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		ln := node.(*LineNumNode)
		if _, err := fmt.Fprintf(w, `<span class="line-marker" id="L%d" data-line="%d"></span>`, ln.Line, ln.Line); err != nil {
			return ast.WalkStop, err
		}
	}
	return ast.WalkContinue, nil
}

// SegmentInfo represents contextual metadata for a sub-guide segment citation.
type SegmentInfo struct {
	GuideID   string
	Title     string
	URL       string
	Adherence string
	Segment   string
	RangeStr  string
	StartLine int
	EndLine   int
}

// SubGuidesContextKey is the Goldmark parser.Context key for passing segment metadata.
var SubGuidesContextKey = parser.NewContextKey()

// SegmentBlockNode represents a container block wrapping the content of a cited sub-guide segment.
type SegmentBlockNode struct {
	ast.BaseBlock
	GuideID   string
	Title     string
	URL       string
	Adherence string
	RangeStr  string
	StartLine int
	EndLine   int
}

// KindSegmentBlock is the Goldmark NodeKind for segment blocks.
var KindSegmentBlock = ast.NewNodeKind("SegmentBlock")

func (n *SegmentBlockNode) Kind() ast.NodeKind {
	return KindSegmentBlock
}

func (n *SegmentBlockNode) Dump(source []byte, level int) {
	_ = level
}

// SegmentRenderer renders SegmentBlockNode elements as HTML section with header chips.
type SegmentRenderer struct{}

func (r *SegmentRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindSegmentBlock, r.renderSegment)
}

func (r *SegmentRenderer) renderSegment(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*SegmentBlockNode)
	if entering {
		tooltip := fmt.Sprintf("Sub-guide: %s", n.Title)
		if n.Adherence != "" && n.RangeStr != "" {
			tooltip = fmt.Sprintf("Sub-guide: %s (%s, %s)", n.Title, n.Adherence, n.RangeStr)
		} else if n.Adherence != "" {
			tooltip = fmt.Sprintf("Sub-guide: %s (%s)", n.Title, n.Adherence)
		} else if n.RangeStr != "" {
			tooltip = fmt.Sprintf("Sub-guide: %s (%s)", n.Title, n.RangeStr)
		}

		href := n.URL
		if href == "" {
			href = "#"
		}

		if _, err := fmt.Fprintf(w, `<section class="segment-section" data-segment="%d-%d" id="segment-L%d">
  <a href="%s" class="segment-indicator" title="%s" aria-label="%s">↳</a>
  <div class="segment-body">
`, n.StartLine, n.EndLine, n.StartLine, html.EscapeString(href), html.EscapeString(tooltip), html.EscapeString(tooltip)); err != nil {
			return ast.WalkStop, err
		}
	} else {
		if _, err := fmt.Fprint(w, `  </div>
</section>
`); err != nil {
			return ast.WalkStop, err
		}
	}
	return ast.WalkContinue, nil
}

func getBlockLines(n ast.Node, getLine func(int) int) (int, int) {
	start := -1
	end := -1
	if lines := n.Lines(); lines != nil && lines.Len() > 0 {
		s := getLine(lines.At(0).Start)
		e := getLine(lines.At(lines.Len() - 1).Stop)
		if start == -1 || s < start {
			start = s
		}
		if e > end {
			end = e
		}
	}
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := c.(*ast.Text); ok {
			s := getLine(t.Segment.Start)
			e := getLine(t.Segment.Stop)
			if start == -1 || s < start {
				start = s
			}
			if e > end {
				end = e
			}
		} else if cb, ok := c.(*ast.FencedCodeBlock); ok {
			lines := cb.Lines()
			if lines != nil && lines.Len() > 0 {
				s := getLine(lines.At(0).Start) - 1
				if s <= 0 {
					s = 1
				}
				e := getLine(lines.At(lines.Len() - 1).Stop)
				if start == -1 || s < start {
					start = s
				}
				if e > end {
					end = e
				}
			}
		}
		return ast.WalkContinue, nil
	})
	if end < start {
		end = start
	}
	return start, end
}

type segmentASTTransformer struct{}

func (st *segmentASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	val := pc.Get(SubGuidesContextKey)
	if val == nil {
		return
	}
	segments, ok := val.([]SegmentInfo)
	if !ok || len(segments) == 0 {
		return
	}

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

	type blockLines struct {
		node  ast.Node
		start int
		end   int
	}

	var blocks []blockLines
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		start, end := getBlockLines(c, getLine)
		if start > 0 {
			blocks = append(blocks, blockLines{node: c, start: start, end: end})
		}
	}

	if len(blocks) == 0 {
		return
	}

	for _, seg := range segments {
		var matched []ast.Node
		for _, b := range blocks {
			if b.start >= seg.StartLine && b.start <= seg.EndLine {
				matched = append(matched, b.node)
			}
		}

		if len(matched) == 0 {
			continue
		}

		firstMatched := matched[0]
		if firstMatched.Parent() != node {
			continue
		}

		segBlock := &SegmentBlockNode{
			GuideID:   seg.GuideID,
			Title:     seg.Title,
			URL:       seg.URL,
			Adherence: seg.Adherence,
			RangeStr:  seg.RangeStr,
			StartLine: seg.StartLine,
			EndLine:   seg.EndLine,
		}

		node.InsertBefore(node, firstMatched, segBlock)
		for _, m := range matched {
			if m.Parent() == node {
				node.RemoveChild(node, m)
				segBlock.AppendChild(segBlock, m)
			}
		}
	}
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
// mathjax formula support, syntax highlighting, link rewrites, segment callouts, and line anchors.
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
				util.Prioritized(&segmentASTTransformer{}, 150),
				util.Prioritized(&lineASTTransformer{}, 200),
			),
		),
		goldmark.WithRendererOptions(
			renderer.WithNodeRenderers(
				util.Prioritized(&SegmentRenderer{}, 400),
				util.Prioritized(&LineNumRenderer{}, 500),
			),
			htmlRenderer.WithHardWraps(),
			htmlRenderer.WithUnsafe(),
		),
	)
}
