# Plan: TOC Navigation from Timing File

## Overview

Add a table-of-contents feature to the HTML viewer. When a timing file (advanced format with I/O type markers) is available alongside `session.log`, record-tui parses it to identify user commands and embeds a navigable TOC in the HTML output.

## Architecture

```
timing file (I/O entries)
        ↓
  internal/timing     → parse entries, group I-entries into commands
        ↓
  internal/toc        → compute line numbers in cleaned output
        ↓
  playback.TOCEntry   → {Label, Line} passed to HTML renderer
        ↓
  HTML template        → floating TOC panel with click-to-scroll
```

## Step 1: `internal/timing` package — Parse advanced timing format

New package with:

- `Entry` struct: `{Type byte, Delay float64, ByteCount int}`
- `Parse(content string) ([]Entry, error)` — parse lines like `I 0.500000 1` and `O 0.009404 16`
- `Command` struct: `{Text string, OutputByteOffset int}` — a grouped user command
- `ExtractCommands(entries []Entry, inputContent []byte) []Command`
  - Walk entries in order, tracking cumulative O byte offset
  - Group consecutive I entries; use byte counts to extract text from `inputContent` (the `session.input` file)
  - Filter: skip single control chars (arrows, Ctrl-C, tab), keep commands terminated by `\r`/`\n`

**Test**: Parse sample timing content, verify entries. Extract commands from timing + input file, verify labels and byte offsets.

## Step 2: `internal/toc` package — Compute TOC entries with line numbers

New package with:

- `Entry` struct: `{Label string, Line int}`
- `FromCommands(commands []timing.Command, rawOutput []byte) []Entry`
  - `rawOutput` = session.log with header/footer already stripped (just the output bytes)
  - For each command, count newlines in `rawOutput[0:command.OutputByteOffset]` → line number
  - These line numbers are approximate (pre-clean-sequence-transformation) but good enough for scroll navigation — clear sequences are structural boundaries anyway

**Test**: Given known output content and command byte offsets, verify line numbers.

## Step 3: Extend `playback` public API

- Add `TOCEntry` type to `playback/types.go`: `{Label string, Line int}`
- Add `TOC []TOCEntry` field to both `Options` and `StreamingOptions`
- Pass through to `internal/html` types and template functions

## Step 4: Extend HTML templates — floating TOC panel

Both `template.go` and `template_streaming.go`:

- Accept TOC entries (JSON-encoded, base64 or escaped)
- Render a collapsible floating panel (top-right):
  - Button to toggle open/closed (e.g., "TOC" or "≡" icon)
  - List of command labels, each clickable
  - Click scrolls the page to `lineNumber * lineHeight + terminalOffset`
  - Highlight current section based on scroll position
- CSS: dark theme matching terminal, semi-transparent background, z-index above terminal
- Only render TOC panel when entries exist (backward compatible — no TOC = no panel)

## Step 5: Update `internal/record/converter.go`

- In `ConvertSessionToHTML` and `ConvertSessionToStreamingHTML`:
  - Check if timing file exists alongside session.log (e.g., `session.timing`)
  - Check if input file exists (e.g., `session.input`)
  - If both exist, parse timing + input → extract commands → compute TOC
  - Pass TOC entries to `playback.RenderHTML` / `playback.RenderStreamingHTML` via options

## Files to create

1. `internal/timing/timing.go` — parsing + command extraction
2. `internal/timing/timing_test.go` — tests
3. `internal/toc/toc.go` — line number computation
4. `internal/toc/toc_test.go` — tests

## Files to modify

1. `playback/types.go` — add TOCEntry, TOC field to Options/StreamingOptions
2. `playback/playback.go` — pass TOC through to internal/html
3. `internal/html/types.go` — add TOCEntry, TOC field to internal types
4. `internal/html/template.go` — render TOC panel in embedded mode
5. `internal/html/template_streaming.go` — render TOC panel in streaming mode
6. `internal/record/converter.go` — detect timing/input files, generate TOC

## Out of scope (for now)

- Exact line-number mapping through clear/alt-screen transformations (approximate is fine for v1)
- macOS BSD `script` input capture (Linux only for now)
- Password scrubbing from input file
- Streaming mode: fetching timing file from browser (TOC is pre-computed server-side)
