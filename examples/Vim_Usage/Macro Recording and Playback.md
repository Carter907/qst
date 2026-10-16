---
prerequisites: [The Dot Command, Vim Registers and System Clipboard]
sub_guides: []
scope: definition
tags: [vim, automation, macros]
---

## Macro Recording and Playback

Macros record arbitrary sequences of Normal, Insert, and Ex commands into registers, turning repetitive multi-step edits into automated one-key scripts.

### Recording Procedure

1. Press `q{register}` (e.g., `qa`) to begin recording keystrokes into register `a`.
2. Execute your editing sequence cleanly, using robust motions (like `0`, `f`, `w`).
3. Press `q` in Normal Mode to stop recording.

### Execution and Loops

- `@a`: Execute the macro stored in register `a` once.
- `@@`: Replay the most recently executed macro.
- `10@a`: Execute the macro stored in register `a` 10 consecutive times.
- Visual execution: Select a block of lines in Visual Line mode (`V`) and run `:normal @a` to run the macro across each line in parallel.
