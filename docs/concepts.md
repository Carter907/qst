# Concepts

Here you'll find all the theoretical foundations for quest and the system that it's based on. This foundation is based largely on the discussions that occurred during the early development of the [Bluelearn](https://github.com/bluelearn-org/bluelearn) project.

## Terms

- Guide - a single markdown file that explains just one concept and may use previous guides as support.
- Prerequisite - a required guide for understanding another guide. Prerequisites must have the *same scope* as the guide that requires them.
- Adherence - Adherence is how closely a sub-guide is presented in a parent guide. For instance, a strict adherence would be roughly identical, word-for-word content.
- Scope - The kind of content goal for a guide. It represents overarching objective for what is how much is to be communicated to the reader.
- Sub-guide - a smaller-scoped guide that is encompassed by a larger-scoped guide.
- Knowledge Graph - a [Directed acyclic graph](https://en.wikipedia.org/wiki/Directed_acyclic_graph) that stores a network of guides. Edges are divided into horizontal connections (`prerequisites`) between guides of the same scope, and vertical connections (`sub_guides`) from larger-scoped guides to smaller ones. This idea is similar to a [Citation graph](https://en.wikipedia.org/wiki/Citation_graph), but instead of any kind of document, guides are used with structural constraints in place for their content and metadata.

## Best Practices

### Graph Normalization

Because the `adherence` property is a functional constraint that forces you to physically summarize `sub_guides` into a parent guide's body text, you must be careful about redundancy. 

If multiple explanation guides share the exact same `sub_guides`, embedding those sub-guides directly will force you to rewrite the exact same summaries in multiple files. 

To avoid this, **normalize your graph**. Strip the shared `sub_guides` out of the individual files, and instead create a new, centralized base guide (of the same scope) to house them. You can then cleanly point this new centralized guide toward the guide that references those concepts.

> [!NOTE]
> This is basically a way of factoring out sub-guides if you have multiple guides that rely on the same sub-guides.

