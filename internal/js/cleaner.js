/**
 * Node.js wrapper for cleaner-core.js.
 * Re-exports core functions and adds self-tests.
 *
 * The core logic lives in cleaner-core.js which is also embedded
 * in template_streaming.go for browser use.
 */

const {
  CLEAR_SEPARATOR,
  clearPattern,
  stripHeader,
  stripFooter,
  createStreamingCleaner
} = require('./cleaner-core.js');

// Re-export for consumers
module.exports = {
  CLEAR_SEPARATOR,
  clearPattern,
  stripHeader,
  stripFooter,
  createStreamingCleaner
};

// Simple self-test when run directly with node
if (require.main === module) {
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
