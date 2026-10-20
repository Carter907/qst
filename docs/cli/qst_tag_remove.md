## qst tag remove

Remove one or more tags from a guide

```
qst tag remove [guide] [tag...] [flags]
```

### Examples

```
# Remove a tag from a guide
qst tag remove "Printf Function" "io" -d examples/C_programming

# Remove a tag specifying file name or path
qst tag remove "Printf Function.md" "io" -d examples/C_programming
qst tag remove examples/C_programming/Printf\ Function.md "io"
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

* [qst tag](qst_tag.md)	 - Manage taxonomy tags on guides in the knowledge graph

