.PHONY: build build-all clean test install info

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
	cp bin/record-tui ~/bin/record-tui
	chmod +x ~/bin/record-tui
	@echo "✓ Installed to ~/bin/record-tui"

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Display binary info
info: build
	@echo "Binary size:"
	@ls -lh bin/record-tui
