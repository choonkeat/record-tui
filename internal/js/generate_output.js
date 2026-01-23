#!/usr/bin/env node
/**
 * Node.js test harness for generating .js.output files.
 * Reads ./recordings/session*.log files, passes them through cleanSessionContent,
 * and writes output to ./recordings-output/*.js.output.
 */

const fs = require('fs');
const path = require('path');
const { cleanSessionContent } = require('./cleaner.js');

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

    // Read file as binary (latin1) to preserve raw bytes like Go's string([]byte)
    // Note: utf8 encoding would replace invalid sequences with U+FFFD replacement char
    let content;
    try {
      content = fs.readFileSync(realPath, 'latin1');
    } catch (err) {
      console.error(`Failed to read file "${realPath}": ${err.message}`);
      continue;
    }

    // Run cleaning pipeline
    const cleanedContent = cleanSessionContent(content);

    // Write output file with input length header for live recording detection
    // Format: "{input_length} bytes\n{cleaned_content}"
    // Write header and content separately to avoid any concatenation issues
    const outputPath = path.join(OUTPUT_DIR, file + '.js.output');
    const header = `${content.length} bytes\n`;
    try {
      fs.writeFileSync(outputPath, header, 'utf8');  // header is ASCII, utf8 is fine
      fs.appendFileSync(outputPath, cleanedContent, 'latin1');  // content as raw bytes
    } catch (err) {
      console.error(`Failed to write output file "${outputPath}": ${err.message}`);
      continue;
    }

    console.log(`Processed ${file} -> ${path.basename(outputPath)} (${content.length} bytes -> ${cleanedContent.length} bytes)`);
  }
}

main();
