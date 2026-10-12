# Usage Guide: End-to-End Knowledge Graph Workflow

This guide covers the complete end-to-end lifecycle of managing a knowledge graph with `qst`, from initialization to publication.

---

## 1. Initializing a Knowledge Graph (`qst new`)

To start a new knowledge graph, use `qst new`:

```sh
qst new Calculus
cd Calculus
```

This creates a new project directory populated with a starter guide and a `manifest.yaml` configuration file.

### Manifest Configuration
The `manifest.yaml` file defines the rules and taxonomy of your knowledge graph:

```yaml
title: Calculus Knowledge Graph
description: A structured curriculum of single-variable and multivariable calculus.
scopes:
  - definition
  - description
  - explanation
  - lesson
adherences:
  - strict
  - detailed
  - introductory
  - vague
relaxed_subguides: false
require_subguides: false
strict_coverage: false
tours:
  - name: Core Foundations
    guides:
      - Limits
      - Derivatives
```

- **`scopes`**: Ordered list of conceptual depth levels, from smallest (`definition`) to largest (`lesson`).
- **`adherences`**: Valid levels of fidelity for how deeply a parent guide rewrites or covers a subguide.
- **`relaxed_subguides`**: When `false`, subguides must strictly sit one scope rung below their parent.
- **`require_subguides`**: When `true`, guides above the base scope must define at least one subguide.
- **`strict_coverage`**: When `true`, every line of a guide's body must map to at least one subguide segment.
- **`tours`**: Curated, linear reading paths for learners following a sequential trajectory.

*(Note: `quest.yaml` is also supported as an alternate configuration filename).*

---

## 2. Adding and Organizing Guides (`qst add`)

Guides are individual Markdown files containing frontmatter and instructional content.

### Adding a New Guide
```sh
qst add "Limits" --scope definition --tags math,calculus
```

This creates `Limits.md` with structured YAML frontmatter:

```markdown
---
prerequisites: []
sub_guides: []
scope: definition
tags:
  - math
  - calculus
---
# Limits

A limit is the value that a function approaches as the input approaches some value.
```

### Subdirectory Support
Guides can be organized into subdirectories for cleaner repository structures:
```
Calculus/
├── manifest.yaml
├── Limits.md
└── differential/
    ├── Derivatives.md
    └── Power Rule.md
```
`qst` parses directory trees recursively, ignoring `public/`, hidden directories, and `.kng` archives.

