package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type frontmatterParts struct {
	lines            []string
	frontmatter      string
	body             string
	newline          string
	frontMatterStart int
	frontMatterEnd   int
}

func extractFrontmatter(content []byte) (*frontmatterParts, error) {
	newline := "\n"
	if strings.Contains(string(content), "\r\n") {
		newline = "\r\n"
	}

	contentStr := strings.ReplaceAll(string(content), "\r\n", "\n")
	lines := strings.Split(contentStr, "\n")
	var frontmatterLines []string
	inFrontmatter := false
	frontMatterStart := -1
	frontMatterEnd := -1

	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				frontMatterStart = i
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

	if frontMatterStart == -1 || frontMatterEnd == -1 || len(frontmatterLines) == 0 {
		return nil, fmt.Errorf("invalid or missing YAML frontmatter")
	}

	body := ""
	if frontMatterEnd+1 < len(lines) {
		body = strings.Join(lines[frontMatterEnd+1:], "\n")
	}

	return &frontmatterParts{
		lines:            lines,
		frontmatter:      strings.Join(frontmatterLines, "\n"),
		body:             body,
		newline:          newline,
		frontMatterStart: frontMatterStart,
		frontMatterEnd:   frontMatterEnd,
	}, nil
}

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

// ParseGuide reads a single markdown file, parses its frontmatter, and constructs a Guide struct.
func ParseGuide(path string) (Guide, error) {
	baseName := filepath.Base(path)
	id := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	content, err := os.ReadFile(path)
	if err != nil {
		return Guide{}, err
	}

	parts, err := extractFrontmatter(content)
	if err != nil {
		return Guide{}, err
	}

	var meta GuideMetadata
	if err := yaml.Unmarshal([]byte(parts.frontmatter), &meta); err != nil {
		return Guide{}, fmt.Errorf("failed to parse yaml: %w", err)
	}

	hasContent := false
	for i := parts.frontMatterEnd + 1; i < len(parts.lines); i++ {
		if strings.TrimSpace(parts.lines[i]) != "" {
			hasContent = true
			break
		}
	}

	return Guide{
		ID:         id,
		Path:       path,
		Metadata:   meta,
		LineCount:  len(parts.lines) - (parts.frontMatterEnd + 1),
		HasContent: hasContent,
		Body:       parts.body,
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

	parts, err := extractFrontmatter(content)
	if err != nil {
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
	newLines = append(newLines, parts.lines[:parts.frontMatterStart+1]...)

	metaStr := strings.TrimRight(string(metaBytes), "\n")
	if metaStr != "" {
		newLines = append(newLines, strings.Split(metaStr, "\n")...)
	}

	newLines = append(newLines, parts.lines[parts.frontMatterEnd:]...)

	return os.WriteFile(path, []byte(strings.Join(newLines, parts.newline)), 0644)
}
