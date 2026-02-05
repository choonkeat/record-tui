package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/choonkeat/record-tui/internal/session"
	"github.com/choonkeat/record-tui/playback"
)

func main() {
	logPath := os.Args[1]

	sessionContent, err := os.ReadFile(logPath)
	if err != nil {
		panic(err)
	}

	rawOutput := session.StripMetadataOnly(string(sessionContent))
	processedOutput := session.StripMetadata(string(sessionContent))
	neutralized, _ := session.NeutralizeAllWithOffsets(rawOutput)

	fmt.Printf("Raw newlines: %d\n", strings.Count(rawOutput, "\n"))
	fmt.Printf("Processed newlines: %d\n", strings.Count(processedOutput, "\n"))
	fmt.Printf("Neutralized newlines: %d\n", strings.Count(neutralized, "\n"))
	fmt.Printf("Processed == Neutralized: %v\n", processedOutput == neutralized)
	fmt.Printf("len(processed)=%d len(neutralized)=%d\n", len(processedOutput), len(neutralized))

	timingPath := strings.TrimSuffix(logPath, ".log") + ".timing"
	inputPath := strings.TrimSuffix(logPath, ".log") + ".input"

	timingFile, err := os.Open(timingPath)
	if err != nil {
		fmt.Printf("No timing file: %v\n", err)
		return
	}
	defer timingFile.Close()

	inputBytes, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Printf("No input file: %v\n", err)
		return
	}

	entries := playback.BuildTOC(timingFile, inputBytes, bytes.NewReader(sessionContent))
	fmt.Printf("\nTOC entries (%d):\n", len(entries))
	for i, e := range entries {
		label := e.Label
		if len(label) > 50 {
			label = label[:50]
		}
		fmt.Printf("  [%d] line=%d label=%q\n", i, e.Line, label)
	}
}
