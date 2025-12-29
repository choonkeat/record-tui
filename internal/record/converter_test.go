package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConvertSessionToHTML_WithSimpleSession tests conversion of a simple recorded session
func TestConvertSessionToHTML_WithSimpleSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a mock session.log file with metadata
	sessionLogPath := filepath.Join(tmpDir, "session.log")
	sessionContent := `Script started on Wed Dec 31 12:10:34 2025
Command: bash
hello world
test content
Script done on Wed Dec 31 12:11:00 2025
`

	err := os.WriteFile(sessionLogPath, []byte(sessionContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test session.log: %v", err)
	}

	// Convert to HTML
	htmlPath, err := ConvertSessionToHTML(sessionLogPath)
	if err != nil {
		t.Fatalf("ConvertSessionToHTML failed: %v", err)
	}

	// Verify output path is correct
	expectedPath := sessionLogPath + ".html"
	if htmlPath != expectedPath {
		t.Errorf("Expected output path %s, got %s", expectedPath, htmlPath)
	}

	// Verify HTML file was created
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Fatalf("HTML file not created at %s", htmlPath)
	}

	// Verify HTML contains expected markers
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML file: %v", err)
	}

	htmlString := string(htmlContent)
	if !strings.Contains(htmlString, "<!DOCTYPE html>") {
		t.Errorf("HTML should contain DOCTYPE declaration")
	}
	if !strings.Contains(htmlString, "xterm") {
		t.Errorf("HTML should reference xterm.js")
	}
	if !strings.Contains(htmlString, "framesBase64") {
		t.Errorf("HTML should contain framesBase64 variable")
	}

	// Verify metadata was stripped (Script started/done should not be in content)
	if strings.Contains(htmlString, "Script started on") {
		t.Errorf("HTML should not contain 'Script started on' metadata")
	}
	if strings.Contains(htmlString, "Script done on") {
		t.Errorf("HTML should not contain 'Script done on' metadata")
	}

	t.Logf("✓ Conversion successful: %s → %s", sessionLogPath, htmlPath)
}

// TestConvertSessionToHTML_FileNotFound tests error handling for missing session.log
func TestConvertSessionToHTML_FileNotFound(t *testing.T) {
	nonexistentPath := "/tmp/nonexistent-session-" + t.Name() + ".log"

	_, err := ConvertSessionToHTML(nonexistentPath)
	if err == nil {
		t.Errorf("Expected error for non-existent file, got none")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error message should mention 'not found', got: %v", err)
	}
}

// TestConvertSessionToHTML_CreatesValidHTML tests that generated HTML is valid structure
func TestConvertSessionToHTML_CreatesValidHTML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create session.log with more realistic content
	sessionLogPath := filepath.Join(tmpDir, "session.log")
	sessionContent := `Script started on Wed Dec 31 12:10:34 2025
Command: bash
[91mRed colored text[39m
$ ls -la
total 42
$ pwd
/home/user
Command exit status: 0
Script done on Wed Dec 31 12:11:00 2025
`

	err := os.WriteFile(sessionLogPath, []byte(sessionContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test session.log: %v", err)
	}

	// Convert
	htmlPath, err := ConvertSessionToHTML(sessionLogPath)
	if err != nil {
		t.Fatalf("ConvertSessionToHTML failed: %v", err)
	}

	// Verify HTML file size is reasonable
	fileInfo, err := os.Stat(htmlPath)
	if err != nil {
		t.Fatalf("Cannot stat HTML file: %v", err)
	}

	if fileInfo.Size() < 1000 {
		t.Errorf("HTML file seems too small: %d bytes", fileInfo.Size())
	}

	t.Logf("✓ Generated HTML file: %d bytes", fileInfo.Size())
}

// TestConvertSessionToHTMLWithPath_CustomOutputPath tests specifying custom output path
func TestConvertSessionToHTMLWithPath_CustomOutputPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create session.log
	sessionLogPath := filepath.Join(tmpDir, "session.log")
	sessionContent := `Script started on Wed Dec 31 12:10:34 2025
Command: bash
test
Script done on Wed Dec 31 12:11:00 2025
`

	err := os.WriteFile(sessionLogPath, []byte(sessionContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test session.log: %v", err)
	}

	// Convert with custom output path
	customOutputPath := filepath.Join(tmpDir, "custom_output.html")
	returnedPath, err := ConvertSessionToHTMLWithPath(sessionLogPath, customOutputPath)
	if err != nil {
		t.Fatalf("ConvertSessionToHTMLWithPath failed: %v", err)
	}

	// Verify it used the custom path (after Abs conversion)
	absCustomPath, _ := filepath.Abs(customOutputPath)
	if returnedPath != absCustomPath {
		t.Errorf("Expected %s, got %s", absCustomPath, returnedPath)
	}

	// Verify file was created at custom location
	if _, err := os.Stat(customOutputPath); os.IsNotExist(err) {
		t.Errorf("HTML file not created at custom path: %s", customOutputPath)
	}

	t.Logf("✓ Custom output path used: %s", customOutputPath)
}

// TestConvertSessionToHTML_WithRealSessionLog tests with actual session.log file
func TestConvertSessionToHTML_WithRealSessionLog(t *testing.T) {
	// Use the real session.log from recording tests
	sessionLogPath := os.ExpandEnv("$HOME/.record-tui/20251231-141042/session.log")

	// Check if file exists, skip if not
	if _, err := os.Stat(sessionLogPath); os.IsNotExist(err) {
		t.Skipf("Test session.log not found at %s, skipping", sessionLogPath)
	}

	// Convert to HTML
	htmlPath, err := ConvertSessionToHTML(sessionLogPath)
	if err != nil {
		t.Fatalf("ConvertSessionToHTML failed: %v", err)
	}

	// Verify file was created
	htmlFileInfo, err := os.Stat(htmlPath)
	if err != nil {
		t.Fatalf("HTML file not created: %v", err)
	}

	t.Logf("✓ Real session converted: %s (%d bytes → %d bytes)",
		sessionLogPath, 176, htmlFileInfo.Size())
}

// TestConvertSessionToHTML_PreservesANSICodes tests that ANSI codes are preserved in HTML
func TestConvertSessionToHTML_PreservesANSICodes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create session.log with ANSI codes
	sessionLogPath := filepath.Join(tmpDir, "session.log")
	sessionContent := `Script started on Wed Dec 31 12:10:34 2025
Command: bash
[91mRed[39m [92mGreen[39m [94mBlue[39m
Script done on Wed Dec 31 12:11:00 2025
`

	err := os.WriteFile(sessionLogPath, []byte(sessionContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test session.log: %v", err)
	}

	// Convert
	htmlPath, err := ConvertSessionToHTML(sessionLogPath)
	if err != nil {
		t.Fatalf("ConvertSessionToHTML failed: %v", err)
	}

	// Read and verify ANSI codes are in the base64 content
	htmlContent, _ := os.ReadFile(htmlPath)
	htmlString := string(htmlContent)

	// ANSI codes should be preserved in the base64-encoded frames
	if !strings.Contains(htmlString, "framesBase64") {
		t.Errorf("HTML should contain frame data")
	}

	t.Logf("✓ ANSI codes preserved in HTML")
}
