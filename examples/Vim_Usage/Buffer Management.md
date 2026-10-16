---
prerequisites: [Command-Line Mode]
sub_guides: []
scope: definition
tags: [vim, buffers, files]
---

## Buffer Management

In Vim, a **buffer** is the in-memory representation of a file. When editing multiple files, Vim loads each into a separate buffer, independent of how windows are displayed on screen.

### Core Buffer Commands

- `:ls` or `:buffers`: List all currently open buffers with their IDs and flags (`%` active, `#` alternate, `+` modified).
- `:b {number}` or `:b {name}`: Switch active viewport to the specified buffer.
- `:bnext` (`:bn`): Switch to the next buffer in sequence.
- `:bprev` (`:bp`): Switch to the previous buffer in sequence.
- `:bd` or `:bdelete`: Unload and close the current buffer.
- `<Ctrl-^>`: Quick toggle between the current buffer and the alternate buffer.
