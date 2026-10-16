package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// ParseGraph reads all markdown files in the specified directory recursively,
// extracts their frontmatter, and returns a map of guide IDs to Guide structs.
func ParseGraph(dirPath string) (map[string]Guide, error) {
	cleanDir := filepath.Clean(dirPath)
	if _, err := os.Stat(cleanDir); err != nil {
		return nil, err
	}

	guides := make(map[string]Guide)

	err := filepath.Walk(cleanDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() {
			if path == cleanDir {
				return nil
			}
			base := info.Name()
			// Skip hidden directories (e.g. .git) and build output (public)
			if strings.HasPrefix(base, ".") || base == "public" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only parse markdown files
		if filepath.Ext(path) != ".md" {
			return nil
		}

		// Skip hidden markdown files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		guide, err := ParseGuide(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		if existing, exists := guides[guide.ID]; exists {
			return fmt.Errorf("duplicate guide ID '%s': found in %s and %s", guide.ID, existing.Path, path)
		}

		guides[guide.ID] = guide
		return nil
	})

	if err != nil {
		return nil, err
	}

	return guides, nil
}

func ParseGuide(path string) (Guide, error) {
	baseName := filepath.Base(path)
	id := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	content, err := os.ReadFile(path)
	if err != nil {
		return Guide{}, err
	}

	contentStr := strings.ReplaceAll(string(content), "\r\n", "\n")
	lines := strings.Split(contentStr, "\n")
	var frontmatterLines []string
	inFrontmatter := false
	frontMatterEnd := 0

	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				frontMatterEnd = i
				break
			}
		}
		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		}
	}

	if len(frontmatterLines) == 0 {
		return Guide{}, fmt.Errorf("invalid or missing YAML frontmatter")
	}

	frontmatter := strings.Join(frontmatterLines, "\n")

	var meta GuideMetadata
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return Guide{}, fmt.Errorf("failed to parse yaml: %w", err)
	}

	hasContent := false
	for i := frontMatterEnd + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			hasContent = true
			break
		}
	}

	body := ""
	if frontMatterEnd+1 < len(lines) {
		body = strings.Join(lines[frontMatterEnd+1:], "\n")
	}

	return Guide{
		ID:         id,
		Path:       path,
		Metadata:   meta,
		LineCount:  len(lines) - (frontMatterEnd + 1),
		HasContent: hasContent,
		Body:       body,
	}, nil
}

// ParseConfig reads the manifest.yaml (or quest.yaml fallback) configuration file in the specified directory.
func ParseConfig(dirPath string) (Manifest, error) {
	configPath := filepath.Join(dirPath, "manifest.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		altPath := filepath.Join(dirPath, "quest.yaml")
		altContent, altErr := os.ReadFile(altPath)
		if altErr != nil {
			return Manifest{}, fmt.Errorf("failed to read manifest.yaml: %w", err)
		}
		content = altContent
	}

	var config Manifest
	if err := yaml.Unmarshal(content, &config); err != nil {
		return Manifest{}, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return config, nil
}

// UpdateConfig writes the manifest.yaml configuration file back to the specified directory.
func UpdateConfig(dirPath string, config Manifest) error {
	configPath := filepath.Join(dirPath, "manifest.yaml")
	content, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest.yaml: %w", err)
	}

	return os.WriteFile(configPath, content, 0644)
}

// UpdateGuideMetadata replaces the YAML frontmatter in the given file while preserving the rest of its content.
func UpdateGuideMetadata(path string, meta GuideMetadata) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	newline := "\n"
	if strings.Contains(string(content), "\r\n") {
		newline = "\r\n"
	}

	contentStr := strings.ReplaceAll(string(content), "\r\n", "\n")
	lines := strings.Split(contentStr, "\n")
	inFrontmatter := false
	frontMatterStart := -1
	frontMatterEnd := -1

	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				frontMatterStart = i
			} else {
				frontMatterEnd = i
				break
			}
		}
	}

	if frontMatterStart == -1 || frontMatterEnd == -1 {
		return fmt.Errorf("invalid or missing YAML frontmatter in %s", path)
	}

	// Ensure nil slices serialize to empty arrays [] in yaml instead of null
	if meta.Prerequisites == nil {
		meta.Prerequisites = []string{}
	}
	if meta.SubGuides == nil {
		meta.SubGuides = []SubGuideRelation{}
	}
	if meta.Tags == nil {
		meta.Tags = []string{}
	}

	metaBytes, err := yaml.Marshal(&meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	var newLines []string
	newLines = append(newLines, lines[:frontMatterStart+1]...)

	metaStr := strings.TrimRight(string(metaBytes), "\n")
	if metaStr != "" {
		newLines = append(newLines, strings.Split(metaStr, "\n")...)
	}

	newLines = append(newLines, lines[frontMatterEnd:]...)

	return os.WriteFile(path, []byte(strings.Join(newLines, newline)), 0644)
}

