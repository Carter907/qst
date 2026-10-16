---
prerequisites: [Arrays of Strings]
sub_guides: []
scope: definition
tags: [c, strings, cli]
---

## Command Line Arguments

In C, command-line arguments are passed into a program via parameters to `main`:

```c
int main(int argc, char *argv[])
```

Key semantics:
- `argc` (**Argument Count**): An integer representing the number of command-line arguments, including the program name.
- `argv` (**Argument Vector**): An array of pointers to null-terminated strings.
  - `argv[0]` contains the program's invocation name or an empty string.
  - `argv[1]` through `argv[argc - 1]` contain the command-line arguments.
  - `argv[argc]` is guaranteed by the C standard to be a null pointer (`NULL`).

Because `argv` is an array of pointers, it can be processed using array indexing (`argv[i]`) or pointer arithmetic (`for (char **p = &argv[1]; *p != NULL; p++)`).
