package session

import (
	"strings"
	"testing"
)

func TestStripMetadata_SimpleCase(t *testing.T) {
	input := `Script started on Wed Dec 31 12:10:34 2025
Command: bash
hello world
Command exit status: 0
Script done on Wed Dec 31 12:11:22 2025
`

	result := StripMetadata(input)

	// Should contain the actual content
	if !strings.Contains(result, "hello world") {
		t.Errorf("Result should contain 'hello world', got: %q", result)
	}

	// Should not contain the metadata
	if strings.Contains(result, "Script started") {
		t.Errorf("Result should not contain 'Script started'")
	}
	if strings.Contains(result, "Script done") {
		t.Errorf("Result should not contain 'Script done'")
	}
	if strings.Contains(result, "Command exit status") {
		t.Errorf("Result should not contain 'Command exit status'")
	}
}

func TestStripMetadata_NoContent(t *testing.T) {
	input := `Script started on Wed Dec 31 12:10:34 2025
Command: bash
Command exit status: 0
Script done on Wed Dec 31 12:11:22 2025
`

	result := StripMetadata(input)
	result = strings.TrimSpace(result)

	if result != "" {
		t.Errorf("Result should be empty for no content, got: %q", result)
	}
}

func TestStripMetadata_WithTrailingEmptyLines(t *testing.T) {
	input := `Script started on Wed Dec 31 12:10:34 2025
Command: bash
test content

Command exit status: 0
Script done on Wed Dec 31 12:11:22 2025
`

	result := StripMetadata(input)

	if !strings.Contains(result, "test content") {
		t.Errorf("Result should contain 'test content'")
	}

	// Trailing empty lines should be removed
	lines := strings.Split(result, "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		t.Errorf("Result should not have trailing empty lines")
	}
}

func TestStripMetadata_MultilineContent(t *testing.T) {
	input := `Script started on Wed Dec 31 12:10:34 2025
Command: bash
line 1
line 2
line 3
Command exit status: 0
Script done on Wed Dec 31 12:11:22 2025
`

	result := StripMetadata(input)

	if !strings.Contains(result, "line 1") {
		t.Errorf("Result should contain 'line 1'")
	}
	if !strings.Contains(result, "line 2") {
		t.Errorf("Result should contain 'line 2'")
	}
	if !strings.Contains(result, "line 3") {
		t.Errorf("Result should contain 'line 3'")
	}
}

func TestStripMetadata_NoHeaderOrFooter(t *testing.T) {
	// Edge case: content without proper header/footer
	input := `some random content
without metadata
`

	result := StripMetadata(input)

	// Should return content as-is
	if !strings.Contains(result, "some random content") {
		t.Errorf("Result should contain 'some random content'")
	}
	if !strings.Contains(result, "without metadata") {
		t.Errorf("Result should contain 'without metadata'")
	}
}

func TestStripMetadata_OnlyCommand(t *testing.T) {
	input := `Script started on Wed Dec 31 12:10:34 2025
Command: /bin/bash
hello world
`

	result := StripMetadata(input)

	if !strings.Contains(result, "hello world") {
		t.Errorf("Result should contain 'hello world'")
	}
}
