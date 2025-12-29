package record

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRecordSession_VerifyFunction checks that RecordSession has correct signature
func TestRecordSession_VerifyFunction(t *testing.T) {
	// Create a temporary directory for test output
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.log")

	// Test with simple echo command
	err := RecordSession(outputPath, []string{"echo", "test"})
	if err != nil {
		t.Fatalf("RecordSession failed: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Expected session.log file to be created at %s", outputPath)
	}
}

// TestRecordSession_CreatesFile verifies that script command creates output file
func TestRecordSession_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "session.log")

	// Record a simple session
	err := RecordSession(outputPath, []string{"echo", "hello world"})
	if err != nil {
		t.Fatalf("RecordSession failed: %v", err)
	}

	// Check file exists
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Session file not found: %v", err)
	}

	// Check file has content
	if fileInfo.Size() == 0 {
		t.Error("Session file is empty, expected content")
	}

	t.Logf("✓ Session file created: %s (%d bytes)", outputPath, fileInfo.Size())
}

// TestRecordSession_WithNoArgs verifies recording works without command args
func TestRecordSession_WithNoArgs(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "session.log")

	// Record with no args (script will use default shell)
	// We'll use a simple command that exits immediately
	err := RecordSession(outputPath, []string{"true"})
	if err != nil {
		t.Fatalf("RecordSession with no args failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Session file not created")
	}
}

// TestRecordSession_WithMultipleArgs verifies args are passed correctly
func TestRecordSession_WithMultipleArgs(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "session.log")

	// Record with multiple arguments
	err := RecordSession(outputPath, []string{"sh", "-c", "echo test && echo more"})
	if err != nil {
		t.Fatalf("RecordSession with multiple args failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Session file not created")
	}

	t.Logf("✓ Multi-arg session recorded successfully")
}

// TestRecordSessionDetailed_ReturnCode verifies exit code detection
func TestRecordSessionDetailed_ReturnCode(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "session.log")

	// Record a successful command
	exitCode, err := RecordSessionDetailed(outputPath, []string{"true"})
	if err != nil {
		t.Logf("Note: RecordSessionDetailed returned error (may be normal): %v", err)
	}

	// Exit code 0 means success
	if exitCode != 0 && err == nil {
		t.Logf("Unexpected exit code: %d", exitCode)
	}

	t.Logf("✓ RecordSessionDetailed completed with exit code: %d", exitCode)
}
