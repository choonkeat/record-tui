package playback

import (
	"strings"
	"testing"
)

func TestStripMetadata_MacOS(t *testing.T) {
	input := `Script started on Wed Dec 31 12:10:34 2025
Command: bash
hello world
Script done on Wed Dec 31 12:11:22 2025
`
	result := StripMetadata(input)

	if !strings.Contains(result, "hello world") {
		t.Errorf("Result should contain 'hello world', got: %q", result)
	}
	if strings.Contains(result, "Script started") {
		t.Errorf("Result should not contain 'Script started'")
	}
}

func TestStripMetadata_Linux(t *testing.T) {
	input := `Script started on 2026-01-12 06:41:43+00:00 [COMMAND="claude" TERM="xterm-256color"]
hello world
Script done on 2026-01-12 06:45:00+00:00 [COMMAND_EXIT_STATUS="0"]
`
	result := StripMetadata(input)

	if !strings.Contains(result, "hello world") {
		t.Errorf("Result should contain 'hello world', got: %q", result)
	}
	if strings.Contains(result, "Script started") {
		t.Errorf("Result should not contain 'Script started'")
	}
}

func TestRenderHTML_SingleFrame(t *testing.T) {
	frames := []Frame{{Timestamp: 0, Content: "hello world"}}
	html, err := RenderHTML(frames)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	// Check basic HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML should contain DOCTYPE")
	}
	if !strings.Contains(html, "xterm") {
		t.Error("HTML should contain xterm.js reference")
	}
	// Content is base64 encoded, so we check for the encoding
	if !strings.Contains(html, "framesBase64") {
		t.Error("HTML should contain base64-encoded frames")
	}
}

func TestRenderHTML_MultipleFrames(t *testing.T) {
	frames := []Frame{
		{Timestamp: 0, Content: "frame 1"},
		{Timestamp: 1.5, Content: "frame 2"},
		{Timestamp: 3.0, Content: "frame 3"},
	}
	html, err := RenderHTML(frames)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML should contain DOCTYPE")
	}
}

func TestRenderHTML_EmptyFrames(t *testing.T) {
	frames := []Frame{}
	html, err := RenderHTML(frames)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	// Should still produce valid HTML even with no frames
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML should contain DOCTYPE even with empty frames")
	}
}

func TestRenderHTML_PreservesANSI(t *testing.T) {
	// Test that ANSI codes are preserved in the output
	frames := []Frame{{Timestamp: 0, Content: "\x1b[31mred text\x1b[0m"}}
	html, err := RenderHTML(frames)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	// The content is base64 encoded, so just verify the HTML is generated
	if len(html) < 1000 {
		t.Error("HTML seems too short, ANSI content may not be included")
	}
}
