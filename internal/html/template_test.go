package html

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderPlaybackHTML_WithFrames(t *testing.T) {
	frames := []PlaybackFrame{
		{
			Timestamp: 0.0,
			Content:   "Hello World",
		},
	}

	html, err := RenderPlaybackHTML(frames, "", FooterLink{})
	if err != nil {
		t.Fatalf("RenderPlaybackHTML failed: %v", err)
	}

	// Check basic structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("HTML should contain DOCTYPE")
	}
	if !strings.Contains(html, "<html lang=\"en\">") {
		t.Errorf("HTML should contain html tag with lang")
	}
	if !strings.Contains(html, "xterm") {
		t.Errorf("HTML should reference xterm.js")
	}
	if !strings.Contains(html, "framesBase64") {
		t.Errorf("HTML should contain framesBase64 variable")
	}
	if !strings.Contains(html, "const frames = JSON.parse") {
		t.Errorf("HTML should parse frames from base64")
	}
}

func TestRenderPlaybackHTML_ValidBase64Encoding(t *testing.T) {
	frames := []PlaybackFrame{
		{
			Timestamp: 0.0,
			Content:   "Test content with unicode: ü ö ä",
		},
	}

	html, err := RenderPlaybackHTML(frames, "", FooterLink{})
	if err != nil {
		t.Fatalf("RenderPlaybackHTML failed: %v", err)
	}

	// Extract base64 string from HTML
	startMarker := "const framesBase64 = '"
	startIdx := strings.Index(html, startMarker)
	if startIdx == -1 {
		t.Fatalf("Could not find framesBase64 in HTML")
	}

	startIdx += len(startMarker)
	endIdx := strings.Index(html[startIdx:], "'")
	if endIdx == -1 {
		t.Fatalf("Could not find end of framesBase64")
	}

	framesBase64 := html[startIdx : startIdx+endIdx]

	// Decode and verify
	decoded, err := base64.StdEncoding.DecodeString(framesBase64)
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	var decodedFrames []PlaybackFrame
	err = json.Unmarshal(decoded, &decodedFrames)
	if err != nil {
		t.Fatalf("Failed to unmarshal frames JSON: %v", err)
	}

	// Verify the content matches
	if len(decodedFrames) != 1 {
		t.Errorf("Expected 1 frame, got %d", len(decodedFrames))
	}
	if decodedFrames[0].Content != "Test content with unicode: ü ö ä" {
		t.Errorf("Content mismatch: expected %q, got %q",
			"Test content with unicode: ü ö ä",
			decodedFrames[0].Content)
	}
}

func TestRenderPlaybackHTML_MultipleFrames(t *testing.T) {
	frames := []PlaybackFrame{
		{
			Timestamp: 0.0,
			Content:   "Frame 1",
		},
		{
			Timestamp: 1.0,
			Content:   "Frame 1\nFrame 2",
		},
		{
			Timestamp: 2.0,
			Content:   "Frame 1\nFrame 2\nFrame 3",
		},
	}

	html, err := RenderPlaybackHTML(frames, "", FooterLink{})
	if err != nil {
		t.Fatalf("RenderPlaybackHTML failed: %v", err)
	}

	// Extract and decode frames
	startMarker := "const framesBase64 = '"
	startIdx := strings.Index(html, startMarker)
	startIdx += len(startMarker)
	endIdx := strings.Index(html[startIdx:], "'")
	framesBase64 := html[startIdx : startIdx+endIdx]

	decoded, _ := base64.StdEncoding.DecodeString(framesBase64)
	var decodedFrames []PlaybackFrame
	json.Unmarshal(decoded, &decodedFrames)

	if len(decodedFrames) != 3 {
		t.Errorf("Expected 3 frames, got %d", len(decodedFrames))
	}
}

