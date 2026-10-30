## qst serve

Start a live-reloading local development server for the knowledge graph

### Synopsis

Start a local HTTP development server that compiles the knowledge graph into an isolated temporary workspace, watches for file changes, validates graph constraints in real time, and automatically reloads connected browsers via Server-Sent Events (SSE).

You can pass either a knowledge graph directory or a packaged .kng archive file.
If the requested port is already in use, serve automatically increments the port until an available port is found.

```
qst serve [directory|archive.kng] [flags]
```

### Examples

```
# Serve the current directory on default port 8080
qst serve

# Serve a specific directory on port 3000 and open the browser
qst serve ./examples/Boolean_Logic --port 3000 --open

# Serve a packaged .kng archive directly in offline mode
qst serve bundle.kng --offline
```

### Options

```
  -h, --help                help for serve
      --host string         Host address to bind to (default "localhost")
      --offline             Serve without downloading external vendor assets (KaTeX, Mermaid)
  -O, --open                Open the default browser automatically on startup
  -p, --port int            Port to listen on (automatically probes next available port if busy) (default 8080)
      --vendor-dir string   Custom directory containing vendor assets (katex, mermaid)
```

### SEE ALSO

* [qst](qst.md)	 - A lossless compression and formatting tool for knowledge

