## qst form

zip a knowledge graph directory into the .kng archive file format

### Synopsis

Validates the formatting constraints of the knowledge graph and packages it losslessly into a compressed .kng archive.

```
qst form [directory] [flags]
```

### Examples

```
# From a directory
qst form my_knowledge

# From Current Directory
qst form

# Specify custom output path
qst form my_knowledge -o dist/bundle.kng
```

### Options

```
  -h, --help            help for form
  -o, --output string   Output path for the .kng archive (default: <directory-name>.kng)
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

