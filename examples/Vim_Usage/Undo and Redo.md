---
prerequisites: [Normal Mode]
sub_guides: []
scope: definition
tags: [vim, history, undo]
---

## Undo and Redo

Vim maintains a comprehensive history of changes, allowing non-linear traversal across past editing states through undo, redo, and undo branch trees.

### History Commands

- `u`: Undo the last change.
- `U`: Restore the current line to its state prior to any modifications made since moving onto it.
- `<Ctrl-r>`: Redo the previously undone change.

### Granularity of Undo

Each discrete entry into Insert mode followed by `<Esc>` represents a single undo step. By breaking editing sessions into focused Insert sessions separated by `<Esc>`, undo actions remain predictable and granular.
