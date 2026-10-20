## qst tag add

Add one or more tags to a guide

```
qst tag add [guide] [tag...] [flags]
```

### Examples

```
# Add tags to a guide
qst tag add "Printf Function" "io" "c-standard" -d examples/C_programming

# Add tags specifying file name or path
qst tag add "Printf Function.md" "io" -d examples/C_programming
qst tag add examples/C_programming/Printf\ Function.md "io"
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

* [qst tag](qst_tag.md)	 - Manage taxonomy tags on guides in the knowledge graph

