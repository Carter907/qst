---
prerequisites: [Delete and Cut Operations, Insert and Append Mode]
sub_guides: []
scope: definition
tags: [vim, editing, change]
---

## Change and Replace Operations

The **change** operator (`c`) deletes specified text and immediately transitions Vim into Insert Mode in a single atomic step, optimizing text replacement workflows.

### Change Operations

- `cw`: Change from cursor to the end of the word.
- `cc`: Change the entire current line (clears line and enters Insert Mode).
- `C`: Change from cursor to the end of the line (shorthand for `c$`).
- `s`: Substitute single character (deletes character under cursor and enters Insert Mode).
- `S`: Substitute entire line (synonym for `cc`).

### Character Replacement and Replace Mode

- `r{char}`: Replace the character under the cursor with `{char}` and immediately return to Normal Mode without entering Insert Mode.
- `R`: Enter **Replace Mode**, where subsequent keystrokes overwrite existing characters until `<Esc>` is pressed.
