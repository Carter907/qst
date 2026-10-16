---
prerequisites: [Null Character]
sub_guides: []
scope: definition
tags: [c, strings, memory]
---

## String Literals

A **String Literal** (or string constant) is a sequence of characters enclosed in double quotation marks, such as `"Hello, World!\n"`.

When the C compiler encounters a string literal, it allocates an array of characters sufficient to contain the characters of the literal plus one additional byte at the end for an automatically appended **Null Character** (`'\0'`). For example, `"hello"` occupies 6 bytes of storage in memory: `'h'`, `'e'`, `'l'`, `'l'`, `'o'`, and `'\0'`.

In modern systems, string literals are typically placed in a read-only memory segment (such as `.rodata` or the text segment). Compilers may also perform **string literal pooling** (interning), where identical string literals share the exact same memory address. Adjacent string literals separated only by whitespace are automatically concatenated by the preprocessor during translation (e.g. `"Hello, " "World!"` becomes `"Hello, World!"`).
