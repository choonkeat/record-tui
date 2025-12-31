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

# Execute the Node.js tool
exec node "$REPO_DIR/cmd/to-pdf/index.js" "$@"
