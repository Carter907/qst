---
prerequisites: [String Variables]
sub_guides: []
scope: definition
tags: [c, strings, initialization]
---

## String Initialization

A character array representing a string variable can be initialized at the point of declaration using string literal shorthand syntax:

```c
char date1[8] = "June 14"; // Elements: 'J', 'u', 'n', 'e', ' ', '1', '4', '\0'
```

C provides several flexible initialization options:

1. **Omitted Length**: If the array dimension is omitted (`char date2[] = "June 14";`), the compiler automatically counts the characters in the literal, adds 1 for `'\0'`, and allocates an array of 8 bytes.
2. **Partial Initialization**: If the declared size exceeds the literal length (e.g. `char date3[10] = "June";`), the compiler copies `"June"`, appends `'\0'`, and zero-fills the remaining elements with additional null characters.
