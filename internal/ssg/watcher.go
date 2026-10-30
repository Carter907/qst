package ssg

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Carter907/qst/internal/graph"
	"github.com/fsnotify/fsnotify"
)

func shouldIgnoreWatchEvent(path string) bool {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") {
		return true
	}
	if strings.HasSuffix(base, "~") ||
		strings.HasSuffix(base, ".swp") ||
		strings.HasSuffix(base, ".swo") ||
		strings.HasSuffix(base, ".tmp") {
		return true
	}
	return false
}

func (s *DevServer) registerWatches() error {
	if s.isArchive {
		parentDir := filepath.Dir(s.targetPath)
		if err := s.watcher.Add(parentDir); err != nil {
			return err
		}
		_ = s.watcher.Add(s.targetPath)
		return nil
	}

	return s.addWatchRecursive(s.srcDir)
}

func (s *DevServer) addWatchRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		absPath, _ := filepath.Abs(path)
		absOut, _ := filepath.Abs(s.outDir)
		if absPath == absOut {
			return filepath.SkipDir
		}
		if path != root && graph.IsIgnoredDir(info.Name()) {
			return filepath.SkipDir
		}
		_ = s.watcher.Add(path)
		return nil
	})
}

func (s *DevServer) watchLoop(ctx context.Context) {
	var timer *time.Timer
	var timerC <-chan time.Time

	triggerDebounce := func() {
		if timer == nil {
			timer = time.NewTimer(debounceDuration)
			timerC = timer.C
		} else {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(debounceDuration)
		}
	}

	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}

			if s.isArchive {
				cleanEvt, _ := filepath.Abs(event.Name)
				if cleanEvt != s.targetPath {
					continue
				}
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
					_ = s.watcher.Add(s.targetPath)
				}
				triggerDebounce()
				continue
			}

			if shouldIgnoreWatchEvent(event.Name) {
				continue
			}
			if strings.EqualFold(filepath.Ext(event.Name), ".kng") {
				continue
			}

			// Check if event is inside an ignored directory
			rel, relErr := filepath.Rel(s.srcDir, event.Name)
			if relErr == nil {
				parts := strings.Split(filepath.ToSlash(rel), "/")
				if slices.ContainsFunc(parts[:len(parts)-1], graph.IsIgnoredDir) {
					continue
				}
			}

			if event.Has(fsnotify.Create) {
				if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
					if !graph.IsIgnoredDir(info.Name()) {
						_ = s.addWatchRecursive(event.Name)
					}
				}
			}

			triggerDebounce()

		case _, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
		case <-timerC:
			timerC = nil
			timer = nil
			s.opts.LogFunc("Change detected, rebuilding knowledge graph...\n")
			s.rebuild(true)
		}
	}
}
