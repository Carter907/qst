---
prerequisites: [Character Motions hjkl]
sub_guides: []
scope: definition
tags: [vim, navigation, lines]
---

## Line Start and End Motions

Navigating within a single line is accelerated by dedicated boundary motions targeting column zero, the first non-blank character, or the end of the line.

### Boundary Motion Keys

- `0`: Jump strictly to the **first character** (column 0) of the line.
- `^`: Jump to the **first non-blank character** (skipping leading indentation spaces/tabs).
- `$`: Jump to the **last character** of the current line.
- `g_`: Jump to the **last non-blank character** of the line (ignoring trailing whitespace).

### Practical Idioms

Coupling line boundary motions with operators yields swift line manipulation:
- `d$`: Delete from cursor to the end of the line (equivalent to `D`).
- `d0`: Delete from cursor back to column 0.
- `c^`: Change from cursor back to the first non-blank character.
