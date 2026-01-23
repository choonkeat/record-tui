package session

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const (
	// compareOutputDir is the directory containing .go.output and .js.output files
	compareOutputDir = "../../recordings-output"
)

// TestCompareGoAndJsOutput compares .go.output files with corresponding .js.output files.
// Any difference is reported with debug-friendly output including:
// - File sizes
// - First differing byte position
// - Hex dump around the difference
// - diff -u output
func TestCompareGoAndJsOutput(t *testing.T) {
	// Check if output directory exists
	if _, err := os.Stat(compareOutputDir); os.IsNotExist(err) {
		t.Skipf("WARNING: output directory %q does not exist, skipping comparison", compareOutputDir)
	}

	// Find all .go.output files
	goPattern := filepath.Join(compareOutputDir, "*.go.output")
	goFiles, err := filepath.Glob(goPattern)
	if err != nil {
		t.Fatalf("failed to glob pattern %q: %v", goPattern, err)
	}

	if len(goFiles) == 0 {
		t.Skipf("WARNING: no .go.output files found in %q, skipping comparison", compareOutputDir)
	}

	// Sort for deterministic ordering
	sort.Strings(goFiles)

	t.Logf("Found %d .go.output files to compare", len(goFiles))

	// Track mismatches, live recordings, and missing files
	var mismatches []string
	var liveRecordings []string
	var missingJS []string
	var missingGo []string

	for _, goFile := range goFiles {
		// Derive corresponding .js.output filename
		baseName := strings.TrimSuffix(filepath.Base(goFile), ".go.output")
		jsFile := filepath.Join(compareOutputDir, baseName+".js.output")

		// Check if JS file exists
		if _, err := os.Stat(jsFile); os.IsNotExist(err) {
			missingJS = append(missingJS, baseName)
			t.Errorf("MISSING JS OUTPUT: %s.js.output (Go output exists)", baseName)
			continue
		}

		// Read both files
		goContent, err := os.ReadFile(goFile)
		if err != nil {
			t.Errorf("failed to read %q: %v", goFile, err)
			continue
		}

		jsContent, err := os.ReadFile(jsFile)
		if err != nil {
			t.Errorf("failed to read %q: %v", jsFile, err)
			continue
		}

		// Parse input lengths from headers (format: "{length} bytes\n...")
		goInputLen, goBody := parseOutputFile(goContent)
		jsInputLen, jsBody := parseOutputFile(jsContent)

		// Compare
		if goInputLen != jsInputLen {
			// Input lengths differ = live recording (file changed between reads)
			liveRecordings = append(liveRecordings, baseName)
			t.Logf("LIVE RECORDING (warning): %s - input sizes differ (Go read %d bytes, JS read %d bytes)",
				baseName, goInputLen, jsInputLen)
		} else if bytes.Equal(goBody, jsBody) {
			t.Logf("MATCH: %s (input: %d bytes, output: %d bytes)", baseName, goInputLen, len(goBody))
		} else {
			mismatches = append(mismatches, baseName)
			reportMismatch(t, baseName, goFile, jsFile, goBody, jsBody)
		}
	}

	// Check for .js.output files without corresponding .go.output
	jsPattern := filepath.Join(compareOutputDir, "*.js.output")
	jsFiles, _ := filepath.Glob(jsPattern)
	for _, jsFile := range jsFiles {
		baseName := strings.TrimSuffix(filepath.Base(jsFile), ".js.output")
		goFile := filepath.Join(compareOutputDir, baseName+".go.output")
		if _, err := os.Stat(goFile); os.IsNotExist(err) {
			missingGo = append(missingGo, baseName)
			t.Errorf("MISSING GO OUTPUT: %s.go.output (JS output exists)", baseName)
		}
	}

	// Summary
	t.Logf("\n=== COMPARISON SUMMARY ===")
	t.Logf("Total .go.output files: %d", len(goFiles))
	t.Logf("Matches: %d", len(goFiles)-len(mismatches)-len(liveRecordings)-len(missingJS))
	t.Logf("Live recordings (skipped): %d", len(liveRecordings))
	t.Logf("Mismatches: %d", len(mismatches))
	t.Logf("Missing .js.output: %d", len(missingJS))
	t.Logf("Missing .go.output: %d", len(missingGo))

	if len(mismatches) > 0 || len(missingJS) > 0 || len(missingGo) > 0 {
		t.Fail()
	}
}

