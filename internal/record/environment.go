package record

import (
	"os"
)

// SetupRecordingEnvironment configures environment variables for proper terminal recording.
// This ensures the recorded session includes color information for better playback.
func SetupRecordingEnvironment() {
	// Enable color output in applications that support it
	os.Setenv("FORCE_COLOR", "1")

	// Enable true color (24-bit) support
	os.Setenv("COLORTERM", "truecolor")
}

// ClearRecordingEnvironment removes recording-specific environment variables.
// Call this if you need to restore original environment after recording.
func ClearRecordingEnvironment() {
	os.Unsetenv("FORCE_COLOR")
	os.Unsetenv("COLORTERM")
}
