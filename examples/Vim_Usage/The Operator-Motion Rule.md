---
prerequisites: [Character Motions hjkl, Word and WORD Navigation]
sub_guides: []
scope: definition
tags: [vim, operators, grammar]
---

## The Operator-Motion Rule

The true power of Vim stems from its composable grammar: actions are not separate hardcoded shortcuts, but sentences composed of **operators** (verbs) and **motions** (nouns).

### The Grammar Equation

$$\text{[count]} + \text{[operator]} + \text{[count]} + \text{[motion]}$$

- **Operator (Verb):** Specifies the action (`d` for delete, `c` for change, `y` for yank, `>` for indent).
- **Motion (Noun):** Specifies the scope and target (`w` for word, `$` for end of line, `G` for end of file, `ip` for paragraph).

### Multiplying Vocabulary

Because operators compose with every motion in Vim, learning 5 operators and 10 motions gives you $5 \times 10 = 50$ distinct commands without memorizing 50 shortcuts.
- `dw`: Delete word.
- `c$`: Change to end of line.
- `yG`: Yank from cursor to end of file.
- `>}`: Indent to end of current paragraph.
