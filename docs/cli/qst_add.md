## qst add

Add a new guide to the knowledge graph.

### Synopsis

Add allows you to insert a guide into the knowledge graph by specifying its prerequisites, and scope. A new markdown file will be inserted into the target directory.

```
qst add [name] [flags]
```

### Examples

```
# Add a new definition
qst add Exponent --scope definition

# Add to a specific directory
qst add Exponent --dir my_graph/
```

### Options

```
  -d, --dir string              Knowledge graph directory (default ".")
  -h, --help                    help for add
  -i, --interactive             Interactive mode
      --prerequisites strings   List of prerequisites
      --scope string            Scope of the guide (e.g. definition, description)
      --tags strings            List of tags
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

