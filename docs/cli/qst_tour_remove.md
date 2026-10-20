## qst tour remove

Remove a guide from a tour in manifest.yaml

```
qst tour remove [tour-name] [guide] [flags]
```

### Examples

```
# Remove a guide from a tour
qst tour remove "Core Fundamentals" "Concept A" -d my_graph/

# Remove a guide specifying file name or path
qst tour remove "Core Fundamentals" "Concept A.md" -d my_graph/
qst tour remove "Core Fundamentals" my_graph/Concept\ A.md
```

### Options

```
  -h, --help   help for remove
```

### Options inherited from parent commands

```
  -d, --dir string   Knowledge graph directory (default ".")
```

### SEE ALSO

* [qst tour](qst_tour.md)	 - Manage curriculum tours in the knowledge graph

