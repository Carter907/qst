---
scope: primitive
prerequisites: []
sub_guides: []
tags: [ssg, server, sse, http]
---

# Server-Sent Events Hub and Middleware Injection

In `internal/ssg/server.go`, `qst serve` implements real-time browser live-reloading and diagnostic error overlays using pure Go standard-library **Server-Sent Events (SSE)** (`http.Flusher`) and on-the-fly HTML middleware injection.

## Event Naming and SSE Broker (`handleSSE`)

The browser connects to `GET /_qst/events` using `new EventSource("/_qst/events")`. The server maintains a mutex-guarded set of subscriber channels (`map[chan sseEvent]struct{}`) and broadcasts three distinct event types:
* `event: reload`: Instructs the browser to persist scroll position and call `window.location.reload()`.
* `event: build_error`: Carries a JSON payload `{"message": "..."}` rendered in a dismissible fixed-position diagnostic overlay.
* `event: clear_error`: Dismisses any active diagnostic overlay.

> **Why `build_error` instead of `error`?** In the browser's native `EventSource` specification, `"error"` is a reserved event name fired whenever the underlying TCP/HTTP stream disconnects. Using a custom `build_error` event prevents transient network reconnects from colliding with graph validation diagnostics.

When a browser first connects (or manually refreshes via `F5`), `handleSSE` immediately reads `s.lastErr` under a read lock and replays any active `build_error` event so validation errors remain visible across manual refreshes.

## Middleware Injection and Split HTTP Caching (`handleStatic`)

Rather than polluting compiled `.html` files on disk with development scripts, `injectLiveReloadScript` splices the `<script data-qst-live-reload>` block immediately before `</body>` in memory when serving HTML responses.

To eliminate flashes of unrendered math or diagrams during live reload, `handleStatic` applies a split caching policy:
* **`/assets/vendor/*`**: Served with `Cache-Control: public, max-age=86400, immutable` so KaTeX fonts and `mermaid.min.js` stay cached in browser memory across reloads.
* **HTML, `style.css`, and guide media**: Served with `Cache-Control: no-store, no-cache, must-revalidate` so edits appear instantaneously.
