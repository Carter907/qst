---
scope: primitive
prerequisites:
  - Guide and Manifest Data Models
sub_guides: []
tags: [graph, validator, segments, coverage]
---

# Segment Range Parser and Overlap Detector

Vertical `sub_guides` links require a `segment` string (such as `"4-8, 12-15"` or `"6"`) that maps the child guide to the exact line ranges of the parent guide's Markdown body. In `internal/graph/validator.go`, `parseSegments` and `validateCoverage` enforce interval validity, non-overlapping ranges, and optional 100% line coverage.

## Parsing Segment Expressions (`parseSegments`)

`parseSegments(segmentStr string) ([][2]int, error)` splits comma-separated clauses using `strings.SplitSeq(segmentStr, ",")` and normalizes both single lines (`"7"` $\to$ `[2]int{7, 7}`) and hyphenated ranges (`"4-10"` $\to$ `[2]int{4, 10}`), rejecting non-positive indices (`start < 1`) or inverted bounds (`start > end`).

## Interval Sorting and Overlap Detection (`validateCoverage`)

Across all sub-guides of a parent guide, `validateSubGuides` aggregates every parsed `[2]int` interval into `allRanges`. `validateCoverage` then sorts the intervals lexicographically by start line (breaking ties by end line):

```go
slices.SortFunc(allRanges, func(a, b [2]int) int {
	if a[0] == b[0] {
		return a[1] - b[1]
	}
	return a[0] - b[0]
})

for i := 1; i < len(allRanges); i++ {
	if allRanges[i][0] <= allRanges[i-1][1] {
		return fmt.Errorf("guide '%s' has overlapping subguide segments: %v and %v",
			guide.ID, allRanges[i-1], allRanges[i])
	}
}
```

Because the intervals are sorted, any overlap—whether partial (`[2, 6]` and `[5, 9]`) or nested (`[2, 10]` and `[4, 6]`)—is guaranteed to manifest between adjacent elements `allRanges[i-1]` and `allRanges[i]` in a single linear pass.

## Strict Coverage and POSIX Trailing Newlines

When `manifest.StrictCoverage` is `true`, every line from `1` to `guide.LineCount` must be covered without gaps (`allRanges[i][0] > allRanges[i-1][1]+1`). Because POSIX text editors append a trailing `\n` at EOF—which creates a trailing empty string after `strings.Split(body, "\n")`—`validateCoverage` decrements `effectiveLineCount` by `1` when `strings.HasSuffix(guide.Body, "\n")` and the final segment ends at `LineCount - 1`.
