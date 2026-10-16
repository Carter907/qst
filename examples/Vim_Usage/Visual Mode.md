---
prerequisites: [Normal Mode]
sub_guides: []
scope: definition
tags: [vim, modes, visual]
---

## Visual Mode

**Visual Mode** provides visual feedback by highlighting a contiguous region of text, enabling operators like delete (`d`), yank (`y`), or change (`c`) to act upon the entire highlighted selection.

### Visual Mode Flavors

- `v`: **Character-wise Visual Mode** — highlights text character by character from the anchor to the cursor.
- `V`: **Line-wise Visual Mode** — highlights full lines including newline terminators.
- `<Ctrl-v>`: **Block-wise (Column) Visual Mode** — highlights a rectangular 2D box of characters across rows.

### Block Selection Power

In block-wise Visual Mode (`<Ctrl-v>`), you can perform column edits simultaneously:
1. Select a vertical block of columns across multiple rows.
2. Press `I` (or `A` for line appending).
3. Type the prefix text.
4. Press `<Esc>`. Vim replicates the typed text across every selected line in the block.
