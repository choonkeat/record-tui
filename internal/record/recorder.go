package record

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// RecordSession executes the `script` command to record a terminal session.
// The `script` command reads from stdin and writes terminal output to a file.
//
// Args:
//   - outputPath: Path to the session.log file to create
//   - args: Command and arguments to execute within the session
//           If empty, script will use the default shell
//
// Returns error if script command fails or cannot be executed
func RecordSession(outputPath string, args []string) error {
	// Build command: script <outputPath> [additional args]
	cmdArgs := []string{outputPath}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command("script", cmdArgs...)

	// Inherit stdin/stdout/stderr so user can interact with the recorded session
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute script command
	err := cmd.Run()
	if err != nil {
		// script returns exit code 0 normally, so any error is a real problem
		return fmt.Errorf("script command failed: %w", err)
	}

	return nil
}

// RecordSessionDetailed is like RecordSession but returns more info about execution
// Returns: exit code, error
func RecordSessionDetailed(outputPath string, args []string) (int, error) {
	cmdArgs := []string{outputPath}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command("script", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), err
			}
		}
		return 1, err
	}

	return 0, nil
}
