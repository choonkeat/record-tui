package record

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/choonkeat/record-tui/playback"
)

// ConvertSessionToHTML reads a session.log file, strips metadata, and generates HTML output.
//
// This function:
// 1. Reads the session.log file
// 2. Strips the session metadata (header/footer from `script` command)
// 3. Creates a playback frame with the cleaned content
// 4. Generates HTML with xterm.js rendering
// 5. Writes to session.log.html
//
// Returns the path to the generated HTML file, or error if any step fails
func ConvertSessionToHTML(sessionLogPath string) (string, error) {
	// Validate input file exists
	if _, err := os.Stat(sessionLogPath); os.IsNotExist(err) {
		return "", fmt.Errorf("session.log not found: %s", sessionLogPath)
	}

	// Read session.log file
	sessionContent, err := os.ReadFile(sessionLogPath)
	if err != nil {
		return "", fmt.Errorf("cannot read session.log: %w", err)
	}

	// Strip session metadata (Script started/done lines from `script` command)
	cleanedContent := playback.StripMetadata(string(sessionContent))
	if cleanedContent == "" {
		return "", fmt.Errorf("session.log appears to be empty after metadata stripping")
	}

	// Create playback frame with all content at timestamp 0.0 (static display)
	frames := []playback.Frame{
		{
			Timestamp: 0.0,
			Content:   cleanedContent,
		},
	}

	// Generate HTML using xterm.js
	htmlContent, err := playback.RenderHTML(frames)
	if err != nil {
		return "", fmt.Errorf("failed to generate HTML: %w", err)
	}

	// Determine output path (same as input but with .html extension)
	outputPath := sessionLogPath + ".html"

	// Write HTML to file
	err = os.WriteFile(outputPath, []byte(htmlContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write HTML file: %w", err)
	}

	return outputPath, nil
}

// ConvertSessionToHTMLWithPath is like ConvertSessionToHTML but allows specifying output path
func ConvertSessionToHTMLWithPath(sessionLogPath string, outputPath string) (string, error) {
	// Validate input file exists
	if _, err := os.Stat(sessionLogPath); os.IsNotExist(err) {
		return "", fmt.Errorf("session.log not found: %s", sessionLogPath)
	}

	// Resolve to absolute paths
	sessionPath, err := filepath.Abs(sessionLogPath)
	if err != nil {
		return "", fmt.Errorf("invalid session path: %w", err)
	}

	outPath, err := filepath.Abs(outputPath)
	if err != nil {
		return "", fmt.Errorf("invalid output path: %w", err)
	}

	// Read session.log
	sessionContent, err := os.ReadFile(sessionPath)
	if err != nil {
		return "", fmt.Errorf("cannot read session.log: %w", err)
	}

	// Strip metadata
	cleanedContent := playback.StripMetadata(string(sessionContent))
	if cleanedContent == "" {
		return "", fmt.Errorf("session.log appears to be empty after metadata stripping")
	}

	// Create playback frame
	frames := []playback.Frame{
		{
			Timestamp: 0.0,
			Content:   cleanedContent,
		},
	}

	// Generate HTML
	htmlContent, err := playback.RenderHTML(frames)
	if err != nil {
		return "", fmt.Errorf("failed to generate HTML: %w", err)
	}

	// Write HTML
	err = os.WriteFile(outPath, []byte(htmlContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write HTML file: %w", err)
	}

	return outPath, nil
}

// ConvertSessionToStreamingHTML generates streaming HTML that fetches session data via JavaScript.
// Unlike ConvertSessionToHTML which embeds all content, this generates lightweight HTML (~10KB)
// that streams content from the log file. The HTML must be served via HTTP (not file://).
//
// Output is written to session.log.streaming.html
func ConvertSessionToStreamingHTML(sessionLogPath string) (string, error) {
	// Validate input file exists
	if _, err := os.Stat(sessionLogPath); os.IsNotExist(err) {
		return "", fmt.Errorf("session.log not found: %s", sessionLogPath)
	}

	// Generate streaming HTML that references the log file
	logFileName := filepath.Base(sessionLogPath)
	htmlContent, err := playback.RenderStreamingHTML(playback.StreamingOptions{
		Title:   logFileName,
		DataURL: "./" + logFileName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate streaming HTML: %w", err)
	}

	// Output path: session.log.streaming.html
	outputPath := sessionLogPath + ".streaming.html"

	// Write HTML to file
	err = os.WriteFile(outputPath, []byte(htmlContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write HTML file: %w", err)
	}

	return outputPath, nil
}
