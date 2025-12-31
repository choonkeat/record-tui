# to-pdf

Convert record-tui HTML output to PDF using Playwright.

## Usage

```bash
node index.js <html-file> [output-pdf] [format] [scale]
```

## Examples

```bash
# Default: A4 landscape, 1.0 scale
node index.js session.log.html session.log.pdf

# Custom format
node index.js session.log.html session.log.pdf A3-landscape

# Custom scale (for wide terminals)
node index.js session.log.html session.log.pdf A4-landscape 0.8
```

## Format Options

- `A4-landscape` (default) - Best for typical terminal recordings
- `A4` - Portrait orientation
- `A3-landscape` - Wider format for very wide terminals
- `A3`, `A2`, `Letter`, `Tabloid` - Additional standard formats

## Scale Options

- `1.0` (default) - No scaling
- `0.8` - 80% scale (fits very wide terminals)
- `0.6` - 60% scale (for extremely wide terminals)

## Performance

- Conversion time: ~2 seconds per file
- Output size: Typically 0.01-0.5MB (highly compressed)
- Requires: Node.js + Playwright browser cache (~250MB downloaded on first use)

## Notes

- xterm.js rendering is captured exactly as displayed in the browser
- Colors, fonts, and formatting are preserved
- PDF is A4 landscape by default to accommodate typical terminal width
