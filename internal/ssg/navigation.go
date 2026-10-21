package ssg

import (
	"sort"
	"strings"

	"github.com/Carter907/qst/internal/graph"
)

// SidebarLink represents an individual guide link within the sidebar.
type SidebarLink struct {
	Title    string
	URL      string
	IsActive bool
	Scope    string
}

// SidebarScopeGroup groups guides belonging to a single scope in the sidebar.
type SidebarScopeGroup struct {
	Scope  string
	Guides []SidebarLink
}

// SidebarTour groups guides sequentially in a tour for sidebar navigation.
type SidebarTour struct {
	Name   string
	Guides []SidebarLink
}

// SidebarData bundles all navigation information required by layout.html.
type SidebarData struct {
	HomeURL     string
	IsHome      bool
	Tours       []SidebarTour
	ScopeGroups []SidebarScopeGroup
}

// TourStepRef provides contextual previous/next links and index within a tour.
type TourStepRef struct {
	TourName string
	Step     int
	Total    int
	Prev     *GuideRef
	Next     *GuideRef
}

// SyllabusGuideRef describes a guide entry on the academic syllabus index page.
type SyllabusGuideRef struct {
	Title         string
	URL           string
	Prerequisites []GuideRef
	Tags          []string
}

// SyllabusScopeSection represents an academic syllabus tier grouped by scope.
type SyllabusScopeSection struct {
	Scope  string
	Guides []SyllabusGuideRef
}

func buildSyllabusScopes(config graph.Manifest, guideFiles []*guideFile, filesByID map[string]*guideFile) []SyllabusScopeSection {
	guidesByScope := make(map[string][]*guideFile)
	for _, gf := range guideFiles {
		sc := gf.guide.Metadata.Scope
		guidesByScope[sc] = append(guidesByScope[sc], gf)
	}

	var orderedScopes []string
	seenScopes := make(map[string]bool)

	// Top-down: Broadest (highest hierarchy index) down to foundational (index 0)
	for i := len(config.Scopes) - 1; i >= 0; i-- {
		sc := config.Scopes[i]
		if len(guidesByScope[sc]) > 0 && !seenScopes[sc] {
			orderedScopes = append(orderedScopes, sc)
			seenScopes[sc] = true
		}
	}

	var otherScopes []string
	for sc := range guidesByScope {
		if !seenScopes[sc] {
			otherScopes = append(otherScopes, sc)
		}
	}
	sort.Strings(otherScopes)
	orderedScopes = append(orderedScopes, otherScopes...)

	var sections []SyllabusScopeSection
	for _, sc := range orderedScopes {
		gfs := guidesByScope[sc]
		sort.Slice(gfs, func(i, j int) bool {
			return gfs[i].guide.ID < gfs[j].guide.ID
		})

		var guides []SyllabusGuideRef
		for _, gf := range gfs {
			var prereqs []GuideRef
			for _, p := range gf.guide.Metadata.Prerequisites {
				prereqGF := filesByID[p]
				if prereqGF != nil {
					prereqs = append(prereqs, GuideRef{
						ID:    p,
						Title: prereqGF.guide.ID,
						URL:   relURL("index.html", prereqGF.outRelPath),
						Scope: prereqGF.guide.Metadata.Scope,
					})
				} else {
					prereqs = append(prereqs, GuideRef{
						ID:    p,
						Title: p,
						URL:   "",
					})
				}
			}

			tags := append([]string(nil), gf.guide.Metadata.Tags...)

			guides = append(guides, SyllabusGuideRef{
				Title:         gf.guide.ID,
				URL:           relURL("index.html", gf.outRelPath),
				Prerequisites: prereqs,
				Tags:          tags,
			})
		}

		scopeName := sc
		if scopeName == "" {
			scopeName = "other"
		}

		sections = append(sections, SyllabusScopeSection{
			Scope:  scopeName,
			Guides: guides,
		})
	}

	return sections
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
