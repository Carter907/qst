## qst add

Add a new guide to the knowledge graph.

### Synopsis

Add allows you to insert a guide into the knowledge graph by specifying its prerequisites, subguides, and scope. A new markdown file will be inserted into the target directory.

```
qst add [name] [flags]
```

### Examples

```
# Add a new definition
qst add Exponent --scope definition

# Add a description with subguides and adherence
qst add "String Operations" --scope description --subguide "String Length, String Copy" --adherence introductory

# Add to a specific directory with initial body content
qst add "Memory Layout" --dir my_graph/ --scope explanation --body "## Memory Layout\n\nExplanation text."
```

### Options

```
      --adherence string        Adherence of the subguide relation
      --body string             Initial markdown body for the guide
  -d, --dir string              Knowledge graph directory (default ".")
      --file string             Path to file containing initial markdown body
  -h, --help                    help for add
  -i, --interactive             Interactive mode
      --prerequisites strings   List of prerequisites
      --scope string            Scope of the guide (e.g. definition, description)
      --segment string          Segment line range in body (e.g. 4-4)
      --sub strings             Alias for --subguide
  -s, --subguide strings        List of subguides to link
      --tags strings            List of tags
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

