.PHONY: build build-all clean test install info install-pdf-tool

# Build record-tui binary
build:
	go build -o bin/record-tui ./cmd/record-tui

# Build for multiple platforms
build-all: build
	GOOS=darwin GOARCH=arm64 go build -o bin/record-tui-darwin-arm64 ./cmd/record-tui
	GOOS=darwin GOARCH=amd64 go build -o bin/record-tui-darwin-amd64 ./cmd/record-tui
	GOOS=linux GOARCH=amd64 go build -o bin/record-tui-linux-amd64 ./cmd/record-tui

# Run all tests
test:
	go test ./internal/... -v

# Install binary to ~/bin
install: build
	rm -f ~/bin/record-tui
	cp bin/record-tui ~/bin/record-tui
	chmod +x ~/bin/record-tui
	@echo "✓ Installed to ~/bin/record-tui"

# Install PDF conversion tool dependencies
install-pdf-tool:
	@command -v node >/dev/null 2>&1 || { echo "Error: Node.js is not installed. Please install Node.js 18 or higher."; exit 1; }
	@NODE_VERSION=$$(node -v | sed 's/v//' | cut -d. -f1); \
	if [ "$$NODE_VERSION" -lt 18 ]; then \
		echo "Error: Node.js 18 or higher is required for PDF export."; \
		echo "Current version: $$(node -v)"; \
		echo "Please upgrade Node.js: https://nodejs.org/"; \
		exit 1; \
	fi
	cd cmd/to-pdf && npm install
	cd cmd/to-pdf && npx playwright install chromium
	sed 's|@REPO_DIR@|'$(PWD)'|g' cmd/to-pdf/to-pdf.sh > ~/bin/to-pdf
	chmod +x ~/bin/to-pdf
	@echo "✓ PDF conversion tool installed to ~/bin/to-pdf"

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Display binary info
info: build
	@echo "Binary size:"
	@ls -lh bin/record-tui
