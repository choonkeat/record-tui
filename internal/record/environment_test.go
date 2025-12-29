package record

import (
	"os"
	"testing"
)

// TestSetupRecordingEnvironment_SetsForceColor verifies FORCE_COLOR is set
func TestSetupRecordingEnvironment_SetsForceColor(t *testing.T) {
	// Clear before test
	os.Unsetenv("FORCE_COLOR")

	// Setup
	SetupRecordingEnvironment()

	// Verify
	value := os.Getenv("FORCE_COLOR")
	if value != "1" {
		t.Errorf("FORCE_COLOR should be '1', got '%s'", value)
	}

	t.Logf("✓ FORCE_COLOR set to: %s", value)
}

// TestSetupRecordingEnvironment_SetsColorterm verifies COLORTERM is set
func TestSetupRecordingEnvironment_SetsColorterm(t *testing.T) {
	// Clear before test
	os.Unsetenv("COLORTERM")

	// Setup
	SetupRecordingEnvironment()

	// Verify
	value := os.Getenv("COLORTERM")
	if value != "truecolor" {
		t.Errorf("COLORTERM should be 'truecolor', got '%s'", value)
	}

	t.Logf("✓ COLORTERM set to: %s", value)
}

// TestSetupRecordingEnvironment_BothVariables verifies both are set together
func TestSetupRecordingEnvironment_BothVariables(t *testing.T) {
	// Clear before test
	os.Unsetenv("FORCE_COLOR")
	os.Unsetenv("COLORTERM")

	// Setup
	SetupRecordingEnvironment()

	// Verify both
	forceColor := os.Getenv("FORCE_COLOR")
	colorTerm := os.Getenv("COLORTERM")

	if forceColor != "1" {
		t.Errorf("FORCE_COLOR not set correctly")
	}
	if colorTerm != "truecolor" {
		t.Errorf("COLORTERM not set correctly")
	}

	t.Logf("✓ Both environment variables set correctly")
}

// TestClearRecordingEnvironment_RemovesVariables verifies variables are cleared
func TestClearRecordingEnvironment_RemovesVariables(t *testing.T) {
	// Setup first
	SetupRecordingEnvironment()

	// Verify they exist
	if os.Getenv("FORCE_COLOR") == "" {
		t.Fatal("FORCE_COLOR should be set before clearing")
	}
	if os.Getenv("COLORTERM") == "" {
		t.Fatal("COLORTERM should be set before clearing")
	}

	// Clear
	ClearRecordingEnvironment()

	// Verify cleared
	if os.Getenv("FORCE_COLOR") != "" {
		t.Errorf("FORCE_COLOR should be cleared")
	}
	if os.Getenv("COLORTERM") != "" {
		t.Errorf("COLORTERM should be cleared")
	}

	t.Logf("✓ Environment variables cleared successfully")
}