// AddPrerequisite adds a prerequisite to the guide at the given path if it's not already present.
func AddPrerequisite(path string, prereq string) error {
	guide, err := ParseGuide(path)
	if err != nil {
		return err
	}
	for _, p := range guide.Metadata.Prerequisites {
		if p == prereq {
			return nil
		}
	}
	guide.Metadata.Prerequisites = append(guide.Metadata.Prerequisites, prereq)
	return UpdateGuideMetadata(path, guide.Metadata)
}

// RemovePrerequisite removes a prerequisite from the guide at the given path.
func RemovePrerequisite(path string, prereq string) error {
	guide, err := ParseGuide(path)
	if err != nil {
		return err
	}
	var updated []string
	found := false
	for _, p := range guide.Metadata.Prerequisites {
		if p == prereq {
			found = true
			continue
		}
		updated = append(updated, p)
	}
	if !found {
		return fmt.Errorf("prerequisite '%s' not found in %s", prereq, guide.ID)
	}
	guide.Metadata.Prerequisites = updated
	return UpdateGuideMetadata(path, guide.Metadata)
}

// UpsertSubGuide adds or updates a subguide relation in the guide at the given path.
// If the relation already exists for relation.Guide, its adherence and segment are updated in place.
func UpsertSubGuide(path string, relation SubGuideRelation) error {
	guide, err := ParseGuide(path)
	if err != nil {
		return err
	}
	found := false
	for i, s := range guide.Metadata.SubGuides {
		if s.Guide == relation.Guide {
			guide.Metadata.SubGuides[i] = relation
			found = true
			break
		}
	}
	if !found {
		guide.Metadata.SubGuides = append(guide.Metadata.SubGuides, relation)
	}
	return UpdateGuideMetadata(path, guide.Metadata)
}

// RemoveSubGuide removes a subguide relation from the guide at the given path.
func RemoveSubGuide(path string, subGuide string) error {
	guide, err := ParseGuide(path)
	if err != nil {
		return err
	}
	var updated []SubGuideRelation
	found := false
	for _, s := range guide.Metadata.SubGuides {
		if s.Guide == subGuide {
			found = true
			continue
		}
		updated = append(updated, s)
	}
	if !found {
		return fmt.Errorf("subguide '%s' not found in %s", subGuide, guide.ID)
	}
	guide.Metadata.SubGuides = updated
	return UpdateGuideMetadata(path, guide.Metadata)
}

// AddTag adds a tag to the guide at the given path if not already present.
func AddTag(path string, tag string) error {
	guide, err := ParseGuide(path)
	if err != nil {
		return err
	}
	if slices.Contains(guide.Metadata.Tags, tag) {
		return nil
	}
	guide.Metadata.Tags = append(guide.Metadata.Tags, tag)
	return UpdateGuideMetadata(path, guide.Metadata)
}

// RemoveTag removes a tag from the guide at the given path.
func RemoveTag(path string, tag string) error {
	guide, err := ParseGuide(path)
	if err != nil {
		return err
	}
	var updated []string
	found := false
	for _, t := range guide.Metadata.Tags {
		if t == tag {
			found = true
			continue
		}
		updated = append(updated, t)
	}
	if !found {
		return fmt.Errorf("tag '%s' not found in %s", tag, guide.ID)
	}
	guide.Metadata.Tags = updated
	return UpdateGuideMetadata(path, guide.Metadata)
}

// CreateTour adds a new tour to manifest.yaml.
func CreateTour(dirPath string, tour TourConfig) error {
	config, err := ParseConfig(dirPath)
	if err != nil {
		return err
	}
	for _, t := range config.Tours {
		if t.Name == tour.Name {
			return fmt.Errorf("tour '%s' already exists", tour.Name)
		}
	}
	if tour.Guides == nil {
		tour.Guides = []string{}
	}
	config.Tours = append(config.Tours, tour)
	return UpdateConfig(dirPath, config)
}

