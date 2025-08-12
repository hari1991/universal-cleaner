# Universal Cleaner Makefile

.PHONY: build clean run test install deps

# Default target
all: build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build for current platform
build: deps
	go build -o universal-cleaner

# Build for all platforms
build-all: build-macos build-windows build-linux

# Build for macOS
build-macos: deps
	GOOS=darwin GOARCH=amd64 go build -o dist/universal-cleaner-macos-amd64
	GOOS=darwin GOARCH=arm64 go build -o dist/universal-cleaner-macos-arm64

# Build for Windows
build-windows: deps
	GOOS=windows GOARCH=amd64 go build -o dist/universal-cleaner-windows-amd64.exe
	GOOS=windows GOARCH=386 go build -o dist/universal-cleaner-windows-386.exe

# Build for Linux
build-linux: deps
	GOOS=linux GOARCH=amd64 go build -o dist/universal-cleaner-linux-amd64
	GOOS=linux GOARCH=386 go build -o dist/universal-cleaner-linux-386
	GOOS=linux GOARCH=arm64 go build -o dist/universal-cleaner-linux-arm64

# Run the application
run: build
	./universal-cleaner

# Test the application
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f universal-cleaner
	rm -rf dist/

# Install to system
install: build
	sudo cp universal-cleaner /usr/local/bin/

# Create distribution directory
dist:
	mkdir -p dist
