// Package playback provides terminal recording playback functionality.
// It supports both macOS and Linux script command output formats.
package playback

// Frame represents a single frame of terminal content at a specific timestamp.
// For static playback, use a single frame with Timestamp 0.
// For animated playback, use multiple frames with increasing timestamps.
type Frame struct {
	Timestamp float64 `json:"timestamp"` // Time in seconds (cumulative from start)
	Content   string  `json:"content"`   // Terminal content (with ANSI codes preserved)
}

// Options configures HTML rendering behavior.
type Options struct {
	Title string // Page title (defaults to "Terminal" if empty)
}
