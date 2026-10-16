---
prerequisites: [Pattern Search Forward and Backward]
sub_guides: []
scope: definition
tags: [vim, substitution, regex]
---

## Line and Global Substitution

The **substitute** command (`:s`) combines regular expressions with range addressing to perform targeted or file-wide string replacements.

### The Substitute Syntax

$$:[\text{range}]s/\text{pattern}/\text{replacement}/[\text{flags}]$$

### Common Ranges

- `:s/old/new/`: Replace the first match of `old` with `new` on the current line.
- `:s/old/new/g`: Replace **all** matches on the current line.
- `:%s/old/new/g`: Replace all matches across the **entire document** (`%` represents lines 1 to `$`).
- `:10,25s/old/new/g`: Replace all matches between lines 10 and 25 inclusive.
- `:'<,'>s/old/new/g`: Replace matches within the active visual selection range.
