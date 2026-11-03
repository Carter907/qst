---
scope: primitive
prerequisites:
  - Guide and Manifest Data Models
sub_guides: []
tags: [graph, archive, security, zip]
---

# Deflate Zip Packaging and Zip Slip Guard

Knowledge graphs in `qst` can be packaged into portable, single-file `.kng` archives via `qst form` and extracted via `qst free` (or served directly via `qst serve`). The implementation in `internal/graph/archive.go` uses standard Deflate compression with strict entry filtering and path traversal protection.

## Archive Entry Filtering (`shouldArchiveEntry`)

When `ArchiveGraph` walks a knowledge graph directory, it skips hidden directories (`.*`) and compiled output (`public/`), and filters files through an explicit whitelist:
* **Manifests**: `manifest.yaml`, `manifest.yml`, `quest.yaml`, `quest.yml`.
* **Guides & Diagrams**: `.md`, `.mmd`, `.mermaid`.
* **Static Media Assets**: `.png`, `.jpg`, `.jpeg`, `.svg`, `.gif`, `.webp`.
* **Excluded**: Hidden files (`.*`) and nested `.kng` archives (preventing recursive archive bloat).

## Cross-Platform Zip Header Normalization

Windows uses backslashes (`\`) for `filepath.Rel`, which violates the ZIP specification if written directly into archive entry headers. `ArchiveGraph` normalizes every entry path to forward slashes and forces `zip.Deflate`:

```go
header.Name = filepath.ToSlash(rel)
header.Method = zip.Deflate
```

## Zip Slip Directory Traversal Protection (`UnarchiveGraph`)

When extracting a `.kng` archive (or a KaTeX `.tar.gz` bundle in `internal/ssg/download.go`), a maliciously crafted archive entry like `../../etc/cron.d/pwn` could escape the destination directory. `UnarchiveGraph` neutralizes Zip Slip before creating any directory or file:

```go
cleanDest := filepath.Clean(destDir)
targetPath := filepath.Join(cleanDest, f.Name)
cleanTarget := filepath.Clean(targetPath)

if cleanTarget != cleanDest && !strings.HasPrefix(cleanTarget, cleanDest+string(os.PathSeparator)) {
	return fmt.Errorf("illegal file path in archive: %s", f.Name)
}
```

Appending `string(os.PathSeparator)` to `cleanDest` is critical: without the trailing separator, a sibling directory with a shared prefix (e.g., `/tmp/graph-evil` when `cleanDest` is `/tmp/graph`) would falsely pass `strings.HasPrefix`.