### Rich Content: Mathematics & Diagrams
Guides support standard GitHub Flavored Markdown (GFM) along with academic enhancements:
- **LaTeX Math**: Inline formulas using `$f'(x)$` and block equations using `$$\lim_{x \to 0} \frac{\sin x}{x} = 1$$`.
- **Mermaid Diagrams**: Fenced code blocks with ` ```mermaid ` are converted into interactive SVG diagrams in the static site compiler.

---

## 3. Managing Relationships (`qst link`, `qst unlink`, `qst rename`)

The strength of `qst` lies in its dual-axis relationship model:
- **Horizontal prerequisites**: Dependencies between guides of the *same* scope.
- **Vertical subguides**: Decomposition of a guide into constituent components of *smaller* scope.

### Linking Prerequisites
To make `Limits` a prerequisite for `Continuity`:
```sh
qst link "Continuity" --prereq "Limits"
```

### Linking Subguides
To declare that `Derivatives` encompasses `Power Rule`:
```sh
qst link "Derivatives" --guide "Power Rule" --adherence detailed --segment 12-25
```
- `--adherence`: How faithfully the subguide is summarized in the parent body.
- `--segment`: Line range(s) in the parent guide where the subguide is discussed.

### Interactive Linking
Run `qst link` without flags to launch an interactive terminal prompt (powered by Huh) that guides you through selecting relationship types, adherences, and line ranges:
```sh
qst link "Derivatives"
```

### Unlinking Relationships
Remove a prerequisite or subguide without manually editing YAML:
```sh
qst unlink "Derivatives" --prereq "Limits"
qst unlink "Derivatives" --guide "Power Rule"
```
You can also run `qst unlink <guide> -i` for an interactive removal selector.

### Renaming Guides
When you rename a concept, manual search-and-replace across a large knowledge graph is prone to broken references. `qst rename` renames the file and cascades the update through all prerequisites, subguides, and manifest tours:
```sh
qst rename "Limits" "Introduction to Limits"
```

---

## 4. Browsing and Inspecting (`qst list`, `qst inspect`, `qst view`)

### Listing Guides
List all guides alphabetically:
```sh
qst list
```

Use `-l, --long` for an aligned tabular overview displaying scope, tags, and relationship counts:
```sh
qst list -l
```

### Inspecting Guide Details
View complete metadata, prerequisites, subguides, adherence levels, and line coverage for an individual guide:
```sh
qst inspect "Derivatives"
```
Add `-c, --coverage` to include a line-by-line coverage preview directly in the terminal output.

### Terminal Mermaid Visualization
Export the graph's structure as a Mermaid.js diagram directly to stdout:
```sh
qst view
```

---

## 5. Coverage Auditing (`qst coverage`)

To ensure parent guides faithfully explain their subguides, `qst` measures body line coverage:

```sh
qst coverage "Derivatives"
```

This outputs a line-by-line breakdown indicating which lines of the guide are mapped to which subguide segments, helping you eliminate dead text and ensure comprehensive explanations.

---

## 6. Validating Graph Integrity (`qst validate`)

Before publishing or sharing a graph, run `qst validate`:

```sh
qst validate
```

`qst validate` verifies:
1. **Acyclicity (DAG)**: Detects cycles and prints the exact loop chain (e.g., `Cycle detected: A -> B -> C -> A`).
2. **Scope Hierarchy**: Ensures horizontal prerequisites share the same scope and vertical subguides descend in scope.
3. **Reference Integrity**: Verifies that all prerequisites and subguides reference existing files.
4. **Adherence & Segments**: Validates that adherence levels match `manifest.yaml` and segment line ranges exist within file boundaries.
5. **Config Schema**: Enforces that all required manifest properties are present.

If any error is found, `qst validate` outputs actionable diagnostic messages and exits with code 1 for CI/CD integration.

---

## 7. Analyzing Graph Health & Metrics (`qst stats`)

Get an architectural overview of your knowledge graph:

```sh
qst stats
```

The statistics dashboard reports:
- **Overview Metrics**: Total guides, total edges, horizontal prerequisites, and vertical subguides.
- **Scope Distribution**: Breakdown of how many guides belong to each scope level.
- **Graph Depth**: The longest prerequisite chain in the graph (identifying conceptual bottlenecks).
- **Tour Coverage**: Percentage of guides integrated into guided tours.
- **Aggregate Line Coverage**: Total proportion of parent guide text mapped to subguide segments.

---

## 8. Compiling the Static Site (`qst build`)

Compile your knowledge graph into a publication-ready, standalone HTML website:

```sh
qst build . --out ./public
```

### Static Site Features
- **Academic Typography**: Clean serif typography, booktabs-style tables, and hairline gutter line numbering matching academic publications.
- **Local KaTeX Math**: LaTeX equations (`$math$` and `$$math$$`) rendered locally without external CDN dependencies.
- **Interactive Knowledge Graph Diagram**:
  - Embedded Mermaid.js DAG overview organized by horizontal scope layers.
  - Interactive pan, zoom, reset, and fullscreen controls.
  - Clickable nodes that navigate directly to guide pages.
- **Collapsible Sidebar with Session Persistence**:
  - Organized into distinct Tours and Scopes sections with color-coded accent indicators.
  - Dropdowns start collapsed by default.
  - Smooth rotating SVG disclosure chevrons.
  - Remembers open/closed states across page navigation using `sessionStorage`.
- **References & Relations Appendix**: Every guide page ends with an appendix detailing immediate prerequisites, dependents, subguides, and parent guides.

---

## 9. Packaging & Distributing (`qst form`, `qst free`)

### Packaging (`qst form`)
When you are ready to distribute your knowledge graph, package it into a `.kng` file:

```sh
qst form -o ./Calculus.kng
```

A `.kng` package is a lossless zip archive housing the manifest, all markdown guides, diagrams, and assets while excluding build artifacts (`public/`) and hidden files.

### Extracting (`qst free`)
To unpack a `.kng` package into an editable knowledge graph directory:

```sh
qst free ./Calculus.kng ./Calculus_Unpacked
```

Parent directories are safely created, Zip Slip directory traversals are strictly blocked, and all nested guides and assets are restored losslessly.
