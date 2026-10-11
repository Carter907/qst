// Package ssg contains logic for the creation of the static site
package ssg

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Carter907/qst/internal/graph"
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

//go:embed templates/*
var templatesFS embed.FS

type GuideRef struct {
	ID        string
	Title     string
	URL       string
	Scope     string
	Adherence string
	Segment   string
}

type SidebarLink struct {
	Title    string
	URL      string
	IsActive bool
	Scope    string
}

type SidebarScopeGroup struct {
	Scope  string
	Guides []SidebarLink
}

type SidebarTour struct {
	Name   string
	Guides []SidebarLink
}

type SidebarData struct {
	HomeURL     string
	IsHome      bool
	Tours       []SidebarTour
	ScopeGroups []SidebarScopeGroup
}

type TourStepRef struct {
	TourName string
	Step     int
	Total    int
	Prev     *GuideRef
	Next     *GuideRef
}

type PageData struct {
	Title         string
	Scope         string
	Tags          []string
	Content       template.HTML
	Prerequisites []GuideRef
	Dependents    []GuideRef
	SubGuides     []GuideRef
	ParentGuides  []GuideRef
	TourSteps     []TourStepRef
	RelativeRoot  string
	Sidebar       SidebarData
}

type IndexPageData struct {
	Title        string
	Description  string
	RelativeRoot string
	Sidebar      SidebarData
	MermaidGraph template.HTML
}

