package ssg

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Carter907/qst/internal/graph"
)

func createTestGraphDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	manifest := `title: Dev Server Graph
description: Testing qst serve
scopes:
  - definition
  - explanation
adherences:
  - strict
tours:
  - name: Intro Tour
    guides:
      - Intro
`
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	intro := `---
scope: definition
prerequisites: []
sub_guides: []
tags: [core]
---

# Welcome to Intro

Initial content v1.
![Diagram](diagram.svg)
`
	if err := os.WriteFile(filepath.Join(dir, "Intro.md"), []byte(intro), 0o644); err != nil {
		t.Fatalf("failed to write Intro.md: %v", err)
	}

	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"></svg>`
	if err := os.WriteFile(filepath.Join(dir, "diagram.svg"), []byte(svg), 0o644); err != nil {
		t.Fatalf("failed to write diagram.svg: %v", err)
	}

	return dir
}

func TestDevServerServeAndLiveReload(t *testing.T) {
	dir := createTestGraphDir(t)

	var openedURL string
	srv, err := NewDevServer(ServeOptions{
		TargetPath: dir,
		Host:       "127.0.0.1",
		Port:       0,
		Offline:    true,
		Open:       true,
		OpenFunc: func(u string) error {
			openedURL = u
			return nil
		},
	})
	if err != nil {
		t.Fatalf("NewDevServer failed: %v", err)
	}

	outDir := srv.outDir
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- srv.Start(ctx)
	}()

	// Wait briefly for server to start and trigger OpenFunc
	time.Sleep(50 * time.Millisecond)
	if openedURL != srv.URL {
		t.Errorf("expected OpenFunc to be called with %s, got %s", srv.URL, openedURL)
	}

	// 1. Fetch index page and verify live reload script injection + no-store header
	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Header.Get("Cache-Control"), "no-store") {
		t.Errorf("expected no-store Cache-Control on HTML, got %s", resp.Header.Get("Cache-Control"))
	}
	if !strings.Contains(string(body), "data-qst-live-reload") {
		t.Errorf("expected injected live reload script in index.html")
	}

	// 2. Fetch guide page with clean URL (/Intro) and .html URL (/Intro.html)
	respGuide, err := http.Get(srv.URL + "/Intro")
	if err != nil {
		t.Fatalf("GET /Intro failed: %v", err)
	}
	guideBody, _ := io.ReadAll(respGuide.Body)
	_ = respGuide.Body.Close()
	if !strings.Contains(string(guideBody), "Initial content v1.") {
		t.Errorf("expected 'Initial content v1.' in guide HTML, got: %s", string(guideBody))
	}

	// 3. Verify static media asset (.svg) was copied and is served
	respSVG, err := http.Get(srv.URL + "/diagram.svg")
	if err != nil {
		t.Fatalf("GET /diagram.svg failed: %v", err)
	}
	svgBytes, _ := io.ReadAll(respSVG.Body)
	_ = respSVG.Body.Close()
	if respSVG.StatusCode != http.StatusOK || !strings.Contains(string(svgBytes), "<svg") {
		t.Errorf("expected 200 OK and SVG content, got status %d: %s", respSVG.StatusCode, string(svgBytes))
	}

	// 4. Connect to SSE endpoint and listen for live reload event upon file edit
	sseCtx, sseCancel := context.WithTimeout(ctx, 5*time.Second)
	defer sseCancel()
	req, err := http.NewRequestWithContext(sseCtx, http.MethodGet, srv.URL+"/_qst/events", nil)
	if err != nil {
		t.Fatalf("failed to create SSE request: %v", err)
	}
	sseResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("SSE connection failed: %v", err)
	}
	defer func() { _ = sseResp.Body.Close() }()

	eventCh := make(chan string, 10)
	go func() {
		scanner := bufio.NewScanner(sseResp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "event: ") {
				eventCh <- strings.TrimPrefix(line, "event: ")
			}
		}
	}()

	// Modify Intro.md
	updatedIntro := `---
scope: definition
prerequisites: []
sub_guides: []
tags: [core]
---

# Welcome to Intro

Updated content v2!
`
	if err := os.WriteFile(filepath.Join(dir, "Intro.md"), []byte(updatedIntro), 0o644); err != nil {
		t.Fatalf("failed to update Intro.md: %v", err)
	}

	select {
	case evt := <-eventCh:
		if evt != "clear_error" && evt != "reload" {
			t.Errorf("unexpected SSE event: %s", evt)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for SSE reload event after file modification")
	}

	// Verify updated content is now served
	respGuide2, err := http.Get(srv.URL + "/Intro.html")
	if err != nil {
		t.Fatalf("GET /Intro.html failed: %v", err)
	}
	guideBody2, _ := io.ReadAll(respGuide2.Body)
	_ = respGuide2.Body.Close()
	if !strings.Contains(string(guideBody2), "Updated content v2!") {
		t.Errorf("expected 'Updated content v2!' after live rebuild, got: %s", string(guideBody2))
	}

	// 5. Introduce a graph validation error (missing prerequisite) and verify build_error SSE event
	invalidIntro := `---
scope: definition
prerequisites: [NonExistentGuide]
sub_guides: []
tags: [core]
---

# Welcome to Intro

Content with broken prereq.
`
	if err := os.WriteFile(filepath.Join(dir, "Intro.md"), []byte(invalidIntro), 0o644); err != nil {
		t.Fatalf("failed to write invalid Intro.md: %v", err)
	}

	gotBuildError := false
	deadline := time.After(3 * time.Second)
	for !gotBuildError {
		select {
		case evt := <-eventCh:
			if evt == "build_error" {
				gotBuildError = true
			}
		case <-deadline:
			t.Fatalf("timed out waiting for build_error SSE event")
		}
	}

	// Shut down server and verify ephemeral outDir is deleted
	cancel()
	if err := <-doneCh; err != nil {
		t.Errorf("DevServer exited with unexpected error: %v", err)
	}
	if _, err := os.Stat(outDir); !os.IsNotExist(err) {
		t.Errorf("expected ephemeral build directory %s to be removed on shutdown", outDir)
	}
}

