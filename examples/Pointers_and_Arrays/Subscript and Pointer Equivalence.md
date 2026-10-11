---
prerequisites: [Array Decay to Pointer]
sub_guides:
    - guide: Array Subscript Equivalence
      adherence: introductory
      segment: 4-4
scope: description
tags: [c, arrays, pointers]
---

## Subscript and Pointer Equivalence

Under the C language specification, **Array Subscript Equivalence** guarantees that bracketed subscript indexing `a[i]` is identical to computing the offset pointer dereference `*(a + i)`.
