## qst delete

Delete a guide and cascade removal of all references to it

### Synopsis

Delete removes a guide file from the knowledge graph and automatically prunes
all incoming references to it from other guides' prerequisites, subguides, and manifest tours.

```
qst delete [guide] [flags]
```

### Examples

```
# Delete a guide and clean up all references across the graph
qst delete "Old Concept" -d my_graph/

# Preview which references would be pruned without deleting
qst delete "Old Concept" -d my_graph/ --dry-run

# Using the rm alias
qst rm "Old Concept"
```

### Options

```
  -d, --dir string   Knowledge graph directory (default ".")
      --dry-run      Preview references that would be pruned without deleting
  -h, --help         help for delete
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

