---
prerequisites: [Word Text Objects, Bracket and Delimiter Matching]
sub_guides: []
scope: definition
tags: [vim, text-objects, quotes, delimiters]
---

## Delimiter and Quote Text Objects

Delimiter text objects provide lightning-fast targeting for quoted strings and nested parentheses, braces, or brackets without requiring manual cursor alignment.

### Common Delimiter Targets

- Double Quotes: `i"` (inside quotes) and `a"` (around quotes including delimiters).
- Single Quotes: `i'` and `a'`.
- Parentheses: `i(` or `ib` (inside parens) and `a(` or `ab` (around parens).
- Curly Braces: `i{` or `iB` (inside braces) and `a{` or `aB` (around braces).
- Square Brackets: `i[` and `a[`.

### Editing Workflows

- `ci"`: Change the contents inside a string literal.
- `da)`: Delete a function call parameter list including parentheses.
- `yi{`: Copy the body of an entire C/Java function block.
