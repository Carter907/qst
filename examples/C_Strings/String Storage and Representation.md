---
prerequisites: []
sub_guides:
    - guide: Null Character
      adherence: strict
      segment: 4-4
    - guide: String Literals
      adherence: detailed
      segment: 6-6
    - guide: Strings v.s. Character
      adherence: introductory
      segment: 8-8
scope: description
tags: [c, strings, fundamentals]
---

## String Storage and Representation

In C, a string is stored as a contiguous sequence of characters concluded by a **Null Character** ('\0'), which serves as an in-band sentinel value for loops and library routines.

Constant text embedded in source code is defined using **String Literals**, which are stored in read-only program memory and automatically appended with a terminating null character.

It is critical to contrast **Strings v.s. Character** constants: while a single character literal 'a' is an integer value occupying one byte, a string literal "a" is an array of two characters evaluated as a pointer to memory.
