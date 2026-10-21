package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Carter907/qst/internal/graph"
	"github.com/spf13/cobra"
)

// cleanGuideName extracts the base guide name from a path or filename,
// stripping any directory components and any case-insensitive .md extension.
func cleanGuideName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	base := filepath.Base(raw)
	if strings.HasSuffix(strings.ToLower(base), ".md") {
		base = base[:len(base)-3]
	}
	return base
}

// resolveGuideAndDir extracts the guide name, and if the "dir" flag was not explicitly
// changed on cmd and raw contains directory components, updates *currentDir with the directory path.
func resolveGuideAndDir(cmd *cobra.Command, rawGuide string, currentDir *string) string {
	cleanName := cleanGuideName(rawGuide)
	if cmd != nil && currentDir != nil {
		dirFlag := cmd.Flags().Lookup("dir")
		if dirFlag == nil || !dirFlag.Changed {
			dir := filepath.Dir(rawGuide)
			if dir != "." && dir != "" {
				*currentDir = dir
			}
		}
	}
	return cleanName
}

// cleanGuideList cleans each guide name in the provided slice.
func cleanGuideList(rawGuides []string) []string {
	if rawGuides == nil {
		return nil
	}
	result := make([]string, len(rawGuides))
	for i, g := range rawGuides {
		result[i] = cleanGuideName(g)
	}
	return result
}

// resolveGuide finds a guide by name or relative path in the specified graph directory,
// searching recursively so nested subdirectories are supported transparently.
func resolveGuide(dir string, rawGuide string) (graph.Guide, map[string]graph.Guide, error) {
	cleanName := cleanGuideName(rawGuide)
	if cleanName == "" {
		return graph.Guide{}, nil, fmt.Errorf("guide name cannot be empty")
	}

	guides, err := graph.ParseGraph(dir)
	if err != nil {
		return graph.Guide{}, nil, fmt.Errorf("error parsing graph: %w", err)
	}

	targetGuide, exists := guides[cleanName]
	if !exists {
		return graph.Guide{}, guides, fmt.Errorf("guide '%s' not found in graph", cleanName)
	}

	return targetGuide, guides, nil
}

// parseCSVOrSlice normalizes an array of string arguments, splitting comma-separated tokens
// and stripping any bracket delimiters and whitespace.
func parseCSVOrSlice(items []string) []string {
	var res []string
	for _, item := range items {
		item = strings.TrimPrefix(item, "[")
		item = strings.TrimSuffix(item, "]")
		for _, part := range strings.Split(item, ",") {
			if trimmed := strings.Trim(part, "[] \t"); trimmed != "" {
				res = append(res, trimmed)
			}
		}
	}
	return res
}

// parseGuideCSVOrSlice normalizes an array of guide names, splitting comma-separated tokens
// and removing file extensions and path prefixes.
func parseGuideCSVOrSlice(items []string) []string {
	raw := parseCSVOrSlice(items)
	res := make([]string, len(raw))
	for i, g := range raw {
		res[i] = cleanGuideName(g)
	}
	return res
}
