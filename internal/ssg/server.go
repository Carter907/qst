package ssg

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Carter907/qst/internal/graph"
	"github.com/fsnotify/fsnotify"
)

const (
	defaultPortAttempts = 10
	debounceDuration    = 150 * time.Millisecond
)

// ServeOptions configures the live-reloading development server.
type ServeOptions struct {
	TargetPath string
	Host       string
	Port       int
	Open       bool
	Offline    bool
	VendorDir  string
	CacheDir   string
	LogFunc    func(format string, args ...any)
	ErrFunc    func(format string, args ...any)
	OpenFunc   func(url string) error
}

// DevServer manages the ephemeral build directory, file watcher, SSE hub, and HTTP server.
type DevServer struct {
	opts       ServeOptions
	targetPath string
	isArchive  bool
	srcDir     string
	outDir     string
	tempDirs   []string

	buildMu sync.RWMutex
	lastErr string

	clientsMu sync.Mutex
	clients   map[chan sseEvent]struct{}

	listener net.Listener
	httpSrv  *http.Server
	watcher  *fsnotify.Watcher
	Addr     string
	URL      string
}

// NewDevServer initializes a new DevServer, prepares its ephemeral workspace,
// performs the initial build and validation pass, and binds the TCP listener.
func NewDevServer(opts ServeOptions) (*DevServer, error) {
	if opts.TargetPath == "" {
		opts.TargetPath = "."
	}
	if opts.Host == "" {
		opts.Host = "localhost"
	}
	if opts.Port < 0 || opts.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", opts.Port)
	}
	if opts.LogFunc == nil {
		opts.LogFunc = func(string, ...any) {}
	}
	if opts.ErrFunc == nil {
		opts.ErrFunc = func(string, ...any) {}
	}
	if opts.OpenFunc == nil {
		opts.OpenFunc = openBrowser
	}

	absTarget, err := filepath.Abs(opts.TargetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	info, err := os.Stat(absTarget)
	if err != nil {
		return nil, fmt.Errorf("target path error: %w", err)
	}

	isArchive := false
	if !info.IsDir() {
		if strings.EqualFold(filepath.Ext(absTarget), ".kng") {
			isArchive = true
		} else {
			return nil, fmt.Errorf("target must be a directory or a .kng archive: %s", opts.TargetPath)
		}
	}

	s := &DevServer{
		opts:       opts,
		targetPath: absTarget,
		isArchive:  isArchive,
		clients:    make(map[chan sseEvent]struct{}),
	}

	outDir, err := os.MkdirTemp("", "qst-serve-build-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary build directory: %w", err)
	}
	s.outDir = outDir
	s.tempDirs = append(s.tempDirs, outDir)

	if isArchive {
		srcDir, mkErr := os.MkdirTemp("", "qst-serve-kng-*")
		if mkErr != nil {
			s.cleanupTempDirs()
			return nil, fmt.Errorf("failed to create temporary archive directory: %w", mkErr)
		}
		s.srcDir = srcDir
		s.tempDirs = append(s.tempDirs, srcDir)

		if unarchErr := graph.UnarchiveGraph(absTarget, srcDir); unarchErr != nil {
			s.cleanupTempDirs()
			return nil, fmt.Errorf("failed to unarchive %s: %w", opts.TargetPath, unarchErr)
		}
	} else {
		s.srcDir = absTarget
	}

	// Prepare vendor assets once at startup
	vendorOpts := VendorOptions{
		Offline:   opts.Offline,
		VendorDir: opts.VendorDir,
		CacheDir:  opts.CacheDir,
		WarnFunc:  opts.ErrFunc,
	}
	if err := EnsureVendorAssets(s.outDir, vendorOpts); err != nil {
		s.cleanupTempDirs()
		return nil, fmt.Errorf("failed to prepare vendor assets: %w", err)
	}

	// Run initial build & validation (resilient: records diagnostic if graph has errors)
	s.rebuild(false)

	// Bind TCP listener with dynamic port fallback
	ln, actualPort, err := bindListener(opts.Host, opts.Port, defaultPortAttempts, opts.LogFunc)
	if err != nil {
		s.cleanupTempDirs()
		return nil, err
	}
	s.listener = ln
	s.Addr = fmt.Sprintf("%s:%d", opts.Host, actualPort)
	s.URL = fmt.Sprintf("http://%s:%d", opts.Host, actualPort)

	mux := http.NewServeMux()
	mux.HandleFunc("/_qst/events", s.handleSSE)
	mux.HandleFunc("/", s.handleStatic)

	s.httpSrv = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		_ = s.listener.Close()
		s.cleanupTempDirs()
		return nil, fmt.Errorf("failed to initialize file watcher: %w", err)
	}
	s.watcher = watcher

	if err := s.registerWatches(); err != nil {
		_ = s.watcher.Close()
		_ = s.listener.Close()
		s.cleanupTempDirs()
		return nil, fmt.Errorf("failed to watch target: %w", err)
	}

	return s, nil
}