func TestRenderPlaybackHTML_ANSICodes(t *testing.T) {
	frames := []PlaybackFrame{
		{
			Timestamp: 0.0,
			Content:   "\x1b[91mRed text\x1b[39m",
		},
	}

	html, err := RenderPlaybackHTML(frames, "", FooterLink{})
	if err != nil {
		t.Fatalf("RenderPlaybackHTML failed: %v", err)
	}

	// Extract and decode frames
	startMarker := "const framesBase64 = '"
	startIdx := strings.Index(html, startMarker)
	startIdx += len(startMarker)
	endIdx := strings.Index(html[startIdx:], "'")
	framesBase64 := html[startIdx : startIdx+endIdx]

	decoded, _ := base64.StdEncoding.DecodeString(framesBase64)
	var decodedFrames []PlaybackFrame
	json.Unmarshal(decoded, &decodedFrames)

	// ANSI codes should be preserved
	if !strings.Contains(decodedFrames[0].Content, "\x1b[91m") {
		t.Errorf("ANSI codes should be preserved in content")
	}
}

func TestRenderPlaybackHTML_EmptyFrames(t *testing.T) {
	frames := []PlaybackFrame{}

	html, err := RenderPlaybackHTML(frames, "", FooterLink{})
	if err != nil {
		t.Fatalf("RenderPlaybackHTML failed: %v", err)
	}

	// Should still generate valid HTML
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("Should generate valid HTML even with empty frames")
	}

	// Extract and verify
	startMarker := "const framesBase64 = '"
	startIdx := strings.Index(html, startMarker)
	startIdx += len(startMarker)
	endIdx := strings.Index(html[startIdx:], "'")
	framesBase64 := html[startIdx : startIdx+endIdx]

	decoded, _ := base64.StdEncoding.DecodeString(framesBase64)
	var decodedFrames []PlaybackFrame
	json.Unmarshal(decoded, &decodedFrames)

	if len(decodedFrames) != 0 {
		t.Errorf("Expected 0 frames, got %d", len(decodedFrames))
	}
}

func TestRenderPlaybackHTML_DefaultFooter(t *testing.T) {
	frames := []PlaybackFrame{{Timestamp: 0, Content: "test"}}

	html, err := RenderPlaybackHTML(frames, "", FooterLink{})
	if err != nil {
		t.Fatalf("RenderPlaybackHTML failed: %v", err)
	}

	// Default footer should have "generated by" with record-tui link
	if !strings.Contains(html, "generated by") {
		t.Errorf("Footer should contain 'generated by'")
	}
	if !strings.Contains(html, `href="https://github.com/choonkeat/record-tui"`) {
		t.Errorf("Footer should link to record-tui repo")
	}
	if !strings.Contains(html, ">record-tui</a>") {
		t.Errorf("Footer should have record-tui as link text")
	}
}

func TestRenderPlaybackHTML_WithFooterLink(t *testing.T) {
	frames := []PlaybackFrame{{Timestamp: 0, Content: "test"}}

	footerLink := FooterLink{
		Text: "swe-swe",
		URL:  "https://github.com/choonkeat/swe-swe",
	}
	html, err := RenderPlaybackHTML(frames, "", footerLink)
	if err != nil {
		t.Fatalf("RenderPlaybackHTML failed: %v", err)
	}

	// Should have both links
	if !strings.Contains(html, ">record-tui</a>") {
		t.Errorf("Footer should have record-tui link")
	}
	if !strings.Contains(html, " x ") {
		t.Errorf("Footer should have ' x ' separator")
	}
	if !strings.Contains(html, `href="https://github.com/choonkeat/swe-swe"`) {
		t.Errorf("Footer should link to custom URL")
	}
	if !strings.Contains(html, ">swe-swe</a>") {
		t.Errorf("Footer should have custom link text")
	}
}

func TestRenderPlaybackHTML_FooterLinkEscaping(t *testing.T) {
	frames := []PlaybackFrame{{Timestamp: 0, Content: "test"}}

	// Test that HTML special chars are escaped
	footerLink := FooterLink{
		Text: "<script>alert('xss')</script>",
		URL:  "https://example.com?a=1&b=2",
	}
	html, err := RenderPlaybackHTML(frames, "", footerLink)
	if err != nil {
		t.Fatalf("RenderPlaybackHTML failed: %v", err)
	}

	// Script tag should be escaped
	if strings.Contains(html, "<script>alert") {
		t.Errorf("Footer text should be HTML escaped")
	}
	// & should be escaped in URL
	if strings.Contains(html, "?a=1&b=2") {
		t.Errorf("Footer URL should be HTML escaped")
	}
}
