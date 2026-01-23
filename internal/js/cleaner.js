/**
 * Browser-compatible session log cleaner.
 * This file replicates the Go logic in internal/session/cleaner.go and clear.go exactly.
 * It can be used in both Node.js (CommonJS) and browsers (copy-paste into HTML).
 */

// Clear sequence separator - must match Go's ClearSeparator in clear.go:9
// Using raw UTF-8 bytes for '─' (U+2500) = 0xe2 0x94 0x80 to ensure byte-level parity with Go
// when processing files as latin1 (raw bytes)
const CLEAR_SEPARATOR = '\n\n\xe2\x94\x80\xe2\x94\x80\xe2\x94\x80\xe2\x94\x80\xe2\x94\x80\xe2\x94\x80\xe2\x94\x80\xe2\x94\x80 terminal cleared \xe2\x94\x80\xe2\x94\x80\xe2\x94\x80\xe2\x94\x80\xe2\x94\x80\xe2\x94\x80\xe2\x94\x80\xe2\x94\x80\n\n';

// Clear sequence pattern - must match Go's clearPattern in clear.go:16
// Matches: \x1b[H\x1b[2J, \x1b[H\x1b[3J, \x1b[2J\x1b[H, \x1b[3J\x1b[H, \x1b[2J, \x1b[3J
const clearPattern = /\x1b\[H\x1b\[[23]J|\x1b\[[23]J\x1b\[H|\x1b\[[23]J/g;

/**
 * Strip header lines from session content.
 * Matches Go's cleaner.go:17-24 exactly.
 * Removes first ~5 lines starting with "Script started on" or "Command:"
 */
function stripHeader(text) {
  const lines = text.split('\n');
  let startIndex = 0;

  // Find where actual content starts (skip header)
  // The header consists of "Script started on..." followed by "Command: ..."
  for (let i = 0; i < lines.length && i < 5; i++) {
    const line = lines[i];
    if (line.startsWith('Script started on') || line.startsWith('Command:')) {
      startIndex = i + 1;
    }
  }

  if (startIndex === 0) {
    return text; // No header found
  }

  return lines.slice(startIndex).join('\n');
}

/**
 * Strip footer lines from session content.
 * Matches Go's cleaner.go:26-48 exactly.
 * Removes trailing lines containing "Saving session", "Command exit status", "Script done on"
 */
function stripFooter(text) {
  const lines = text.split('\n');
  let endIndex = lines.length;

  // Find where actual content ends (skip footer)
  // Footer can contain "Saving session", "Command exit status", "Script done on" in any order
  // Work backwards from end of file
  let footerStartIndex = lines.length;
  for (let i = lines.length - 1; i >= 0; i--) {
    const line = lines[i];
    // Check if this line is part of the footer
    if (line.includes('Saving session') ||
        line.includes('Command exit status') ||
        line.includes('Script done on') ||
        (line.trim() === '' && i > 0)) {
      footerStartIndex = i;
    } else if (footerStartIndex < lines.length) {
      // We've found content before the footer, stop looking
      break;
    }
  }
  endIndex = footerStartIndex;

  // Trim any trailing empty lines from the content
  while (endIndex > 0 && lines[endIndex - 1].trim() === '') {
    endIndex--;
  }

  if (endIndex >= lines.length) {
    return text; // No footer found
  }

  return lines.slice(0, endIndex).join('\n');
}

/**
 * Create a streaming cleaner for processing chunked data.
 * Processes: stripHeader -> (streaming content with clear sequence handling) -> stripFooter
 *
 * @param {function(string): void} onOutput - Callback invoked with cleaned chunks
 * @returns {{write: function(string): void, end: function(): void}}
 *
 * Usage:
 *   const cleaner = createStreamingCleaner((chunk) => process.stdout.write(chunk));
 *   cleaner.write(chunk1);
 *   cleaner.write(chunk2);
 *   cleaner.end();
 */
function createStreamingCleaner(onOutput) {
  // Header state - buffer first few lines to detect and strip header
  let headerBuffer = '';
  let headerStripped = false;
  const HEADER_LINES_THRESHOLD = 5;

  // Clear sequence state - track whether we need to emit separator before next content
  let hasEmittedContent = false;
  let pendingSeparator = false;

  // Buffer for whitespace that follows a clear sequence
  // This whitespace should be prepended to the next non-empty content (after the separator)
  let pendingWhitespace = '';

  // Buffer for incomplete escape sequences at chunk boundaries
  let escapeBuffer = '';

  // Trailing buffer for footer detection
  let trailingBuffer = '';
  const TRAILING_SIZE = 500;

  /**
   * Process text for clear sequences, respecting streaming state.
   * Updates hasEmittedContent and pendingSeparator as side effects.
   */
  function processForClears(text) {
    if (!text) return '';

    clearPattern.lastIndex = 0;
    const matches = [];
    let m;
    while ((m = clearPattern.exec(text)) !== null) {
      matches.push([m.index, m.index + m[0].length]);
    }

    if (matches.length === 0) {
      // No clears - check if we need to emit pending separator
      if (text.trim() !== '') {
        if (pendingSeparator) {
          // Emit separator, any pending whitespace, then this content
          const result = CLEAR_SEPARATOR + pendingWhitespace + text;
          pendingSeparator = false;
          pendingWhitespace = '';
          hasEmittedContent = true;
          return result;
        }
        hasEmittedContent = true;
        return text;
      }
      // Text is whitespace-only
      if (pendingSeparator) {
        // Buffer whitespace to prepend after separator when we see non-empty content
        pendingWhitespace += text;
        return '';
      }
      return text;
    }

    let result = '';
    let lastEnd = 0;

    for (const [start, end] of matches) {
      const before = text.slice(lastEnd, start);

      if (before.trim() !== '') {
        if (pendingSeparator) {
          result += CLEAR_SEPARATOR + pendingWhitespace;
          pendingSeparator = false;
          pendingWhitespace = '';
        }
        result += before;
        hasEmittedContent = true;
      } else if (pendingSeparator) {
        // Whitespace-only before a clear - buffer it
        pendingWhitespace += before;
      }

      // After seeing a clear, if we had content before, we might need separator
      if (hasEmittedContent) {
        pendingSeparator = true;
        // Discard any pending whitespace - it was before this clear, so should be dropped
        // (batch logic: whitespace-only content before a clear is not included)
        pendingWhitespace = '';
      }

      lastEnd = end;
    }

    // Handle remaining after last clear
    const remaining = text.slice(lastEnd);
    if (remaining.trim() !== '') {
      if (pendingSeparator) {
        result += CLEAR_SEPARATOR + pendingWhitespace;
        pendingSeparator = false;
        pendingWhitespace = '';
      }
      result += remaining;
      hasEmittedContent = true;
    } else if (remaining && pendingSeparator) {
      // Whitespace-only after the last clear - buffer it for next chunk
      pendingWhitespace += remaining;
    }

    return result;
  }

  /**
   * Feed a chunk of data to the cleaner.
   * May invoke onOutput zero or more times.
   */
  function write(chunk) {
    // Prepend any buffered incomplete escape sequence
    let text = escapeBuffer + chunk;
    escapeBuffer = '';

    // Check for incomplete escape sequence at end (escape sequences are typically <10 bytes)
    const lastEsc = text.lastIndexOf('\x1b');
    if (lastEsc >= 0 && lastEsc > text.length - 10) {
      // Might be incomplete, buffer it for next chunk
      escapeBuffer = text.slice(lastEsc);
      text = text.slice(0, lastEsc);
    }

    // Handle header - buffer until we have enough lines
    if (!headerStripped) {
      headerBuffer += text;
      text = '';

      // Count newlines to determine if we have enough for header detection
      const newlineCount = (headerBuffer.match(/\n/g) || []).length;
      if (newlineCount >= HEADER_LINES_THRESHOLD) {
        headerBuffer = stripHeader(headerBuffer);
        headerStripped = true;
        text = headerBuffer;
        headerBuffer = '';
      } else {
        return; // Need more data for header detection
      }
    }

    // Add to trailing buffer
    text = trailingBuffer + text;
    trailingBuffer = '';

    // Keep trailing portion for footer detection at end
    if (text.length > TRAILING_SIZE) {
      const toEmit = text.slice(0, -TRAILING_SIZE);
      trailingBuffer = text.slice(-TRAILING_SIZE);
      const processed = processForClears(toEmit);
      if (processed) onOutput(processed);
    } else {
      trailingBuffer = text;
    }
  }

  /**
   * Signal end of stream. Processes remaining buffered data.
   * Must be called to flush final content.
   */
  function end() {
    // Combine all remaining buffers
    let text = trailingBuffer + escapeBuffer;

    // If header wasn't stripped yet (very small input), strip it now
    if (!headerStripped) {
      text = headerBuffer + text;
      text = stripHeader(text);
    }

    // Strip footer from final content
    text = stripFooter(text);

    // Process for clears and emit
    const processed = processForClears(text);
    if (processed) onOutput(processed);
  }

  return { write, end };
}

// Export for Node.js (CommonJS)
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    CLEAR_SEPARATOR,
    clearPattern,
    stripHeader,
    stripFooter,
    createStreamingCleaner
  };
}

// Simple self-test when run directly with node
if (typeof require !== 'undefined' && require.main === module) {
  console.log('Running cleaner.js self-test...');

  // Test stripHeader
  const headerTest = 'Script started on Wed Dec 31 12:10:34 2025\nCommand: bash\nhello world';
  const headerResult = stripHeader(headerTest);
  console.log('stripHeader test:', headerResult === 'hello world' ? 'PASS' : 'FAIL');

  // Test stripFooter
  const footerTest = 'hello world\n\nScript done on Wed Dec 31 12:11:22 2025';
  const footerResult = stripFooter(footerTest);
  console.log('stripFooter test:', footerResult === 'hello world' ? 'PASS' : 'FAIL');

  // === Streaming Tests ===
  console.log('\nRunning streaming cleaner tests...');

  // Helper: process with streaming at given chunk size
  function testStreaming(input, chunkSize) {
    const chunks = [];
    const cleaner = createStreamingCleaner((c) => chunks.push(c));
    for (let i = 0; i < input.length; i += chunkSize) {
      cleaner.write(input.slice(i, i + chunkSize));
    }
    cleaner.end();
    return chunks.join('');
  }

  // Helper: verify streaming output matches expected
  function verifyStreaming(input, chunkSize, expected, testName) {
    const result = testStreaming(input, chunkSize);
    const pass = result === expected;
    console.log(`${testName}: ${pass ? 'PASS' : 'FAIL'}`);
    if (!pass) {
      console.log(`  Expected (${expected.length} bytes): ${JSON.stringify(expected.slice(0, 100))}...`);
      console.log(`  Got (${result.length} bytes): ${JSON.stringify(result.slice(0, 100))}...`);
    }
    return pass;
  }

  // Helper: verify same result with different chunk sizes
  function verifyChunkIndependence(input, expected, testName) {
    const result1 = testStreaming(input, 1);      // byte-by-byte
    const result2 = testStreaming(input, 7);      // small chunks
    const result3 = testStreaming(input, 1024);   // large chunks
    const pass = result1 === expected && result2 === expected && result3 === expected;
    console.log(`${testName}: ${pass ? 'PASS' : 'FAIL'}`);
    if (!pass) {
      console.log(`  Expected: ${JSON.stringify(expected.slice(0, 50))}...`);
      console.log(`  Chunk 1: ${JSON.stringify(result1.slice(0, 50))}...`);
      console.log(`  Chunk 7: ${JSON.stringify(result2.slice(0, 50))}...`);
      console.log(`  Chunk 1024: ${JSON.stringify(result3.slice(0, 50))}...`);
    }
    return pass;
  }

  // Test 1: Basic clear sequence
  verifyChunkIndependence(
    'before\x1b[2Jafter',
    'before' + CLEAR_SEPARATOR + 'after',
    'streaming basic clear'
  );

  // Test 2: Clear at start (no separator)
  verifyChunkIndependence(
    '\x1b[2Jcontent',
    'content',
    'streaming clear at start'
  );

  // Test 3: Clear at end (no separator)
  verifyChunkIndependence(
    'content\x1b[2J',
    'content',
    'streaming clear at end'
  );

  // Test 4: Multiple clears
  verifyChunkIndependence(
    'a\x1b[2Jb\x1b[2Jc',
    'a' + CLEAR_SEPARATOR + 'b' + CLEAR_SEPARATOR + 'c',
    'streaming multiple clears'
  );

  // Test 5: Whitespace after clear preserved
  verifyChunkIndependence(
    'before\x1b[2J   after',
    'before' + CLEAR_SEPARATOR + '   after',
    'streaming whitespace after clear'
  );

  // Test 6: Whitespace between clears discarded
  verifyChunkIndependence(
    'a\x1b[2J   \x1b[2Jb',
    'a' + CLEAR_SEPARATOR + 'b',
    'streaming whitespace between clears'
  );

  // Test 7: With header/footer
  const fullInput = 'Script started on Wed Dec 31 12:10:34 2025\nCommand: bash\nhello\x1b[2Jworld\n\nScript done on Wed Dec 31 12:11:22 2025';
  verifyChunkIndependence(
    fullInput,
    'hello' + CLEAR_SEPARATOR + 'world',
    'streaming with header/footer'
  );

  // Test 8: No clears
  verifyChunkIndependence(
    'just some content',
    'just some content',
    'streaming no clears'
  );

  console.log('\nSelf-test complete.');
}
