---
prerequisites: [String Storage and Representation]
sub_guides:
    - guide: String Variables
      adherence: introductory
      segment: 4-4
    - guide: String Initialization
      adherence: detailed
      segment: 6-6
    - guide: Pointers to String Literals
      adherence: strict
      segment: 8-8
scope: description
tags: [c, strings, initialization]
---

## Declaring and Initializing Strings

A mutable string in C is declared using **String Variables**, which are one-dimensional character arrays allocated with at least N + 1 elements to accommodate both text and the terminating null byte.

Convenient syntax allows **String Initialization** at declaration time using string literals, with the compiler automatically sizing arrays or zero-filling any omitted trailing elements.

Alternatively, developers may declare **Pointers to String Literals** (char *p = "text";), but these point to read-only memory segments and must not have their contents mutated.
