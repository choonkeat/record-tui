# Streaming Cleaner JS API

## Goal

Modify `cleaner.js` to support streaming/chunked data processing so it can be used by `RenderStreamingHTML` without waiting for entire files to be assembled. The streaming API must produce **byte-identical output** to the batch `cleanSessionContent()` function, verified by `make test`.

## Constraints

- **No Go code changes in Phases 1-5** - only JS modifications until streaming is verified
- **Phase 6 modifies Go file** - but only the embedded JS string in `template_streaming.go`
- **`make test` must pass** - byte-level parity with Go output

## Phases Overview

1. ✅ **Phase 1: Streaming API Implementation** - Add `createStreamingCleaner()` to `cleaner.js`
2. ✅ **Phase 2: Update Node.js Test Harness** - Modify `generate_output.js` to use streaming API with randomized chunks
3. ✅ **Phase 3: Verify Parity & Fix Discrepancies** - Run `make test`, debug any mismatches
4. ✅ **Phase 4: Add Self-Tests for Streaming** - Add streaming-specific unit tests to `cleaner.js`
5. ✅ **Phase 5: Remove Orphaned Code** - Delete unused batch functions
6. ✅ **Phase 6: Update RenderStreamingHTML** - Replace inline JS in `template_streaming.go` with tested streaming cleaner

---

## Phase 1: Streaming API Implementation

### What will be achieved
A `createStreamingCleaner(onOutput)` function in `cleaner.js` that:
- Accepts arbitrary chunks of data via `write(chunk)`
- Emits cleaned data via callback as it becomes available
- Handles `end()` to flush remaining buffered content
- Produces identical output to batch `cleanSessionContent()`

### Small steps

1. **Review current partial implementation** - Already added `createStreamingCleaner()`, verify logic handles all edge cases

2. **Header handling**:
   - Buffer first ~5 lines worth of content
   - Call `stripHeader()` once we have enough lines
   - Pass through remaining content

3. **Clear sequence handling with state machine**:
   - Track `hasEmittedContent` (have we output any non-empty content?)
   - Track `pendingSeparator` (did we see a clear after content, waiting for more content?)
   - When new non-empty content arrives with `pendingSeparator=true`, emit SEPARATOR first

4. **Escape sequence boundary handling**:
   - If chunk ends with `\x1b` or partial escape, buffer it for next chunk
   - Prevents splitting escape sequences across chunks

5. **Footer handling**:
   - Keep trailing ~500 bytes in buffer
   - At `end()`, call `stripFooter()` on the trailing buffer

### Verification (TDD)

1. **Red**: The current implementation may have bugs - `make test` will likely fail after Phase 2 switches to streaming
2. **Green**: Fix any discrepancies found in Phase 3
3. **Refactor**: Clean up code after tests pass

---

## Phase 2: Update Node.js Test Harness

### What will be achieved
Modify `generate_output.js` to use `createStreamingCleaner()` with randomized chunk sizes, proving the streaming API produces identical output regardless of how data is chunked.

### Small steps

1. **Import `createStreamingCleaner`** from `cleaner.js`

2. **Add random chunk size generator**:
   ```javascript
   // Allow reproducible runs via SEED env var
   const seed = process.env.SEED ? parseInt(process.env.SEED) : Date.now();
   console.log(`Random seed: ${seed} (reproduce with: SEED=${seed} make test)`);

   // Simple seeded PRNG
   let rngState = seed;
   function random() {
     rngState = (rngState * 1103515245 + 12345) & 0x7fffffff;
     return rngState / 0x7fffffff;
   }

   function getRandomChunkSize() {
     return Math.floor(random() * (16384 - 64)) + 64;  // 64 bytes to 16KB
   }
   ```

3. **Implement streaming processor with random chunks**:
   ```javascript
   function processWithStreaming(content) {
     const chunks = [];
     const cleaner = createStreamingCleaner((c) => chunks.push(c));

     let offset = 0;
     while (offset < content.length) {
       const size = getRandomChunkSize();
       cleaner.write(content.slice(offset, offset + size));
       offset += size;
     }
     cleaner.end();

     return chunks.join('');
   }
   ```

