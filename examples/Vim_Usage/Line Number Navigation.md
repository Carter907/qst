---
prerequisites: [Character Motions hjkl]
sub_guides: []
scope: definition
tags: [vim, navigation, line-numbers]
---

## Line Number Navigation

Line number navigation provides absolute and relative jumping capabilities across the entire document length, fundamental for inspecting compiler errors and navigating large codebases.

### Direct Jump Commands

- `gg`: Jump to the **first line** of the file.
- `G`: Jump to the **last line** of the file.
- `{count}G` or `{count}gg`: Jump directly to line number `{count}` (e.g., `42G` jumps to line 42).
- `:{number}<CR>`: Jump to line `{number}` via Command-Line Mode.

### Cursor Querying

- `<Ctrl-g>`: Display the current file path, modification status, total line count, and percentage progress in the command bar.
- `g<Ctrl-g>`: Display detailed statistics including word count, byte count, and column position.
