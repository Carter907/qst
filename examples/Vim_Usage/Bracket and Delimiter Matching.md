---
prerequisites: [Word and WORD Navigation]
sub_guides: []
scope: definition
tags: [vim, navigation, brackets]
---

## Bracket and Delimiter Matching

The **`%`** command jumps between matching pairs of structural delimiters, providing instant traversal across nested code blocks, control structures, and conditional preprocessor macros.

### Supported Delimiters

- Parentheses: `(` and `)`
- Curly braces: `{` and `}`
- Square brackets: `[` and `]`
- Preprocessor blocks: `#if`, `#else`, `#endif` (when `matchit` plugin is active).

### Navigation and Deletion

- When positioned on any delimiter, pressing `%` teleports the cursor directly to its corresponding opening or closing counterpart.
- `d%`: Delete from the current delimiter up to and including its matching delimiter.
- `v%`: Visually select the entire block bounded by the matching pair.
