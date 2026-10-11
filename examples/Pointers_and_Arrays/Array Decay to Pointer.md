---
prerequisites: []
sub_guides:
    - guide: Array Decay
      adherence: introductory
      segment: 4-4
scope: description
tags: [c, arrays, pointers]
---

## Array Decay to Pointer

Whenever an array name appears in an expression without a dereference or address-of operator, **Array Decay** automatically transforms it into a pointer targeting its initial element (`&a[0]`).
