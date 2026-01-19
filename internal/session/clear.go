package session

import (
	"regexp"
	"strings"
)

// ClearSeparator is the visual separator used to replace clear sequences
const ClearSeparator = "\n\n──────── terminal cleared ────────\n\n"

// clearPatterns matches terminal clear sequences:
// - \x1b[2J - Clear entire screen
// - \x1b[3J - Clear entire screen including scrollback
// - \x1b[H - Cursor home (only when combined with clear)
// Combined patterns: \x1b[H\x1b[2J or \x1b[2J\x1b[H
var clearPattern = regexp.MustCompile(`\x1b\[H\x1b\[[23]J|\x1b\[[23]J\x1b\[H|\x1b\[[23]J`)

// NeutralizeClearSequences replaces terminal clear sequences with a visual separator.
// This preserves content before and after clear commands in the rendered output.
// Clear sequences at the very start or end are simply stripped (no separator needed).
// Other ANSI sequences (colors, cursor movement, etc.) are preserved.
func NeutralizeClearSequences(content string) string {
	// Find all clear sequences
	matches := clearPattern.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return content
	}

	var result strings.Builder
	lastEnd := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		// Get content before this clear
		before := content[lastEnd:start]

		// Only add separator if there's non-empty content before
		if strings.TrimSpace(before) != "" {
			result.WriteString(before)

			// Check if there's content after this clear
			remaining := content[end:]
			if strings.TrimSpace(remaining) != "" {
				result.WriteString(ClearSeparator)
			}
		}

		lastEnd = end
	}

	// Add remaining content after the last clear
	remaining := content[lastEnd:]
	if strings.TrimSpace(remaining) != "" {
		// If we haven't written anything yet (clears were at start), just write content
		if result.Len() == 0 {
			result.WriteString(remaining)
		} else {
			result.WriteString(remaining)
		}
	}

	return result.String()
}