type guideFile struct {
	guide      graph.Guide
	path       string
	outPath    string
	outRelPath string
	relRoot    string
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

type LineNumNode struct {
	ast.BaseInline
	Line int
}

var KindLineNum = ast.NewNodeKind("LineNum")

func (n *LineNumNode) Kind() ast.NodeKind {
	return KindLineNum
}

func (n *LineNumNode) Dump(source []byte, level int) {
	_ = level
}

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

func relURL(fromOutRelPath, toOutRelPath string) string {
	fromDir := filepath.Dir(fromOutRelPath)
	rel, err := filepath.Rel(fromDir, toOutRelPath)
	if err != nil {
		return toOutRelPath
	}
	return filepath.ToSlash(rel)
}

func BuildSite(dirPath string, outDir string) error {
	config, err := graph.ParseConfig(dirPath)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Ensure outDir exists
	if err := os.MkdirAll(outDir, 0o755); err != nil {
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
	if writeErr := os.WriteFile(filepath.Join(outDir, "assets", "style.css"), styleData, 0o644); writeErr != nil {
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

	// 4. Parse Templates
	indexTmpl, err := template.ParseFS(templatesFS, "templates/layout.html", "templates/index.html")
	if err != nil {
		return fmt.Errorf("failed to parse index templates: %w", err)
	}
	guideTmpl, err := template.ParseFS(templatesFS, "templates/layout.html", "templates/guide.html")
	if err != nil {
		return fmt.Errorf("failed to parse guide templates: %w", err)
	}

	// 5. Discover and Index Markdown Files
	var guideFiles []*guideFile
	filesByID := make(map[string]*guideFile)

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
			name := info.Name()
			if path != dirPath && (strings.HasPrefix(name, ".") || name == "vendor" || name == "assets" || name == "public") {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(info.Name(), ".md") {
			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}

			baseName := filepath.Base(relPath)
			adjustedRel := relPath
			if strings.ToLower(baseName) == "index.md" {
				dirName := filepath.Dir(relPath)
				if dirName == "." {
					adjustedRel = "index-guide.md"
				} else {
					adjustedRel = filepath.Join(dirName, "index-guide.md")
				}
			}

			outRelPath := strings.TrimSuffix(adjustedRel, ".md") + ".html"
			outPath := filepath.Join(outDir, outRelPath)

			outRelDir := filepath.Dir(outRelPath)
			relRoot := ""
			if outRelDir != "." && outRelDir != "" {
				rel, err := filepath.Rel(outRelDir, ".")
				if err == nil {
					relRoot = filepath.ToSlash(rel) + "/"
				}
			}

			g, err := graph.ParseGuide(path)
			if err != nil {
				// Fallback for markdown files without valid frontmatter
				src, readErr := os.ReadFile(path)
				if readErr != nil {
					return readErr
				}
				id := strings.TrimSuffix(info.Name(), ".md")
				g = graph.Guide{
					ID:         id,
					Path:       path,
					Body:       string(src),
					HasContent: len(src) > 0,
				}
			}

			gf := &guideFile{
				guide:      g,
				path:       path,
				outPath:    outPath,
				outRelPath: outRelPath,
				relRoot:    relRoot,
			}
			guideFiles = append(guideFiles, gf)
			filesByID[g.ID] = gf
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 6. Build Inverted Relations (Dependents and Parent Guides)
	dependentsMap := make(map[string][]string)
	parentGuidesMap := make(map[string][]string)

	for _, gf := range guideFiles {
		for _, prereq := range gf.guide.Metadata.Prerequisites {
			dependentsMap[prereq] = append(dependentsMap[prereq], gf.guide.ID)
		}
		for _, sub := range gf.guide.Metadata.SubGuides {
			parentGuidesMap[sub.Guide] = append(parentGuidesMap[sub.Guide], gf.guide.ID)
		}
	}

	// 7. Render Guides
	for _, gf := range guideFiles {
		var prereqs []GuideRef
		for _, prereqID := range gf.guide.Metadata.Prerequisites {
			ref := GuideRef{
				ID:    prereqID,
				Title: prereqID,
			}
			if target, ok := filesByID[prereqID]; ok {
				ref.URL = relURL(gf.outRelPath, target.outRelPath)
				ref.Scope = target.guide.Metadata.Scope
			}
			prereqs = append(prereqs, ref)
		}

		var dependents []GuideRef
		depIDs := append([]string(nil), dependentsMap[gf.guide.ID]...)
		sort.Strings(depIDs)
		for _, depID := range depIDs {
			ref := GuideRef{
				ID:    depID,
				Title: depID,
			}
			if target, ok := filesByID[depID]; ok {
				ref.URL = relURL(gf.outRelPath, target.outRelPath)
				ref.Scope = target.guide.Metadata.Scope
			}
			dependents = append(dependents, ref)
		}

		var subGuides []GuideRef
		for _, sub := range gf.guide.Metadata.SubGuides {
			ref := GuideRef{
				ID:        sub.Guide,
				Title:     sub.Guide,
				Adherence: sub.Adherence,
				Segment:   sub.Segment,
			}
			if target, ok := filesByID[sub.Guide]; ok {
				ref.URL = relURL(gf.outRelPath, target.outRelPath)
				ref.Scope = target.guide.Metadata.Scope
			}
			subGuides = append(subGuides, ref)
		}

		var parentGuides []GuideRef
		parentIDs := append([]string(nil), parentGuidesMap[gf.guide.ID]...)
		sort.Strings(parentIDs)
		for _, parentID := range parentIDs {
			ref := GuideRef{
				ID:    parentID,
				Title: parentID,
			}
			if target, ok := filesByID[parentID]; ok {
				ref.URL = relURL(gf.outRelPath, target.outRelPath)
				ref.Scope = target.guide.Metadata.Scope
			}
			parentGuides = append(parentGuides, ref)
		}

		var buf bytes.Buffer
		if convErr := md.Convert([]byte(gf.guide.Body), &buf); convErr != nil {
			return convErr
		}

		if mkdirErr := os.MkdirAll(filepath.Dir(gf.outPath), 0o755); mkdirErr != nil {
			return mkdirErr
		}

		sidebar := buildSidebar(gf.outRelPath, gf.guide.ID, config, guideFiles, filesByID)
		tourSteps := buildTourSteps(gf, config, filesByID)

		data := PageData{
			Title:         gf.guide.ID,
			Scope:         gf.guide.Metadata.Scope,
			Tags:          gf.guide.Metadata.Tags,
			Content:       template.HTML(buf.String()),
			Prerequisites: prereqs,
			Dependents:    dependents,
			SubGuides:     subGuides,
			ParentGuides:  parentGuides,
			TourSteps:     tourSteps,
			RelativeRoot:  gf.relRoot,
			Sidebar:       sidebar,
		}

		f, err := os.Create(gf.outPath)
		if err != nil {
			return err
		}

		if tmplErr := guideTmpl.ExecuteTemplate(f, "layout.html", data); tmplErr != nil {
			_ = f.Close()
			return tmplErr
		}
		_ = f.Close()
	}

	// 8. Generate Index Page
	indexSidebar := buildSidebar("index.html", "", config, guideFiles, filesByID)
	indexMermaid := generateIndexMermaid(guideFiles, filesByID)
	indexData := IndexPageData{
		Title:        config.Title,
		Description:  config.Description,
		RelativeRoot: "",
		Sidebar:      indexSidebar,
		MermaidGraph: template.HTML(indexMermaid),
	}

	idxOutPath := filepath.Join(outDir, "index.html")
	idxF, err := os.Create(idxOutPath)
	if err != nil {
		return fmt.Errorf("failed to create index.html: %w", err)
	}
	defer func() { _ = idxF.Close() }()

	if tmplErr := indexTmpl.ExecuteTemplate(idxF, "layout.html", indexData); tmplErr != nil {
		return fmt.Errorf("failed to execute template for index.html: %w", tmplErr)
	}

	return nil
}

func buildSidebar(fromOutRelPath string, currentID string, config graph.Manifest, guideFiles []*guideFile, filesByID map[string]*guideFile) SidebarData {
	homeURL := relURL(fromOutRelPath, "index.html")

	// 1. Build Tours
	var sidebarTours []SidebarTour
	for _, tour := range config.Tours {
		var tourGuides []SidebarLink
		for _, gID := range tour.Guides {
			gID = strings.TrimSpace(gID)
			if gID == "" {
				continue
			}
			target, exists := filesByID[gID]
			link := SidebarLink{
				Title:    gID,
				IsActive: gID == currentID,
			}
			if exists {
				link.URL = relURL(fromOutRelPath, target.outRelPath)
				link.Scope = target.guide.Metadata.Scope
			}
			tourGuides = append(tourGuides, link)
		}
		if len(tourGuides) > 0 {
			sidebarTours = append(sidebarTours, SidebarTour{
				Name:   tour.Name,
				Guides: tourGuides,
			})
		}
	}

	// 2. Build Scope Groups
	guidesByScope := make(map[string][]*guideFile)
	for _, gf := range guideFiles {
		sc := gf.guide.Metadata.Scope
		guidesByScope[sc] = append(guidesByScope[sc], gf)
	}

	var scopeGroups []SidebarScopeGroup
	seenScopes := make(map[string]bool)

	for _, sc := range config.Scopes {
		gfs := guidesByScope[sc]
		if len(gfs) == 0 {
			continue
		}
		seenScopes[sc] = true

		sortedGfs := append([]*guideFile(nil), gfs...)
		sort.Slice(sortedGfs, func(i, j int) bool {
			return sortedGfs[i].guide.ID < sortedGfs[j].guide.ID
		})

		var links []SidebarLink
		for _, gf := range sortedGfs {
			links = append(links, SidebarLink{
				Title:    gf.guide.ID,
				URL:      relURL(fromOutRelPath, gf.outRelPath),
				IsActive: gf.guide.ID == currentID,
				Scope:    sc,
			})
		}
		scopeGroups = append(scopeGroups, SidebarScopeGroup{
			Scope:  sc,
			Guides: links,
		})
	}

	var otherScopes []string
	for sc := range guidesByScope {
		if !seenScopes[sc] {
			otherScopes = append(otherScopes, sc)
		}
	}
	sort.Strings(otherScopes)
	for _, sc := range otherScopes {
		gfs := guidesByScope[sc]
		if len(gfs) == 0 {
			continue
		}
		sortedGfs := append([]*guideFile(nil), gfs...)
		sort.Slice(sortedGfs, func(i, j int) bool {
			return sortedGfs[i].guide.ID < sortedGfs[j].guide.ID
		})

		var links []SidebarLink
		for _, gf := range sortedGfs {
			links = append(links, SidebarLink{
				Title:    gf.guide.ID,
				URL:      relURL(fromOutRelPath, gf.outRelPath),
				IsActive: gf.guide.ID == currentID,
				Scope:    sc,
			})
		}
		groupName := sc
		if groupName == "" {
			groupName = "other"
		}
		scopeGroups = append(scopeGroups, SidebarScopeGroup{
			Scope:  groupName,
			Guides: links,
		})
	}

	return SidebarData{
		HomeURL:     homeURL,
		IsHome:      currentID == "",
		Tours:       sidebarTours,
		ScopeGroups: scopeGroups,
	}
}

func buildTourSteps(currentGf *guideFile, config graph.Manifest, filesByID map[string]*guideFile) []TourStepRef {
	var steps []TourStepRef
	for _, tour := range config.Tours {
		var tourGuideIDs []string
		for _, gID := range tour.Guides {
			gID = strings.TrimSpace(gID)
			if gID != "" {
				tourGuideIDs = append(tourGuideIDs, gID)
			}
		}

		currentIndex := -1
		for i, gID := range tourGuideIDs {
			if gID == currentGf.guide.ID {
				currentIndex = i
				break
			}
		}

		if currentIndex == -1 {
			continue
		}

		stepRef := TourStepRef{
			TourName: tour.Name,
			Step:     currentIndex + 1,
			Total:    len(tourGuideIDs),
		}

		if currentIndex > 0 {
			prevID := tourGuideIDs[currentIndex-1]
			ref := GuideRef{
				ID:    prevID,
				Title: prevID,
			}
			if target, ok := filesByID[prevID]; ok {
				ref.URL = relURL(currentGf.outRelPath, target.outRelPath)
				ref.Scope = target.guide.Metadata.Scope
			}
			stepRef.Prev = &ref
		}

		if currentIndex < len(tourGuideIDs)-1 {
			nextID := tourGuideIDs[currentIndex+1]
			ref := GuideRef{
				ID:    nextID,
				Title: nextID,
			}
			if target, ok := filesByID[nextID]; ok {
				ref.URL = relURL(currentGf.outRelPath, target.outRelPath)
				ref.Scope = target.guide.Metadata.Scope
			}
			stepRef.Next = &ref
		}

		steps = append(steps, stepRef)
	}
	return steps
}
