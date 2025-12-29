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

	html, err := RenderPlaybackHTML(frames)
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

	html, err := RenderPlaybackHTML(frames)
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

	html, err := RenderPlaybackHTML(frames)
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

	html, err := RenderPlaybackHTML(frames)
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

	html, err := RenderPlaybackHTML(frames)
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
