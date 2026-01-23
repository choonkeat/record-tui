# Streaming HTML for Large Terminal Recordings

## Problem

Large terminal recordings (multi-megabyte) cause the browser to hang while parsing the embedded base64 data. The current approach embeds all frame data as base64 in the HTML, which must be fully parsed before any rendering begins.

## Solution

Add a streaming mode that generates a lightweight HTML shell + fetches session data separately. This allows:
- Progressive rendering as bytes arrive
- Faster time-to-first-render
- Same final output as embedded mode

## Scope

- **record-tui CLI**: Continues using embedded HTML (works from Finder/file://)
- **swe-swe**: Uses new streaming HTML (served via Go web server)

## Test Data

- `sample.log` (12MB) — provided for testing, DO NOT commit to git

---

## Phase 1: Streaming JavaScript Implementation

### Goal
Create browser-side streaming logic with header/footer stripping and clear sequence neutralization.

### Steps

1. [x] **Create test harness** (`public/streaming-test.html`)
   - Minimal HTML that loads xterm.js from CDN
   - Contains streaming implementation for manual browser testing
   - Fetches `./sample.log` (or configurable URL)

2. [x] **Implement `stripHeader(text)` function**
   - Match Go's logic: remove first ~5 lines starting with "Script started on" or "Command:"
   - Return cleaned text

3. [x] **Implement `stripFooter(text)` function**
   - Match Go's logic: remove trailing lines containing "Saving session", "Script done on", "Command exit status"
   - Return cleaned text

4. [x] **Implement `neutralizeClearSequences(text)` function**
   - Regex: `/\x1b\[H\x1b\[[23]J|\x1b\[[23]J\x1b\[H|\x1b\[[23]J/g`
   - Replace with: `\n\n──────── terminal cleared ────────\n\n`
   - Return cleaned text

5. [x] **Implement `streamSession(url, xterm)` function**
   ```javascript
   async function streamSession(url, xterm) {
     const response = await fetch(url);
     const reader = response.body.getReader();
     const decoder = new TextDecoder();

     let buffer = '';
     let headerStripped = false;
     const TRAILING_SIZE = 500;

     while (true) {
       const {done, value} = await reader.read();
       if (done) break;

       buffer += decoder.decode(value, {stream: true});

       if (!headerStripped) {
         buffer = stripHeader(buffer);
         headerStripped = true;
       }

       if (buffer.length > TRAILING_SIZE) {
         let toWrite = buffer.slice(0, -TRAILING_SIZE);
         buffer = buffer.slice(-TRAILING_SIZE);

         // Don't split escape sequences
         const lastEsc = toWrite.lastIndexOf('\x1b');
         if (lastEsc > toWrite.length - 10) {
           buffer = toWrite.slice(lastEsc) + buffer;
           toWrite = toWrite.slice(0, lastEsc);
         }

         xterm.write(neutralizeClearSequences(toWrite));
       }
     }

     // Final: strip footer, neutralize, write
     xterm.write(neutralizeClearSequences(stripFooter(buffer)));
   }
   ```

6. [x] **Add loading indicator**
   - Show "Loading..." in terminal div before streaming starts
   - Clear it on first xterm.write()

7. [x] **Handle terminal dimensions**
   - Initialize with reasonable defaults (cols: 120, rows: 50)
   - Resize after streaming completes based on actual content

### Verification

- [ ] Manual browser test: serve `sample.log` on port 3000, open test harness
- [ ] Content renders progressively (not all at once)
- [ ] Header lines don't appear in output
- [ ] Footer lines don't appear in output
- [ ] Clear sequences show separator (not actual clear)
- [ ] No JavaScript console errors

---

## Phase 2: Go API for Streaming HTML Generation

### Goal
Add `RenderStreamingHTML()` function to the `playback` package.

### Steps

1. **Add `StreamingOptions` to `playback/types.go`**
   ```go
   type StreamingOptions struct {
       Title      string     // Page title (defaults to "Terminal")
       DataURL    string     // URL to fetch session data from
       FooterLink FooterLink // Optional co-branding link
   }
   ```

2. **Create `internal/html/template_streaming.go`**
   - HTML shell matching current styling
   - Embedded streaming JavaScript from Phase 1
   - Template substitution for title, dataURL, footer
   - Loading indicator div

3. **Add `RenderStreamingHTML()` to `playback/playback.go`**
   ```go
   func RenderStreamingHTML(opts StreamingOptions) (string, error) {
       // Validate DataURL is provided
       // Call internal template renderer
       // Return HTML string
   }
   ```

4. **Add unit tests in `playback/playback_test.go`**
   - Test with various options
   - Test default values
   - Test special character escaping in title/URL

### Verification

- [ ] `go test ./playback/...` passes
- [ ] `go test ./internal/...` passes
- [ ] Existing `RenderHTML()` tests still pass (no regression)
- [ ] Generated HTML contains correct title, dataURL, JavaScript

---

## Phase 3: Integration Testing

### Goal
End-to-end verification with actual large session file served from Go.

### Steps

1. **Create test server** (`internal/html/testserver_test.go`)
   - Serves streaming HTML at `/`
   - Serves session.log at `/session.log`
   - Runs on port 3000

2. **Manual browser verification**
   - Start test server: `go test -run TestStreamingServer -v ./internal/html/...`
   - Open http://localhost:3000 in browser
   - Verify progressive rendering
   - Verify correct content (no header/footer, clear separators)

3. **Performance documentation**
   - Note time-to-first-render improvement
   - Note total load time comparison

4. **Edge case testing**
   - Empty session file
   - Session with only header/footer
   - Network error handling (fetch fails)

### Verification

- [ ] Browser shows "Loading..." then progressive content
- [ ] No JavaScript console errors
- [ ] Network tab shows chunked streaming (not single large response)
- [ ] Final output matches embedded version visually

---

## Phase 4: Documentation & Loading Indicator for Embedded

### Goal
Add loading indicator to embedded HTML, update documentation.

### Steps

1. **Add loading indicator to `internal/html/template.go`**
   - Show "Loading..." before base64 parsing
   - Replace with terminal content after xterm.write()

2. **Update `README.md`**
   - Document `RenderStreamingHTML()` in Library Usage section
   - Explain when to use each mode:
     - `RenderHTML()`: embedded, works offline, best for < 1MB
     - `RenderStreamingHTML()`: streaming, requires server, best for large files
   - Add code examples for both

3. **Add code comments**
   - Document streaming JavaScript logic
   - Explain escape sequence boundary handling
   - Explain trailing buffer for footer detection

### Verification

- [ ] Embedded HTML shows "Loading..." before content (test with large file)
- [ ] README examples are copy-paste runnable
- [ ] `make test` passes

---

## File Changes Summary

### New Files
- `public/streaming-test.html` — Manual test harness (not for production)
- `internal/html/template_streaming.go` — Streaming HTML template
- `internal/html/testserver_test.go` — Integration test server

### Modified Files
- `playback/types.go` — Add `StreamingOptions`
- `playback/playback.go` — Add `RenderStreamingHTML()`
- `playback/playback_test.go` — Add streaming tests
- `internal/html/template.go` — Add loading indicator
- `README.md` — Document streaming mode

---

## Success Criteria

1. 12MB `sample.log` renders progressively in browser (visible content within 1-2 seconds)
2. Final rendered output matches embedded version
3. All existing tests pass
4. No JavaScript console errors
5. Works when served from Go web server on port 3000
