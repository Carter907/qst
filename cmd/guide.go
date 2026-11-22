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

// loadGraphLenient parses the knowledge graph leniently, printing a concise warning to stderr
// if any guide files were skipped due to parse errors or duplicate IDs.
func loadGraphLenient(cmd *cobra.Command, dir string) (graph.ParseResult, error) {
	res, err := graph.ParseGraphLenient(dir)
	if err != nil {
		return graph.ParseResult{}, err
	}
	if len(res.Errors) > 0 && cmd != nil {
		cmd.PrintErrf("warning: skipped %d invalid guide(s); run 'qst validate' for details\n", len(res.Errors))
	}
	return res, nil
}

// lookupGuide resolves a guide ID from a ParseResult, returning the file's specific parse or
// duplicate error if that ID failed to parse instead of a generic "not found" error.
func lookupGuide(res graph.ParseResult, guideID string) (graph.Guide, error) {
	if pe := res.ErrorForID(guideID); pe != nil {
		return graph.Guide{}, pe.Err
	}
	g, exists := res.Guides[guideID]
	if !exists {
		return graph.Guide{}, fmt.Errorf("guide '%s' not found in graph", guideID)
	}
	return g, nil
}

// resolveGuide finds a guide by name or relative path in the specified graph directory,
// searching recursively so nested subdirectories are supported transparently.
func resolveGuide(cmd *cobra.Command, dir string, rawGuide string) (graph.Guide, map[string]graph.Guide, error) {
	cleanName := cleanGuideName(rawGuide)
	if cleanName == "" {
		return graph.Guide{}, nil, fmt.Errorf("guide name cannot be empty")
	}

	res, err := loadGraphLenient(cmd, dir)
	if err != nil {
		return graph.Guide{}, nil, fmt.Errorf("error parsing graph: %w", err)
	}

	targetGuide, err := lookupGuide(res, cleanName)
	if err != nil {
		return graph.Guide{}, res.Guides, err
	}

	return targetGuide, res.Guides, nil
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
