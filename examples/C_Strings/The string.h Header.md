---
prerequisites: []
sub_guides: []
scope: definition
tags: [c, strings, stdlib]
---

## The string.h Header

The standard C library provides string and memory manipulation declarations in the header `<string.h>`.

All string-handling functions in `<string.h>` assume as a fundamental precondition that their string arguments are valid **null-terminated** sequences of characters. Passing pointers to unterminated arrays causes these functions to read beyond array boundaries, causing undefined behavior.

Key characteristics:
- Functions that inspect or return sizes (such as `strlen`) use the unsigned integer type `size_t`, defined by the standard to hold the maximum possible object size in memory.
- Function parameters representing read-only source strings are declared with `const char *` qualifiers, preventing accidental in-place modification.
