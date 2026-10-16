---
prerequisites: []
sub_guides: []
scope: definition
tags: [c, strings, arrays]
---

## Arrays of Strings

Storing a collection of strings in C can be achieved using two distinct data structures:

1. **Two-Dimensional Character Array (Rectangular Array)**:
   ```c
   char planets[8][10] = {"Mercury", "Venus", "Earth", ...};
   ```
   All rows have identical length (10 bytes). Memory is stored in contiguous row-major order. If strings vary widely in length, memory is wasted by unused zero-padding.
2. **Array of Pointers to Characters (Ragged Array)**:
   ```c
   char *planets[] = {"Mercury", "Venus", "Earth", ...};
   ```
   Each element is a pointer (`char *`) to a string literal (or dynamically allocated string). Strings can have arbitrary lengths without wasted space, requiring only additional pointer storage for each row.
