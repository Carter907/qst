---
prerequisites: [Normal Mode]
sub_guides: []
scope: definition
tags: [vim, modes, command-line]
---

## Command-Line Mode

**Command-Line Mode** (also known as Ex Mode) allows users to execute complex editor operations, file management commands, search queries, and external shell processes via a prompt line at the bottom of the screen.

### Activation Characters

- `:`: Enters Command-Line Mode for Ex commands (e.g., `:w` to save, `:q` to quit, `:e {file}` to edit).
- `/`: Initiates forward pattern searching.
- `?`: Initiates backward pattern searching.

### Essential Ex Operations

- `:w`: Write (save) buffer changes to disk.
- `:q`: Quit current window (fails if buffer has unsaved edits).
- `:q!`: Force quit, discarding all unsaved modifications.
- `:wq` or `:x`: Save changes and quit.
- `q:`: Open the command-line history window as an editable buffer.
