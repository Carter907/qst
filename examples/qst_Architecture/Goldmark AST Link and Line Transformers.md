---
scope: primitive
prerequisites:
  - Line-Preserving Frontmatter Parser
sub_guides: []
tags: [ssg, markdown, goldmark, ast]
---

# Goldmark AST Link and Line Transformers

In `internal/ssg/markdown.go`, `qst` configures the Goldmark Markdown engine with two custom Abstract Syntax Tree (AST) transformers (`linkASTTransformer` at priority `100` and `lineASTTransformer` at priority `200`) and a custom HTML node renderer (`LineNumRenderer`).

## Relative Markdown Link Rewriting (`linkASTTransformer`)

Authors write inter-guide links in Markdown pointing to `.md` files. During AST traversal, `linkASTTransformer` inspects every `*ast.Link` whose destination ends with `.md` and does not start with `http`:
* Standard `.md` targets are rewritten to `.html` (`GuideA.md` $\to$ `GuideA.html`).
* **`index.md` Collision Avoidance**: Because the static site generator always synthesizes `index.html` at the root as the academic course syllabus, any guide named `index.md` is compiled to `index-guide.html`, and links to `index.md` are automatically rewritten to `index-guide.html`.

## Binary-Searched Line Anchors (`lineASTTransformer`)

Because `SubGuideRelation.Segment` references specific post-frontmatter line numbers, rendered HTML guides display interactive left-margin line numbers (`#L1`, `#L2`, ...).

Goldmark AST nodes store byte offsets (`Segment.Start`) rather than line numbers. `lineASTTransformer` precomputes the byte offset of every newline in the Markdown source slice `lineStarts` and resolves any byte offset `pos` to its 1-based line number in $O(\log N)$ time using binary search:

```go
getLine := func(pos int) int {
	return sort.Search(len(lineStarts), func(i int) bool {
		return lineStarts[i] > pos
	})
}
```

When walking `*ast.Text` or `*ast.FencedCodeBlock` nodes, the transformer climbs up to the enclosing block container (`target.Parent().Type() != ast.TypeBlock`) and inserts a synthetic `&LineNumNode{Line: lineNo}` before the block. `LineNumRenderer` then emits:

```html
<span class="line-marker" id="L12" data-line="12"><span class="line-num">12</span></span>
```
