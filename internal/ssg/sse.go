package ssg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// sseEvent represents a server-sent event broadcast to connected browsers.
type sseEvent struct {
	Event string
	Data  string
}

func ensureFallbackIndex(outDir string) {
	idxPath := filepath.Join(outDir, "index.html")
	if _, err := os.Stat(idxPath); err == nil {
		return
	}
	fallback, err := templatesFS.ReadFile("templates/fallback_index.html")
	if err != nil {
		return
	}
	_ = os.WriteFile(idxPath, fallback, 0o644)
}

func (s *DevServer) broadcastError(msg string) {
	payload, err := json.Marshal(struct {
		Message string `json:"message"`
	}{Message: msg})
	if err != nil {
		return
	}
	s.broadcastEvent(sseEvent{
		Event: "build_error",
		Data:  string(payload),
	})
}

func (s *DevServer) broadcastEvent(evt sseEvent) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for ch := range s.clients {
		select {
		case ch <- evt:
		default:
		}
	}
}

func (s *DevServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := make(chan sseEvent, 16)
	s.clientsMu.Lock()
	s.clients[ch] = struct{}{}
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		if _, exists := s.clients[ch]; exists {
			delete(s.clients, ch)
			close(ch)
		}
		s.clientsMu.Unlock()
	}()

	// Send initial comment to establish stream and replay any active build/validation error
	_, _ = fmt.Fprintf(w, ": connected\n\n")
	s.buildMu.RLock()
	activeErr := s.lastErr
	s.buildMu.RUnlock()
	if activeErr != "" {
		if payload, err := json.Marshal(struct {
			Message string `json:"message"`
		}{Message: activeErr}); err == nil {
			_, _ = fmt.Fprintf(w, "event: build_error\ndata: %s\n\n", string(payload))
		}
	}
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case evt, open := <-ch:
			if !open {
				return
			}
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Event, evt.Data)
			flusher.Flush()
		}
	}
}

func injectLiveReloadScript(htmlBytes []byte) []byte {
	scriptBytes, err := templatesFS.ReadFile("templates/livereload.html")
	if err != nil {
		return htmlBytes
	}
	closingTag := []byte("</body>")
	idx := bytes.LastIndex(bytes.ToLower(htmlBytes), closingTag)
	if idx == -1 {
		var buf bytes.Buffer
		buf.Write(htmlBytes)
		buf.Write(scriptBytes)
		return buf.Bytes()
	}
	var buf bytes.Buffer
	buf.Grow(len(htmlBytes) + len(scriptBytes))
	buf.Write(htmlBytes[:idx])
	buf.Write(scriptBytes)
	buf.Write(htmlBytes[idx:])
	return buf.Bytes()
}
