// Package toc generates table-of-contents entries from timing data.
// It maps user commands (extracted from timing files) to line numbers
// in the terminal output, enabling navigation in the HTML viewer.
package toc

import (
	"bufio"
	"io"
	"sort"
	"strings"

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

// isScriptHeader returns true for lines added by the `script` command at the top.
func isScriptHeader(line string) bool {
	return strings.HasPrefix(line, "Script started on") || strings.HasPrefix(line, "Command:")
}

// FromCommandsReader computes TOC entries by streaming through an io.Reader
// instead of requiring the full session content in memory. Skips script
// header lines. Line numbers are approximate (no alt-screen/clear neutralization)
// but accurate enough for TOC navigation.
//
// Performance: O(max_offset) in streaming I/O, constant memory.
func FromCommandsReader(commands []timing.Command, r io.Reader) []Entry {
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

	// Find max offset to know when we can stop reading
	maxOffset := sorted[len(sorted)-1].cmd.OutputByteOffset

	// Stream through the reader, counting newlines and tracking byte position
	entries := make([]Entry, len(commands))
	bytePos := 0
	lineCount := 0
	cmdIdx := 0
	inHeader := true

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip script header lines at the start
		if inHeader && isScriptHeader(line) {
			// Account for the bytes consumed (line + newline)
			bytePos += len(line) + 1
			continue
		}
		inHeader = false

		// Process bytes in this line + the newline
		lineEnd := bytePos + len(line) + 1

		// Assign entries for any commands whose offset falls within this line
		for cmdIdx < len(sorted) && sorted[cmdIdx].cmd.OutputByteOffset < lineEnd {
			entries[sorted[cmdIdx].origIndex] = Entry{
				Label: sorted[cmdIdx].cmd.Text,
				Line:  lineCount,
			}
			cmdIdx++
		}

		lineCount++
		bytePos = lineEnd

		// Stop early if we've passed all command offsets
		if cmdIdx >= len(sorted) || bytePos > maxOffset {
			break
		}
	}

	// Handle any remaining commands beyond EOF
	for ; cmdIdx < len(sorted); cmdIdx++ {
		entries[sorted[cmdIdx].origIndex] = Entry{
			Label: sorted[cmdIdx].cmd.Text,
			Line:  lineCount,
		}
	}

	return entries
}
