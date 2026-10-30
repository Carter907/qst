Contributing to this project is greatly appreciated as you help make learning more accessible for everyone.

## Exploring the Codebase Architecture

`qst` models its own internal architecture as a self-hosted knowledge graph in [`examples/qst_Architecture/`](examples/qst_Architecture/). New contributors can launch the live development server to browse the system architecture and follow the **Onboarding & Architecture Overview** tour:

```sh
go run . serve examples/qst_Architecture --open
```

## Example Procedure

Contributions are done through issues and pull requests.

1. Create an issue
2. Get feedback and assignment
3. Create a fork of the repository if you haven't already
4. Sync the main branch, and begin work on a separate branch
5. Create a pull request early
6. If all checks pass, then your branch will be merged


## Commits

Commits should be signed and follow [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/#summary).

You can sign a commit by passing in the `-s` flag when committing through git.

Here is a list of additional commit types that can be used in this repository:

- chore - periodic work that must be done to maintain the project.
- docs - any changes made under the `/docs` folder, README, or other documentation (including this file).
- refactor - reimplementation of an already established feature. The behavior should largely be unaffected.

> message scopes are optional.


## Code Conventions

Comments should generally be capitalized, short, informative messages. Keep comments to a minimum if possible.
