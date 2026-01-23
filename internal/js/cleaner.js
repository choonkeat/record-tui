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
 * Neutralize clear sequences by replacing them with visible separator.
 * Matches Go's clear.go:22-64 exactly (sophisticated logic).
 * Only adds separator if there's non-empty content before AND after the clear.
 */
function neutralizeClearSequences(text) {
  // Find all clear sequences
  const matches = [];
  let match;
  // Reset lastIndex since we're reusing the regex
  clearPattern.lastIndex = 0;
  while ((match = clearPattern.exec(text)) !== null) {
    matches.push([match.index, match.index + match[0].length]);
  }

  if (matches.length === 0) {
    return text;
  }

  let result = '';
  let lastEnd = 0;

  for (const [start, end] of matches) {
    // Get content before this clear
    const before = text.slice(lastEnd, start);

    // Only add separator if there's non-empty content before
    if (before.trim() !== '') {
      result += before;

      // Check if there's content after this clear
      const remaining = text.slice(end);
      if (remaining.trim() !== '') {
        result += CLEAR_SEPARATOR;
      }
    }

    lastEnd = end;
  }

  // Add remaining content after the last clear
  const remaining = text.slice(lastEnd);
  if (remaining.trim() !== '') {
    // If we haven't written anything yet (clears were at start), just write content
    // (This matches Go's logic at clear.go:56-60)
    result += remaining;
  }

  return result;
}

/**
 * Clean session content by stripping metadata and neutralizing clear sequences.
 * This is the main entry point that matches the Go pipeline.
 * Pipeline: text -> stripHeader -> stripFooter -> neutralizeClearSequences
 */
function cleanSessionContent(text) {
  // Step 1: Strip header
  let content = stripHeader(text);
  // Step 2: Strip footer
  content = stripFooter(content);
  // Step 3: Neutralize clear sequences
  content = neutralizeClearSequences(content);
  return content;
}

/**
 * Create a streaming cleaner for processing chunked data.
 * Produces identical output to cleanSessionContent() but processes data incrementally.
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
    neutralizeClearSequences,
    cleanSessionContent,
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

  // Test neutralizeClearSequences
  const clearTest = 'before\x1b[2Jafter';
  const clearResult = neutralizeClearSequences(clearTest);
  const expectedClear = 'before' + CLEAR_SEPARATOR + 'after';
  console.log('neutralizeClearSequences test:', clearResult === expectedClear ? 'PASS' : 'FAIL');

  // Test clear at start (no separator should be added)
  const clearStartTest = '\x1b[2Jafter';
  const clearStartResult = neutralizeClearSequences(clearStartTest);
  console.log('neutralizeClearSequences (start) test:', clearStartResult === 'after' ? 'PASS' : 'FAIL');

  // Test clear at end (no separator should be added)
  const clearEndTest = 'before\x1b[2J';
  const clearEndResult = neutralizeClearSequences(clearEndTest);
  console.log('neutralizeClearSequences (end) test:', clearEndResult === 'before' ? 'PASS' : 'FAIL');

  // === Streaming Tests ===
  console.log('\nRunning streaming cleaner tests...');

  // Helper: process with streaming at given chunk size
  function testStreaming(input, chunkSize, testName) {
    const chunks = [];
    const cleaner = createStreamingCleaner((c) => chunks.push(c));
    for (let i = 0; i < input.length; i += chunkSize) {
      cleaner.write(input.slice(i, i + chunkSize));
    }
    cleaner.end();
    return chunks.join('');
  }

  // Helper: compare streaming vs batch
  function compareStreamingVsBatch(input, chunkSize, testName) {
    const batchResult = cleanSessionContent(input);
    const streamResult = testStreaming(input, chunkSize, testName);
    const pass = batchResult === streamResult;
    console.log(`${testName}: ${pass ? 'PASS' : 'FAIL'}`);
    if (!pass) {
      console.log(`  Expected (${batchResult.length} bytes): ${JSON.stringify(batchResult.slice(0, 100))}...`);
      console.log(`  Got (${streamResult.length} bytes): ${JSON.stringify(streamResult.slice(0, 100))}...`);
    }
    return pass;
  }

  // Test 1: Basic streaming matches batch
  compareStreamingVsBatch('before\x1b[2Jafter', 1024, 'streaming basic');

  // Test 2: Byte-by-byte (stress test)
  compareStreamingVsBatch('before\x1b[2Jafter', 1, 'streaming byte-by-byte');

  // Test 3: Clear at chunk boundary
  compareStreamingVsBatch('content\x1b[2Jmore', 7, 'streaming clear at boundary');

  // Test 4: Multiple clears
  compareStreamingVsBatch('a\x1b[2Jb\x1b[2Jc', 1, 'streaming multiple clears');

  // Test 5: Clear at start (no separator)
  compareStreamingVsBatch('\x1b[2Jcontent', 1, 'streaming clear at start');

  // Test 6: Clear at end (no separator)
  compareStreamingVsBatch('content\x1b[2J', 1, 'streaming clear at end');

  // Test 7: Whitespace after clear
  compareStreamingVsBatch('before\x1b[2J   after', 5, 'streaming whitespace after clear');

  // Test 8: Only whitespace between clears
  compareStreamingVsBatch('a\x1b[2J   \x1b[2Jb', 3, 'streaming whitespace between clears');

  // Test 9: With header/footer
  const fullInput = 'Script started on Wed Dec 31 12:10:34 2025\nCommand: bash\nhello\x1b[2Jworld\n\nScript done on Wed Dec 31 12:11:22 2025';
  compareStreamingVsBatch(fullInput, 10, 'streaming with header/footer');

  // Test 10: Large chunks (approaching batch)
  compareStreamingVsBatch('before\x1b[2Jafter', 10000, 'streaming large chunks');

  console.log('\nSelf-test complete.');
}
