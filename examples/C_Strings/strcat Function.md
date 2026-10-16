---
prerequisites: [strcpy Function]
sub_guides: []
scope: definition
tags: [c, strings, stdlib]
---

## strcat Function

The `strcat` (string concatenation) function appends the contents of one string to the end of another:

```c
char *strcat(char *dest, const char *src);
```

Key behaviors:
1. `strcat` locates the terminating null character of the `dest` string.
2. It begins copying characters from `src` directly over `dest`'s original terminating null character.
3. Once all characters from `src` are copied, `strcat` appends a new null character at the end.
4. **Capacity Precondition**: The array pointed to by `dest` must be large enough to hold the sum of both strings plus one byte for the terminating null character: `strlen(dest) + strlen(src) + 1`.
5. `strcat` returns `dest`, supporting chained operations.
