#!/usr/bin/env node
/**
 * Node.js test harness for generating .js.output files.
 * Reads ./recordings/session*.log files, passes them through cleanSessionContent,
 * and writes output to ./recordings-output/*.js.output.
 */

const fs = require('fs');
const path = require('path');
const { createStreamingCleaner } = require('./cleaner.js');

// Seeded PRNG for reproducible random chunk sizes
// Allow reproducible runs via SEED env var
const seed = process.env.SEED ? parseInt(process.env.SEED) : Date.now();
console.log(`Random seed: ${seed} (reproduce with: SEED=${seed} make test)`);

let rngState = seed;
function random() {
  rngState = (rngState * 1103515245 + 12345) & 0x7fffffff;
  return rngState / 0x7fffffff;
}

function getRandomChunkSize() {
  return Math.floor(random() * (16384 - 64)) + 64;  // 64 bytes to 16KB
}

/**
 * Process content using streaming cleaner with random chunk sizes.
 * This proves the streaming API produces identical output regardless of chunking.
 */
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

const RECORDINGS_DIR = './recordings';
const OUTPUT_DIR = './recordings-output';

function main() {
  // Check if recordings directory exists
  if (!fs.existsSync(RECORDINGS_DIR)) {
    console.log(`WARNING: recordings directory "${RECORDINGS_DIR}" does not exist, skipping`);
    process.exit(0);
  }

  // Create output directory if it doesn't exist
  if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  }

  // Find all session*.log files
  const files = fs.readdirSync(RECORDINGS_DIR)
    .filter(f => f.startsWith('session') && f.endsWith('.log'))
    .sort(); // Sort for deterministic ordering

  if (files.length === 0) {
    console.log(`WARNING: no session*.log files found in "${RECORDINGS_DIR}", skipping`);
    process.exit(0);
  }

  console.log(`Found ${files.length} session*.log files to process`);

  for (const file of files) {
    const inputPath = path.join(RECORDINGS_DIR, file);

    // Follow symlinks by using realpath
    let realPath;
    try {
      realPath = fs.realpathSync(inputPath);
    } catch (err) {
      console.error(`Failed to resolve symlink "${inputPath}": ${err.message}`);
      continue;
    }

    // Read file as UTF-8 to match Go's string handling (Go strings are UTF-8)
    // Session logs are text output from terminals, which use UTF-8
    let content;
    try {
      content = fs.readFileSync(realPath, 'utf8');
    } catch (err) {
      console.error(`Failed to read file "${realPath}": ${err.message}`);
      continue;
    }

    // Run cleaning pipeline using streaming with random chunks
    const cleanedContent = processWithStreaming(content);

    // Write output file with input length header for live recording detection
    // Format: "{input_length} bytes\n{cleaned_content}"
    // Write header and content separately to avoid any concatenation issues
    const outputPath = path.join(OUTPUT_DIR, file + '.js.output');
    // Use Buffer.byteLength to get byte count (not character count) to match Go's len([]byte)
    const inputByteLength = Buffer.byteLength(content, 'utf8');
    const header = `${inputByteLength} bytes\n`;
    try {
      fs.writeFileSync(outputPath, header, 'utf8');
      fs.appendFileSync(outputPath, cleanedContent, 'utf8');
    } catch (err) {
      console.error(`Failed to write output file "${outputPath}": ${err.message}`);
      continue;
    }

    console.log(`Processed ${file} -> ${path.basename(outputPath)} (${content.length} bytes -> ${cleanedContent.length} bytes)`);
  }
}

main();