// parseOutputFile parses the header from output file content.
// Format: "{input_length} bytes\n{body}"
// Returns input length and body content.
func parseOutputFile(content []byte) (inputLen int, body []byte) {
	// Find first newline
	idx := bytes.IndexByte(content, '\n')
	if idx == -1 {
		// No header found, treat whole content as body with unknown input length
		return -1, content
	}

	header := string(content[:idx])
	body = content[idx+1:]

	// Parse "{length} bytes"
	var length int
	_, err := fmt.Sscanf(header, "%d bytes", &length)
	if err != nil {
		return -1, content
	}

	return length, body
}

// reportMismatch outputs detailed debug information about a mismatch
func reportMismatch(t *testing.T, baseName, goFile, jsFile string, goContent, jsContent []byte) {
	t.Helper()

	t.Errorf("\n=== MISMATCH: %s ===", baseName)
	t.Errorf("Go output: %d bytes (%s)", len(goContent), goFile)
	t.Errorf("JS output: %d bytes (%s)", len(jsContent), jsFile)

	// Find first differing byte
	minLen := len(goContent)
	if len(jsContent) < minLen {
		minLen = len(jsContent)
	}

	diffPos := -1
	for i := 0; i < minLen; i++ {
		if goContent[i] != jsContent[i] {
			diffPos = i
			break
		}
	}

	if diffPos == -1 && len(goContent) != len(jsContent) {
		// Files are same up to minLen, but different lengths
		diffPos = minLen
		t.Errorf("First difference at byte %d (length mismatch)", diffPos)
	} else if diffPos >= 0 {
		t.Errorf("First difference at byte %d", diffPos)

		// Show hex dump around the difference
		showHexDump(t, "Go", goContent, diffPos)
		showHexDump(t, "JS", jsContent, diffPos)
	}

	// Run diff -u for text comparison
	runDiff(t, goFile, jsFile)
}

// showHexDump shows a hex dump around the specified position
func showHexDump(t *testing.T, label string, content []byte, pos int) {
	t.Helper()

	// Show 16 bytes before and 16 bytes after
	start := pos - 16
	if start < 0 {
		start = 0
	}
	end := pos + 16
	if end > len(content) {
		end = len(content)
	}

	if start >= len(content) {
		t.Errorf("  %s: (position %d is beyond content length %d)", label, pos, len(content))
		return
	}

	slice := content[start:end]

	// Build hex string
	var hexParts []string
	for _, b := range slice {
		hexParts = append(hexParts, fmt.Sprintf("0x%02x", b))
	}

	// Build printable string
	var printable strings.Builder
	for _, b := range slice {
		if b >= 32 && b < 127 {
			printable.WriteByte(b)
		} else if b == '\n' {
			printable.WriteString("\\n")
		} else if b == '\r' {
			printable.WriteString("\\r")
		} else if b == '\t' {
			printable.WriteString("\\t")
		} else {
			printable.WriteString(".")
		}
	}

	t.Errorf("  %s: [%s]", label, strings.Join(hexParts, " "))
	t.Errorf("  %s: %q", label, printable.String())
}

// runDiff runs diff -u on the two files and reports output
func runDiff(t *testing.T, goFile, jsFile string) {
	t.Helper()

	cmd := exec.Command("diff", "-u", goFile, jsFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// diff returns exit code 1 when files differ, which is expected
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// Show first 100 lines of diff output
			lines := strings.Split(string(output), "\n")
			if len(lines) > 100 {
				lines = lines[:100]
				lines = append(lines, fmt.Sprintf("... (%d more lines)", len(strings.Split(string(output), "\n"))-100))
			}
			t.Errorf("\n--- diff -u output (first 100 lines) ---\n%s", strings.Join(lines, "\n"))
		} else {
			t.Errorf("diff command failed: %v", err)
		}
	}
}
