---
prerequisites: [Normal Mode]
sub_guides: []
scope: definition
tags: [vim, modes, insert]
---

## Insert and Append Mode

**Insert Mode** is the state in which keystrokes are directly inserted into the buffer as text. Vim provides multiple specialized command keys to enter Insert Mode depending on where the cursor should begin writing.

### Common Entry Points

- `i`: Insert immediately **before** the cursor.
- `I`: Insert at the **first non-blank character** of the current line.
- `a`: Append immediately **after** the cursor.
- `A`: Append at the **very end** of the current line.
- `o`: Open a new blank line **below** the current line and enter Insert Mode.
- `O`: Open a new blank line **above** the current line and enter Insert Mode.

### Insert Mode Shortcuts

- `<Ctrl-w>`: Delete the word immediately preceding the cursor without leaving Insert Mode.
- `<Ctrl-u>`: Delete from the cursor back to the start of the line.
- `<Ctrl-o>`: Execute a single Normal Mode command and instantly return to Insert Mode.
- `<Esc>` or `<Ctrl-[>`: Leave Insert Mode and return to Normal Mode.
