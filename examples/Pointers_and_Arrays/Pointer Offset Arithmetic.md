---
prerequisites: []
sub_guides:
    - guide: Pointer Arithmetic
      adherence: introductory
      segment: 4-4
scope: description
tags: [c, pointers, arithmetic]
---

## Pointer Offset Arithmetic

Adding or subtracting an integer offset `j` to a pointer executes scaled **Pointer Arithmetic**, advancing or retreating by `j * sizeof(*p)` bytes to point directly to neighbouring array elements.
