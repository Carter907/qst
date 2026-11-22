package graph

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// ParseError records a non-fatal error encountered while parsing a specific guide file
// (such as missing/invalid frontmatter, YAML syntax errors, or duplicate guide IDs).
type ParseError struct {
	ID   string
	Path string
	Err  error
}

func (e ParseError) Error() string {
	return e.Err.Error()
}

func (e ParseError) Unwrap() error {
	return e.Err
}

// ParseResult holds the valid guides and any per-file parse errors collected during a lenient graph walk.
type ParseResult struct {
	Guides map[string]Guide
	Errors []ParseError
}

// ErrorForID returns the first ParseError associated with the given guide ID, or nil if none exists.
func (r ParseResult) ErrorForID(id string) *ParseError {
	for i := range r.Errors {
		if r.Errors[i].ID == id {
			return &r.Errors[i]
		}
	}
	return nil
}

// CombinedError returns a single error joining all collected ParseErrors, or nil if there are none.
func (r ParseResult) CombinedError() error {
	if len(r.Errors) == 0 {
		return nil
	}
	errs := make([]error, len(r.Errors))
	for i, pe := range r.Errors {
		errs[i] = pe
	}
	return errors.Join(errs...)
}

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

// IsIgnoredDir returns true if a directory name should be skipped during graph traversal,
// archiving, static site generation, or file watching.
func IsIgnoredDir(name string) bool {
	return strings.HasPrefix(name, ".") || name == "public" || name == "vendor" || name == "assets"
}

// IsStaticAssetExt returns true if the file extension is a supported diagram or media asset.
func IsStaticAssetExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".png", ".jpg", ".jpeg", ".svg", ".gif", ".webp", ".mmd", ".mermaid":
		return true
	default:
		return false
	}
}

// normalizeMetadataSlices ensures nil slices in GuideMetadata serialize as empty YAML arrays [] rather than null.
func normalizeMetadataSlices(meta *GuideMetadata) {
	if meta.Prerequisites == nil {
		meta.Prerequisites = []string{}
	}
	if meta.SubGuides == nil {
		meta.SubGuides = []SubGuideRelation{}
	}
	if meta.Tags == nil {
		meta.Tags = []string{}
	}
}

// ParseGraphLenient reads all markdown files in the specified directory recursively,
// collecting per-file parse errors and duplicate guide ID conflicts into ParseResult.Errors
// instead of aborting on the first invalid file. Only directory/filesystem walk errors are fatal.
func ParseGraphLenient(dirPath string) (ParseResult, error) {
	cleanDir := filepath.Clean(dirPath)
	if _, err := os.Stat(cleanDir); err != nil {
		return ParseResult{}, err
	}

	guides := make(map[string]Guide)
	seenPaths := make(map[string]string)
	var parseErrs []ParseError

	err := filepath.Walk(cleanDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() {
			if path == cleanDir {
				return nil
			}
			if IsIgnoredDir(info.Name()) {
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

		baseName := filepath.Base(path)
		id := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		guide, err := ParseGuide(path)
		if err != nil {
			parseErrs = append(parseErrs, ParseError{
				ID:   id,
				Path: path,
				Err:  fmt.Errorf("failed to parse %s: %w", path, err),
			})
			return nil
		}

		if firstPath, seen := seenPaths[guide.ID]; seen {
			delete(guides, guide.ID)
			parseErrs = append(parseErrs, ParseError{
				ID:   guide.ID,
				Path: path,
				Err:  fmt.Errorf("duplicate guide ID '%s': found in %s and %s", guide.ID, firstPath, path),
			})
			return nil
		}

		seenPaths[guide.ID] = path
		guides[guide.ID] = guide
		return nil
	})

	if err != nil {
		return ParseResult{}, err
	}

	return ParseResult{
		Guides: guides,
		Errors: parseErrs,
	}, nil
}

// ParseGraph reads all markdown files in the specified directory recursively,
// extracts their frontmatter, and returns a map of guide IDs to Guide structs.
// If any files fail to parse or have duplicate IDs, all collected errors are returned.
func ParseGraph(dirPath string) (map[string]Guide, error) {
	res, err := ParseGraphLenient(dirPath)
	if err != nil {
		return nil, err
	}
	if combinedErr := res.CombinedError(); combinedErr != nil {
		return nil, combinedErr
	}
	return res.Guides, nil
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

// ParseGuidePermissive parses a markdown guide file, falling back to a plain body Guide
// if the file does not contain YAML frontmatter.
func ParseGuidePermissive(path string) (Guide, error) {
	if g, err := ParseGuide(path); err == nil {
		return g, nil
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return Guide{}, err
	}
	baseName := filepath.Base(path)
	id := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	return Guide{
		ID:         id,
		Path:       path,
		Body:       string(src),
		HasContent: len(src) > 0,
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

	normalizeMetadataSlices(&meta)

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
