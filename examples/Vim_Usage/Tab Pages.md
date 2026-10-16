---
prerequisites: [Split Window Navigation]
sub_guides: []
scope: definition
tags: [vim, tabs, workspaces]
---

## Tab Pages

In Vim, a **tab page** is not a representation of a single file as in traditional web browsers or GUI editors; it is an independent workspace layout containing a distinct collection of split windows.

### Tab Commands

- `:tabnew {file}` or `:tabe {file}`: Open a new tab page (optionally loading `{file}`).
- `gt`: Cycle forward to the **next** tab page.
- `gT`: Cycle backward to the **previous** tab page.
- `{count}gt`: Jump directly to tab page number `{count}`.
- `:tabclose`: Close the current tab page and its viewports.
- `:tabonly`: Close all other tab pages except the currently focused one.
