#!/usr/bin/env node

import { chromium } from 'playwright';
import { fileURLToPath } from 'url';
import { dirname, resolve } from 'path';
import fs from 'fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

async function convertHtmlToPdf(htmlPath, outputPath, format = 'A4-landscape', scale = '1.0') {
  if (!htmlPath) {
    console.error('Usage: node index.js <html-file> [output-pdf] [format] [scale]');
    console.error('Formats: A4-landscape (default), A4, A3, A3-landscape, A2, Letter, Tabloid');
    console.error('Scale: 0.1-1.0 (default: 1.0) - scales content to fit page if needed');
    process.exit(1);
  }

  // Resolve to absolute paths
  const inputPath = resolve(htmlPath);
  const outPath = outputPath ? resolve(outputPath) : inputPath.replace(/\.html$/, '.pdf');

  // Check input file exists
  if (!fs.existsSync(inputPath)) {
    console.error(`Error: File not found: ${inputPath}`);
    process.exit(1);
  }

  console.error(`Converting: ${inputPath}`);
  console.error(`Output:     ${outPath}`);
  console.error(`Format:     ${format}`);

  const startTime = performance.now();

  let browser;
  try {
    // Launch browser
    browser = await chromium.launch();
    const page = await browser.newPage();

    // Load HTML file
    await page.goto(`file://${inputPath}`, { waitUntil: 'networkidle' });

    // Wait for xterm.js to render
    await page.waitForLoadState('networkidle');

    // Inject print CSS to scale content to fit page width
    await page.addStyleTag({
      content: `
        @media print {
          body {
            width: 100% !important;
            overflow: visible !important;
          }
          #terminal {
            width: 100% !important;
            transform: none !important;
            margin: 0 !important;
            padding: 0 !important;
          }
          .xterm {
            width: 100% !important;
            height: auto !important;
            transform-origin: top left;
          }
          #footer {
            width: 100% !important;
          }
        }
      `,
    });

    // Build PDF options
    const pdfOptions = {
      path: outPath,
      margin: {
        top: '0.25in',
        right: '0.25in',
        bottom: '0.25in',
        left: '0.25in',
      },
      scale: parseFloat(scale),
    };

    // Set format/dimensions
    if (format.includes('landscape')) {
      const baseFormat = format.replace('-landscape', '');
      pdfOptions.format = baseFormat;
      pdfOptions.landscape = true;
    } else {
      pdfOptions.format = format;
    }

    // Export as PDF
    await page.pdf(pdfOptions);

    const endTime = performance.now();
    const duration = ((endTime - startTime) / 1000).toFixed(2);
    const stats = fs.statSync(outPath);
    const fileSizeMB = (stats.size / 1024 / 1024).toFixed(2);

    console.error(`✓ PDF generated in ${duration}s (${fileSizeMB}MB)`);
    console.log(outPath);
  } catch (error) {
    console.error(`Error: ${error.message}`);
    process.exit(1);
  } finally {
    if (browser) {
      await browser.close();
    }
  }
}

const [, , ...args] = process.argv;

// Handle help flags
if (args.length === 0 || args.includes('-h') || args.includes('--help')) {
  console.error('Usage: to-pdf <html-file> [output-pdf] [format] [scale]');
  console.error('Formats: A4-landscape (default), A4, A3, A3-landscape, A2, Letter, Tabloid');
  console.error('Scale: 0.1-1.0 (default: 1.0) - scales content to fit page if needed');
  console.error('');
  console.error('Examples:');
  console.error('  to-pdf session.log.html');
  console.error('  to-pdf session.log.html output.pdf A4-landscape');
  console.error('  to-pdf session.log.html output.pdf A4-landscape 0.8');
  process.exit(args.length === 0 ? 1 : 0);
}

const [htmlPath, outputPath, format, scale] = args;
convertHtmlToPdf(htmlPath, outputPath, format, scale);
