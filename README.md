# `qst`: A Command-line Knowledge System

Quest is a command-line tool for storing and structuring knowledge. You use markdown files to explain a certain idea or practice. These files are related to each other through a directed acyclic graph. This tool helps you manipulate, verify, and package this structured material into a single file format.

These files can be distributed and unpacked in markdown editors like Obsidian. Integration with this tool in your workflow is made easy through its simple yet comprehensive command-line interface.


## Why Quest?

Quest facilitates the distribution and management of knowledge.

You can use the file format on education platforms for curriculum building and learning path creation (walks on the graph).

Being able to structure information using a standard format is important. Quest helps decrease uncertainty in the learning process by proposing and helping to enforce a structural standard for markdown content. Quest enforces this standard format using `qst validate`, which checks many kinds of inconsistencies in the knowledge graph structure (you can learn more [here](/docs/cli/qst_validate.md)).

By offloading validation of content to a concrete system, you can focus on comprehending and building upon learning material.


## Installation & Getting Started

**Using `go install`**

```sh
go install github.com/Carter907/qst@latest
```

Visit [installation.md](/docs/installation.md) for more details.

## Usage

**Running `qst`**:

```sh
qst --help
```

For a basic usage example of the command, please read [usage.md](/docs/usage.md)

## Concepts

Check out [concepts.md](/docs/concepts.md) for more theoretical foundations.

## Format

Find formatting constraints and metadata specifics in [format.md](/docs/format.md)

## CLI Docs

Visit [the cli docs](/docs/cli/) for generated docs on the command-line interface (`qst` functionality).
