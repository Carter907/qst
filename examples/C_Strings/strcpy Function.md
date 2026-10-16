---
prerequisites: [strlen Function]
sub_guides: []
scope: definition
tags: [c, strings, stdlib]
---

## strcpy Function

In C, array variables cannot be copied using the standard assignment operator (`s1 = s2;`). Instead, the `strcpy` (string copy) function is used:

```c
char *strcpy(char *dest, const char *src);
```

Key behaviors:
1. `strcpy` copies the string pointed to by `src` into the memory array pointed to by `dest`, copying character-by-character until the terminating null character is copied.
2. **Buffer Capacity**: `strcpy` performs no bounds checking. The caller must guarantee that `dest` points to an allocated buffer of at least `strlen(src) + 1` bytes. Failure to provide sufficient space causes buffer overflow.
3. **Return Value**: `strcpy` returns `dest`, allowing function calls to be chained (e.g. `printf("%s\n", strcpy(buf, "text"));`).
