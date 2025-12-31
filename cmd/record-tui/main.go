package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/choonkeat/record-tui/internal/record"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: record-tui [command ...]

Start a terminal recording session and automatically convert to HTML.

Arguments:
  [command ...]  Command to execute in the recorded session (optional)
                 If omitted, starts an interactive shell

Examples:
  record-tui                  # Start interactive shell recording
  record-tui echo hello       # Record specific command
  record-tui /bin/bash        # Record bash session
`)
}

// getRecordingDir creates and returns the recording directory path
// Format: ~/.record-tui/YYYYMMDD-HHMMSS/
func getRecordingDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".record-tui")
	timestamp := time.Now().Format("20060102-150405")
	recordingDir := filepath.Join(baseDir, timestamp)

	// Create directory with permissions 0755
	err = os.MkdirAll(recordingDir, 0755)
	if err != nil {
		return "", fmt.Errorf("cannot create recording directory %s: %w", recordingDir, err)
	}

	return recordingDir, nil
}

// isInteractiveTerminal checks if stdout is connected to a TTY
func isInteractiveTerminal() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	// Check if stdout is a character device (terminal)
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// isSSHSession checks if running over SSH
func isSSHSession() bool {
	return os.Getenv("SSH_CLIENT") != ""
}

// openRecordingDir opens the recording directory in the file explorer
// Only opens if running in an interactive terminal and not over SSH
func openRecordingDir(dir string) {
	// Skip if not interactive
	if !isInteractiveTerminal() {
		return
	}
	// Skip if over SSH
	if isSSHSession() {
		return
	}
	// Open directory (silently ignore errors)
	exec.Command("open", dir).Run()
}

func main() {
	flag.Parse()
	args := flag.Args()

	// Setup environment for color recording
	record.SetupRecordingEnvironment()

	// Create recording directory
	recordingDir, err := getRecordingDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create recording directory: %v\n", err)
		os.Exit(1)
	}

	sessionLogPath := filepath.Join(recordingDir, "session.log")

	// Record the session
	fmt.Fprintf(os.Stderr, "Recording started. Press Ctrl-D to exit.\n")
	err = record.RecordSession(sessionLogPath, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Recording failed: %v\n", err)
		os.Exit(1)
	}

	// Verify session.log was created
	if _, err := os.Stat(sessionLogPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: session.log was not created\n")
		os.Exit(1)
	}

	// Convert session.log to HTML
	htmlPath, err := record.ConvertSessionToHTML(sessionLogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: HTML conversion failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Note: session.log was recorded successfully\n")
		// Don't exit - recording was successful even if conversion failed
	} else {
		fmt.Fprintf(os.Stderr, "✓ HTML generated: %s\n", htmlPath)
	}

	// Success message
	fmt.Fprintf(os.Stderr, "✓ Recording saved to: %s/\n", recordingDir)

	// Open directory in file explorer (interactive terminals only, skip SSH)
	openRecordingDir(recordingDir)

	os.Exit(0)
}
