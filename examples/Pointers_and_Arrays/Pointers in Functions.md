---
prerequisites: [Pointer Operators and Assignment]
sub_guides:
    - guide: Pointers as Function Arguments
      adherence: introductory
      segment: 4-4
    - guide: Const Pointer Parameters
      adherence: introductory
      segment: 6-6
    - guide: Returning Pointers from Functions
      adherence: introductory
      segment: 8-8
scope: explanation
tags: [c, functions, pointers]
---

## Pointers in Functions

Pointers provide a crucial communication bridge between functions through **Pointers as Function Arguments**, allowing subroutines to write output values directly back to caller variables.

When arguments should remain protected from unintended modification, developers designate them using **Const Pointer Parameters** for compile-time safety checks.

Finally, functions can output references via **Returning Pointers from Functions**, while taking care to avoid dangerous dangling references to local stack frames.
