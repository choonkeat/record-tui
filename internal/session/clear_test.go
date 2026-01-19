package session

import (
	"strings"
	"testing"
)

// ClearSeparator is the visual separator used to replace clear sequences
const testClearSeparator = "\n\n──────── terminal cleared ────────\n\n"

func TestNeutralizeClearSequences_SimpleClear(t *testing.T) {
	// \x1b[2J is "clear entire screen"
	input := "first half\x1b[2Jsecond half"

	result := NeutralizeClearSequences(input)

	// Should contain both halves
	if !strings.Contains(result, "first half") {
		t.Errorf("Result should contain 'first half', got: %q", result)
	}
	if !strings.Contains(result, "second half") {
		t.Errorf("Result should contain 'second half', got: %q", result)
	}

	// Should contain separator
	if !strings.Contains(result, "terminal cleared") {
		t.Errorf("Result should contain separator 'terminal cleared', got: %q", result)
	}

	// Should NOT contain the escape sequence
	if strings.Contains(result, "\x1b[2J") {
		t.Errorf("Result should not contain clear escape sequence")
	}
}

func TestNeutralizeClearSequences_ClearWithHome(t *testing.T) {
	// \x1b[2J\x1b[H is common "clear and home" combination
	input := "first half\x1b[2J\x1b[Hsecond half"

	result := NeutralizeClearSequences(input)

	if !strings.Contains(result, "first half") {
		t.Errorf("Result should contain 'first half', got: %q", result)
	}
	if !strings.Contains(result, "second half") {
		t.Errorf("Result should contain 'second half', got: %q", result)
	}
	if !strings.Contains(result, "terminal cleared") {
		t.Errorf("Result should contain separator, got: %q", result)
	}

	// Should NOT contain the escape sequences
	if strings.Contains(result, "\x1b[2J") {
		t.Errorf("Result should not contain clear escape sequence")
	}
}

func TestNeutralizeClearSequences_HomeThenClear(t *testing.T) {
	// \x1b[H\x1b[2J is "home then clear" combination
	input := "first half\x1b[H\x1b[2Jsecond half"

	result := NeutralizeClearSequences(input)

	if !strings.Contains(result, "first half") {
		t.Errorf("Result should contain 'first half', got: %q", result)
	}
	if !strings.Contains(result, "second half") {
		t.Errorf("Result should contain 'second half', got: %q", result)
	}
	if !strings.Contains(result, "terminal cleared") {
		t.Errorf("Result should contain separator, got: %q", result)
	}
}

func TestNeutralizeClearSequences_ClearScrollback(t *testing.T) {
	// \x1b[3J is "clear screen including scrollback"
	input := "first half\x1b[3Jsecond half"

	result := NeutralizeClearSequences(input)

	if !strings.Contains(result, "first half") {
		t.Errorf("Result should contain 'first half', got: %q", result)
	}
	if !strings.Contains(result, "second half") {
		t.Errorf("Result should contain 'second half', got: %q", result)
	}
	if !strings.Contains(result, "terminal cleared") {
		t.Errorf("Result should contain separator, got: %q", result)
	}
}

func TestNeutralizeClearSequences_MultipleClears(t *testing.T) {
	// Multiple clears should each get a separator (or collapse to one)
	input := "part1\x1b[2Jpart2\x1b[2Jpart3"

	result := NeutralizeClearSequences(input)

	if !strings.Contains(result, "part1") {
		t.Errorf("Result should contain 'part1', got: %q", result)
	}
	if !strings.Contains(result, "part2") {
		t.Errorf("Result should contain 'part2', got: %q", result)
	}
	if !strings.Contains(result, "part3") {
		t.Errorf("Result should contain 'part3', got: %q", result)
	}
}

func TestNeutralizeClearSequences_NoClear(t *testing.T) {
	// Content without clear sequences should pass through unchanged
	input := "normal content\nwith newlines\nno clears"

	result := NeutralizeClearSequences(input)

	if result != input {
		t.Errorf("Result should be unchanged, got: %q", result)
	}
}

func TestNeutralizeClearSequences_ClearAtStart(t *testing.T) {
	// Clear at start should just be stripped (no separator needed)
	input := "\x1b[2Jcontent after clear"

	result := NeutralizeClearSequences(input)

	if !strings.Contains(result, "content after clear") {
		t.Errorf("Result should contain 'content after clear', got: %q", result)
	}
	// Should not have separator at the very start
	if strings.HasPrefix(strings.TrimSpace(result), "────") {
		t.Errorf("Result should not start with separator, got: %q", result)
	}
}

func TestNeutralizeClearSequences_ClearAtEnd(t *testing.T) {
	// Clear at end should just be stripped (no separator needed)
	input := "content before clear\x1b[2J"

	result := NeutralizeClearSequences(input)

	if !strings.Contains(result, "content before clear") {
		t.Errorf("Result should contain 'content before clear', got: %q", result)
	}
	// Should not have separator at the very end
	if strings.HasSuffix(strings.TrimSpace(result), "────") {
		t.Errorf("Result should not end with separator, got: %q", result)
	}
}

func TestNeutralizeClearSequences_OnlyClears(t *testing.T) {
	// Content that is only clear sequences should return empty or minimal
	input := "\x1b[2J\x1b[H\x1b[3J"

	result := NeutralizeClearSequences(input)

	// Result should be empty or just whitespace
	if strings.TrimSpace(result) != "" {
		t.Errorf("Result should be empty for only clears, got: %q", result)
	}
}

func TestNeutralizeClearSequences_PreservesOtherANSI(t *testing.T) {
	// Other ANSI sequences (colors, etc.) should be preserved
	input := "\x1b[31mred text\x1b[0m\x1b[2J\x1b[32mgreen text\x1b[0m"

	result := NeutralizeClearSequences(input)

	// Color sequences should remain
	if !strings.Contains(result, "\x1b[31m") {
		t.Errorf("Result should preserve red color sequence, got: %q", result)
	}
	if !strings.Contains(result, "\x1b[32m") {
		t.Errorf("Result should preserve green color sequence, got: %q", result)
	}
	// But clear should be replaced
	if strings.Contains(result, "\x1b[2J") {
		t.Errorf("Result should not contain clear sequence")
	}
}
