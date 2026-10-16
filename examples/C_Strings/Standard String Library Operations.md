---
prerequisites: [String Input and Output Operations]
sub_guides:
    - guide: The string.h Header
      adherence: introductory
      segment: 4-4
    - guide: strlen Function
      adherence: detailed
      segment: 6-6
    - guide: strcpy Function
      adherence: detailed
      segment: 8-8
    - guide: strcat Function
      adherence: detailed
      segment: 10-10
    - guide: strcmp Function
      adherence: detailed
      segment: 12-12
scope: description
tags: [c, strings, stdlib]
---

## Standard String Library Operations

The primary functions for text manipulation are declared in **The string.h Header**, which mandates that passed parameters point to valid null-terminated character sequences.

Measuring string size is performed using the **strlen Function**, which returns the number of characters preceding the null byte in O(N) linear time.

Assigning and copying text buffers is handled by the **strcpy Function**, which copies bytes from source to destination until the null terminator is transferred.

Combining multiple strings end-to-end is accomplished through the **strcat Function**, which appends source characters directly onto the end of an existing destination string.

Evaluating whether two strings are identical or determining their dictionary order relies on the **strcmp Function**, which performs unsigned byte-by-byte lexicographical comparison.