func TestDevServerPortFallbackAndResilientStartup(t *testing.T) {
	// Create a directory with a broken manifest initially
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte("title: [unclosed"), 0o644); err != nil {
		t.Fatalf("failed to write broken manifest: %v", err)
	}

	// Occupy a port to test dynamic port fallback
	blocker, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on blocker port: %v", err)
	}
	defer func() { _ = blocker.Close() }()
	busyPort := blocker.Addr().(*net.TCPAddr).Port

	var logs strings.Builder
	srv, err := NewDevServer(ServeOptions{
		TargetPath: dir,
		Host:       "127.0.0.1",
		Port:       busyPort,
		Offline:    true,
		LogFunc: func(format string, args ...any) {
			_, _ = fmt.Fprintf(&logs, format, args...)
		},
	})
	if err != nil {
		t.Fatalf("NewDevServer should succeed resiliently even with broken manifest: %v", err)
	}
	defer srv.Close()

	if srv.Addr == fmt.Sprintf("127.0.0.1:%d", busyPort) {
		t.Errorf("expected server to increment away from busy port %d", busyPort)
	}
	if !strings.Contains(logs.String(), "is in use, using next available port") {
		t.Errorf("expected port fallback log message, got: %s", logs.String())
	}
	if srv.lastErr == "" {
		t.Errorf("expected initial build diagnostic error to be recorded for broken manifest")
	}
}

func TestDevServerKngArchive(t *testing.T) {
	srcDir := createTestGraphDir(t)
	kngPath := filepath.Join(t.TempDir(), "test.kng")

	if err := graph.ArchiveGraph(srcDir, kngPath); err != nil {
		t.Fatalf("failed to create .kng archive: %v", err)
	}

	srv, err := NewDevServer(ServeOptions{
		TargetPath: kngPath,
		Host:       "127.0.0.1",
		Port:       0,
		Offline:    true,
	})
	if err != nil {
		t.Fatalf("NewDevServer failed for .kng archive: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	doneCh := make(chan error, 1)
	go func() {
		doneCh <- srv.Start(ctx)
	}()

	resp, err := http.Get(srv.URL + "/Intro.html")
	if err != nil {
		t.Fatalf("GET /Intro.html from .kng server failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if !strings.Contains(string(body), "Initial content v1.") {
		t.Errorf("expected archive guide content to be served, got: %s", string(body))
	}

	cancel()
	if err := <-doneCh; err != nil {
		t.Errorf("DevServer exited with unexpected error: %v", err)
	}
}
