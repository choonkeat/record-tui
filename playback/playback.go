package playback

import (
	"github.com/choonkeat/record-tui/internal/html"
	"github.com/choonkeat/record-tui/internal/session"
)

// StripMetadata removes script command header/footer metadata from session log content.
// Supports both macOS and Linux script output formats:
//
// macOS format:
//
//	Script started on Wed Dec 31 12:10:34 2025
//	Command: bash
//	[content]
//	Script done on Wed Dec 31 12:11:22 2025
//
// Linux format:
//
//	Script started on 2026-01-12 06:41:43+00:00 [COMMAND="bash" TERM="xterm-256color" ...]
//	[content]
//	Script done on 2026-01-12 06:45:00+00:00 [COMMAND_EXIT_STATUS="0"]
func StripMetadata(content string) string {
	return session.StripMetadata(content)
}

// RenderHTML generates a standalone HTML page with terminal playback using xterm.js.
// The generated HTML is self-contained and can be viewed in any modern browser.
//
// For static display, pass a single Frame with Timestamp 0.
// The HTML includes proper ANSI color rendering, automatic dimension calculation,
// and responsive layout.
//
// Options can be used to customize the output (e.g., page title).
func RenderHTML(frames []Frame, opts ...Options) (string, error) {
	// Convert public Frame to internal PlaybackFrame
	internalFrames := make([]html.PlaybackFrame, len(frames))
	for i, f := range frames {
		internalFrames[i] = html.PlaybackFrame{
			Timestamp: f.Timestamp,
			Content:   f.Content,
		}
	}

	// Extract title from options
	title := "Terminal"
	if len(opts) > 0 && opts[0].Title != "" {
		title = opts[0].Title
	}

	return html.RenderPlaybackHTML(internalFrames, title)
}
