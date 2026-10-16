---
prerequisites: []
sub_guides: []
scope: definition
tags: [c, strings, variables]
---

## String Variables

In C, a **String Variable** is represented as a one-dimensional array of `char` capable of storing a null-terminated sequence of characters.

Because strings must terminate with the null character (`'\0'`), an array declared to hold a string of maximum length $N$ must have a declared dimension of at least $N + 1$:

```c
char str[21]; // Can hold a string of up to 20 characters plus '\0'
```

If a string variable is declared without an initializer (e.g. `char s[10];`), its elements contain uninitialized garbage values. Placing a null character at the very beginning (`s[0] = '\0';`) turns it into an empty string of length 0.
