# Usage

## Working with `qst`

### 1. Create a new knowledge graph with `qst new`

```sh
qst new Exponents
```

This will create a new directory called "Exponents" and populate it with a `manifest.yaml` file and a **starter markdown file**.

The main configuration file is the manifest.yaml

The `manifest.yaml` file has the following content: 

```yaml
title: Your New Knowledge
description: What can you say about this knowledge?
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
require_subguides: true
strict_coverage: true
tours:
  - name: Tour 1
    guides:
      - Guide Filename 1
      - Guide Filename 2
  - name: Tour 2
    guides:
      -
```

### 2. Add your first guides

```sh
qst add "Understanding Exponents" --scope lesson --tags math,algebra
```

This command creates a markdown file with populated front-matter. You can also specify prerequisites if you already know the structure you're going for.

If you want to link an existing sub-guide later on, you can use the `qst link` command:

```sh
qst link "Understanding Exponents" --guide "Powers" --adherence detailed --segment 1-5
```

### 3. Validate the structure

As you create more and more guides and start to draft out a presentable network of guides, you want to make sure the graph is valid under the constraints:

```sh
qst validate
```

`qst validate` will check for the following:

- Cycles
- Non-existent prerequisite
- Non-existent sub-guides
- Missing properties


### 4. Package your knowledge graph

Once you are ready to distribute your knowledge, you use the `qst form` command:

```sh
qst form
```

The current directory is chosen if you don't specify one. When you run this command, a `.kng` file with the same name as the chosen directory will be created. This file is a zip archive housing all markdown files and the `manifest.yaml` metafile.

## Summary

This should give you an idea of what the workflow of curating knowledge graphs looks like. Getting the pedagogical structure of a knowledge right takes much more than `qst validate` and a single person. The reason why this CLI is focused on packaging knowledge is because we believe knowledge should be open, free, and easily comprehensible. Sharing `.kng` files is the most direct way of getting insight on how build something presentable.
