package cmd

import (
	"path/filepath"
	"strings"

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
