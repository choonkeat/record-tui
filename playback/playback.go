package playback

import (
	"io"

	"github.com/choonkeat/record-tui/internal/html"
	"github.com/choonkeat/record-tui/internal/session"
	"github.com/choonkeat/record-tui/internal/timing"
	"github.com/choonkeat/record-tui/internal/toc"
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
	var tocEntries []html.TOCEntry
	if len(opts) > 0 {
		if opts[0].Title != "" {
			title = opts[0].Title
		}
		footerLink = html.FooterLink{
			Text: opts[0].FooterLink.Text,
			URL:  opts[0].FooterLink.URL,
		}
		for _, e := range opts[0].TOC {
			tocEntries = append(tocEntries, html.TOCEntry{
				Label: e.Label,
				Line:  e.Line,
			})
		}
	}

	return html.RenderPlaybackHTML(internalFrames, title, footerLink, tocEntries)
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
// BuildTOC generates table-of-contents entries from a timing file, input file,
// and session log content. Returns nil if parsing fails or no commands are found.
//
// Parameters:
//   - timingReader: reader for the timing file (advanced format with I/O/H/S markers)
//   - inputContent: raw bytes from the session.input file (may include script metadata)
//   - sessionContent: raw bytes from the session.log file (may include script metadata)
//
// Both inputContent and sessionContent have their script header/footer stripped
// automatically before processing.
//
// Example:
//
//	timingFile, _ := os.Open("session.timing")
//	inputBytes, _ := os.ReadFile("session.input")
//	sessionBytes, _ := os.ReadFile("session.log")
//	tocEntries := playback.BuildTOC(timingFile, inputBytes, sessionBytes)
func BuildTOC(timingReader io.Reader, inputContent []byte, sessionContent []byte) []TOCEntry {
	entries, err := timing.Parse(timingReader)
	if err != nil {
		return nil
	}

	strippedInput := []byte(session.StripMetadataOnly(string(inputContent)))
	commands := timing.ExtractCommands(entries, strippedInput)
	if len(commands) == 0 {
		return nil
	}

	// Get raw output (metadata stripped, but escape sequences preserved)
	rawOutput := session.StripMetadataOnly(string(sessionContent))

	// Apply neutralization (alt-screen + clear) with offset tracking.
	// The HTML template applies these same transformations before feeding
	// content to xterm.js, so line numbers must be computed from the
	// processed content to match the rendered output.
	processedOutput, mapOffset := session.NeutralizeAllWithOffsets(rawOutput)

	// Map each command's raw byte offset to the processed content offset
	mappedCommands := make([]timing.Command, len(commands))
	copy(mappedCommands, commands)
	for i := range mappedCommands {
		mappedCommands[i].OutputByteOffset = mapOffset(mappedCommands[i].OutputByteOffset)
	}

	tocRaw := toc.FromCommands(mappedCommands, []byte(processedOutput))

	result := make([]TOCEntry, len(tocRaw))
	for i, e := range tocRaw {
		result[i] = TOCEntry{Label: e.Label, Line: e.Line}
	}
	return result
}

func RenderStreamingHTML(opts StreamingOptions) (string, error) {
	var tocEntries []html.TOCEntry
	for _, e := range opts.TOC {
		tocEntries = append(tocEntries, html.TOCEntry{
			Label: e.Label,
			Line:  e.Line,
		})
	}

	internalOpts := html.StreamingOptions{
		Title:   opts.Title,
		DataURL: opts.DataURL,
		FooterLink: html.FooterLink{
			Text: opts.FooterLink.Text,
			URL:  opts.FooterLink.URL,
		},
		Cols:    opts.Cols,
		MaxRows: opts.MaxRows,
		TOC:     tocEntries,
	}
	return html.RenderStreamingPlaybackHTML(internalOpts)
}