// Start runs the file watcher loop and serves HTTP requests until ctx is cancelled.
func (s *DevServer) Start(ctx context.Context) error {
	defer s.Close()

	watchDone := make(chan struct{})
	go func() {
		defer close(watchDone)
		s.watchLoop(ctx)
	}()

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpSrv.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	s.opts.LogFunc("Serving knowledge graph from '%s' at %s (Press Ctrl+C to stop)\n", s.opts.TargetPath, s.URL)

	if s.opts.Open {
		if err := s.opts.OpenFunc(s.URL); err != nil {
			s.opts.ErrFunc("Warning: could not open browser automatically: %v\n", err)
		}
	}

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = s.httpSrv.Shutdown(shutdownCtx)
		<-watchDone
		return nil
	case err := <-errCh:
		return err
	}
}

// Close shuts down the HTTP server, closes the file watcher and SSE streams, and removes ephemeral directories.
func (s *DevServer) Close() {
	if s.watcher != nil {
		_ = s.watcher.Close()
	}
	if s.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = s.httpSrv.Shutdown(ctx)
		cancel()
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}

	s.clientsMu.Lock()
	for ch := range s.clients {
		delete(s.clients, ch)
		close(ch)
	}
	s.clientsMu.Unlock()

	s.cleanupTempDirs()
}

func (s *DevServer) cleanupTempDirs() {
	for _, dir := range s.tempDirs {
		if dir != "" {
			_ = os.RemoveAll(dir)
		}
	}
	s.tempDirs = nil
}

func bindListener(host string, startPort int, maxAttempts int, logf func(string, ...any)) (net.Listener, int, error) {
	if startPort == 0 {
		ln, err := net.Listen("tcp", fmt.Sprintf("%s:0", host))
		if err != nil {
			return nil, 0, fmt.Errorf("failed to bind port: %w", err)
		}
		addr, ok := ln.Addr().(*net.TCPAddr)
		if !ok {
			_ = ln.Close()
			return nil, 0, fmt.Errorf("unexpected listener address type")
		}
		return ln, addr.Port, nil
	}

	var lastErr error
	for i := range maxAttempts {
		port := startPort + i
		if port > 65535 {
			break
		}
		ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
		if err == nil {
			if i > 0 && logf != nil {
				logf("Port %d is in use, using next available port %d\n", startPort, port)
			}
			return ln, port, nil
		}
		lastErr = err
	}
	return nil, 0, fmt.Errorf("failed to bind port (tried %d..%d): %w", startPort, startPort+maxAttempts-1, lastErr)
}

// cleanGeneratedOutput removes previously generated HTML and copied media files from outDir
// while preserving outDir/assets/vendor so KaTeX and Mermaid do not need to be re-copied.
func cleanGeneratedOutput(outDir string) {
	vendorDir := filepath.Join(outDir, "assets", "vendor")
	assetsDir := filepath.Join(outDir, "assets")

	entries, err := os.ReadDir(outDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		fullPath := filepath.Join(outDir, entry.Name())
		if fullPath == assetsDir {
			// Clean everything inside assets except vendor
			assetEntries, aErr := os.ReadDir(assetsDir)
			if aErr == nil {
				for _, ae := range assetEntries {
					ap := filepath.Join(assetsDir, ae.Name())
					if ap != vendorDir {
						_ = os.RemoveAll(ap)
					}
				}
			}
			continue
		}
		_ = os.RemoveAll(fullPath)
	}
}

