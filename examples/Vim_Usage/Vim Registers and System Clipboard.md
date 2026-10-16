---
prerequisites: [Yank and Put Mechanics]
sub_guides: []
scope: definition
tags: [vim, registers, clipboard]
---

## Vim Registers and System Clipboard

Vim does not have a single clipboard; it maintains multiple distinct **registers** dedicated to deletions, yanks, named storage, and system clipboard interoperability.

### Register Types

- `""` (**unnamed register**): Default destination for all yank and delete operations.
- `"0` (**yank register**): Retains the text of the last yank, immune to subsequent deletions.
- `"a` to `"z` (**named registers**): 26 registers for persistent user-specified storage.
- `"+` and `"*` (**clipboard registers**): Interface directly with the operating system desktop clipboard.
- `"_` (**black hole register**): Discards text completely without polluting any registers (e.g., `"_dd`).

### Register Usage Syntax

- `"ayw`: Yank word into register `a`.
- `"ap`: Paste contents of register `a`.
- `"+y`: Copy selected text to the OS system clipboard.
- `"+p`: Paste OS system clipboard content into Vim.
