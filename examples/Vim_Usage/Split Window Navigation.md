---
prerequisites: [Buffer Management]
sub_guides: []
scope: definition
tags: [vim, windows, splits]
---

## Split Window Navigation

Windows are viewports into buffers. Vim allows tiling multiple windows horizontally or vertically to view and edit different files (or different parts of the same file) simultaneously.

### Creating Splits

- `:split` or `:sp` or `<Ctrl-w> s`: Split viewport **horizontally**.
- `:vsplit` or `:vs` or `<Ctrl-w> v`: Split viewport **vertically**.
- `:close` or `<Ctrl-w> c`: Close the active split window.

### Navigating Split Viewports

- `<Ctrl-w> h`: Move focus to the window on the **left**.
- `<Ctrl-w> j`: Move focus to the window **below**.
- `<Ctrl-w> k`: Move focus to the window **above**.
- `<Ctrl-w> l`: Move focus to the window on the **right**.
- `<Ctrl-w> =`: Equalize dimensions of all split windows.
