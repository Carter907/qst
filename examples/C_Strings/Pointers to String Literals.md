---
prerequisites: [String Initialization]
sub_guides: []
scope: definition
tags: [c, strings, pointers]
---

## Pointers to String Literals

A pointer of type `char *` can be initialized to point directly to a string literal:

```c
char *p = "abc";
```

This declaration differs significantly from an initialized character array (`char s[] = "abc";`):

- In `char s[] = "abc";`, an array `s` is allocated in writable memory and initialized with a copy of the literal. The elements of `s` can be freely modified (`s[0] = 'x';`), but the array identifier `s` cannot be reassigned to point elsewhere.
- In `char *p = "abc";`, `p` is a pointer pointing directly to the read-only string literal. The pointer `p` can be reassigned to point to different strings, but attempting to modify the characters through `p` causes undefined behavior. Defensive C code uses `const char *p = "abc";` to enforce immutability at compile time.
