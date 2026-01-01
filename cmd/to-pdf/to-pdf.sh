#!/bin/bash
# Wrapper script for to-pdf Node.js tool
# This file is a template - it gets replaced with the actual REPO_DIR during install

# REPO_DIR will be substituted during make install-pdf-tool
REPO_DIR="@REPO_DIR@"

set -e

if [ ! -d "$REPO_DIR/cmd/to-pdf" ]; then
  echo "Error: Cannot find record-tui at: $REPO_DIR" >&2
  echo "Please run: make install-pdf-tool" >&2
  exit 1
fi

if [ ! -d "$REPO_DIR/cmd/to-pdf/node_modules" ]; then
  echo "Error: to-pdf dependencies not installed" >&2
  echo "Run: cd $REPO_DIR && make install-pdf-tool" >&2
  exit 1
fi

# Check Node.js version (Playwright requires Node.js 18+)
NODE_VERSION=$(node -v | sed 's/v//' | cut -d. -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
  echo "Error: Node.js 18 or higher is required for PDF export." >&2
  echo "Current version: $(node -v)" >&2
  echo "Please upgrade Node.js: https://nodejs.org/" >&2
  exit 1
fi

# Execute the Node.js tool
exec node "$REPO_DIR/cmd/to-pdf/index.js" "$@"
