---
prerequisites: [Const Pointer Parameters]
sub_guides:
    - guide: Dangling Pointer
      adherence: introductory
      segment: 4-4
scope: description
tags: [c, functions, memory, pointers]
---

## Returning Pointers from Functions

While C functions can return pointers to dynamic or static objects, returning the address of a local stack variable produces an invalid **Dangling Pointer** resulting in undefined behavior.