4. **Replace batch call with streaming**:
   - Change `cleanSessionContent(content)` to `processWithStreaming(content)`

5. **Keep test harness output format identical** (only in `generate_output.js`):
   - Same header: `"{input_length} bytes\n{cleaned_content}"`

6. **Add `clean-compare-output` Makefile target**:
   ```makefile
   clean-compare-output:
   	rm -rf ./recordings-output/
   ```

7. **Update `test` target ordering**:
   ```makefile
   test: clean-compare-output test-go test-js compare-output
   ```

### Verification (TDD)

1. **Red**: Run `make test` - likely fails if streaming logic has bugs
2. **Green**: Phase 3 will fix any discrepancies

---

## Phase 3: Verify Parity & Fix Discrepancies

### What will be achieved
Run `make test` and fix any byte-level differences between streaming JS output and Go output. This is the debugging/fixing phase where we ensure the streaming implementation is correct.

### Small steps

1. **Run `make test`** - expect potential failures

2. **If failures occur, analyze the diff output**:
   - `TestCompareGoAndJsOutput` provides detailed debug info:
     - File sizes (Go vs JS)
     - First differing byte position
     - Hex dump around difference
     - `diff -u` text output

3. **Common issues to check**:
   - **Chunk boundary bugs**: Clear sequence split across chunks
   - **Header detection timing**: Not enough lines buffered before stripping
   - **Trailing buffer size**: Footer partially emitted before `end()`
   - **State machine bugs**: `pendingSeparator` not set/cleared correctly
   - **Whitespace handling**: Empty lines treated differently than batch

4. **Fix issues in `createStreamingCleaner()`**:
   - Adjust buffer sizes if needed
   - Fix state transitions
   - Ensure escape sequence buffering is correct

5. **Re-run `make test`** until all files pass

6. **Test with different chunk sizes** (optional verification):
   - Very small chunks (e.g., 64 bytes) - stress tests boundary handling
   - Very large chunks (e.g., 1MB) - approaches batch behavior
   - Verify output is identical regardless of chunk size

### Verification (TDD)

1. **Red**: `make test` fails with detailed diff
2. **Green**: Fix streaming logic until `make test` passes
3. **Refactor**: Clean up any debug code added during fixing

### Regression guarantee
- Go output is the source of truth (unchanged)
- Every fix is verified against Go output
- No changes to Go code

---

## Phase 4: Add Self-Tests for Streaming

### What will be achieved
Add streaming-specific unit tests to the self-test section of `cleaner.js` (run via `node cleaner.js`). These tests verify the streaming API works correctly with various edge cases.

### Small steps

1. **Add streaming test helper**:
   ```javascript
   function testStreaming(input, chunkSize, expected, testName) {
     const chunks = [];
     const cleaner = createStreamingCleaner((c) => chunks.push(c));
     for (let i = 0; i < input.length; i += chunkSize) {
       cleaner.write(input.slice(i, i + chunkSize));
     }
     cleaner.end();
     const result = chunks.join('');
     console.log(testName + ':', result === expected ? 'PASS' : 'FAIL');
   }
   ```

2. **Add test cases**:
   - **Basic streaming**: Content with no clears, verify output matches
   - **Clear at chunk boundary**: Clear sequence split across two chunks
   - **Multiple clears**: Verify separators added correctly
   - **Clear at start**: No separator (no content before)
   - **Clear at end**: No separator (no content after)
   - **Small chunks**: Process byte-by-byte (stress test)
   - **Header/footer stripping**: Verify metadata removed in streaming mode

3. **Verify streaming matches batch** (key test):
   ```javascript
   // For each test input, verify streaming output === batch output
   const batchResult = cleanSessionContent(input);  // Before we remove it
   const streamResult = processWithStreaming(input);
   assert(batchResult === streamResult);
   ```

4. **Run self-tests**: `node internal/js/cleaner.js`

### Verification (TDD)

1. **Red**: Write tests first, some may fail initially
2. **Green**: Tests pass after Phase 3 fixes
3. **Refactor**: Clean up test code

---

## Phase 5: Remove Orphaned Code

