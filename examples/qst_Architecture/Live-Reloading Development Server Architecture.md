---
scope: component
prerequisites:
  - Static Site Compilation and Academic Layout
sub_guides:
  - guide: Server-Sent Events Hub and Middleware Injection
    adherence: strict
    segment: 6-10
  - guide: Recursive fsnotify Watcher and RWMutex Build Lock
    adherence: detailed
    segment: 12-18
tags: [ssg, server, live-reload, dev-server]
---

# Live-Reloading Development Server Architecture

In `internal/ssg/server.go`, `DevServer` provides an isolated, live-reloading local authoring environment (`qst serve`) that combines instant HTML previews with real-time graph validation.

## Ephemeral Isolation, SSE Streaming, and Resilient Startup

When `NewDevServer` initializes, it creates an ephemeral build directory via `os.MkdirTemp("", "qst-serve-build-*")` so development previews never pollute the repository's working tree or overwrite `./public`. If the target is a `.kng` archive, it also creates an ephemeral extraction directory (`qst-serve-kng-*`) and unpacks the archive via `graph.UnarchiveGraph`.

If the requested TCP port (default `8080`) is occupied, `bindListener` automatically probes subsequent ports (`8081..8089`). Even if the initial `manifest.yaml` or guide frontmatter contains a syntax error at startup, `NewDevServer` starts resiliently, writing a fallback `index.html` and streaming the active error over `/_qst/events` so saving a fix in the editor automatically recovers the browser view.

## Continuous Rebuild and Validation Lifecycle

When `watchLoop` detects a debounced filesystem change, `rebuild(true)` executes a four-stage transactional pipeline under `s.buildMu.Lock()`:
1. **Manifest Guard (`graph.ParseConfig`)**: Validates `manifest.yaml` before touching `outDir` so a broken manifest edit leaves the last known good HTML build intact while broadcasting a `build_error` overlay.
2. **Selective Cleanup (`cleanGeneratedOutput`)**: Purges stale `.html` and media files while preserving `outDir/assets/vendor`.
3. **Fast HTML Compilation (`BuildSiteWithOptions`)**: Recompiles Markdown guides with `SkipVendor: true` in a few milliseconds.
4. **Live Graph Validation (`graph.ParseGraph` + `graph.ValidateGraph`)**: If structural or frontmatter rules fail, the browser still reloads to show the updated Markdown content *and* displays the validation diagnostics in the non-blocking SSE overlay.
