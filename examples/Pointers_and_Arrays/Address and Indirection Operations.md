---
prerequisites: []
sub_guides:
    - guide: Address Operator
      adherence: introductory
      segment: 4-4
    - guide: Indirection Operator
      adherence: introductory
      segment: 6-6
scope: description
tags: [c, operators, pointers]
---

## Address and Indirection Operations

A pointer is initialized with a target variable's location using the **Address Operator** (`&i`), creating a pointer that references `i`.

Once initialized, the program accesses and modifies the pointed-to value through the **Indirection Operator** (`*p`), treating `*p` as an alias for `i`.
