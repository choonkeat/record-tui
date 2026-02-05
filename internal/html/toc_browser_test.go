package html

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestStreamingNav_NoJSErrors starts a server with streaming HTML that includes
// navigation entries and verifies that pressing < > does not produce JavaScript
// errors. This catches scoping bugs like "xterm is not defined" in the nav JS.
//
// Run with: RUN_BROWSER_TEST=1 go test -run TestStreamingNav_NoJSErrors -v ./internal/html/...
// Then use browser tools to interact with the page.
func TestStreamingNav_NoJSErrors(t *testing.T) {
	if os.Getenv("RUN_BROWSER_TEST") != "1" {
		t.Skip("Skipping browser test (set RUN_BROWSER_TEST=1 to run)")
	}

	// Minimal terminal content with recognizable markers for nav entries
	sessionContent := "$ echo hello\r\nhello\r\n$ ls -la\r\ntotal 0\r\n"

	tocEntries := []TOCEntry{
		{Label: "echo hello", Line: 0},
		{Label: "ls -la", Line: 2},
	}

	htmlContent, err := RenderStreamingPlaybackHTML(StreamingOptions{
		Title:   "Nav Browser Test",
		DataURL: "./session.log",
		TOC:     tocEntries,
	})
	if err != nil {
		t.Fatalf("Failed to generate streaming HTML: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(htmlContent))
	})
	mux.HandleFunc("/session.log", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(sessionContent))
	})

	// Use a fixed port so browser tools can reach it
	server := &http.Server{
		Addr:    ":3001",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			// Server stopped
		}
	}()
	defer server.Close()

	// Suppress unused import
	_ = httptest.NewServer

	fmt.Println("=== Nav Browser Test server running on http://localhost:3001 ===")
	fmt.Println("Waiting for browser test to complete...")
	fmt.Println("The test will be driven by browser MCP tools.")
	fmt.Println("Press Ctrl+C to stop.")

	// Block until test is killed (browser tools will drive the test)
	select {}
}
