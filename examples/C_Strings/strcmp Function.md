---
prerequisites: [strcat Function]
sub_guides: []
scope: definition
tags: [c, strings, stdlib]
---

## strcmp Function

In C, the relational operators (`==`, `!=`, `<`, `>`) cannot be used to compare string contents because arrays decay to pointers, comparing memory addresses rather than character data. The `strcmp` function performs lexicographical comparison:

```c
int strcmp(const char *s1, const char *s2);
```

Return values:
- **Negative Integer (`< 0`)**: `s1` is lexicographically less than `s2` (its first differing character has a smaller numerical value).
- **Zero (`0`)**: `s1` and `s2` are identical in length and content.
- **Positive Integer (`> 0`)**: `s1` is lexicographically greater than `s2`.

Characters are compared as `unsigned char` values, ensuring consistent ordering across signed character platforms. The standard idiom for testing equality is:
```c
if (strcmp(s1, s2) == 0) { /* strings are equal */ }
```
