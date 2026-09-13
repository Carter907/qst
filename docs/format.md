# File Format

Knowledge graphs are archived into `.kng` files which are zip archives containing markdown files (`.md`) denoted as "Guides". These guides have a particular format that makes them ideal for housing and structuring knowledge in a Directed Acyclic Graph (DAG).

## Guide Metadata

These guides require a YAML frontmatter that carries their metadata. When formatting these markdown files you should make sure all fields specified below are included.

> [!NOTE]
> When creating metadata about language itself, it's important to realize that we are rubbing up against philosophical and linguistic barriers. There is no universally accepted way of measuring "scope" in a piece of text. Therefore, defining these terms qualitatively gives us some breathing room for their inherent fuzziness.

### Properties

#### `prerequisites`

A list of required guides.

**Strict Rule:** Horizontal edges in the knowledge graph can *only* exist between guides of the exact identical `scope` (horizontal relationship).

#### `sub_guides`

An optional list of sub-guide relation objects.

>[!NOTE] These sub-guides must be exactly one scope level smaller than the current guide, unless `relaxed_subguides: true` is set in the `manifest.yaml` which allows any smaller scope.

Each object must contain:
- `guide`: The name of the sub-guide being referenced.
- `adherence`: The specific constraint applied to this sub-guide reference, dictating exactly how the sub-guide must be tangibly represented and rewritten inside the body text (checked against the `adherences` list in `manifest.yaml`). Common examples:
  - **strict**: The author must explicitly use the exact or near-exact content of the sub-guide within the text.
  - **detailed**: The author must explicitly write closely summarized versions of the sub-guide within the text.
  - **introductory**: The author only needs to generalize the sub-guide, pointing to the concept without exact reproduction.
  - **vague**: The author may reference the sub-guide loosely, with the least exact reproduction of content.
- `segment`: The line range(s) in the current guide's text that cover this sub-guide (e.g., "10-20, 25-30").

#### `scope`

How much content is covered in a guide; how many concepts or things were explained. Scope is qualitative:
- **definition**: Smallest scope (singular term). *Example: "Exponent"*
- **description**: Smaller scope, but slightly larger than a definition, focusing heavily on providing examples and comparisons. *Example: "Graphs of Exponential Function"*
- **explanation**: Medium guide that explicitly relies on multiple descriptions to explain a mechanic. *Example: "Real-valued Functions"*
- **lesson**: Largest content type that aggregates and teaches multiple explanations. *Example: "Understanding Graphing in Algebra"*

### Tags

Tags are subjects or topic descriptors that summarize the content of the entire guide.

### Title

The title of a guide is defined by its file name.


## Knowledge Graph Metadata and Configuration

At the root of all knowledge graphs is a `manifest.yaml` file, which includes all constraint configuration and metadata of the knowledge graph.

### Descriptive Properties

These properties describe the knowledge graph as a whole.

#### `title`

The name given to the entire knowledge graph

#### `description`

A small description of whatever your knowledge graph seeks to explain

### Categories

#### `scopes`

An ordered list defining the valid "sizes" or hierarchical levels of guides, ordered strictly from smallest to largest. This establishes the vertical layers of your graph.

#### `adherences`

Adherence is a measure of fidelity: it defines how closely and deeply a parent guide rewrites or summarizes a sub-guide in its own body text.

### Constriants

#### `relaxed_subguides`

When false (default): Sub-guides must be strictly one step down on the scope ladder (e.g., an explanation can only contain
      descriptions).

When true: Allows skipping rungs on the ladder (e.g., a lesson can directly encompass a definition).

#### `require_subguides`

When enabled, any guide above the lowest scope level must define at least one sub-guide. It prevents authors from creating high-level, complex guides without breaking them down into smaller constituent concepts.

#### `strict_coverage`

When enabled, every single line of a guide's body text must be mapped to at least one sub-guide segment. It ensures that no text exists in the guide that isn't explicitly tied to a tracked concept.

### Tours

While the knowledge graph is an interconnected network, tours define curated, step-by-step reading sequences (e.g., "Beginner Track" or "Chapter 1 Walkthrough") for learners who need a linear trajectory.
