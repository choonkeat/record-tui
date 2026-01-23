# Go/JS Output Parity Test Infrastructure

## Goal

Create a cross-language test infrastructure to verify that Go and JavaScript implementations of session log cleaning produce **identical output**, ensuring the browser-side JS (used in streaming HTML) matches the server-side Go logic exactly.

## Phases Overview

1. **Phase 1: Go Test Infrastructure** - Create Go test that generates `.go.output` files ✅ COMPLETE
2. **Phase 2: Browser-Compatible JS Module** - Extract/create standalone JS cleaning functions ✅ COMPLETE
3. **Phase 3: Node.js Test Harness** - Create Node.js script that generates `.js.output` files ✅ COMPLETE
4. **Phase 4: Output Comparison & Makefile** - Go-based comparison test + Makefile restructuring

---

## Phase 1: Go Test Infrastructure

### What will be achieved
A Go test that processes `./recordings/session*.log` files through the same cleaning pipeline used by `RenderHTML` (steps 1-4: read bytes → string → strip metadata → neutralize clears) and writes results to `./recordings-output/*.go.output`.

### Small steps

1. **Create test file** `internal/session/output_test.go` with a test function `TestGenerateGoOutput`

2. **Implement directory handling**:
   - Define a const for the recordings directory (initially `./recordings-123` for TDD)
   - Check if directory exists
   - If missing, log warning with `t.Skip()`
   - Create `./recordings-output` directory if it doesn't exist

3. **Implement file discovery**:
   - Use `filepath.Glob` to find `{recordingsDir}/session*.log`
   - Use `filepath.EvalSymlinks` to follow symlinks
   - Sort results for deterministic ordering

4. **For each file, run the cleaning pipeline**:
   - `os.ReadFile()` → `[]byte`
   - `string()` cast
   - `session.StripMetadata()` (which internally calls `NeutralizeClearSequences`)
   - This already shares code with `RenderHTML` via `playback.StripMetadata()` → `session.StripMetadata()`

5. **Write output**:
   - Write to `recordings-output/{basename}.go.output`
   - Use `os.WriteFile` with `0644` permissions

### Verification (TDD)

