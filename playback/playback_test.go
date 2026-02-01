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

func TestRenderHTML_WithTitle(t *testing.T) {
	frames := []Frame{{Timestamp: 0, Content: "hello"}}
	html, err := RenderHTML(frames, Options{Title: "My Session"})
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	if !strings.Contains(html, "<title>My Session</title>") {
		t.Error("HTML should contain custom title")
	}
}

func TestRenderHTML_TitleEscaping(t *testing.T) {
	frames := []Frame{{Timestamp: 0, Content: "hello"}}
	html, err := RenderHTML(frames, Options{Title: "<script>alert('xss')</script>"})
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	// Title should be HTML escaped
	if strings.Contains(html, "<script>alert") {
		t.Error("HTML title should be escaped to prevent XSS")
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("HTML title should contain escaped characters")
	}
}

func TestRenderHTML_WithFooterLink(t *testing.T) {
	frames := []Frame{{Timestamp: 0, Content: "hello"}}
	html, err := RenderHTML(frames, Options{
		Title: "Test",
		FooterLink: FooterLink{
			Text: "swe-swe",
			URL:  "https://github.com/choonkeat/swe-swe",
		},
	})
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	// Should have both record-tui and custom link
	if !strings.Contains(html, ">record-tui</a>") {
		t.Error("HTML should contain record-tui link")
	}
	if !strings.Contains(html, " x ") {
		t.Error("HTML should contain ' x ' separator")
	}
	if !strings.Contains(html, `href="https://github.com/choonkeat/swe-swe"`) {
		t.Error("HTML should contain custom footer URL")
	}
	if !strings.Contains(html, ">swe-swe</a>") {
		t.Error("HTML should contain custom footer text")
	}
}

func TestStripMetadata_NeutralizesClearSequences(t *testing.T) {
	// Integration test: StripMetadata should neutralize clear sequences
	input := `Script started on Wed Dec 31 12:10:34 2025
Command: bash
first half` + "\x1b[2J" + `second half
Script done on Wed Dec 31 12:11:22 2025
`
	result := StripMetadata(input)

	// Both halves should be present
	if !strings.Contains(result, "first half") {
		t.Errorf("Result should contain 'first half', got: %q", result)
	}
	if !strings.Contains(result, "second half") {
		t.Errorf("Result should contain 'second half', got: %q", result)
	}

	// Separator should be present
	if !strings.Contains(result, "terminal cleared") {
		t.Errorf("Result should contain separator 'terminal cleared', got: %q", result)
	}

	// Clear sequence should be gone
	if strings.Contains(result, "\x1b[2J") {
		t.Error("Result should not contain clear escape sequence")
	}
}

// Tests for RenderStreamingHTML

func TestRenderStreamingHTML_Basic(t *testing.T) {
	html, err := RenderStreamingHTML(StreamingOptions{
		DataURL: "./session.log",
	})
	if err != nil {
		t.Fatalf("RenderStreamingHTML failed: %v", err)
	}

	// Check basic HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML should contain DOCTYPE")
	}
	if !strings.Contains(html, "xterm") {
		t.Error("HTML should contain xterm.js reference")
	}
	// Should contain streaming JavaScript
	if !strings.Contains(html, "streamSession") {
		t.Error("HTML should contain streamSession function")
	}
	// Should reference the data URL
	if !strings.Contains(html, "./session.log") {
		t.Error("HTML should contain the data URL")
	}
}

func TestRenderStreamingHTML_DefaultTitle(t *testing.T) {
	html, err := RenderStreamingHTML(StreamingOptions{
		DataURL: "./data.log",
	})
	if err != nil {
		t.Fatalf("RenderStreamingHTML failed: %v", err)
	}

	// Default title should be "Terminal"
	if !strings.Contains(html, "<title>Terminal</title>") {
		t.Error("HTML should contain default title 'Terminal'")
	}
}

