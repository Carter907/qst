---
prerequisites: [The Operator-Motion Rule]
sub_guides: []
scope: definition
tags: [vim, productivity, repetition]
---

## The Dot Command

The **`.`** (dot) command repeats the last modifying change. It is considered one of the highest leverage productivity features in modal text editing.

### Definition of a Change

In Vim, a "change" is defined as any single operation that modifies buffer content, from entering Insert mode to returning to Normal mode, or an operator-motion command like `dw` or `cw`.

### The Dot Formula Idiom

1. Identify a repetitive target.
2. Execute a single clean change on the first occurrence (e.g., `cwreplacement<Esc>`).
3. Navigate to subsequent targets using a repeatable motion (e.g., `n` to jump to next search match).
4. Press `.` to replicate the change instantly.

By structuring edits into motion + dot pairs (`n.` or `;`), repetitive refactoring becomes effortless and error-free.
