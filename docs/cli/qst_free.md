## qst free

Unzip a .kng file so you can read or edit the knowledge graph

### Synopsis

Free is how you start learning new knowledge. Unzipped archives are placed in a directory with the same name as the file (or a specified destination). All markdown guides, manifests, and assets are losslessly decompressed.

```
qst free <archive.kng> [destination] [flags]
```

### Examples

```
# Free the Knowledge to default directory
qst free my-graph.kng

# Free to a custom destination directory
qst free my-graph.kng custom_dir
qst free my-graph.kng -o custom_dir
```

### Options

```
  -h, --help            help for free
  -o, --output string   Destination directory for unpacked graph (default: <archive-name>)
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

