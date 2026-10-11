---
prerequisites: []
sub_guides:
    - guide: Row Major Order
      adherence: introductory
      segment: 4-4
scope: description
tags: [c, arrays, memory, multidimensional]
---

## Multidimensional Array Layout and Traversal

Because 2D matrices follow contiguous **Row Major Order** in memory, code can traverse all matrix cells linearly with a single simple pointer iterating from `&a[0][0]` up through the final element.
