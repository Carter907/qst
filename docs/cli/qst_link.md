## qst link

Link a subguide or prerequisite to an existing guide.

### Synopsis

Link allows you to connect guides together in the knowledge graph.
- Use --guide to link a hierarchical subguide (with --adherence and optional --segment).
  If the subguide is already linked, its adherence and segment will be updated in place.
- Use --prereq to link a horizontal prerequisite of the exact same scope.
- Use --interactive (-i) for interactive guided selection with constraint checking.

```
qst link [guide] [flags]
```

### Examples

```
# Link a subguide with adherence and line segment
qst link "The Compilation Process" --guide "Compiler" --adherence introductory --segment 6-6

# Link a prerequisite of identical scope
qst link "Scanf Function" --prereq "Printf Function"

# Interactive link selection
qst link "Variables and Assignment" -i
```

### Options

```
      --adherence string   Adherence of the subguide relation
  -d, --dir string         Knowledge graph directory (default ".")
      --guide string       Name of the subguide to link
  -h, --help               help for link
  -i, --interactive        Interactive mode with guided selection
      --prereq string      Name of the prerequisite guide to link
      --segment string     Segment line range in body (e.g. 4-4)
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

