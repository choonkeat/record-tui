package session

import (
	"os"
	"strings"
	"testing"
)

// TestStripMetadata_RealSessionLog tests the cleaner on the actual recorded session.log
// This test reads from the file if it exists, otherwise it's skipped
func TestStripMetadata_RealSessionLog(t *testing.T) {
	sessionPath := os.ExpandEnv("$HOME/.record-tui/20251231-121034/session.log")

	// Check if file exists
	_, err := os.Stat(sessionPath)
	if err != nil {
		t.Skipf("Test session.log not found at %s, skipping real file test", sessionPath)
	}

	// Read the file
	content, err := os.ReadFile(sessionPath)
	if err != nil {
		t.Fatalf("Failed to read session.log: %v", err)
	}

	sessionContent := string(content)

	// Run the stripper
	result := StripMetadata(sessionContent)

	// Verify results
	if result == "" {
		t.Errorf("Result should not be empty")
	}

	// Should not contain metadata markers
	if strings.Contains(result, "Script started on") {
		t.Errorf("Result should not contain 'Script started on'")
	}
	if strings.Contains(result, "Script done on") {
		t.Errorf("Result should not contain 'Script done on'")
	}
	if strings.Contains(result, "Command exit status") {
		t.Errorf("Result should not contain 'Command exit status'")
	}

	// Should contain actual terminal content (ANSI codes preserved)
	if !strings.Contains(result, "[91m") {
		t.Logf("Warning: No ANSI color codes found in result (expected in real recordings)")
	}

	t.Logf("✓ Successfully stripped real session.log: %d bytes → %d bytes", len(sessionContent), len(result))
}
