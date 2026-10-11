---
prerequisites: [Pointers as Function Arguments]
sub_guides:
    - guide: Const Pointer
      adherence: introductory
      segment: 4-4
scope: description
tags: [c, functions, pointers]
---

## Const Pointer Parameters

To pass large structures or arrays efficiently without permitting accidental mutations, parameters can be declared as a **Const Pointer** (`const int *p`) to enforce read-only semantics.
