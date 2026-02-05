package toc

import (
	"strings"
	"testing"

	"github.com/choonkeat/record-tui/internal/timing"
)

func TestFromCommands_Basic(t *testing.T) {
	content := "$ \nls output\nfile1\nfile2\n$ \nnpm test output\nPASS\n"

	commands := []timing.Command{
		{Text: "ls", OutputByteOffset: 2},
		{Text: "npm test", OutputByteOffset: 26},
	}

	entries := FromCommands(commands, strings.NewReader(content))
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
	entries := FromCommands(nil, strings.NewReader("some output"))
	if entries != nil {
		t.Errorf("expected nil, got %v", entries)
	}
}

func TestFromCommands_OffsetBeyondContent(t *testing.T) {
	commands := []timing.Command{
		{Text: "cmd", OutputByteOffset: 1000},
	}

	entries := FromCommands(commands, strings.NewReader("short\n"))
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// Should count all newlines in the content
	if entries[0].Line != 1 {
		t.Errorf("got line %d, want 1", entries[0].Line)
	}
}

func TestFromCommands_ZeroOffset(t *testing.T) {
	commands := []timing.Command{
		{Text: "ls", OutputByteOffset: 0},
	}

	entries := FromCommands(commands, strings.NewReader("$ ls\noutput\n"))
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Line != 0 {
		t.Errorf("got line %d, want 0", entries[0].Line)
	}
}

func TestFromCommands_SkipsScriptHeader(t *testing.T) {
	content := "Script started on 2026-01-12 06:41:43+00:00\n$ \nls output\n"

	commands := []timing.Command{
		{Text: "ls", OutputByteOffset: 45},
	}

	entries := FromCommands(commands, strings.NewReader(content))
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Line != 0 {
		t.Errorf("got line %d, want 0", entries[0].Line)
	}
}

func TestFromCommands_LargeContent(t *testing.T) {
	var b strings.Builder
	b.WriteString("Script started on 2026-01-12\n")
	for i := 0; i < 10000; i++ {
		b.WriteString("output line\n")
	}
	commands := []timing.Command{
		{Text: "halfway", OutputByteOffset: 60029},
	}

	entries := FromCommands(commands, strings.NewReader(b.String()))
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Line < 4990 || entries[0].Line > 5010 {
		t.Errorf("got line %d, want ~5000", entries[0].Line)
	}
}
