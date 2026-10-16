## qst link

Link a subguide or prerequisite to an existing guide.

### Synopsis

Link allows you to connect guides together in the knowledge graph.
- Use --subguide (or --sub, -s) to link one or more hierarchical subguides (with --adherence and optional --segment).
  If a subguide is already linked, its adherence and segment will be updated in place.
- Use --prereq to link one or more horizontal prerequisites of the exact same scope.
- Use --preview (-p) to preview numbered body lines for accurate segment range selection.
- Use --interactive (-i) for interactive guided selection with constraint checking.

```
qst link [guide] [flags]
```

### Examples

```
# Link a subguide with adherence and line segment
qst link "The Compilation Process" --subguide "Compiler" --adherence introductory --segment 6-6

# Link multiple subguides at once
qst link "Chapter 13 Strings" --subguide "String Fundamentals, String IO" --adherence introductory

# Link a prerequisite of identical scope
qst link "Scanf Function" --prereq "Printf Function"

# Preview body lines before linking
qst link "Variables and Assignment" --preview

# Interactive link selection
qst link "Variables and Assignment" -i
```

### Options

```
      --adherence string   Adherence of the subguide relation
  -d, --dir string         Knowledge graph directory (default ".")
  -h, --help               help for link
  -i, --interactive        Interactive mode with guided selection
      --prereq strings     Name(s) of the prerequisite guide(s) to link
  -p, --preview            Preview numbered body lines of the guide
      --segment string     Segment line range in body (e.g. 4-4)
      --sub strings        Alias for --subguide
  -s, --subguide strings   Name(s) of the subguide(s) to link
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

