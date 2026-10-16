---
prerequisites: [Strings v.s. Character]
sub_guides: []
scope: definition
tags: [c, strings, safety]
---

## Undefined Behavior in Strings

**Undefined Behavior** occurs in C when a program violates the rules of the C standard, causing unpredictable execution, crashes, silent memory corruption, or security vulnerabilities.

In C string programming, the most common causes of undefined behavior include:

1. **Modifying String Literals**: Because string literals are stored in read-only memory, attempting to write to them (for example, `char *p = "read-only"; p[0] = 'R';`) triggers segmentation faults or operating system memory protection traps.
2. **Missing Null Terminator**: Passing a non-null-terminated character array to functions like `strlen`, `strcpy`, or `printf("%s")` causes the function to read past the allocated buffer bounds until a stray zero byte is found elsewhere in memory.
3. **Buffer Overflows**: Writing more characters into a destination buffer than its declared capacity (e.g. copying a 20-character string into a 10-byte buffer) overwrites adjacent variables or stack frame return addresses.
