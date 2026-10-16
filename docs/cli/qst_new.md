## qst new

create a new knowledge graph directory

### Synopsis

You start a new knowledge graph with the new command. You only have to specify the name of the knowledge graph; a directory will be created. If you don't specify a directory, the current directory will be used. You can optionally include a starter guide template.

```
qst new [flags]
```

### Examples

```
# Specify New Directory
qst new learn-cpp

# Use Current Directory
qst new

# Create with starter guide
qst new --starter
```

### Options

```
  -h, --help              help for new
  -s, --starter           Initialize with a starter markdown guide
      --strict-coverage   Enable strict coverage mode in manifest.yaml
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

