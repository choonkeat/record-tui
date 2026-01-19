# Neutralize Terminal Clear Sequences

## Goal

Neutralize terminal clear sequences in recorded sessions so that content before a "clear" is preserved in the generated HTML, rather than being wiped when xterm.js interprets the escape sequences.

---

## Phase 1: Research & Test Setup

### What will be achieved
A comprehensive understanding of the clear escape sequences we need to handle, and a failing test that demonstrates the current problem.

### Steps

1. **Research clear escape sequences** — Document the common sequences:
   - `\x1b[2J` — Clear entire screen
   - `\x1b[H` — Cursor home (often paired with clear)
   - `\x1b[3J` — Clear entire screen including scrollback
   - `\x1b[J` / `\x1b[0J` — Clear from cursor to end of screen
   - Combined sequences like `\x1b[2J\x1b[H` or `\x1b[H\x1b[2J`

2. **Create test file** — Add `internal/session/clear_test.go` with tests:
   - Test that detects a simple clear sequence mid-content
   - Test with multiple clears
   - Test with no clear (passthrough)
   - Test with clear at start/end of content

3. **Write failing tests first (red)** — Tests call a not-yet-implemented `NeutralizeClearSequences(content string) string` function that should replace clears with a separator

### Verification
- Tests exist and fail with "function not found" or "unexpected output"
- Running `make test` shows the new tests failing
- This confirms our test setup is correct before implementation

---

## Phase 2: Core Implementation

### What will be achieved
A working `NeutralizeClearSequences` function in the `session` package that replaces clear sequences with a visual separator.

### Steps

1. **Create `internal/session/clear.go`** — New file with:
   - `NeutralizeClearSequences(content string) string` function
   - A constant for the separator text (e.g., `\n\n─── terminal cleared ───\n\n`)
   - Regex or string matching for clear sequences

2. **Handle the common patterns**:
   - `\x1b[2J` alone → replace with separator
   - `\x1b[H\x1b[2J` (home then clear) → replace entire combo with separator
   - `\x1b[2J\x1b[H` (clear then home) → replace entire combo with separator
   - `\x1b[3J` (clear with scrollback) → replace with separator
   - Be careful NOT to touch `\x1b[H` alone (cursor home without clear is common)

3. **Collapse consecutive clears** — If multiple clear sequences appear back-to-back, emit only one separator

4. **Make tests pass (green)**

### Verification
- `make test` passes
- Tests from Phase 1 go from red to green
- Add a manual verification: create a test content string with clear sequence, run through function, confirm separator appears

---

## Phase 3: Integration

### What will be achieved
The `NeutralizeClearSequences` function is called at the right point in the pipeline so that generated HTML preserves content before clears.

### Steps

1. **Identify integration point** — The function should be called in `session.StripMetadata` after stripping metadata but before returning

2. **Update `internal/session/strip.go`** — Call `NeutralizeClearSequences` at the end of `StripMetadata`

3. **Add integration test** — Test in `playback/playback_test.go` that verifies clear sequences are neutralized when using the public API

### Verification
- `make test` passes (no regression in existing tests)
- New integration test confirms end-to-end behavior
- Manual test: create a `session.log` with clear sequence, run through `playback.RenderHTML`, inspect output HTML

---

## Phase 4: Edge Cases & Polish

### What will be achieved
Handle edge cases gracefully and ensure the visual separator looks good in the rendered HTML.

### Steps

1. **Handle cursor repositioning after clear** — Strip `\x1b[H` immediately after clear since it's redundant after our separator

2. **Test edge cases**:
   - Clear at very start of content (no separator needed, just strip)
   - Clear at very end of content (no separator needed, just strip)
   - Content that is only clear sequences (return empty or minimal output)
   - Clear followed by `\x1b[H` (strip both, emit one separator)

3. **Visual separator styling** — Use Unicode box-drawing character for a clean look:
   - `\n\n──────── terminal cleared ────────\n\n`
   - Ensure it's visible but not distracting

4. **Update existing tests if needed** — If any existing tests in `session` or `playback` packages assumed clear sequences pass through unchanged, update them

5. **Manual end-to-end test** — Record a real session with `clear` command, generate HTML, verify both halves are visible

### Verification
- `make test` passes
- All edge case tests pass
- Manual inspection of generated HTML shows separator renders nicely in xterm.js
- No regression: existing test HTML files still look correct

---

## Summary

| Phase | Description | Key Deliverable |
|-------|-------------|-----------------|
| 1 | Research & Test Setup | Failing tests in `internal/session/clear_test.go` |
| 2 | Core Implementation | Working `NeutralizeClearSequences` function in `internal/session/clear.go` |
| 3 | Integration | `StripMetadata` calls `NeutralizeClearSequences`, integration test passes |
| 4 | Edge Cases & Polish | Handle edge cases, nice separator, manual verification |
