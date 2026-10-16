---
prerequisites: [Normal Mode]
sub_guides: []
scope: definition
tags: [vim, navigation, motions]
---

## Character Motions hjkl

The directional keys **`h`**, **`j`**, **`k`**, and **`l`** form the primary navigational bedrock of Vim. Positioned directly under the right hand on the home row of standard QWERTY keyboards, they eliminate the need to move your hand to arrow keys.

### Directional Mapping

- `h`: Move cursor one character **left**.
- `j`: Move cursor one line **down**.
- `k`: Move cursor one line **up**.
- `l`: Move cursor one character **right**.

### Numerical Counts

All motion commands accept numeric count prefixes to multiply the movement distance:
- `5j`: Move 5 lines down.
- `10l`: Move 10 characters right.
- `3k`: Move 3 lines up.
