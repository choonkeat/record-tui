# record-tui

Record terminal sessions and convert them to beautiful standalone HTML with xterm.js rendering.

**See [example output](https://record-tui.netlify.app)** - recorded `claude` session with ANSI colors and terminal emulation.

## Features

- **Records & converts**: Terminal sessions recorded with `script` command, automatically converted to HTML
- **Single Go binary**: No Node.js, no npm, no external dependencies
- **Statically linked**: Works anywhere, includes xterm.js via CDN
- **Color support**: Preserves ANSI color codes in recordings

## Installation

```bash
make install  # Compile and install to ~/bin/record-tui
```

### Cross-Platform Build

```bash
make build-all  # Build for darwin-arm64, darwin-amd64, linux-amd64
```

## Usage

Record a terminal session and automatically generate HTML:

```bash
# Record interactive session
record-tui

# Run specific command and record
record-tui claude

# After recording, files are in ~/.record-tui/YYYYMMDD-HHMMSS/
# - session.log (raw session file)
# - session.log.html (static HTML with xterm.js rendering)
```

**Features:**
- Creates timestamped directory: `~/.record-tui/YYYYMMDD-HHMMSS/`
- Automatically enables color support (FORCE_COLOR=1, COLORTERM=truecolor)
- Converts session.log to HTML immediately after recording
- Press Ctrl-D to exit recording

## Recording with script

If you want to record manually:

```bash
# Simple recording
script session.log
# ... run commands ...
# Ctrl-D to exit

# Record with timing (for future playback features)
script -t 2>timing.log session.log
# ... run commands ...
# Ctrl-D to exit
```

## Development

### Build

```bash
make build      # Compile binaries
make build-all  # Cross-platform compilation
make clean      # Remove bin/
```

### Test

```bash
make test  # Run all tests (27 tests, all passing)
```

### Project Structure

```
cmd/
└── record-tui/     # CLI: records and converts in one step

internal/
├── record/         # Recording and conversion logic
│   ├── recorder.go     # Execute script command
│   ├── converter.go    # Convert session.log to HTML
│   └── environment.go  # Setup color environment vars
├── html/           # HTML generation with xterm.js
└── session/        # Session.log parsing and cleanup
```

### Code Organization

- **cmd/record-tui/main.go**: CLI entry point
- **internal/record/recorder.go**: Wraps `script` command execution
- **internal/record/converter.go**: Converts session.log to HTML
- **internal/record/environment.go**: Sets FORCE_COLOR=1, COLORTERM=truecolor
- **internal/html/**: HTML template with embedded xterm.js
- **internal/session/**: Strips script command metadata from session logs

### Testing Strategy

```bash
# Run all tests with verbose output
go test ./internal/... -v

# Test categories:
# - Recorder tests (5): Session recording with various arguments
# - Converter tests (6): HTML generation and file handling
# - Environment tests (4): Color support variable setup
# - HTML template tests (5): Base64 encoding, ANSI codes
# - Session cleaner tests (7): Metadata removal
```

### Binary Info

```bash
make info  # Show binary size
```

Output:
```
record-tui: 2.9M (single binary, includes everything)
```

## Architecture

### record-tui Command Flow

```
record-tui (CLI)
├─ SetupRecordingEnvironment()      # FORCE_COLOR=1, COLORTERM=truecolor
├─ Create directory ~/.record-tui/YYYYMMDD-HHMMSS/
├─ RecordSession()                  # Execute `script` command
│  └─ os/exec.Command("script", ...) # Direct system call
├─ ConvertSessionToHTML()           # Convert log to HTML
│  ├─ Read session.log
│  ├─ session.StripMetadata()       # Remove script wrapper
│  ├─ html.RenderPlaybackHTML()     # Render final output as HTML with xterm.js
│  └─ Write session.log.html
└─ Success message with file path
```

### Internal Package Reuse

Both binaries use the same internal packages:
- **session**: Parse and clean session.log files
- **html**: Generate standalone HTML with xterm.js

No subprocess overhead—direct Go function calls.

## Features

### Supported Input

- Terminal sessions from `script` command (POSIX standard)
- ANSI escape codes (colors, bold, italic, underline)
- Mixed line terminators (CRLF, CR, LF)
- Long lines (tested with 800+ character lines)

### Output

- Completely standalone HTML file
- Static rendering of final session output
- xterm.js for terminal emulation (color, formatting, ANSI codes)
- Works in any modern web browser
- Responsive design
- Dark theme by default

### ANSI Color Support

Full support for standard ANSI colors:
- 8 base colors (black, red, green, yellow, blue, magenta, cyan, white)
- Bright variants of each
- Foreground and background colors
- Bold, italic, underline formatting

## Examples

### Quick Start

```bash
# Record and convert in one step
record-tui

# Or record a specific command
record-tui sh -c "ls -la"

# Navigate to the generated HTML
open ~/.record-tui/*/session.log.html
```

### Multiple Sessions

```bash
# View all your recordings
ls ~/.record-tui/

# Open a specific session
open ~/.record-tui/20251231-144256/session.log.html
```

## Limitations

- Interactive programs (vim, less, etc.) display as control sequences
- Very large sessions may produce large HTML files (typical: <5MB)
- Terminal emulation via JavaScript has browser performance limits

## Requirements

- Go 1.21+ (for building from source)
- POSIX-compliant shell (bash, zsh, sh, etc.)
- No Node.js required for runtime

## Technology Stack

- **Language**: Go 1.21+
- **Terminal Recording**: POSIX `script` command
- **Terminal Emulation**: xterm.js (CDN-hosted in HTML)
- **Session Processing**: Custom Go parsers
- **Testing**: Go `testing` package

## Building for Different Platforms

```bash
# Build for current platform
make build

# Build for all supported platforms (darwin-arm64, darwin-amd64, linux-amd64)
make build-all

# Cross-compile for specific platform
GOOS=linux GOARCH=amd64 go build -o bin/record-tui-linux-amd64 ./cmd/record-tui
```

## Contributing

Contributions welcome! Please:
1. Write tests for new features
2. Ensure all tests pass: `make test`
3. Commit only relevant files (avoid committing `bin/`)
4. Follow Go conventions

## License

MIT
