---
prerequisites: [String Literals]
sub_guides: []
scope: definition
tags: [c, strings, types]
---

## Strings v.s. Character

In C, there is a fundamental syntactic, structural, and semantic distinction between a single character constant and a string literal containing a single character:

- `'a'` is a **Character Constant**. It represents a single numerical character code of type `char` (or `int` in C expressions) and occupies exactly 1 byte.
- `"a"` is a **String Literal**. It represents an array of two characters: `'a'` followed immediately by `'\0'`. It occupies 2 bytes of storage in memory.

When used in expressions, `"a"` evaluates to a pointer (`char *` or `const char *`) pointing to the memory address of the first character (`'a'`). Attempting to use a string literal where an integer character is expected (such as passing `"a"` to a function expecting `char`), or vice versa (such as passing `'a'` to `printf("%s", ...)`), results in type mismatches or fatal memory segmentation faults.
