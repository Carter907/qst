---
prerequisites: [The Operator-Motion Rule]
sub_guides: []
scope: definition
tags: [vim, editing, deletion]
---

## Delete and Cut Operations

Deletion in Vim functions as a cut operation: removed text is not permanently discarded, but automatically stored in the unnamed register, ready to be pasted with `p`.

### Single Character and Line Deletions

- `x`: Delete (cut) the single character directly under the cursor.
- `X`: Delete (cut) the character preceding the cursor.
- `dd`: Delete (cut) the entire current line.
- `D`: Delete (cut) from cursor position to the end of the line (shorthand for `d$`).

### Motion-Coupled Deletions

- `dw`: Delete from cursor to the start of the next word.
- `de`: Delete from cursor to the end of the current word.
- `d0`: Delete from cursor back to column 0.
- `dG`: Delete from cursor down to the very end of the document.
- `3dd`: Delete 3 consecutive lines starting from cursor.
