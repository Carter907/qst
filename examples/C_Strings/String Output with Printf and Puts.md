---
prerequisites: [String Input with Gets and Fgets]
sub_guides: []
scope: definition
tags: [c, strings, io]
---

## String Output with Printf and Puts

C standard I/O provides two primary functions for writing strings to standard output:

- `printf("%s", str)`: Writes characters sequentially starting from the address pointed to by `str` until encountering a null character (`'\0'`).
  - Field width (`%20s`): Right-aligns the string in a field of 20 columns.
  - Left-alignment (`%-20s`): Left-aligns the string in a field of 20 columns.
  - Precision (`%.*s` or `%.5s`): Limits output to at most the specified number of characters, even if the underlying string is longer.
- `puts(str)`: Writes the null-terminated string to `stdout` and automatically appends an explicit newline character (`'\n'`). `puts` is simpler and faster than `printf("%s\n", str)` because it avoids format string parsing overhead.