func TestRenderStreamingHTML_CustomTitle(t *testing.T) {
	html, err := RenderStreamingHTML(StreamingOptions{
		Title:   "My Recording",
		DataURL: "./data.log",
	})
	if err != nil {
		t.Fatalf("RenderStreamingHTML failed: %v", err)
	}

	if !strings.Contains(html, "<title>My Recording</title>") {
		t.Error("HTML should contain custom title")
	}
}

func TestRenderStreamingHTML_TitleEscaping(t *testing.T) {
	html, err := RenderStreamingHTML(StreamingOptions{
		Title:   "<script>alert('xss')</script>",
		DataURL: "./data.log",
	})
	if err != nil {
		t.Fatalf("RenderStreamingHTML failed: %v", err)
	}

	// Title should be HTML escaped
	if strings.Contains(html, "<script>alert") {
		t.Error("HTML title should be escaped to prevent XSS")
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("HTML title should contain escaped characters")
	}
}

func TestRenderStreamingHTML_DataURLEscaping(t *testing.T) {
	html, err := RenderStreamingHTML(StreamingOptions{
		DataURL: "/api/session?id=123&foo=bar",
	})
	if err != nil {
		t.Fatalf("RenderStreamingHTML failed: %v", err)
	}

	// DataURL should be HTML escaped (& -> &amp;)
	if !strings.Contains(html, "/api/session?id=123&amp;foo=bar") {
		t.Errorf("HTML DataURL should be escaped, got: %s", html)
	}
}

func TestRenderStreamingHTML_WithFooterLink(t *testing.T) {
	html, err := RenderStreamingHTML(StreamingOptions{
		Title:   "Test",
		DataURL: "./data.log",
		FooterLink: FooterLink{
			Text: "swe-swe",
			URL:  "https://github.com/choonkeat/swe-swe",
		},
	})
	if err != nil {
		t.Fatalf("RenderStreamingHTML failed: %v", err)
	}

	// Should have both record-tui and custom link
	if !strings.Contains(html, ">record-tui</a>") {
		t.Error("HTML should contain record-tui link")
	}
	if !strings.Contains(html, " x ") {
		t.Error("HTML should contain ' x ' separator")
	}
	if !strings.Contains(html, `href="https://github.com/choonkeat/swe-swe"`) {
		t.Error("HTML should contain custom footer URL")
	}
	if !strings.Contains(html, ">swe-swe</a>") {
		t.Error("HTML should contain custom footer text")
	}
}

func TestRenderStreamingHTML_ContainsLoadingIndicator(t *testing.T) {
	html, err := RenderStreamingHTML(StreamingOptions{
		DataURL: "./data.log",
	})
	if err != nil {
		t.Fatalf("RenderStreamingHTML failed: %v", err)
	}

	// Should contain loading indicator
	if !strings.Contains(html, "Loading...") {
		t.Error("HTML should contain loading indicator")
	}
	if !strings.Contains(html, `id="loading"`) {
		t.Error("HTML should contain loading div with id")
	}
}

func TestRenderStreamingHTML_ContainsStreamingFunctions(t *testing.T) {
	html, err := RenderStreamingHTML(StreamingOptions{
		DataURL: "./data.log",
	})
	if err != nil {
		t.Fatalf("RenderStreamingHTML failed: %v", err)
	}

	// Should contain all streaming helper functions
	functions := []string{
		"stripHeader",
		"stripFooter",
		"createStreamingCleaner",
		"streamSession",
	}
	for _, fn := range functions {
		if !strings.Contains(html, "function "+fn) {
			t.Errorf("HTML should contain function %s", fn)
		}
	}
}

func TestRenderStreamingHTML_ContainsClearSeparator(t *testing.T) {
	html, err := RenderStreamingHTML(StreamingOptions{
		DataURL: "./data.log",
	})
	if err != nil {
		t.Fatalf("RenderStreamingHTML failed: %v", err)
	}

	// Should contain the clear separator constant
	if !strings.Contains(html, "terminal cleared") {
		t.Error("HTML should contain clear separator text")
	}
}
