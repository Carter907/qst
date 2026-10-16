---
prerequisites: [Pattern Search Forward and Backward]
sub_guides: []
scope: definition
tags: [vim, search, options]
---

## Search Options and Highlighting

Vim provides several configuration options that dramatically improve the responsiveness and clarity of search operations.

### Recommended Options

- `:set hlsearch` (`:set hls`): Highlight all matches matching the current search pattern.
- `:set nohlsearch`: Disable match highlighting.
- `:nohlsearch` (`:noh`): Temporarily turn off current highlights until the next search query.
- `:set incsearch` (`:set is`): Enable incremental search, jumping to the first match dynamically as you type.
- `:set ignorecase` (`:set ic`): Make search case-insensitive.
- `:set smartcase` (`:set scs`): Automatically switch to case-sensitive search whenever an uppercase letter is entered.
