package session

import (
	"regexp"
	"strings"
)

// ClearSeparator is the visual separator used to replace clear sequences
const ClearSeparator = "\n\n──────── terminal cleared ────────\n\n"

// AltScreenSeparator is the visual separator used when exiting the alternate screen buffer
const AltScreenSeparator = "\n\n──────── alternate screen ────────\n\n"

// altScreenPattern matches alternate screen buffer sequences:
// - \x1b[?1049h / \x1b[?1049l - xterm alternate screen (most common)
// - \x1b[?47h / \x1b[?47l - older alternate screen
// - \x1b[?1047h / \x1b[?1047l - alternate screen variant
var altScreenPattern = regexp.MustCompile(`\x1b\[\?(1049|47|1047)[hl]`)

// clearPatterns matches terminal clear sequences:
// - \x1b[2J - Clear entire screen
// - \x1b[3J - Clear entire screen including scrollback
// - \x1b[J / \x1b[0J - Erase from cursor to end of screen
// - \x1b[H or \x1b[1;1H - Cursor home (only when combined with erase)
// Combined patterns: \x1b[H\x1b[2J, \x1b[2J\x1b[H, \x1b[1;1H\x1b[J, etc.
var clearPattern = regexp.MustCompile(`\x1b\[(?:1;1)?H\x1b\[(?:0?J|[23]J)|\x1b\[[23]J\x1b\[H|\x1b\[[23]J`)

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

// NeutralizeAltScreenSequences removes alternate screen buffer regions and the
// TUI content that precedes them. Content between enter (\x1b[?1049h) and leave
// (\x1b[?1049l) is discarded. Additionally, content from the last clear sequence
// before the enter is also discarded, since it contains cursor-positioned TUI
// redraws that would corrupt the rendering of content after the leave.
// A separator is inserted at the transition point when there is content on both sides.
//
// This function should be called BEFORE NeutralizeClearSequences so it can find
// the clear sequences that precede alt screen transitions.
func NeutralizeAltScreenSequences(content string) string {
	altMatches := altScreenPattern.FindAllStringSubmatchIndex(content, -1)
	if len(altMatches) == 0 {
		return content
	}

	// Find all clear sequences to identify TUI redraw boundaries
	clearMatches := clearPattern.FindAllStringIndex(content, -1)

	var result strings.Builder
	lastEnd := 0
	inAltScreen := false

	for _, match := range altMatches {
		start, end := match[0], match[1]
		isEnter := content[end-1] == 'h'

		if isEnter && !inAltScreen {
			// Find the first clear sequence before this enter (after lastEnd)
			// Everything from that clear to the alt screen leave is TUI content
			// (the TUI app clears the screen and redraws repeatedly)
			stripFrom := start
			for _, cm := range clearMatches {
				if cm[0] >= lastEnd && cm[1] <= start {
					stripFrom = cm[0]
					break // Use the first clear, not the last
				}
			}

			// Keep content before the strip point
			before := content[lastEnd:stripFrom]
			result.WriteString(before)
			inAltScreen = true
		} else if !isEnter && inAltScreen {
			// Leaving alt screen — insert separator if content on both sides
			inAltScreen = false
			beforeContent := result.String()
			remaining := content[end:]
			if strings.TrimSpace(beforeContent) != "" && strings.TrimSpace(remaining) != "" {
				result.WriteString(AltScreenSeparator)
			}
		}

		lastEnd = end
	}

	// Add remaining content after the last match (only if not inside alt screen)
	if !inAltScreen {
		remaining := content[lastEnd:]
		result.WriteString(remaining)
	}

	return result.String()
}
