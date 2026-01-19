package session

import (
	"strings"
)

// StripMetadata removes the session header and footer from raw session.log content.
// Removes patterns like:
// - Header: "Script started on ..." and "Command: ..."
// - Footer: "Saving session", "Command exit status", "Script done on"
func StripMetadata(content string) string {
	lines := strings.Split(content, "\n")

	startIndex := 0
	endIndex := len(lines)

	// Find where actual content starts (skip header)
	// The header consists of "Script started on..." followed by "Command: ..."
	for i := 0; i < len(lines) && i < 5; i++ {
		line := lines[i]
		if strings.HasPrefix(line, "Script started on") || strings.HasPrefix(line, "Command:") {
			startIndex = i + 1
		}
	}

	// Find where actual content ends (skip footer)
	// Footer can contain "Saving session", "Command exit status", "Script done on" in any order
	// Work backwards from end of file
	footerStartIndex := len(lines)
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		// Check if this line is part of the footer
		if strings.Contains(line, "Saving session") ||
			strings.Contains(line, "Command exit status") ||
			strings.Contains(line, "Script done on") ||
			(strings.TrimSpace(line) == "" && i > 0) {
			footerStartIndex = i
		} else if footerStartIndex < len(lines) {
			// We've found content before the footer, stop looking
			break
		}
	}
	endIndex = footerStartIndex

	// Trim any trailing empty lines from the content
	for endIndex > startIndex && strings.TrimSpace(lines[endIndex-1]) == "" {
		endIndex--
	}

	// Return the sliced content
	if startIndex >= len(lines) || startIndex >= endIndex {
		return ""
	}
	content = strings.Join(lines[startIndex:endIndex], "\n")

	// Neutralize clear sequences so content before clears is preserved
	content = NeutralizeClearSequences(content)

	return content
}