1. **Red**: Set const to `./recordings-123` (doesn't exist) → run test → should skip with warning message
2. **Green**: Change const to `./recordings` (exists) → run test → should produce `./recordings-output/*.go.output` files
3. **Verify**: Manually check output files exist and contain expected cleaned content

### Regression guarantee
- No changes to existing `session.StripMetadata()` or `NeutralizeClearSequences()`
- New test file only, no modifications to production code
- Existing `make test` (soon `test-go`) continues to pass

---

## Phase 2: Browser-Compatible JS Module

### What will be achieved
A standalone JavaScript file containing the session cleaning logic that:
- Works identically in both Node.js and browsers
- Uses no Node.js-specific APIs (no `fs`, `path`, `Buffer`, etc.)
- Can be copy-pasted into HTML templates as-is
- Produces byte-for-byte identical output to the Go implementation

### Small steps

1. **Create JS file** `internal/js/cleaner.js` with the cleaning functions

2. **Implement functions matching Go's logic exactly**:
   - `CLEAR_SEPARATOR` constant: `'\n\n──────── terminal cleared ────────\n\n'`
   - `clearPattern` regex: `/\x1b\[H\x1b\[[23]J|\x1b\[[23]J\x1b\[H|\x1b\[[23]J/g`
   - `stripHeader(text)` - matches Go's `cleaner.go:17-24`
   - `stripFooter(text)` - matches Go's `cleaner.go:26-48`
   - `neutralizeClearSequences(text)` - matches Go's `clear.go:22-64` **exactly** (sophisticated logic, not simple replace)

3. **Create main entry function** `cleanSessionContent(text)`:
   - Calls: `text` → `stripHeader()` → `stripFooter()` → `neutralizeClearSequences()`
   - Returns cleaned string

4. **Add module exports** compatible with both environments:
   ```javascript
   if (typeof module !== 'undefined' && module.exports) {
     module.exports = { cleanSessionContent, stripHeader, stripFooter, neutralizeClearSequences };
   }
   ```

### Critical: JS must match Go's sophisticated `neutralizeClearSequences`

The current JS in `template_streaming.go` uses simple `replace()`:
```javascript
function neutralizeClearSequences(text) {
  return text.replace(clearPattern, CLEAR_SEPARATOR);
}
```

But Go's `clear.go:22-64` only adds separator if there's non-empty content before AND after. The JS must replicate this logic exactly.

### Verification (TDD)

1. **Red**: Create `cleaner.js` with deliberate difference → Phase 4's comparison will fail
2. **Green**: Fix JS to match Go exactly → comparison passes
3. **Unit test**: Add inline test runnable with `node cleaner.js` for basic sanity check

### Regression guarantee
- No changes to existing Go code
- No changes to `template_streaming.go` yet
- New file only

---

## Phase 3: Node.js Test Harness

### What will be achieved
A Node.js script that reads `./recordings/session*.log` files, passes content through the browser-compatible JS cleaning function, and writes output to `./recordings-output/*.js.output`.

### Small steps

1. **Create Node.js test script** `internal/js/generate_output.js`

2. **Implement directory handling**:
   - Check if `./recordings` exists
   - If missing, log warning and exit gracefully (exit code 0)
   - Create `./recordings-output` directory if it doesn't exist

3. **Implement file discovery**:
   - Use `fs.readdirSync()` to list `./recordings`
   - Filter for files matching `session*.log`
   - Use `fs.realpathSync()` to follow symlinks
   - Sort results for deterministic ordering

4. **For each file, run the cleaning pipeline**:
   - `fs.readFileSync(path)` → `Buffer`
   - `.toString('utf8')` → string
   - `cleanSessionContent(text)` from `cleaner.js`

5. **Write output**:
   - Write to `recordings-output/{basename}.js.output`
   - Use `fs.writeFileSync()` with UTF-8 encoding

6. **Add execution entry point**:
   - Script runs when invoked directly
   - Print progress (file count, files processed)

### Verification (TDD)

1. **Red**: Set directory const to `./recordings-123` → should log warning and exit cleanly
2. **Green**: Change const to `./recordings` → should produce `./recordings-output/*.js.output` files
3. **Verify**: Check `.js.output` files exist alongside `.go.output` files

### Regression guarantee
- No changes to existing code
- New file only: `internal/js/generate_output.js`

---

## Phase 4: Output Comparison & Makefile

### What will be achieved
A Go-based comparison test with debug-friendly output, plus Makefile restructuring.

### Small steps

1. **Rename existing Makefile target**:
   - Rename `test` to `test-go`
   - Keep same command: `go test ./internal/... -v`

2. **Add `test-js` target**:
   - Run `node internal/js/generate_output.js`

3. **Create Go comparison test** `internal/session/compare_output_test.go`:
   - Function `TestCompareGoAndJsOutput`
   - List all `*.go.output` files in `./recordings-output`
   - For each, find corresponding `*.js.output`
   - Binary comparison: `bytes.Equal()` on raw bytes
   - On mismatch, provide debug-friendly output:
     - File sizes (Go vs JS)
     - First differing byte position
     - Hex dump of bytes around difference (±16 bytes)
     - Character representation where printable
     - Run `diff -u` via `exec.Command` for text diff
   - Fail if `.go.output` has no corresponding `.js.output` (or vice versa)
   - Skip if `./recordings-output` doesn't exist or is empty

4. **Add `compare-output` target**:
   - Run `go test ./internal/session -run TestCompareGoAndJsOutput -v`

5. **Add combined `test` target**:
   ```makefile
   test: test-go test-js compare-output
   ```

6. **Add `.gitignore` entries**:
   - `recordings/`
   - `recordings-output/`

### Debug-friendly output example
```
=== MISMATCH: session-example.log ===
Go output:  1234 bytes (recordings-output/session-example.log.go.output)
JS output:  1230 bytes (recordings-output/session-example.log.js.output)

First difference at byte 456:
  Go:  [0x0a 0x0a 0xe2 0x94 0x80 0xe2 0x94 0x80] "──"
  JS:  [0x0a 0xe2 0x94 0x80 0xe2 0x94 0x80 0xe2] "─"

--- diff -u output ---
@@ -15,3 +15,2 @@
 some content
-
 ──────── terminal cleared ────────
```

### Verification (TDD)

1. **Red**: Temporarily tweak `cleaner.js` (e.g., change separator) → `make test` fails with clear byte-level diff
2. **Green**: Fix `cleaner.js` to match Go → `make test` passes
3. **Verify**: Intentionally break something, confirm debug output is clear

### Regression guarantee
- Existing test behavior preserved (renamed to `test-go`)
- `make test-go` runs exactly what `make test` used to run
- New comparison test in separate function

---

## Files to Create/Modify

### New Files
- `internal/session/output_test.go` - Go test generating `.go.output`
- `internal/session/compare_output_test.go` - Go comparison test
- `internal/js/cleaner.js` - Browser-compatible cleaning functions
- `internal/js/generate_output.js` - Node.js test harness

### Modified Files
- `Makefile` - Rename `test` to `test-go`, add `test-js`, `compare-output`, `test`
- `.gitignore` - Add `recordings/` and `recordings-output/`

### Not Modified (source of truth)
- `internal/session/cleaner.go`
- `internal/session/clear.go`
- `internal/html/template_streaming.go` (JS will be updated separately later)
