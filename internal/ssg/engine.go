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
)

//go:embed templates/*
var templatesFS embed.FS

// GuideRef represents a linked reference to another guide with metadata.
type GuideRef struct {
	ID        string
	Title     string
	URL       string
	Scope     string
	Adherence string
	Segment   string
}

// PageData contains the rendering context for a guide's page.
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
	MermaidGraph  template.HTML
}

// IndexPageData contains the rendering context for index.html.
type IndexPageData struct {
	Title        string
	Description  string
	RelativeRoot string
	Sidebar      SidebarData
	Scopes       []SyllabusScopeSection
}

type guideFile struct {
	guide      graph.Guide
	path       string
	outPath    string
	outRelPath string
	relRoot    string
}

func relURL(fromOutRelPath, toOutRelPath string) string {
	fromDir := filepath.Dir(fromOutRelPath)
	rel, err := filepath.Rel(fromDir, toOutRelPath)
	if err != nil {
		return toOutRelPath
	}
	return filepath.ToSlash(rel)
}

// BuildOptions defines configuration options for compiling a static site.
type BuildOptions struct {
	DirPath   string
	OutDir    string
	Offline   bool
	VendorDir string
	CacheDir  string
	WarnFunc  func(format string, args ...any)
}

// BuildSite compiles the static site from dirPath into outDir with default options.
func BuildSite(dirPath string, outDir string) error {
	return BuildSiteWithOptions(BuildOptions{
		DirPath: dirPath,
		OutDir:  outDir,
	})
}

// BuildSiteWithOptions compiles the static site using the provided options.
func BuildSiteWithOptions(opts BuildOptions) error {
	dirPath := opts.DirPath
	outDir := opts.OutDir

	config, err := graph.ParseConfig(dirPath)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Ensure outDir exists
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create outDir: %w", err)
	}

	// 1. Prepare vendor assets in outDir/assets/vendor
	vendorOpts := VendorOptions{
		Offline:   opts.Offline,
		VendorDir: opts.VendorDir,
		CacheDir:  opts.CacheDir,
		WarnFunc:  opts.WarnFunc,
	}
	if err := EnsureVendorAssets(outDir, vendorOpts); err != nil {
		return fmt.Errorf("failed to prepare vendor assets: %w", err)
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
	md := newMarkdownParser()

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

	// 6. Build Inverted Reference Index
	guidesMap := make(map[string]graph.Guide, len(guideFiles))
	for _, gf := range guideFiles {
		guidesMap[gf.guide.ID] = gf.guide
	}
	graphIndex := graph.BuildIndex(guidesMap, config)

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
		depIDs := append([]string(nil), graphIndex.GetDependents(gf.guide.ID)...)
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
		parentIDs := append([]string(nil), graphIndex.GetParentGuides(gf.guide.ID)...)
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

		guideMermaid := generateGuideMermaid(gf, config, filesByID)

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
			MermaidGraph:  template.HTML(guideMermaid),
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
	syllabusScopes := buildSyllabusScopes(config, guideFiles, filesByID)
	indexData := IndexPageData{
		Title:        config.Title,
		Description:  config.Description,
		RelativeRoot: "",
		Sidebar:      indexSidebar,
		Scopes:       syllabusScopes,
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
