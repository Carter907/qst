---
prerequisites: [String Input with Scanf]
sub_guides: []
scope: definition
tags: [c, strings, io, safety]
---

## String Input with Gets and Fgets

Reading entire lines of text containing embedded spaces requires dedicated line-reading functions:

- `gets(str)`: Historically read a line from `stdin` until a newline character was reached. Because `gets` has no mechanism to limit the number of characters read, it was inherently vulnerable to buffer overflows and was officially deprecated in C99 and removed completely in C11.
- `fgets(str, sizeof(str), stdin)`: The standard, secure function for reading a line. It takes the destination array, maximum buffer size $N$, and an input stream (`stdin`). It reads at most $N - 1$ characters, stores the newline character (`'\n'`) if encountered, and always appends the terminating `'\0'`.

To strip an unwanted trailing newline left by `fgets`, the standard C idiom uses:
```c
str[strcspn(str, "\n")] = '\0';
```
