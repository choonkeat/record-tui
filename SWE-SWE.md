# swe-swe Integration Guide

## Streaming HTML for Large Recordings

swe-swe recordings can be multi-megabyte. Use `RenderStreamingHTML` instead of `RenderHTML` to avoid browser hangs during page load.

### Usage

```go
import "github.com/choonkeat/record-tui/playback"

// Generate lightweight HTML shell that streams from a URL
html, err := playback.RenderStreamingHTML(playback.StreamingOptions{
    Title:   "Session Recording",
    DataURL: "/session.log", // URL path to fetch raw session data
    FooterLink: playback.FooterLink{
        Text: "swe-swe",
        URL:  "https://github.com/anthropics/swe-swe",
    },
})
```

### Server Setup

The Go server must serve two endpoints:

1. **HTML page** (e.g., `/recordings/{id}`) — returns the streaming HTML
2. **Raw session data** (e.g., `/recordings/{id}/session.log`) — returns the raw `session.log` file

The `DataURL` in `StreamingOptions` should be the path to the raw session data, relative to the HTML page or absolute.

### Example Handler

```go
func handleRecording(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id") // or mux.Vars(r)["id"]

    // Serve raw session.log
    if strings.HasSuffix(r.URL.Path, "/session.log") {
        http.ServeFile(w, r, filepath.Join(recordingsDir, id, "session.log"))
        return
    }

    // Serve streaming HTML
    html, err := playback.RenderStreamingHTML(playback.StreamingOptions{
        Title:   "Recording " + id,
        DataURL: "./session.log", // relative to this page
    })
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(html))
}
```

### What the Streaming HTML Does

The generated HTML:
1. Shows "Loading..." indicator
2. Fetches `DataURL` via `fetch()` with streaming
3. Strips header lines (`Script started on`, `Command:`)
4. Strips footer lines (`Script done on`, `Saving session`)
5. Replaces clear sequences with visible `──────── terminal cleared ────────` separator
6. Renders progressively to xterm.js as bytes arrive
7. Resizes terminal to fit content after streaming completes

### When NOT to Use Streaming

Use regular `RenderHTML` (embedded mode) when:
- File size < 1MB
- Offline viewing required (file:// URLs)
- Single self-contained HTML file needed

Streaming mode requires HTTP(S) — won't work with `file://` URLs.
