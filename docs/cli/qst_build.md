## qst build

Compile the knowledge graph into a static HTML website

### Synopsis

Compile the knowledge graph into a static HTML website.

By default, the compiler fetches KaTeX and Mermaid vendor assets over HTTP and caches them in your local user cache directory (~/.cache/qst/vendor).
Use --offline to skip all network requests and compile using local assets.
Use --vendor-dir or the QST_VENDOR_DIR environment variable to provide custom pre-downloaded vendor assets.

```
qst build [directory] [flags]
```

### Examples

```
# Standard build with online assets
qst build . --out ./public

# Offline build (reuses cached assets or builds without external downloads)
qst build . --out ./public --offline

# Offline build using a specific vendor assets directory
qst build . --out ./public --offline --vendor-dir /path/to/vendor
```

### Options

```
  -h, --help                help for build
      --offline             Build site without downloading external vendor assets (KaTeX, Mermaid)
  -o, --out string          Output directory for the compiled site (default "public")
      --vendor-dir string   Custom directory containing vendor assets (katex, mermaid)
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

