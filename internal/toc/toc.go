// Package toc generates table-of-contents entries from timing data.
// It maps user commands (extracted from timing files) to line numbers
// in the terminal output, enabling navigation in the HTML viewer.
package toc

import (
	"sort"

	"github.com/choonkeat/record-tui/internal/timing"
)

// Entry represents a single TOC navigation point.
type Entry struct {
	Label string // What the user typed (e.g., "npm test")
	Line  int    // Line number in the output (0-indexed, for xterm.js scrolling)
}

// FromCommands computes TOC entries from timing commands and raw output content.
// rawOutput should be the session.log content with header/footer already stripped
// (just the terminal output bytes, before clear/alt-screen transformations).
// Line numbers are computed by counting newlines up to each command's output byte offset.
//
// Performance: O(len(rawOutput)) regardless of number of commands — newlines are
// counted in a single incremental pass.
func FromCommands(commands []timing.Command, rawOutput []byte) []Entry {
	if len(commands) == 0 {
		return nil
	}

	// Sort commands by offset for incremental newline counting
	type indexedCmd struct {
		origIndex int
		cmd       timing.Command
	}
	sorted := make([]indexedCmd, len(commands))
	for i, cmd := range commands {
		sorted[i] = indexedCmd{origIndex: i, cmd: cmd}
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].cmd.OutputByteOffset < sorted[j].cmd.OutputByteOffset
	})

	// Single pass: count newlines incrementally
	entries := make([]Entry, len(commands))
	prevOffset := 0
	lineCount := 0
	for _, sc := range sorted {
		offset := sc.cmd.OutputByteOffset
		if offset > len(rawOutput) {
			offset = len(rawOutput)
		}
		for i := prevOffset; i < offset; i++ {
			if rawOutput[i] == '\n' {
				lineCount++
			}
		}
		prevOffset = offset
		entries[sc.origIndex] = Entry{
			Label: sc.cmd.Text,
			Line:  lineCount,
		}
	}

	return entries
}
