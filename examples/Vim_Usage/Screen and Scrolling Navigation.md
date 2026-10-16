---
prerequisites: [Character Motions hjkl]
sub_guides: []
scope: definition
tags: [vim, navigation, scrolling]
---

## Screen and Scrolling Navigation

Screen and scrolling navigation commands reposition the cursor relative to the visible viewport or scroll the viewport itself without requiring repetitive line-by-line stepping.

### Screen Positioning Motions

- `H`: Jump cursor to the **High** (top) line of the visible screen.
- `M`: Jump cursor to the **Middle** line of the visible screen.
- `L`: Jump cursor to the **Low** (bottom) line of the visible screen.

### Viewport Scrolling Commands

- `<Ctrl-u>`: Scroll viewport **Up** by half a screen.
- `<Ctrl-d>`: Scroll viewport **Down** by half a screen.
- `<Ctrl-b>`: Scroll viewport **Backward** by a full screen page.
- `<Ctrl-f>`: Scroll viewport **Forward** by a full screen page.
- `zz`: Redraw screen with the cursor centered vertically.
- `zt`: Redraw screen with the cursor at the top.
- `zb`: Redraw screen with the cursor at the bottom.
