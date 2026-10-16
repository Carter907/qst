---
prerequisites: [Delimiter and Quote Text Objects]
sub_guides: []
scope: definition
tags: [vim, text-objects, markup, blocks]
---

## Block and Tag Text Objects

Block and tag text objects extend structural editing to HTML/XML markups, prose paragraphs, and natural language sentences.

### Markup Tags

- `it` (**inner tag**): Targets the content contained strictly between HTML/XML opening and closing tags.
- `at` (**around tag**): Targets the content **and** both the opening `<tag>` and closing `</tag>` elements.
- Example: `cit` inside `<p>Hello World</p>` removes `Hello World` and enters Insert Mode.

### Paragraphs and Sentences

- `ip` / `ap`: Inner and around paragraph blocks separated by blank lines.
- `is` / `as`: Inner and around sentences bounded by periods and punctuation.
- Example: `gqap` formats and word-wraps the entire current paragraph.
