package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// ParseGraph reads all markdown files in the specified directory,
// extracts their frontmatter, and returns a map of guide IDs to Guide structs.
func ParseGraph(dirPath string) (map[string]Guide, error) {
	guides := make(map[string]Guide)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {

		// ignore all other files
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		path := filepath.Join(dirPath, entry.Name())
		guide, err := ParseGuide(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", path, err)
		}

		guides[guide.ID] = guide
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

	return Guide{
		ID:         id,
		Path:       path,
		Metadata:   meta,
		LineCount:  len(lines) - (frontMatterEnd + 1),
		HasContent: hasContent,
	}, nil
}

// ParseConfig reads the manifest.yaml configuration file in the specified directory.
func ParseConfig(dirPath string) (Manifest, error) {
	configPath := filepath.Join(dirPath, "manifest.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("failed to read manifest.yaml: %w", err)
	}

	var config Manifest
	if err := yaml.Unmarshal(content, &config); err != nil {
		return Manifest{}, fmt.Errorf("failed to parse manifest.yaml: %w", err)
	}

	return config, nil
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
