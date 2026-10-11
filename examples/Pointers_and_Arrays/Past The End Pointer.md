---
prerequisites: []
sub_guides: []
scope: definition
tags: [c, pointers, arrays]
---

## Past The End Pointer

A **Past The End Pointer** is a pointer positioned one index past the final element of an array (`&a[N]`), which C permits for comparison as a loop sentinel, though dereferencing it is undefined behavior.
