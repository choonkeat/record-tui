# record-tui

Record terminal sessions and convert them to beautiful standalone HTML with xterm.js rendering.

**See [example output](https://record-tui.netlify.app)** - recorded `claude` session with ANSI colors and terminal emulation.

## Requirements

**macOS only** — requires `script` and `open` commands (both built into macOS).

## Installation

```bash
make install  # Compile and install to ~/bin/record-tui
```

Optional: PDF export support

```bash
make install-pdf-tool  # Enable automatic PDF generation
```

## Usage

Record a terminal session:

```bash
# Interactive shell
record-tui

# Record a specific command
record-tui claude
record-tui npm test
record-tui sh -c "ls -la"
```

Files are saved to `~/.record-tui/YYYYMMDD-HHMMSS/`:
- `session.log` — raw session file
- `session.log.html` — standalone HTML with ANSI colors and terminal emulation
- `session.log.pdf` — printable PDF (A4 landscape, requires `make install-pdf-tool`)

Recording stops when:
- Command exits (if you specified one)
- You press **Ctrl-D** or type `exit` (if running interactive shell)

The directory opens automatically in Finder when recording completes (unless over SSH).

## Features

- ✅ **One command**: Records and converts automatically
- ✅ **Standalone HTML**: No external dependencies, works offline after generation
- ✅ **PDF export**: Generate printable PDFs via `make install-pdf-tool` (optional)
- ✅ **Colors preserved**: Full ANSI color support (8 colors + bright variants)
- ✅ **Auto-open**: Directory opens in Finder on completion
- ✅ **Single binary**: No Node.js, npm, or external dependencies (except optional PDF tool)
- ✅ **Instant**: Fast recording and HTML generation (~2s more for PDF)

## What Gets Recorded

Everything typed in the terminal session:
- ✅ Command output with colors
- ✅ Interactive commands (runs fully)
- ✅ Text and code with formatting
- ⚠️ Full-screen TUIs (vim, htop, etc.) display as control sequences

## Examples

```bash
# View all recordings
ls ~/.record-tui/

# Open a specific session
open ~/.record-tui/20251231-144256/session.log.html
```

## Library Usage

record-tui can be imported as a Go library for generating HTML from terminal recordings:

```bash
go get github.com/choonkeat/record-tui@latest
```

```go
package main

import (
    "os"

    "github.com/choonkeat/record-tui/playback"
)

func main() {
    // Read session.log content (from macOS or Linux `script` command)
    content, _ := os.ReadFile("session.log")

    // Strip script header/footer metadata
    cleaned := playback.StripMetadata(string(content))

    // Create playback frames (single frame for static display)
    frames := []playback.Frame{
        {Timestamp: 0, Content: cleaned},
    }

    // Generate standalone HTML with xterm.js
    html, _ := playback.RenderHTML(frames, playback.Options{
        Title: "My Session", // Optional: custom page title
    })
    os.WriteFile("output.html", []byte(html), 0644)
}
```

Supports both macOS and Linux `script` command output formats.

## Development

### Build

```bash
make build   # Compile binary
make test    # Run tests
make clean   # Remove build artifacts
```

### Project Structure

```
cmd/record-tui/    # CLI entry point
playback/          # Public API (import this package)
internal/
├── record/        # Recording and conversion logic
├── html/          # HTML generation with xterm.js
└── session/       # Session log parsing and cleanup
```

## License

MIT
