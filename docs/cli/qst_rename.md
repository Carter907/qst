## qst rename

Rename a guide and update all references to it

```
qst rename [old_name] [new_name] [flags]
```

### Examples

```
# Rename a guide in the graph
qst rename "Old Name" "New Name" -d my_graph/

# Rename specifying file names or paths
qst rename "Old Name.md" "New Name.md"
qst rename my_graph/Old\ Name.md "New Name"
```

### Options

```
  -d, --dir string   Knowledge graph directory (default ".")
  -h, --help         help for rename
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

