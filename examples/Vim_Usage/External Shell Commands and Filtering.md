---
prerequisites: [Command-Line Mode]
sub_guides: []
scope: definition
tags: [vim, shell, filter]
---

## External Shell Commands and Filtering

Vim integrates directly with Unix command-line utilities, enabling external process execution, shell output insertion, and buffer range filtering.

### Executing Shell Commands

- `:!{cmd}`: Execute `{cmd}` in an external shell and display its output (e.g., `:!git status`, `:!make`).
- `:r !{cmd}`: Execute `{cmd}` and insert its standard output into the buffer directly below the cursor (e.g., `:r !date`).

### Filtering Buffer Text through Shell Programs

- `:[range]!{cmd}`: Pipe the specified buffer range as standard input to `{cmd}` and replace the range with its standard output.
- `:%!sort`: Sort all lines in the current file using the system `sort` utility.
- `:%!jq .`: Format and pretty-print an entire JSON buffer using `jq`.
