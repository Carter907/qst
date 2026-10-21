package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

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

// RenameGuide renames a guide and cascades updates across all incoming prerequisites,
// subguides, and manifest tours. It also renames the guide markdown file on disk.
// It returns the total number of references updated, or an error.
func RenameGuide(dirPath string, oldID string, newID string) (int, error) {
	guides, err := ParseGraph(dirPath)
	if err != nil {
		return 0, fmt.Errorf("error parsing graph: %w", err)
	}

	targetGuide, exists := guides[oldID]
	if !exists {
		return 0, fmt.Errorf("guide '%s' does not exist in graph", oldID)
	}
	if _, exists := guides[newID]; exists {
		return 0, fmt.Errorf("guide '%s' already exists in graph", newID)
	}

	updatedCount := 0
	for _, guide := range guides {
		needsUpdate := false
		meta := guide.Metadata

		for i, p := range meta.Prerequisites {
			if p == oldID {
				meta.Prerequisites[i] = newID
				needsUpdate = true
			}
		}

		for i, s := range meta.SubGuides {
			if s.Guide == oldID {
				meta.SubGuides[i].Guide = newID
				needsUpdate = true
			}
		}

		if needsUpdate {
			if err := UpdateGuideMetadata(guide.Path, meta); err != nil {
				return updatedCount, fmt.Errorf("failed to update guide '%s': %w", guide.ID, err)
			}
			updatedCount++
		}
	}

	// Update references in manifest.yaml tours
	config, err := ParseConfig(dirPath)
	if err == nil {
		configUpdated := false
		for i, tour := range config.Tours {
			for j, g := range tour.Guides {
				if g == oldID {
					config.Tours[i].Guides[j] = newID
					configUpdated = true
					updatedCount++
				}
			}
		}
		if configUpdated {
			if err := UpdateConfig(dirPath, config); err != nil {
				return updatedCount, fmt.Errorf("failed to update manifest.yaml: %w", err)
			}
		}
	}

	// Rename file on disk
	newPath := filepath.Join(filepath.Dir(targetGuide.Path), newID+".md")
	if err := os.Rename(targetGuide.Path, newPath); err != nil {
		return updatedCount, fmt.Errorf("failed to rename file: %w", err)
	}

	return updatedCount, nil
}
