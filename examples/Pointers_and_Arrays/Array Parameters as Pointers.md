---
prerequisites: [Subscript and Pointer Equivalence]
sub_guides:
    - guide: Array Parameter
      adherence: introductory
      segment: 4-4
scope: description
tags: [c, functions, arrays, pointers]
---

## Array Parameters as Pointers

Declaring a function parameter with array notation establishes an **Array Parameter** that C rewrites internally as a pointer, passing address information rather than duplicating array memory.
