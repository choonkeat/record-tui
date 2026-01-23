package session

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

const (
	// recordingsDir is the directory containing session*.log files to process
	recordingsDir = "../../recordings"
	// outputDir is the directory where cleaned output files are written
	outputDir = "../../recordings-output"
)

// TestGenerateGoOutput reads session*.log files from recordingsDir,
// runs them through the cleaning pipeline (StripMetadata which includes
// NeutralizeClearSequences), and writes output to outputDir.
// This test generates .go.output files for comparison with JS output.
func TestGenerateGoOutput(t *testing.T) {
	// Check if recordings directory exists
	if _, err := os.Stat(recordingsDir); os.IsNotExist(err) {
		t.Skipf("WARNING: recordings directory %q does not exist, skipping", recordingsDir)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output directory %q: %v", outputDir, err)
	}

	// Find all session*.log files
	pattern := filepath.Join(recordingsDir, "session*.log")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("failed to glob pattern %q: %v", pattern, err)
	}

	if len(matches) == 0 {
		t.Skipf("WARNING: no session*.log files found in %q, skipping", recordingsDir)
	}

	// Sort for deterministic ordering
	sort.Strings(matches)

	t.Logf("Found %d session*.log files to process", len(matches))

	for _, match := range matches {
		// Follow symlinks
		realPath, err := filepath.EvalSymlinks(match)
		if err != nil {
			t.Errorf("failed to resolve symlink %q: %v", match, err)
			continue
		}

		// Read file bytes
		content, err := os.ReadFile(realPath)
		if err != nil {
			t.Errorf("failed to read file %q: %v", realPath, err)
			continue
		}

		// Convert to string and run cleaning pipeline
		// This is steps 1-4: read bytes -> string -> strip metadata -> neutralize clears
		cleanedContent := StripMetadata(string(content))

		// Write output file
		basename := filepath.Base(match)
		outputPath := filepath.Join(outputDir, basename+".go.output")
		if err := os.WriteFile(outputPath, []byte(cleanedContent), 0644); err != nil {
			t.Errorf("failed to write output file %q: %v", outputPath, err)
			continue
		}

		t.Logf("Processed %s -> %s (%d bytes -> %d bytes)",
			basename, filepath.Base(outputPath), len(content), len(cleanedContent))
	}
}
