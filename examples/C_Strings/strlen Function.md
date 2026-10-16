---
prerequisites: [The string.h Header]
sub_guides: []
scope: definition
tags: [c, strings, stdlib]
---

## strlen Function

The `strlen` function returns the length of a null-terminated string:

```c
size_t strlen(const char *s);
```

Key characteristics:
- **Length Counting**: `strlen` counts characters sequentially up to, but not including, the terminating null character (`'\0'`). For example, `strlen("abc")` returns `3`, even though the literal occupies 4 bytes of memory.
- **Complexity**: Because C strings do not store length in a header, `strlen` must scan characters one by one, operating in $O(N)$ linear time.
- **Performance Pitfall**: Calling `strlen(str)` inside a loop condition:
  ```c
  for (size_t i = 0; i < strlen(str); i++) // Inefficient O(N^2)
  ```
  causes `strlen` to be re-evaluated on every single iteration. Saving the length to a variable prior to the loop achieves $O(N)$ linear performance.
