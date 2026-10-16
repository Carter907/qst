---
prerequisites: [Command-Line Mode, Split Window Navigation]
sub_guides: []
scope: definition
tags: [vim, help, documentation]
---

## Built-in Help System

Vim includes a comprehensive, hyperlinked, authoritative documentation system covering every command, option, function, and editing concept.

### Accessing Documentation

- `:help` or `<F1>`: Open the root help directory in a new split window.
- `:help {subject}`: Search documentation for a specific subject (e.g., `:help user-manual`, `:help 'hlsearch'`).
- `:help {mode}_{cmd}`: Query mode-specific syntax (e.g., `:help i_CTRL-W`, `:help v_d`).

### Navigating Hypertext Tags

- `<Ctrl-]>`: Jump to the tag definition under the cursor.
- `<Ctrl-t>` or `<Ctrl-o>`: Jump back to the previous tag position.
- `:helptags {dir}`: Rebuild help indices for third-party plugins.
