package html

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestStreamingServer starts a test server on port 3000 for manual browser testing.
// Run with: go test -run TestStreamingServer -v ./internal/html/...
// Then open http://localhost:3000 in a browser.
//
// The server serves:
//   - / : Streaming HTML that fetches ./session.log
//   - /session.log : The sample.log file from the project root
//
// Press Ctrl+C or wait 5 minutes for automatic shutdown.
func TestStreamingServer(t *testing.T) {
	// Skip in normal test runs - only run when explicitly requested
	if os.Getenv("RUN_STREAMING_SERVER") != "1" {
		t.Skip("Skipping test server (set RUN_STREAMING_SERVER=1 to run)")
	}

	// Find sample.log in project root
	projectRoot := findProjectRoot(t)
	sampleLogPath := filepath.Join(projectRoot, "sample.log")
	if _, err := os.Stat(sampleLogPath); os.IsNotExist(err) {
		t.Fatalf("sample.log not found at %s - please create it first", sampleLogPath)
	}

	// Generate streaming HTML
	htmlContent, err := RenderStreamingPlaybackHTML(StreamingOptions{
		Title:   "Streaming Test",
		DataURL: "./session.log",
		FooterLink: FooterLink{
			Text: "test",
			URL:  "http://localhost:3000",
		},
	})
	if err != nil {
		t.Fatalf("Failed to generate streaming HTML: %v", err)
	}

	// Create HTTP handlers
	mux := http.NewServeMux()

	// Serve streaming HTML at /
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(htmlContent))
	})

	// Serve sample.log as session.log
	mux.HandleFunc("/session.log", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		// Don't buffer - stream directly
		w.Header().Set("X-Accel-Buffering", "no")
		http.ServeFile(w, r, sampleLogPath)
	})

	// Start server
	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

	fmt.Println("Starting streaming test server on http://localhost:3000")
	fmt.Println("Open this URL in a browser to test streaming HTML")
	fmt.Println("Server will auto-shutdown in 5 minutes")
	fmt.Println("Press Ctrl+C to stop earlier")

	// Auto-shutdown after 5 minutes
	go func() {
		time.Sleep(5 * time.Minute)
		fmt.Println("\nAuto-shutting down after 5 minutes...")
		server.Close()
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		t.Fatalf("Server error: %v", err)
	}
}

// findProjectRoot walks up directories to find the project root (where go.mod is)
func findProjectRoot(t *testing.T) string {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (go.mod)")
		}
		dir = parent
	}
}
