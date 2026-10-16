---
prerequisites: [Character Motions hjkl]
sub_guides: []
scope: definition
tags: [vim, navigation, words]
---

## Word and WORD Navigation

Word motions allow rapid horizontal navigation across words rather than stepping character-by-character. Vim makes an important pedagogical distinction between **words** (lowercase) and **WORDs** (uppercase).

### Word Navigation (lowercase)

A *word* consists of letters, digits, and underscores, or a sequence of punctuation characters:
- `w`: Jump forward to the start of the next word.
- `b`: Jump backward to the start of the previous word.
- `e`: Jump forward to the end of the current or next word.
- `ge`: Jump backward to the end of the previous word.

### WORD Navigation (uppercase)

A *WORD* consists of non-whitespace characters separated strictly by whitespace. This enables jumping across compound expressions and identifiers containing punctuation (e.g., `user.profile.get_id()`):
- `W`: Jump forward to the start of the next whitespace-delimited WORD.
- `B`: Jump backward to the start of the previous whitespace-delimited WORD.
- `E`: Jump forward to the end of the current or next whitespace-delimited WORD.
