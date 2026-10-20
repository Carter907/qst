## qst tour add

Add one or more guides to a tour in manifest.yaml

```
qst tour add [tour-name] [guide...] [flags]
```

### Examples

```
# Add guides to a tour
qst tour add "Core Fundamentals" "Concept A" "Concept B" -d my_graph/

# Add guides specifying file names or paths
qst tour add "Core Fundamentals" "Concept A.md" "Concept B.md" -d my_graph/
qst tour add "Core Fundamentals" my_graph/Concept\ A.md
```

### Options

```
  -h, --help   help for add
```

### Options inherited from parent commands

```
  -d, --dir string   Knowledge graph directory (default ".")
```

### SEE ALSO

* [qst tour](qst_tour.md)	 - Manage curriculum tours in the knowledge graph

