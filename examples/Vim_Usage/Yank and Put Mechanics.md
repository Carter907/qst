---
prerequisites: [Delete and Cut Operations]
sub_guides: []
scope: definition
tags: [vim, editing, yank-put]
---

## Yank and Put Mechanics

In Vim terminology, copying is called **yanking** (`y`) and pasting is called **putting** (`p`). Yanking preserves buffer text while populating registers for subsequent insertion.

### Yanking Operations

- `yy` or `Y`: Yank the entire current line.
- `yw`: Yank from cursor to the start of the next word.
- `y$`: Yank from cursor to the end of the line.
- `y%`: Yank the entire block between matching delimiters.

### Putting Operations

- `p`: Put text **after** the cursor (if character-wise) or **below** the current line (if line-wise).
- `P`: Put text **before** the cursor (if character-wise) or **above** the current line (if line-wise).
- `gp` / `gP`: Like `p` / `P`, but leaves the cursor positioned after the inserted text.
