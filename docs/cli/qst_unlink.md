## qst unlink

Unlink a subguide or prerequisite from an existing guide.

### Synopsis

Unlink allows you to sever relationships between guides in the knowledge graph.
- Use --guide to remove a subguide relation.
- Use --prereq to remove a prerequisite relation.
- Use --interactive (-i) to choose from existing links to remove.

```
qst unlink [guide] [flags]
```

### Examples

```
# Unlink a subguide
qst unlink "The Compilation Process" --guide "Compiler"

# Unlink a prerequisite
qst unlink "Scanf Function" --prereq "Printf Function"

# Interactive link removal
qst unlink "Variables and Assignment" -i
```

### Options

```
  -d, --dir string      Knowledge graph directory (default ".")
      --guide string    Name of the subguide to unlink
  -h, --help            help for unlink
  -i, --interactive     Interactive mode with guided selection
      --prereq string   Name of the prerequisite guide to unlink
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

