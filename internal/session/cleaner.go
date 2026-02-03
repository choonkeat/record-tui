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
	hasFooterMarker := false
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		// Check if this line is a footer marker (must start with the marker text)
		if strings.HasPrefix(line, "Saving session") ||
			strings.HasPrefix(line, "Command exit status") ||
			strings.HasPrefix(line, "Script done on") {
			hasFooterMarker = true
			footerStartIndex = i
		} else if hasFooterMarker && strings.TrimSpace(line) == "" {
			// Only treat empty lines as footer if we already found a footer marker
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

	// Neutralize alternate screen buffer sequences first (before clear handling)
	// so it can find clear sequences that precede alt screen transitions
	content = NeutralizeAltScreenSequences(content)

	// Neutralize clear sequences so content before clears is preserved
	content = NeutralizeClearSequences(content)

	return content
}

// StripMetadataOnly removes only the session header and footer, without
// neutralizing clear or alt-screen sequences. This preserves the raw terminal
// output bytes, which is needed for byte-offset-based line number computation
// (e.g., TOC generation from timing files).
func StripMetadataOnly(content string) string {
	lines := strings.Split(content, "\n")

	startIndex := 0
	endIndex := len(lines)

	for i := 0; i < len(lines) && i < 5; i++ {
		line := lines[i]
		if strings.HasPrefix(line, "Script started on") || strings.HasPrefix(line, "Command:") {
			startIndex = i + 1
		}
	}

	footerStartIndex := len(lines)
	hasFooterMarker := false
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if strings.HasPrefix(line, "Saving session") ||
			strings.HasPrefix(line, "Command exit status") ||
			strings.HasPrefix(line, "Script done on") {
			hasFooterMarker = true
			footerStartIndex = i
		} else if hasFooterMarker && strings.TrimSpace(line) == "" {
			footerStartIndex = i
		} else if footerStartIndex < len(lines) {
			break
		}
	}
	endIndex = footerStartIndex

	for endIndex > startIndex && strings.TrimSpace(lines[endIndex-1]) == "" {
		endIndex--
	}

	if startIndex >= len(lines) || startIndex >= endIndex {
		return ""
	}
	return strings.Join(lines[startIndex:endIndex], "\n")
}