// DeleteTour removes a tour from manifest.yaml.
func DeleteTour(dirPath string, tourName string) error {
	config, err := ParseConfig(dirPath)
	if err != nil {
		return err
	}
	var updated []TourConfig
	found := false
	for _, t := range config.Tours {
		if t.Name == tourName {
			found = true
			continue
		}
		updated = append(updated, t)
	}
	if !found {
		return fmt.Errorf("tour '%s' not found", tourName)
	}
	config.Tours = updated
	return UpdateConfig(dirPath, config)
}

// AddGuideToTour appends a guide to a tour in manifest.yaml if not already present.
func AddGuideToTour(dirPath string, tourName string, guideID string) error {
	config, err := ParseConfig(dirPath)
	if err != nil {
		return err
	}
	tourIdx := -1
	for i, t := range config.Tours {
		if t.Name == tourName {
			tourIdx = i
			break
		}
	}
	if tourIdx == -1 {
		return fmt.Errorf("tour '%s' not found", tourName)
	}
	if slices.Contains(config.Tours[tourIdx].Guides, guideID) {
		return nil
	}
	config.Tours[tourIdx].Guides = append(config.Tours[tourIdx].Guides, guideID)
	return UpdateConfig(dirPath, config)
}

// RemoveGuideFromTour removes a guide from a tour in manifest.yaml.
func RemoveGuideFromTour(dirPath string, tourName string, guideID string) error {
	config, err := ParseConfig(dirPath)
	if err != nil {
		return err
	}
	tourIdx := -1
	for i, t := range config.Tours {
		if t.Name == tourName {
			tourIdx = i
			break
		}
	}
	if tourIdx == -1 {
		return fmt.Errorf("tour '%s' not found", tourName)
	}
	var updated []string
	found := false
	for _, g := range config.Tours[tourIdx].Guides {
		if g == guideID {
			found = true
			continue
		}
		updated = append(updated, g)
	}
	if !found {
		return fmt.Errorf("guide '%s' not found in tour '%s'", guideID, tourName)
	}
	config.Tours[tourIdx].Guides = updated
	return UpdateConfig(dirPath, config)
}

// DeleteGuide removes a guide and prunes all incoming references across prerequisites, subguides, and manifest tours.
// It returns the total number of references pruned from other guides and tours.
func DeleteGuide(dirPath string, guideID string) (int, error) {
	guides, err := ParseGraph(dirPath)
	if err != nil {
		return 0, fmt.Errorf("error parsing graph: %w", err)
	}

	targetGuide, exists := guides[guideID]
	if !exists {
		return 0, fmt.Errorf("guide '%s' does not exist in graph", guideID)
	}

	prunedCount := 0

	for id, g := range guides {
		if id == guideID {
			continue
		}
		meta := g.Metadata
		needsUpdate := false

		// Prune from prerequisites
		var updatedPrereqs []string
		for _, p := range meta.Prerequisites {
			if p == guideID {
				needsUpdate = true
				prunedCount++
			} else {
				updatedPrereqs = append(updatedPrereqs, p)
			}
		}
		if needsUpdate {
			meta.Prerequisites = updatedPrereqs
		}

		// Prune from subguides
		var updatedSubs []SubGuideRelation
		subUpdated := false
		for _, s := range meta.SubGuides {
			if s.Guide == guideID {
				subUpdated = true
				prunedCount++
			} else {
				updatedSubs = append(updatedSubs, s)
			}
		}
		if subUpdated {
			meta.SubGuides = updatedSubs
			needsUpdate = true
		}

		if needsUpdate {
			if err := UpdateGuideMetadata(g.Path, meta); err != nil {
				return prunedCount, fmt.Errorf("failed to update guide '%s': %w", id, err)
			}
		}
	}

	// Prune from tours in manifest.yaml
	config, err := ParseConfig(dirPath)
	if err == nil {
		configUpdated := false
		for i, tour := range config.Tours {
			var updatedTourGuides []string
			for _, g := range tour.Guides {
				if g == guideID {
					configUpdated = true
					prunedCount++
				} else {
					updatedTourGuides = append(updatedTourGuides, g)
				}
			}
			config.Tours[i].Guides = updatedTourGuides
		}
		if configUpdated {
			if err := UpdateConfig(dirPath, config); err != nil {
				return prunedCount, fmt.Errorf("failed to update manifest.yaml: %w", err)
			}
		}
	}

	// Remove file from disk
	if err := os.Remove(targetGuide.Path); err != nil {
		return prunedCount, fmt.Errorf("failed to remove guide file: %w", err)
	}

	return prunedCount, nil
}

