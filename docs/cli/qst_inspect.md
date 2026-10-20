## qst inspect

Display metadata, scope, relations, and coverage for a guide

### Synopsis

Inspect displays a concise summary of a guide's scope, tags, prerequisites, subguides, coverage, and body statistics.
Use --coverage (-c) to expand the line-by-line body annotation.

```
qst inspect [guide] [flags]
```

### Examples

```
# Inspect a guide in the current directory
qst inspect "Variables and Assignment"

# Inspect specifying file name or path
qst inspect "Variables and Assignment.md" -c
qst inspect examples/C_programming/Scanf\ Function.md -c

# Inspect a guide in a specific directory
qst inspect "Scanf Function" -d examples/C_programming -c
```

### Options

```
  -c, --coverage     Show detailed line-by-line body coverage
  -d, --dir string   Knowledge graph directory (default ".")
  -h, --help         help for inspect
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

