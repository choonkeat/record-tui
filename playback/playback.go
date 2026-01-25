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

	// Extract options
	title := "Terminal"
	var footerLink html.FooterLink
	if len(opts) > 0 {
		if opts[0].Title != "" {
			title = opts[0].Title
		}
		footerLink = html.FooterLink{
			Text: opts[0].FooterLink.Text,
			URL:  opts[0].FooterLink.URL,
		}
	}

	return html.RenderPlaybackHTML(internalFrames, title, footerLink)
}

// RenderStreamingHTML generates an HTML page that streams terminal data from a URL.
// Unlike RenderHTML which embeds all data in the HTML, this version fetches data
// via JavaScript fetch() and streams it to xterm.js for progressive rendering.
//
// This is ideal for large recordings (multi-megabyte) where embedding data in HTML
// would cause slow page loads. The generated HTML is lightweight (~10KB) and
// progressively renders content as it streams from DataURL.
//
// Requirements:
//   - The HTML must be served via HTTP(S), not file:// (fetch() requires a server)
//   - DataURL should point to raw session.log content (the JS handles metadata stripping)
//
// Example:
//
//	html, err := playback.RenderStreamingHTML(playback.StreamingOptions{
//	    Title:   "My Recording",
//	    DataURL: "./session.log",
//	})
func RenderStreamingHTML(opts StreamingOptions) (string, error) {
	internalOpts := html.StreamingOptions{
		Title:   opts.Title,
		DataURL: opts.DataURL,
		FooterLink: html.FooterLink{
			Text: opts.FooterLink.Text,
			URL:  opts.FooterLink.URL,
		},
		Cols:    opts.Cols,
		MaxRows: opts.MaxRows,
	}
	return html.RenderStreamingPlaybackHTML(internalOpts)
}
