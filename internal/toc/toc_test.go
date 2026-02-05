package toc

import (
	"strings"
	"testing"

	"github.com/choonkeat/record-tui/internal/timing"
)

func TestFromCommands_Basic(t *testing.T) {
	rawOutput := []byte("$ \nls output\nfile1\nfile2\n$ \nnpm test output\nPASS\n")
	//                    0123 456789012 34567 890123 45678 901234567890123 456789 0

	commands := []timing.Command{
		{Text: "ls", OutputByteOffset: 2},       // at "$ " (line 0)
		{Text: "npm test", OutputByteOffset: 26}, // at second "$ " (line 4)
	}

	entries := FromCommands(commands, rawOutput)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Label != "ls" || entries[0].Line != 0 {
		t.Errorf("entry 0: got {%q, %d}, want {\"ls\", 0}", entries[0].Label, entries[0].Line)
	}
	if entries[1].Label != "npm test" || entries[1].Line != 4 {
		t.Errorf("entry 1: got {%q, %d}, want {\"npm test\", 4}", entries[1].Label, entries[1].Line)
	}
}

func TestFromCommands_Empty(t *testing.T) {
	entries := FromCommands(nil, []byte("some output"))
	if entries != nil {
		t.Errorf("expected nil, got %v", entries)
	}
}

func TestFromCommands_OffsetBeyondContent(t *testing.T) {
	rawOutput := []byte("short\n")
	commands := []timing.Command{
		{Text: "cmd", OutputByteOffset: 1000},
	}

	entries := FromCommands(commands, rawOutput)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// Should count all newlines in the content
	if entries[0].Line != 1 {
		t.Errorf("got line %d, want 1", entries[0].Line)
	}
}

func TestFromCommands_ZeroOffset(t *testing.T) {
	rawOutput := []byte("$ ls\noutput\n")
	commands := []timing.Command{
		{Text: "ls", OutputByteOffset: 0},
	}

	entries := FromCommands(commands, rawOutput)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Line != 0 {
		t.Errorf("got line %d, want 0", entries[0].Line)
	}
}

func TestFromCommandsReader_Basic(t *testing.T) {
	content := "$ \nls output\nfile1\nfile2\n$ \nnpm test output\nPASS\n"

	commands := []timing.Command{
		{Text: "ls", OutputByteOffset: 2},
		{Text: "npm test", OutputByteOffset: 26},
	}

	entries := FromCommandsReader(commands, strings.NewReader(content))
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Label != "ls" || entries[0].Line != 0 {
		t.Errorf("entry 0: got {%q, %d}, want {\"ls\", 0}", entries[0].Label, entries[0].Line)
	}
	if entries[1].Label != "npm test" || entries[1].Line != 4 {
		t.Errorf("entry 1: got {%q, %d}, want {\"npm test\", 4}", entries[1].Label, entries[1].Line)
	}
}

func TestFromCommandsReader_SkipsScriptHeader(t *testing.T) {
	content := "Script started on 2026-01-12 06:41:43+00:00\n$ \nls output\n"

	commands := []timing.Command{
		// Offset 45 = right after header line (44 bytes + newline)
		{Text: "ls", OutputByteOffset: 45},
	}

	entries := FromCommandsReader(commands, strings.NewReader(content))
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// Line 0 because header is skipped
	if entries[0].Line != 0 {
		t.Errorf("got line %d, want 0", entries[0].Line)
	}
}

func TestFromCommandsReader_Empty(t *testing.T) {
	entries := FromCommandsReader(nil, strings.NewReader("some output"))
	if entries != nil {
		t.Errorf("expected nil, got %v", entries)
	}
}

func TestFromCommandsReader_LargeContent(t *testing.T) {
	// Simulate a large recording with many lines
	var b strings.Builder
	b.WriteString("Script started on 2026-01-12\n")
	for i := 0; i < 10000; i++ {
		b.WriteString("output line\n")
	}
	// Command at ~halfway point (header=29 bytes, each line=12 bytes)
	// Line 5000 starts at byte 29 + 5000*12 = 60029
	commands := []timing.Command{
		{Text: "halfway", OutputByteOffset: 60029},
	}

	entries := FromCommandsReader(commands, strings.NewReader(b.String()))
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// Should be around line 5000 (header skipped)
	if entries[0].Line < 4990 || entries[0].Line > 5010 {
		t.Errorf("got line %d, want ~5000", entries[0].Line)
	}
}
