/**
 * Core session log cleaner functions.
 * This file is the single source of truth used by both:
 * - Node.js test harness (via cleaner.js)
 * - Browser streaming HTML (embedded by Go via go:embed)
 *
 * Environment-agnostic: no Node.js or browser-specific code.
 */

// Clear sequence separator - must match Go's ClearSeparator in clear.go:9
// Using Unicode escape \u2500 for box-drawing character '─' (U+2500)
// Both Go strings and JS strings are UTF-8/UTF-16, so this works correctly in:
// - Node.js test harness (reading files as UTF-8)
// - Browser streaming (TextDecoder with UTF-8)
const CLEAR_SEPARATOR = '\n\n\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500 terminal cleared \u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\n\n';

// Clear sequence pattern - must match Go's clearPattern in clear.go
// Matches: \x1b[H\x1b[2J, \x1b[H\x1b[3J, \x1b[2J\x1b[H, \x1b[3J\x1b[H, \x1b[2J, \x1b[3J
// Also: \x1b[1;1H\x1b[J, \x1b[H\x1b[J, \x1b[1;1H\x1b[0J (home + erase to end = effective clear)
const clearPattern = /\x1b\[(?:1;1)?H\x1b\[(?:0?J|[23]J)|\x1b\[[23]J\x1b\[H|\x1b\[[23]J/g;

// Alt screen separator - must match Go's AltScreenSeparator in clear.go
const ALT_SCREEN_SEPARATOR = '\n\n\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500 alternate screen \u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\n\n';

// Alt screen pattern - must match Go's altScreenPattern in clear.go
// Matches: \x1b[?1049h, \x1b[?1049l, \x1b[?47h, \x1b[?47l, \x1b[?1047h, \x1b[?1047l
const altScreenPattern = /\x1b\[\?(1049|47|1047)([hl])/g;

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
  let hasFooterMarker = false;
  for (let i = lines.length - 1; i >= 0; i--) {
    const line = lines[i];
    // Check if this line is a footer marker (must start with the marker text)
    if (line.startsWith('Saving session') ||
        line.startsWith('Command exit status') ||
        line.startsWith('Script done on')) {
      hasFooterMarker = true;
      footerStartIndex = i;
    } else if (hasFooterMarker && line.trim() === '') {
      // Only treat empty lines as footer if we already found a footer marker
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

  // Alt screen state - when inside alt screen, discard all content
  let inAltScreen = false;
  let altScreenHadContentBefore = false;

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
   * Process text for alternate screen sequences, respecting streaming state.
   * Content between enter and leave is discarded (TUI cursor-positioned content
   * would corrupt the main screen). A separator is inserted at the leave point
   * when there's content on both sides.
   */
  function processForAltScreen(text) {
    if (!text) return '';

    // If we're inside alt screen, check for leave sequence
    if (inAltScreen) {
      altScreenPattern.lastIndex = 0;
      const matches = [];
      let m;
      while ((m = altScreenPattern.exec(text)) !== null) {
        if (m[2] === 'l') {
          matches.push({ start: m.index, end: m.index + m[0].length });
        }
      }

      if (matches.length === 0) {
        // Still inside alt screen — discard everything
        return '';
      }

      // Found leave — discard everything before it, emit separator + rest
      const firstLeave = matches[0];
      inAltScreen = false;
      const remaining = text.slice(firstLeave.end);

      // Recursively process remaining (might have more enter/leave pairs)
      const processed = processForAltScreen(remaining);
      if (altScreenHadContentBefore && processed.trim() !== '') {
        return ALT_SCREEN_SEPARATOR + processed;
      }
      return processed;
    }

    // Not inside alt screen — look for enter sequence
    altScreenPattern.lastIndex = 0;
    const matches = [];
    let m;
    while ((m = altScreenPattern.exec(text)) !== null) {
      if (m[2] === 'h') {
        matches.push({ start: m.index, end: m.index + m[0].length });
        break; // Only need the first enter
      }
    }

    if (matches.length === 0) {
      // No alt screen sequences
      if (text.trim() !== '') {
        altScreenHadContentBefore = true;
      }
      return text;
    }

    // Found enter — keep content before, discard after until leave
    const before = text.slice(0, matches[0].start);
    if (before.trim() !== '') {
      altScreenHadContentBefore = true;
    }
    inAltScreen = true;

    // Process remaining after enter (might contain leave in same chunk)
    const afterEnter = text.slice(matches[0].end);
    const processed = processForAltScreen(afterEnter);
    return before + processed;
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
      let processed = processForClears(toEmit);
      processed = processForAltScreen(processed);
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

    // Process for clears and alt screen, then emit
    let processed = processForClears(text);
    processed = processForAltScreen(processed);
    if (processed) onOutput(processed);
  }

  return { write, end };
}

// Export for Node.js (CommonJS) - ignored in browser
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    CLEAR_SEPARATOR,
    ALT_SCREEN_SEPARATOR,
    clearPattern,
    altScreenPattern,
    stripHeader,
    stripFooter,
    createStreamingCleaner
  };
}
