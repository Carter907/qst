## qst path

Generate a topological learning path for a guide or the entire graph

### Synopsis

Path computes an optimal, prerequisite-satisfying sequence to learn a specific guide
or the entire knowledge graph using topological sorting.

By default, learning paths resolve both horizontal prerequisites (same scope) and
vertical subguide components (lower scopes), ensuring foundational building blocks
are mastered before high-level topics.

When no guide argument is supplied, a complete curriculum for the entire knowledge
graph is generated and grouped by manifest scopes.

```
qst path [guide] [flags]
```

### Examples

```
# Discover learning path for a specific guide in current directory
qst path "Variables and Assignment"

# Discover learning path in a specific knowledge graph directory
qst path "Chapter 2 C Fundamentals" -d examples/C_programming

# Output learning path as a Mermaid diagram
qst path "Variables and Assignment" -d examples/C_programming --format mermaid

# Discover what topics a foundational guide unlocks
qst path "Variable" -d examples/C_programming --unlocks

# Generate full syllabus for the entire knowledge graph
qst path -d examples/C_programming

# Save the resulting sequence directly as an official tour in manifest.yaml
qst path "Chapter 2 C Fundamentals" -d examples/C_programming --save-tour "Chapter 2 Learning Path"
```

### Options

```
  -d, --dir string         Knowledge graph directory (default ".")
      --direct             Only include immediate 1-hop prerequisites
      --format string      Output format: text, json, mermaid (default "text")
  -h, --help               help for path
      --prereqs-only       Only follow horizontal prerequisites, ignoring subguides
      --save-tour string   Save the resulting path as a tour in manifest.yaml
      --unlocks            Discover downstream guides unlocked by the target
  -v, --verbose            Show dependency reasons and requirements per step
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

