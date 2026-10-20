## qst coverage

Visualize line-by-line subguide coverage of a guide's body

### Synopsis

Coverage displays the markdown body of a guide with 1-based line numbers prefixed.
Each line is annotated with the subguide that covers it, highlighting unmapped and overlapping lines.
This is a direct shortcut for 'qst inspect <guide> --coverage'.

```
qst coverage [guide] [flags]
```

### Examples

```
# Inspect body line coverage
qst coverage "Variables and Assignment"

# Inspect specifying file name or path
qst coverage "Variables and Assignment.md"
qst coverage examples/C_programming/The\ Compilation\ Process.md

# Inspect in specific directory
qst coverage "The Compilation Process" -d examples/C_programming
```

### Options

```
  -d, --dir string   Knowledge graph directory (default ".")
  -h, --help         help for coverage
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

