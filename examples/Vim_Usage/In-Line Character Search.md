---
prerequisites: [Line Start and End Motions]
sub_guides: []
scope: definition
tags: [vim, navigation, inline-search]
---

## In-Line Character Search

In-line character searches allow instantaneous targeting of specific characters on the current line, serving both as high-speed navigation motions and precise operator targets.

### Search Commands

- `f{char}`: Jump forward and land **on** the next occurrence of `{char}`.
- `F{char}`: Jump backward and land **on** the previous occurrence of `{char}`.
- `t{char}`: Jump forward to land **till** (one character before) the next occurrence of `{char}`.
- `T{char}`: Jump backward to land **till** (one character after) the previous occurrence of `{char}`.

### Repeating and Reversing

- `;`: Repeat the last `f`, `F`, `t`, or `T` motion in the **same direction**.
- `,`: Repeat the last `f`, `F`, `t`, or `T` motion in the **opposite direction**.

### Combining with Operators

- `df;`: Delete from cursor up to and including the next semicolon.
- `ct)`: Change from cursor up until (excluding) the closing parenthesis.
