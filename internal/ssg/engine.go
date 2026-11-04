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
	"github.com/yuin/goldmark/parser"
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
	Title           string
	Scope           string
	Tags            []string
	Content         template.HTML
	Prerequisites   []GuideRef
	Dependents      []GuideRef
	SubGuides       []GuideRef
	ParentGuides    []GuideRef
	TourSteps       []TourStepRef
	RelativeRoot    string
	Sidebar         SidebarData
	MermaidGraph    template.HTML
	HorizonPath     []PathStepRef
	FullTrajectory  []PathStepRef
	HasDistantSteps bool
	LearningPath    []PathStepRef
	IsFoundational  bool
	SubsequentPath  []PathStepRef
	IsTerminal      bool
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
	DirPath    string
	OutDir     string
	Offline    bool
	VendorDir  string
	CacheDir   string
	SkipVendor bool
	WarnFunc   func(format string, args ...any)
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
	if !opts.SkipVendor {
		vendorOpts := VendorOptions{
			Offline:   opts.Offline,
			VendorDir: opts.VendorDir,
			CacheDir:  opts.CacheDir,
			WarnFunc:  opts.WarnFunc,
		}
		if err := EnsureVendorAssets(outDir, vendorOpts); err != nil {
			return fmt.Errorf("failed to prepare vendor assets: %w", err)
		}
	} else {
		if err := os.MkdirAll(filepath.Join(outDir, "assets"), 0o755); err != nil {
			return fmt.Errorf("failed to create assets dir: %w", err)
		}
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
			if path != dirPath && graph.IsIgnoredDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasPrefix(info.Name(), ".") {
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

			g, err := graph.ParseGuidePermissive(path)
			if err != nil {
				return err
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
			return nil
		}

		if graph.IsStaticAssetExt(filepath.Ext(info.Name())) {
			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}
			return copyFile(path, filepath.Join(outDir, relPath))
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

		pc := parser.NewContext()
		var segmentList []SegmentInfo
		for _, s := range gf.guide.Metadata.SubGuides {
			if s.Segment == "" {
				continue
			}
			ranges, err := graph.ParseSegments(s.Segment)
			if err != nil {
				continue
			}
			subTitle := s.Guide
			subURL := ""
			if target, ok := filesByID[s.Guide]; ok {
				subTitle = target.guide.ID
				subURL = relURL(gf.outRelPath, target.outRelPath)
			}
			for _, r := range ranges {
				rangeStr := fmt.Sprintf("lines %d–%d", r[0], r[1])
				if r[0] == r[1] {
					rangeStr = fmt.Sprintf("line %d", r[0])
				}
				segmentList = append(segmentList, SegmentInfo{
					GuideID:   s.Guide,
					Title:     subTitle,
					URL:       subURL,
					Adherence: s.Adherence,
					Segment:   s.Segment,
					RangeStr:  rangeStr,
					StartLine: r[0],
					EndLine:   r[1],
				})
			}
		}
		sort.Slice(segmentList, func(i, j int) bool {
			return segmentList[i].StartLine < segmentList[j].StartLine
		})
		pc.Set(SubGuidesContextKey, segmentList)

		var buf bytes.Buffer
		if convErr := md.Convert([]byte(gf.guide.Body), &buf, parser.WithContext(pc)); convErr != nil {
			return convErr
		}

		if mkdirErr := os.MkdirAll(filepath.Dir(gf.outPath), 0o755); mkdirErr != nil {
			return mkdirErr
		}

		sidebar := buildSidebar(gf.outRelPath, gf.guide.ID, config, guideFiles, filesByID)
		tourSteps := buildTourSteps(gf, config, filesByID)

		guideMermaid := generateGuideMermaid(gf, config, filesByID, graphIndex)

		var learningPath []PathStepRef
		var upstreamRefs []PathStepRef
		isFoundational := false

		// 1. Calculate Upstream Hop Distances (BFS backwards)
		visitedUp := map[string]int{gf.guide.ID: 0}
		qUp := []string{gf.guide.ID}
		for len(qUp) > 0 {
			curr := qUp[0]
			qUp = qUp[1:]
			depth := visitedUp[curr]
			currG, ok := guidesMap[curr]
			if !ok {
				continue
			}
			var deps []string
			deps = append(deps, currG.Metadata.Prerequisites...)
			for _, sub := range currG.Metadata.SubGuides {
				deps = append(deps, sub.Guide)
			}
			for _, dep := range deps {
				if _, seen := visitedUp[dep]; !seen {
					visitedUp[dep] = depth + 1
					qUp = append(qUp, dep)
				}
			}
		}

		pathResult, err := graph.PrerequisitePath(gf.guide.ID, guidesMap, config, graph.PathOptions{
			IncludeSubguides: true,
		})
		if err != nil {
			if opts.WarnFunc != nil {
				opts.WarnFunc("warning: skipping prerequisite path for %s: %v\n", gf.guide.ID, err)
			}
		} else if pathResult != nil {
			if len(pathResult.Steps) <= 1 {
				isFoundational = true
			} else {
				subGuidesOfCurrent := make(map[string]bool)
				for _, sub := range gf.guide.Metadata.SubGuides {
					subGuidesOfCurrent[sub.Guide] = true
				}

				for _, step := range pathResult.Steps {
					isCurr := step.GuideID == gf.guide.ID
					hop := 0
					offset := template.HTML("Current")
					relation := "Current"
					if !isCurr {
						depth := visitedUp[step.GuideID]
						if depth == 0 {
							depth = 1
						}
						hop = -depth
						offset = template.HTML(fmt.Sprintf("-%d", depth))
						if depth == 1 {
							if subGuidesOfCurrent[step.GuideID] {
								relation = "Sub-guide"
							} else {
								relation = "Prerequisite"
							}
						} else {
							relation = "Prerequisite"
						}
					}
					ref := PathStepRef{
						Step:      step.Index,
						Hop:       hop,
						Offset:    offset,
						Relation:  relation,
						Scope:     step.Scope,
						Title:     step.GuideID,
						IsCurrent: isCurr,
					}
					if target, ok := filesByID[step.GuideID]; ok {
						ref.URL = relURL(gf.outRelPath, target.outRelPath)
					}
					learningPath = append(learningPath, ref)
					if !isCurr {
						upstreamRefs = append(upstreamRefs, ref)
					}
				}
			}
		}

		// 2. Calculate Downstream Hop Distances (BFS forwards)
		visitedDown := map[string]int{gf.guide.ID: 0}
		qDown := []string{gf.guide.ID}
		for len(qDown) > 0 {
			curr := qDown[0]
			qDown = qDown[1:]
			depth := visitedDown[curr]
			var downs []string
			downs = append(downs, graphIndex.GetDependents(curr)...)
			downs = append(downs, graphIndex.GetParentGuides(curr)...)
			for _, down := range downs {
				if _, seen := visitedDown[down]; !seen {
					visitedDown[down] = depth + 1
					qDown = append(qDown, down)
				}
			}
		}

		var subsequentPath []PathStepRef
		var downstreamRefs []PathStepRef
		isTerminal := false

		subsequentResult, err := graph.SubsequentPath(gf.guide.ID, guidesMap, config, graph.PathOptions{
			IncludeSubguides: true,
		})
		if err != nil {
			if opts.WarnFunc != nil {
				opts.WarnFunc("warning: skipping subsequent path for %s: %v\n", gf.guide.ID, err)
			}
		} else if subsequentResult != nil {
			if len(subsequentResult.Steps) <= 1 {
				isTerminal = true
			} else {
				parentsOfCurrent := make(map[string]bool)
				for _, p := range graphIndex.GetParentGuides(gf.guide.ID) {
					parentsOfCurrent[p] = true
				}

				for _, step := range subsequentResult.Steps {
					isCurr := step.GuideID == gf.guide.ID
					hop := 0
					offset := template.HTML("Current")
					relation := "Current"
					if !isCurr {
						depth := visitedDown[step.GuideID]
						if depth == 0 {
							depth = 1
						}
						hop = depth
						offset = template.HTML(fmt.Sprintf("+%d", depth))
						if depth == 1 {
							if parentsOfCurrent[step.GuideID] {
								relation = "Parent"
							} else {
								relation = "Subsequent"
							}
						} else {
							relation = "Subsequent"
						}
					}
					ref := PathStepRef{
						Step:      step.Index,
						Hop:       hop,
						Offset:    offset,
						Relation:  relation,
						Scope:     step.Scope,
						Title:     step.GuideID,
						IsCurrent: isCurr,
					}
					if target, ok := filesByID[step.GuideID]; ok {
						ref.URL = relURL(gf.outRelPath, target.outRelPath)
					}
					subsequentPath = append(subsequentPath, ref)
					if !isCurr {
						downstreamRefs = append(downstreamRefs, ref)
					}
				}
			}
		}

		// 3. Assemble Unified Trajectory and Horizon Paths
		var horizonPath []PathStepRef
		var fullTrajectory []PathStepRef
		hasDistantSteps := false

		if len(upstreamRefs) > 0 || len(downstreamRefs) > 0 {
			currentRef := PathStepRef{
				Offset:    template.HTML("Current"),
				Hop:       0,
				Relation:  "Current",
				Scope:     gf.guide.Metadata.Scope,
				Title:     gf.guide.ID,
				IsCurrent: true,
			}

			// Full Trajectory: Upstream -> Current -> Downstream
			fullTrajectory = append(fullTrajectory, upstreamRefs...)
			fullTrajectory = append(fullTrajectory, currentRef)
			fullTrajectory = append(fullTrajectory, downstreamRefs...)

			// Horizon: Direct Prereqs (Hop == -1) -> Current -> Direct Subsequent (Hop == 1)
			for _, ref := range upstreamRefs {
				if ref.Hop == -1 {
					horizonPath = append(horizonPath, ref)
				}
			}
			horizonPath = append(horizonPath, currentRef)
			for _, ref := range downstreamRefs {
				if ref.Hop == 1 {
					horizonPath = append(horizonPath, ref)
				}
			}

			hasDistantSteps = len(fullTrajectory) > len(horizonPath)
		}

		data := PageData{
			Title:           gf.guide.ID,
			Scope:           gf.guide.Metadata.Scope,
			Tags:            gf.guide.Metadata.Tags,
			Content:         template.HTML(buf.String()),
			Prerequisites:   prereqs,
			Dependents:      dependents,
			SubGuides:       subGuides,
			ParentGuides:    parentGuides,
			TourSteps:       tourSteps,
			RelativeRoot:    gf.relRoot,
			Sidebar:         sidebar,
			MermaidGraph:    template.HTML(guideMermaid),
			HorizonPath:     horizonPath,
			FullTrajectory:  fullTrajectory,
			HasDistantSteps: hasDistantSteps,
			LearningPath:    learningPath,
			IsFoundational:  isFoundational,
			SubsequentPath:  subsequentPath,
			IsTerminal:      isTerminal,
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
