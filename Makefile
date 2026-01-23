.PHONY: build build-all clean clean-compare-output test test-go test-js compare-output install info install-pdf-tool

# Build record-tui binary
build:
	go build -o bin/record-tui ./cmd/record-tui

# Build for multiple platforms
build-all: build
	GOOS=darwin GOARCH=arm64 go build -o bin/record-tui-darwin-arm64 ./cmd/record-tui
	GOOS=darwin GOARCH=amd64 go build -o bin/record-tui-darwin-amd64 ./cmd/record-tui
	GOOS=linux GOARCH=amd64 go build -o bin/record-tui-linux-amd64 ./cmd/record-tui

# Clean comparison output files (ensures fresh comparison)
clean-compare-output:
	rm -rf ./recordings-output/

# Run all tests (clean, Go tests, JS output generation, then compare)
test: clean-compare-output test-go test-js compare-output

# Run Go tests (generates .go.output files in recordings-output/)
test-go:
	go test ./internal/... -v

# Run JS output generation (generates .js.output files in recordings-output/)
test-js:
	node internal/js/generate_output.js

# Compare Go and JS outputs (fails if any differ)
compare-output:
	go test ./internal/session -run TestCompareGoAndJsOutput -v

# Install binary to ~/bin
install: build
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
