# Agent & Developer Development Guide

This guide defines the version control procedures, CI requirements, and architectural standards for contributing to the `qst` codebase. Follow this workflow before creating commits or pushing to remote repositories.

---

## 1. Pre-Push Verification Pipeline

The repository uses GitHub Actions CI ([`.github/workflows/go.yml`](file:///home/speer/Code/golang/qst/.github/workflows/go.yml)) which enforces static analysis, linting, documentation parity, and test coverage.

### One-Shot Verification Command
Run this command in the repository root prior to pushing:

```bash
go build -v ./... && \
errcheck ./... && \
go vet ./... && \
go run ./internal/tools/docgen -out ./docs/cli -format markdown && \
git diff --exit-code docs/cli/ && \
go test -count=1 ./...
```

### Pipeline Steps Explained

1. **Compilation Check**:
   ```bash
   go build -v ./...
   ```
   Ensures all packages and tools compile without warnings.

2. **Unhandled Error Detection (`errcheck`)**:
   ```bash
   errcheck ./...
   ```
   *Crucial for CI*: All functions returning errors (especially `fmt.Fprintf` writing to `io.Writer` or `util.BufWriter`) must have their errors checked or handled. Neglecting this will fail remote CI.

3. **Go Vet**:
   ```bash
   go vet ./...
   ```
   Catches suspicious constructs and formatting bugs.

4. **CLI Documentation Parity**:
   ```bash
   go run ./internal/tools/docgen -out ./docs/cli -format markdown
   git diff --exit-code docs/cli/
   ```
   *Mandatory when modifying `cmd/`*: Whenever commands, flags, aliases, or descriptions are added or altered, the CLI documentation in `docs/cli/` must be regenerated and committed. If `git diff --exit-code docs/cli/` produces any output, CI will fail.

5. **Full Test Suite Execution**:
   ```bash
   go test -count=1 ./...
   ```
   Runs all unit tests, graph validation suites, and archive roundtrips without relying on cached results.

---

## 2. Git & Commit Guidelines

1. **Conventional Commits**:
   Prefix commit messages with standard conventional types and scopes:
   - `feat(cmd): ...` (New commands, flags, user-facing behavior)
   - `feat(graph): ...` (Graph parser, validator, mutations)
   - `feat(ssg): ...` (Static site generator logic, templates, assets)
   - `fix(ssg): ...` (Bug fixes, linter resolutions)
   - `refactor(examples): ...` (Manifest updates, example corrections)
   - `docs: ...` (Documentation and docgen output updates)
   - `chore: ...` (Build configs, `.gitignore`, dependency updates)

2. **Signed Commits (`-s`)**:
   All commits **must be signed-off** using `git commit -s`. This automatically appends the `Signed-off-by:` trailer matching repository history.

3. **Commit Body Style**:
   Keep commit bodies **minimal or empty** (subject line + sign-off line is standard). Avoid lengthy changelogs in commit bodies unless strictly required for breaking changes.

4. **Atomic Commits**:
   Keep commits focused by subsystem (`cmd/`, `internal/graph/`, `internal/ssg/`, `docs/cli/`). Do not bundle unrelated features together.

---

## 3. Knowledge Graph Constraints & Examples

Any modifications to knowledge graphs in [`examples/`](file:///home/speer/Code/golang/qst/examples/) must satisfy the following:

1. **Manifest File**:
   - Every graph must contain a `manifest.yaml` (or fallback `quest.yaml`).
   - Required fields: `title`, `scopes`, `adherences`, and `tours`.
2. **Guide Frontmatter**:
   - Markdown files must have valid frontmatter (`scope`, `prerequisites`, `sub_guides`, `tags`).
   - Horizontal dependencies (`prerequisites`) must have identical scopes.
   - Vertical dependencies (`sub_guides`) must adhere to scope hierarchy rules (child scope strictly smaller than parent scope).
   - If `segment` line ranges are defined, they must fit within the parent guide's body line count.
3. **Acyclic Graphs**:
   - Graphs must remain strictly acyclic (checked via `graph.CheckAcyclic`).
4. **Asset Exclusions**:
   - Do **not** commit generated directories (`public/`, `examples/*/public/`).
   - Do **not** commit large reference PDFs (`*.pdf`) or packaged archives (`*.kng`).

---

## 4. CLI Architecture Best Practices

When adding or modifying Cobra commands in [`cmd/`](file:///home/speer/Code/golang/qst/cmd/):

1. **Standardized Output Channels**:
   - Use `cmd.Printf` / `cmd.Println` for standard user output.
   - Use `cmd.PrintErrf` / `cmd.PrintErrln` for error messages.
   - Avoid uncaptured `fmt.Printf` so command outputs can be validated in unit tests via `cmd.SetOut()` and `cmd.SetErr()`.

2. **Exit Handling in Tests**:
   - Use `exitFunc(1)` (defined via `var exitFunc = os.Exit` in [`cmd/root.go`](file:///home/speer/Code/golang/qst/cmd/root.go)) instead of raw `os.Exit(1)`.
   - This allows unit tests in [`cmd/commands_test.go`](file:///home/speer/Code/golang/qst/cmd/commands_test.go) to intercept exit codes using `executeCommandWithExit` without killing the test runner process.

3. **Flag State Resetting**:
   - Commands should check `cmd.Flags().Changed("flagname")` when determining whether optional filters or overrides were provided by the user.

---

## 5. Archiving & Packaging Standards (`.kng`)

When working with graph archives ([`internal/graph/archive.go`](file:///home/speer/Code/golang/qst/internal/graph/archive.go)):
- Package files losslessly using `zip.Deflate`.
- Normalize zip entry path separators using `filepath.ToSlash(relPath)` for cross-platform compatibility.
- Ensure `manifest.yaml`, diagrams (`.mmd`), and images are preserved while ignoring `public/`, hidden files (`.*`), and `.kng` files.
- Always enforce path validation (`strings.HasPrefix`) in `UnarchiveGraph` to prevent Zip Slip directory traversal vulnerabilities.
