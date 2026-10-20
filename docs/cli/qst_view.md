## qst view

Export the knowledge graph to a Mermaid.js diagram

### Synopsis

The view command reads the knowledge directory and exports its structure as a Mermaid.js graph.
Prerequisite links are shown as solid arrows, and sub-guide relationships are shown as dotted arrows.
Use --guide (-g) to export a localized forward-only subgraph centered on a specific guide.

```
qst view [directory] [flags]
```

### Examples

```
# Output global mermaid to stdout
qst view my_knowledge_graph_dir/

# Output localized subgraph for a specific guide
qst view . --guide "Printf Function"

# Output localized subgraph specifying file name or path
qst view . --guide "Printf Function.md"
qst view -g examples/C_programming/Printf\ Function.md

# Save to a markdown file
qst view . > graph.md
```

### Options

```
  -g, --guide string   Export localized subgraph centered on a specific guide
  -h, --help           help for view
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

