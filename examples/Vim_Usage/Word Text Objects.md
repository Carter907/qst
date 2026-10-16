---
prerequisites: [The Operator-Motion Rule]
sub_guides: []
scope: definition
tags: [vim, text-objects, words]
---

## Word Text Objects

Text objects enable operations to target structured linguistic and syntactic constructs directly, regardless of where the cursor is positioned within the object.

### Inner vs Around Words

- `iw` (**inner word**): Selects or operates strictly on the alphanumeric characters of the word, excluding surrounding whitespace.
- `aw` (**a word**): Selects or operates on the word **plus** its trailing whitespace boundary.

### Practical Examples

- `ciw`: Change inner word (clears word and enters Insert Mode, even if cursor is in the middle of the word).
- `daw`: Delete a word cleanly, preserving correct spacing between adjacent words.
- `yiw`: Yank the word under the cursor into the unnamed register.
- `viw`: Visually select the current word.
