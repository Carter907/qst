---
scope: primitive
prerequisites: []
sub_guides: []
tags: [cli, cobra, testing, architecture]
---

# Testable Cobra Output Channels and Exit Interception

CLI applications written in Go frequently fail unit testability when subcommands write directly to `os.Stdout` via `fmt.Printf` or terminate the process mid-test via `os.Exit(1)`. In `cmd/root.go` and `cmd/commands_test.go`, `qst` enforces three architectural patterns that make every CLI command 100% deterministic and unit-testable in-process.

## Standardized Cobra Output Streams

Every command in `cmd/` routes user-facing output exclusively through Cobra's bound `io.Writer` streams:
* **Standard Output**: `cmd.Printf(...)` and `cmd.Println(...)`.
* **Error Diagnostics**: `cmd.PrintErrf(...)` and `cmd.PrintErrln(...)`.

In `cmd/commands_test.go`, `executeCommandWithExit` binds both streams to an in-memory `*bytes.Buffer` via `rootCmd.SetOut(buf)` and `rootCmd.SetErr(buf)` to inspect exact command output.

## Process Exit Interception (`exitFunc`)

`cmd/root.go` defines a package-level replaceable exit hook:

```go
var exitFunc = os.Exit
```

When a command encounters a validation or I/O error, it calls `exitFunc(1)` followed immediately by `return`. During unit tests, `executeCommandWithExit` swaps `exitFunc` with a closure that records `exitCode = code` without terminating the `go test` runner process.

## Flag and Context State Resetting

Because `rootCmd` and its subcommands are package-level singletons in `cmd/`, flag mutations and command contexts persist across consecutive `rootCmd.ExecuteContext(ctx)` calls in the same test binary. `executeCommandWithExit` explicitly resets every flag (`f.Changed = false`, `f.Value.Set(f.DefValue)`) and propagates `c.SetContext(ctx)` across all subcommands before each test invocation.
