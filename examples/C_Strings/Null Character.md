---
prerequisites: []
sub_guides: []
scope: definition
tags: [c, strings, fundamentals]
---

## Null Character

A **Null Character** (written as the escape sequence `'\0'` in C) is a byte with all bits set to zero (numerical integer value `0`).

In C, strings are not stored as distinct primitive data types with built-in length headers. Instead, a string is stored as a contiguous sequence of characters marked at its conclusion by the null character. The null character serves as an in-band **sentinel value**, allowing standard library routines and pointer loops to traverse the string until the terminating byte is reached.

It is important to distinguish the null character `'\0'` (code 0) from the character literal `'0'` (ASCII code 48), the macro `NULL` (a null pointer constant typically defined as `((void *)0)`), and the integer constant `0`.
