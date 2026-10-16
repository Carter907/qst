---
prerequisites: [strcmp Function]
sub_guides: []
scope: definition
tags: [c, strings, safety, stdlib]
---

## Bounded String Functions

The standard C library provides length-bounded counterparts to mitigate buffer overflows: `strncpy`, `strncat`, and `strncmp`.

Key functions and critical caveats:

1. `char *strncpy(char *dest, const char *src, size_t n);`: Copies at most `n` characters. **Caveat**: If `strlen(src) >= n`, `strncpy` does **not** append a terminating null character. Defensive programmers must enforce null termination manually:
   ```c
   strncpy(dest, src, sizeof(dest) - 1);
   dest[sizeof(dest) - 1] = '\0';
   ```
2. `char *strncat(char *dest, const char *src, size_t n);`: Appends at most `n` characters from `src` and **always** appends a null character. Therefore, `dest` must have room for its current length plus `n + 1` characters.
3. `int strncmp(const char *s1, const char *s2, size_t n);`: Compares at most `n` characters of `s1` and `s2`. Useful for prefix matching without modifying strings.