### What will be achieved
Clean up `cleaner.js` by removing functions that are no longer used after switching to the streaming-only approach.

### Small steps

1. **Identify orphaned code**:
   - `cleanSessionContent()` - was batch entry point, now unused
   - `neutralizeClearSequences()` - only called by `cleanSessionContent()`

2. **Keep required functions**:
   - `stripHeader()` - used by `createStreamingCleaner()`
   - `stripFooter()` - used by `createStreamingCleaner()`
   - `createStreamingCleaner()` - the new primary API
   - `CLEAR_SEPARATOR` - used by streaming cleaner
   - `clearPattern` - used by streaming cleaner

3. **Remove orphaned functions**:
   - Delete `neutralizeClearSequences()` function
   - Delete `cleanSessionContent()` function

4. **Update exports**:
   ```javascript
   module.exports = {
     CLEAR_SEPARATOR,
     clearPattern,
     stripHeader,
     stripFooter,
     createStreamingCleaner
   };
   ```

5. **Update self-tests**:
   - Remove tests that use `cleanSessionContent()`
   - Remove tests that use `neutralizeClearSequences()` directly
   - Keep/update streaming tests from Phase 4

6. **Run `make test`** to verify nothing broke

### Verification (TDD)

1. **Red**: If we accidentally remove something needed, `make test` fails
2. **Green**: All tests pass with cleaned-up code
3. **Verify**: `node internal/js/cleaner.js` self-tests still pass

### Regression guarantee
- `make test` passes before AND after cleanup
- Streaming behavior unchanged (we're only removing unused code)

---

## Phase 6: Update RenderStreamingHTML

### What will be achieved
Replace the simplified inline JS in `internal/html/template_streaming.go` with the tested `createStreamingCleaner()` implementation, ensuring the browser uses the same sophisticated clear sequence logic as Go.

### Small steps

1. **Extract browser-ready JS from `cleaner.js`**:
   - Copy `CLEAR_SEPARATOR`, `clearPattern`, `stripHeader`, `stripFooter`
   - Copy `createStreamingCleaner()` (including `processForClears`)
   - Remove Node.js-specific code (module.exports, require.main check)

2. **Update `template_streaming.go` embedded JS**:
   - Replace existing `stripHeader()`, `stripFooter()`, `neutralizeClearSequences()` with new functions
   - Add `createStreamingCleaner()` function

3. **Update `streamSession()` to use streaming cleaner**:
   - Replace manual header/footer/clear handling with:
   ```javascript
   const cleaner = createStreamingCleaner((chunk) => {
     if (firstWrite) {
       loadingDiv.style.display = 'none';
       firstWrite = false;
     }
     xterm.write(chunk);
   });

   while (true) {
     const result = await reader.read();
     if (result.done) break;
     cleaner.write(decoder.decode(result.value, { stream: true }));
   }
   cleaner.end();
   ```

4. **Remove now-unused code from template**:
   - Remove old `neutralizeClearSequences()` (simple replace version)
   - Remove manual trailing buffer logic (now handled by cleaner)

5. **Manual browser testing**:
   - Build: `make build`
   - Generate streaming HTML for a recording
   - Open in browser, verify it renders correctly
   - Check browser console for errors

### Verification

1. **Visual verification**: Recording plays correctly in browser
2. **Console check**: No JavaScript errors
3. **Compare with embedded HTML**: Same recording rendered via embedded HTML should look identical
4. **`make test` still passes**: Go/JS parity tests unaffected (they test `cleaner.js`, not template)

### Regression guarantee
- Streaming HTML output should render identically to before
- `make test` continues to pass
- Browser console has no errors

---

## Files to Modify

### Phases 1-5 (JS only)
- `internal/js/cleaner.js` - Add streaming API, later remove orphaned code
- `internal/js/generate_output.js` - Switch to streaming with random chunks
- `Makefile` - Add `clean-compare-output` target, update `test` target ordering

### Phase 6 (Go file with embedded JS)
- `internal/html/template_streaming.go` - Replace inline JS with tested streaming cleaner

### Not Modified
- `internal/session/*.go` (Go source of truth)
- `playback/*.go`
