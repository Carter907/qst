---
prerequisites: []
sub_guides: []
scope: definition
tags: [c, arrays, pointers]
---

## Array Decay

**Array Decay** is the implicit conversion in C where an array identifier automatically decays into a pointer pointing to its first element (`a` becomes `&a[0]`), in almost all expression contexts.
