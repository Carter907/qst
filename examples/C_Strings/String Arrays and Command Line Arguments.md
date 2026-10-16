---
prerequisites: [Safe String Handling and Bounded Operations]
sub_guides:
    - guide: Arrays of Strings
      adherence: detailed
      segment: 4-4
    - guide: Command Line Arguments
      adherence: detailed
      segment: 6-6
scope: description
tags: [c, strings, cli, arrays]
---

## String Arrays and Command Line Arguments

Collections of text are represented using **Arrays of Strings**, either as fixed 2D character matrices with uniform dimensions or as ragged arrays of pointers to individual character arrays.

The ragged pointer array architecture forms the exact mechanism used to receive **Command Line Arguments** in main via the argc count and argv vector.
