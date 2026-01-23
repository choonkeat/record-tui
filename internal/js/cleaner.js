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

// Export for Node.js (CommonJS)
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    CLEAR_SEPARATOR,
    clearPattern,
    stripHeader,
    stripFooter,
    neutralizeClearSequences,
    cleanSessionContent
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

  console.log('Self-test complete.');
}
