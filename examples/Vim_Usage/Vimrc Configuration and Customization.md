---
prerequisites: [Built-in Help System, Search Options and Highlighting]
sub_guides: []
scope: definition
tags: [vim, config, vimrc]
---

## Vimrc Configuration and Customization

Vim's runtime environment, default behavior, option settings, and key combinations are configured through the initialization file `~/.vimrc` (or `init.lua` / `init.vim` in Neovim).

### Essential Configuration Directives

```vim
" Sensible Editor Defaults
set nocompatible         " Disable vi compatibility
set number relativenumber" Hybrid line numbering
set expandtab tabstop=4  " Spaces over hard tabs
set shiftwidth=4         " Indent steps
set ignorecase smartcase " Intelligent search casing
set incsearch hlsearch   " Real-time highlighted search
set scrolloff=8          " Keep 8 lines visible above/below cursor
```

### Non-Recursive Key Mapping

- `let mapleader = " "`: Set the global `<Leader>` shortcut to Space.
- `nnoremap <Leader>w :w<CR>`: Map `<Space>w` to quickly save files in Normal Mode.
- Use `noremap` variants (`nnoremap`, `inoremap`, `vnoremap`) to avoid recursive evaluation traps.
