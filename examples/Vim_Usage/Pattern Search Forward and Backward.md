---
prerequisites: [Command-Line Mode]
sub_guides: []
scope: definition
tags: [vim, search, patterns]
---

## Pattern Search Forward and Backward

Pattern search is the primary long-distance navigation tool in Vim, allowing quick cursor positioning using regular expressions and exact string matching.

### Initiation and Navigation

- `/{pattern}<CR>`: Search **forward** for `{pattern}`.
- `?{pattern}<CR>`: Search **backward** for `{pattern}`.
- `n`: Jump to the **next** match in the current search direction.
- `N`: Jump to the **previous** match in the opposite search direction.

### Cursor Word Matching

- `*`: Instantly search forward for the word directly under the cursor.
- `#`: Instantly search backward for the word directly under the cursor.
- Combine `*` with `cw` and `n.` for rapid variable refactoring across a file.