func (s *DevServer) rebuild(broadcast bool) {
	s.buildMu.Lock()

	if s.isArchive && broadcast {
		_ = os.RemoveAll(s.srcDir)
		_ = os.MkdirAll(s.srcDir, 0o755)
		if err := graph.UnarchiveGraph(s.targetPath, s.srcDir); err != nil {
			msg := fmt.Sprintf("Failed to unarchive %s: %v", filepath.Base(s.targetPath), err)
			s.lastErr = msg
			ensureFallbackIndex(s.outDir)
			s.buildMu.Unlock()
			s.opts.ErrFunc("Build error: %s\n", msg)
			s.broadcastError(msg)
			return
		}
	}

	// 1. Validate manifest first so we don't wipe the last known good HTML build if manifest is broken
	config, cfgErr := graph.ParseConfig(s.srcDir)
	if cfgErr != nil {
		msg := cfgErr.Error()
		s.lastErr = msg
		ensureFallbackIndex(s.outDir)
		s.buildMu.Unlock()
		s.opts.ErrFunc("Build error: %s\n", msg)
		if broadcast {
			s.broadcastError(msg)
		}
		return
	}

	// 2. Clean stale generated HTML/media while keeping vendor assets intact
	cleanGeneratedOutput(s.outDir)

	// 3. Build static site HTML and copy media assets
	buildErr := BuildSiteWithOptions(BuildOptions{
		DirPath:    s.srcDir,
		OutDir:     s.outDir,
		Offline:    s.opts.Offline,
		VendorDir:  s.opts.VendorDir,
		CacheDir:   s.opts.CacheDir,
		SkipVendor: true,
	})
	if buildErr != nil {
		msg := buildErr.Error()
		s.lastErr = msg
		ensureFallbackIndex(s.outDir)
		s.buildMu.Unlock()
		s.opts.ErrFunc("Build error: %s\n", msg)
		if broadcast {
			s.broadcastError(msg)
		}
		return
	}

	// 4. Parse and validate knowledge graph structure & frontmatter
	var diagIssues []string
	guides, parseErr := graph.ParseGraph(s.srcDir)
	if parseErr != nil {
		diagIssues = append(diagIssues, parseErr.Error())
	} else {
		if valErr := graph.ValidateGraph(guides, config); valErr != nil {
			diagIssues = append(diagIssues, valErr.Error())
		}
	}

	if len(diagIssues) > 0 {
		msg := strings.Join(diagIssues, "\n")
		s.lastErr = msg
		s.buildMu.Unlock()
		s.opts.ErrFunc("Validation issue:\n%s\n", msg)
		if broadcast {
			// Reload so updated HTML content is visible, and send error overlay
			s.broadcastEvent(sseEvent{Event: "reload", Data: "reload"})
			s.broadcastError(msg)
		}
		return
	}

	s.lastErr = ""
	s.buildMu.Unlock()

	if broadcast {
		s.opts.LogFunc("Rebuild complete.\n")
		s.broadcastEvent(sseEvent{Event: "clear_error", Data: "{}"})
		s.broadcastEvent(sseEvent{Event: "reload", Data: "reload"})
	}
}

func (s *DevServer) handleStatic(w http.ResponseWriter, r *http.Request) {
	s.buildMu.RLock()
	defer s.buildMu.RUnlock()

	if strings.HasPrefix(r.URL.Path, "/assets/vendor/") {
		w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
		http.FileServer(http.Dir(s.outDir)).ServeHTTP(w, r)
		return
	}

	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

	cleanRoot := filepath.Clean(s.outDir)
	relPath := filepath.FromSlash(strings.TrimPrefix(pathClean(r.URL.Path), "/"))
	target := filepath.Clean(filepath.Join(cleanRoot, relPath))

	// Path traversal guard
	if target != cleanRoot && !strings.HasPrefix(target, cleanRoot+string(os.PathSeparator)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	info, err := os.Stat(target)
	if err == nil && info.IsDir() {
		indexCandidate := filepath.Join(target, "index.html")
		if idxInfo, idxErr := os.Stat(indexCandidate); idxErr == nil && !idxInfo.IsDir() {
			target = indexCandidate
			info = idxInfo
			err = nil
		}
	} else if err != nil && filepath.Ext(target) == "" {
		htmlCandidate := target + ".html"
		if htmlInfo, htmlErr := os.Stat(htmlCandidate); htmlErr == nil && !htmlInfo.IsDir() {
			target = htmlCandidate
			info = htmlInfo
			err = nil
		}
	}

	if err == nil && !info.IsDir() && strings.EqualFold(filepath.Ext(target), ".html") {
		rawHTML, readErr := os.ReadFile(target)
		if readErr != nil {
			http.Error(w, "Failed to read page", http.StatusInternalServerError)
			return
		}
		injected := injectLiveReloadScript(rawHTML)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(injected)
		return
	}

	http.FileServer(http.Dir(s.outDir)).ServeHTTP(w, r)
}

func pathClean(p string) string {
	if p == "" || p == "/" {
		return "/"
	}
	p = filepath.ToSlash(filepath.Clean(p))
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
