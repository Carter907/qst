package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func setupTestGraph(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	manifest := `title: Test Graph
scopes:
  - definition
  - description
adherences:
  - strict
  - introductory
relaxed_subguides: false
strict_coverage: false
tours: []
`
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	def1 := `---
scope: definition
prerequisites: []
sub_guides: []
tags: [basic]
---

## Def 1
Definition content here.
`
	if err := os.WriteFile(filepath.Join(dir, "Def 1.md"), []byte(def1), 0644); err != nil {
		t.Fatalf("failed to write def1: %v", err)
	}

	def2 := `---
scope: definition
prerequisites: []
sub_guides: []
tags: [basic]
---

## Def 2
Second definition.
`
	if err := os.WriteFile(filepath.Join(dir, "Def 2.md"), []byte(def2), 0644); err != nil {
		t.Fatalf("failed to write def2: %v", err)
	}

	desc1 := `---
scope: description
prerequisites: []
sub_guides: []
tags: [concept]
---

## Desc 1

Line 3 has Def 1 reference.
`
	if err := os.WriteFile(filepath.Join(dir, "Desc 1.md"), []byte(desc1), 0644); err != nil {
		t.Fatalf("failed to write desc1: %v", err)
	}

	return dir
}

type exitPanic int

type commandResult struct {
	Stdout   string
	Stderr   string
	Combined string
	ExitCode int
	Err      error
}

type syncCaptureWriter struct {
	mu       *sync.Mutex
	stream   *bytes.Buffer
	combined *bytes.Buffer
	onWrite  func(streamContent string)
}

func (w *syncCaptureWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	w.stream.Write(p)
	w.combined.Write(p)
	var snapshot string
	if w.onWrite != nil {
		snapshot = w.stream.String()
	}
	w.mu.Unlock()

	if w.onWrite != nil {
		w.onWrite(snapshot)
	}
	return len(p), nil
}

func resetCommandTree(c *cobra.Command, ctx context.Context) {
	c.SetContext(ctx)
	resetFlag := func(f *pflag.Flag) {
		f.Changed = false
		if sv, ok := f.Value.(pflag.SliceValue); ok {
			_ = sv.Replace(nil)
			if f.DefValue != "" && f.DefValue != "[]" {
				_ = f.Value.Set(strings.Trim(f.DefValue, "[]"))
			}
		} else {
			_ = f.Value.Set(f.DefValue)
		}
	}
	c.Flags().VisitAll(resetFlag)
	c.PersistentFlags().VisitAll(resetFlag)
	for _, sub := range c.Commands() {
		resetCommandTree(sub, ctx)
	}
}

func executeCommandWithHook(ctx context.Context, onStdout func(string), args ...string) (res commandResult) {
	if ctx == nil {
		ctx = context.Background()
	}

	oldExit := exitFunc
	exitFunc = func(code int) {
		panic(exitPanic(code))
	}
	defer func() {
		exitFunc = oldExit
		rootCmd.SetOut(os.Stdout)
		rootCmd.SetErr(os.Stderr)
		rootCmd.SetArgs(nil)
		resetCommandTree(rootCmd, context.Background())
	}()

	var mu sync.Mutex
	var stdoutBuf, stderrBuf, combinedBuf bytes.Buffer
	outW := &syncCaptureWriter{mu: &mu, stream: &stdoutBuf, combined: &combinedBuf, onWrite: onStdout}
	errW := &syncCaptureWriter{mu: &mu, stream: &stderrBuf, combined: &combinedBuf}

	rootCmd.SetOut(outW)
	rootCmd.SetErr(errW)
	rootCmd.SetArgs(args)
	resetCommandTree(rootCmd, ctx)

	defer func() {
		if r := recover(); r != nil {
			if code, ok := r.(exitPanic); ok {
				res.ExitCode = int(code)
			} else {
				panic(r)
			}
		}
		mu.Lock()
		res.Stdout = stdoutBuf.String()
		res.Stderr = stderrBuf.String()
		res.Combined = combinedBuf.String()
		mu.Unlock()
	}()

	res.Err = rootCmd.ExecuteContext(ctx)
	return res
}

func executeCommandContext(ctx context.Context, args ...string) commandResult {
	return executeCommandWithHook(ctx, nil, args...)
}

func executeCommandWithExit(args ...string) (string, int, error) {
	res := executeCommandContext(context.Background(), args...)
	return res.Combined, res.ExitCode, res.Err
}

func executeCommand(args ...string) (string, error) {
	res := executeCommandContext(context.Background(), args...)
	if res.Err == nil && res.ExitCode != 0 {
		res.Err = fmt.Errorf("command exited with code %d: %s", res.ExitCode, res.Combined)
	}
	return res.Combined, res.Err
}

var _ io.Writer = (*syncCaptureWriter)(nil)
