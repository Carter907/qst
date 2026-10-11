---
prerequisites: [Multidimensional Array Layout and Traversal]
sub_guides:
    - guide: Pointer to Array
      adherence: introductory
      segment: 4-4
scope: description
tags: [c, arrays, pointers, multidimensional]
---

## Pointers to Arrays

Passing multidimensional matrix rows to helper functions requires declaring a **Pointer to Array** (`int (*p)[NUM_COLS]`), maintaining row dimension semantics across function boundaries.
