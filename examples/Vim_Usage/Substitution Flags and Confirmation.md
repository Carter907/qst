---
prerequisites: [Line and Global Substitution]
sub_guides: []
scope: definition
tags: [vim, substitution, flags]
---

## Substitution Flags and Confirmation

Substitution flags control matching behavior and introduce interactive prompts to review and verify replacements before applying them.

### Critical Flags

- `g` (global): Replace all occurrences on each addressed line, rather than just the first.
- `i` (ignore case): Perform case-insensitive pattern matching.
- `I` (case sensitive): Force case-sensitive matching.
- `c` (confirm): Prompt for confirmation on every individual match.

### Interactive Confirmation Prompt (`/c`)

When running `:%s/old/new/gc`, Vim pauses at each match with options:
- `y`: Substitute this match.
- `n`: Skip this match.
- `a`: Substitute this and all remaining matches without asking.
- `q`: Quit substitution immediately.
- `l`: Substitute this match and then stop (last).
