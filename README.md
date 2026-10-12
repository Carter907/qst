# `qst`: A Command-line Knowledge System

`qst` (Quest) is a command-line tool, graph validator, and static site generator for curating, structuring, and distributing knowledge graphs. 

In `qst`, knowledge is written as Markdown **Guides** connected in a **Directed Acyclic Graph (DAG)**. Concepts connect horizontally through prerequisite dependencies and vertically through hierarchical subguide decompositions. `qst` provides tools to author, validate, analyze, package, and publish these resource graphs as standalone static learning sites.

---

## Key Capabilities

- **DAG Knowledge Modeling**: Enforces acyclic graphs, strict scope hierarchies, and horizontal/vertical relationship constraints with deterministic cycle logging.
- **Academic Static Site Generator (`qst build`)**: Generates responsive, standalone documentation sites that can be shared or deployed as pedagogical resources for learners.
- **Graph Health & Analytics (`qst stats`)**: Measures graph depth (longest prerequisite chains), scope distributions, tour coverage, and aggregate body line coverage metrics.
- **Line Coverage & Auditing (`qst inspect`, `qst coverage`)**: Validates that parent guide texts accurately encompass their subguides with line-by-line coverage breakdowns.
- **Relationship Management (`qst link`, `qst unlink`, `qst rename`)**: Effortlessly connect, disconnect, or rename guides with automated reference cascading across your entire knowledge graph.
- **Lossless Archiving (`qst form`, `qst free`)**: Distribute and unpack knowledge graphs via `.kng` archives preserving manifests, nested directories, and media assets.

---

## Installation

### Prerequisites
- Go 1.22 or later

### Install via `go install`
```sh
go install github.com/Carter907/qst@latest
```

### Install from Source
```sh
git clone https://github.com/Carter907/qst.git
cd qst
go build -o $(go env GOPATH)/bin/qst .
```

For platform-specific instructions, see [installation.md](docs/installation.md).

---

## Quickstart

Get a knowledge graph running in under two minutes:

```sh
# 1. Initialize a new knowledge graph
qst new Calculus
cd Calculus

# 2. Add your first guide
qst add "Limits" --scope definition --tags math,calculus

# 3. Add a dependent guide and link prerequisites
qst add "Derivatives" --scope definition --tags math,calculus
qst link "Derivatives" --prereq "Limits"

# 4. Check graph structure and verify acyclicity
qst validate

# 5. Compile to a static site
qst build . --out ./public

# 6. Package for distribution as a .kng archive
qst form
```

---

## CLI Command Reference

`qst` includes a comprehensive command suite. Full documentation for each command is available in [the CLI docs](docs/cli/):

| Command | Description | Documentation |
| :--- | :--- | :--- |
| `qst new <name>` | Create a new knowledge graph directory with starter configuration | [docs/cli/qst_new.md](docs/cli/qst_new.md) |
| `qst add <title>` | Add a new guide with frontmatter to the knowledge graph | [docs/cli/qst_add.md](docs/cli/qst_add.md) |
| `qst link <guide>` | Link a prerequisite or subguide to an existing guide | [docs/cli/qst_link.md](docs/cli/qst_link.md) |
| `qst unlink <guide>` | Unlink a prerequisite or subguide from an existing guide | [docs/cli/qst_unlink.md](docs/cli/qst_unlink.md) |
| `qst rename <old> <new>` | Rename a guide and automatically update all inbound references | [docs/cli/qst_rename.md](docs/cli/qst_rename.md) |
| `qst list` | List all guides (with optional `-l` aligned tabular details) | [docs/cli/qst_list.md](docs/cli/qst_list.md) |
| `qst inspect <guide>` | Display metadata, scope, relations, and coverage for a guide | [docs/cli/qst_inspect.md](docs/cli/qst_inspect.md) |
| `qst coverage <guide>` | Visualize line-by-line subguide coverage across a guide body | [docs/cli/qst_coverage.md](docs/cli/qst_coverage.md) |
| `qst view` | Export the knowledge graph to a Mermaid.js diagram in terminal | [docs/cli/qst_view.md](docs/cli/qst_view.md) |
| `qst stats` | Display graph metrics, depth, tour coverage, and scope distributions | [docs/cli/qst_stats.md](docs/cli/qst_stats.md) |
| `qst validate` | Validate the knowledge directory for cycles and constraint errors | [docs/cli/qst_validate.md](docs/cli/qst_validate.md) |
| `qst build <dir>` | Compile the knowledge graph into a static HTML documentation site | [docs/cli/qst_build.md](docs/cli/qst_build.md) |
| `qst form` | Package a knowledge graph directory into a `.kng` archive | [docs/cli/qst_form.md](docs/cli/qst_form.md) |
| `qst free <archive.kng>` | Extract a `.kng` archive into an editable knowledge graph folder | [docs/cli/qst_free.md](docs/cli/qst_free.md) |

---

## Documentation

- **[Usage Guide](docs/usage.md)**: End-to-end tutorial covering graph authoring, validation, inspection, and compilation.
- **[File Format & Schema](docs/format.md)**: Manifest configuration, frontmatter schemas, adherence levels, and scopes.
- **[Concepts & Theory](docs/concepts.md)**: Theoretical foundations of curriculum graphs, normalization, and pedagogical tours.
- **[Installation](docs/installation.md)**: Installation methods, building from source, and shell configuration.
- **[CLI Reference](docs/cli/)**: Full documentation for each Cobra command and flag.

---

## License

This project uses a strictly copy-left, free-software license. see [LICENSE](/LICENSE)
