---
scope: primitive
prerequisites:
  - Server-Sent Events Hub and Middleware Injection
sub_guides: []
tags: [ssg, server, fsnotify, concurrency]
---

# Recursive fsnotify Watcher and RWMutex Build Lock

Operating system file notification APIs (`inotify` on Linux, `kqueue` on macOS, `ReadDirectoryChangesW` on Windows) exposed via `github.com/fsnotify/fsnotify` are non-recursive by default and emit bursts of events when editors perform atomic save-and-rename operations. `internal/ssg/server.go` wraps `fsnotify` with recursive directory registration, timer debouncing, selective output cleaning, and read-write mutex synchronization.

## Recursive and Dynamic Directory Registration

1. **Startup Walk (`addWatchRecursive`)**: Walks `srcDir` and registers every non-hidden directory (skipping `.git`, `public`, `vendor`, and `assets`).
2. **Dynamic Subdirectory Registration**: When `watchLoop` observes an `fsnotify.Create` event for a newly created directory, it immediately calls `addWatchRecursive(event.Name)` so files created inside new subfolders are tracked without restarting `qst serve`.
3. **Swap File Filtering & `150ms` Debounce**: Events on hidden files (`.*`), editor backup/swap files (`*~`, `*.swp`, `*.swo`, `*.tmp`), and `Chmod`-only operations are ignored. Valid write/create/remove/rename events reset a `150ms` `time.Timer`, coalescing multi-file saves into a single rebuild.

## Selective Output Cleaning (`cleanGeneratedOutput`)

Deleting the entire temporary build directory on every save would force `EnsureVendorAssets` to re-copy megabytes of KaTeX fonts and Mermaid bundles on every keystroke. Conversely, never cleaning `outDir` would leave orphaned `.html` files behind when a guide is renamed or deleted. `cleanGeneratedOutput` removes all generated `.html` and media files in `outDir` while preserving `outDir/assets/vendor`.

## Atomic Build Synchronization (`sync.RWMutex`)

`DevServer.buildMu` coordinates concurrent rebuilds and HTTP requests:
* `rebuild()` acquires `s.buildMu.Lock()` (write lock) during the ~2ms HTML generation pass.
* `handleStatic()` acquires `s.buildMu.RLock()` (read lock) while reading `.html` files, guaranteeing that a browser never receives a truncated or partially written document.
