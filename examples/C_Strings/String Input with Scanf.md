---
prerequisites: []
sub_guides: []
scope: definition
tags: [c, strings, io]
---

## String Input with Scanf

The standard library function `scanf` uses the `%s` conversion specification to read a string from standard input into a character array:

```c
char str[20];
scanf("%s", str); // Note: no & operator needed because array name decays to a pointer
```

When reading with `%s`:

1. `scanf` skips any leading whitespace characters (spaces, tabs, newlines).
2. It reads non-whitespace characters sequentially and stores them in the array.
3. Reading stops as soon as the next whitespace character is encountered.
4. `scanf` automatically appends a terminating null character (`'\0'`).

Without a field width limit, unbounded `scanf("%s", str)` is vulnerable to buffer overflow. Specifying a maximum width matching buffer capacity minus one (e.g. `scanf("%19s", str);`) guarantees bounds safety.
