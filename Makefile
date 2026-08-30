# Universal Cleaner Makefile
# Build and package the app for installation on macOS, Windows, and Linux.

APP_NAME := universal-cleaner
# Fyne requires semver x.y.z for appVersion; fall back to 0.1.0 if no tag.
VERSION  := $(shell git describe --tags --abbrev=0 2>/dev/null || echo 0.1.0)
DIST     := dist
ICON     := Icon.png

.PHONY: all build clean run test vet fmt deps \
        package package-macos package-windows package-linux \
        build-macos build-windows build-linux \
        install

# Default target
all: build

# ---- Dependencies ----

deps:
	go mod download

# ---- Build ----

# Build for current platform (raw binary)
build: deps
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(APP_NAME) .

# Run the application
run: build
	./$(APP_NAME)

# Run tests
test:
	go test -v ./...

# Run go vet
vet:
	go vet ./...

# Format code
fmt:
	gofmt -s -w .

# ---- Cross-compile raw binaries (no GUI packaging) ----

build-macos: deps
	mkdir -p $(DIST)
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(DIST)/$(APP_NAME)-macos-arm64
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(DIST)/$(APP_NAME)-macos-amd64

build-windows: deps
	mkdir -p $(DIST)
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(DIST)/$(APP_NAME)-windows-amd64.exe
	GOOS=windows GOARCH=386 go build -ldflags="-s -w" -o $(DIST)/$(APP_NAME)-windows-386.exe

build-linux: deps
	mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(DIST)/$(APP_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(DIST)/$(APP_NAME)-linux-arm64

build-all: build-macos build-windows build-linux

# ---- Fyne packaging (native installers) ----
#
# These targets use `fyne package` to produce platform-native installers:
#   macOS:   .app bundle and .dmg disk image
#   Windows: .msi installer (requires WiX Toolset on Windows)
#   Linux:   .tar.xz portable archive and .deb package
#
# Prerequisites:
#   - fyne CLI:  go install fyne.io/fyne/v2/cmd/fyne@latest
#   - macOS:     Xcode command-line tools
#   - Windows:   WiX Toolset 3.x (for .msi)
#   - Linux:     libgl1-mesa-dev, xorg-dev, gcc, pkg-config

# Package for the current host platform (auto-detects OS)
package: deps
	fyne package -icon $(ICON) -name "$(APP_NAME)" -appVersion $(VERSION)

# Package for macOS (.app / .dmg) — run on macOS
package-macos: deps
	fyne package -os darwin -icon $(ICON) -name "UniversalCleaner" -appVersion $(VERSION) -appID "com.universalcleaner.app"
	mkdir -p $(DIST)
	# Create .dmg disk image from the .app bundle
	if [ -d "UniversalCleaner.app" ]; then \
	  hdiutil create -volname "UniversalCleaner" -srcfolder "UniversalCleaner.app" -ov -format UDZO "$(DIST)/UniversalCleaner-$(VERSION).dmg"; \
	  zip -r "$(DIST)/UniversalCleaner-$(VERSION).app.zip" "UniversalCleaner.app"; \
	fi
	rm -rf UniversalCleaner.app

# Package for Windows (.msi) — run on Windows or cross-compile host
package-windows: deps
	fyne package -os windows -icon $(ICON) -name "$(APP_NAME)" -appVersion $(VERSION)
	mkdir -p $(DIST)
	mv *.msi $(DIST)/ 2>/dev/null || true
	mv *.exe $(DIST)/ 2>/dev/null || true

# Package for Linux (.tar.xz / .deb) — run on Linux
package-linux: deps
	fyne package -os linux -icon $(ICON) -name "$(APP_NAME)" -appVersion $(VERSION)
	mkdir -p $(DIST)
	mv *.tar.xz $(DIST)/ 2>/dev/null || true
	mv *.deb $(DIST)/ 2>/dev/null || true

# Package for all platforms (requires running on each OS or using CI)
package-all: package-macos package-windows package-linux

# ---- Install ----

# Install the raw binary to /usr/local/bin (Linux/macOS)
install: build
	cp $(APP_NAME) /usr/local/bin/

# ---- Clean ----

clean:
	rm -f $(APP_NAME)
	rm -rf $(DIST)
	rm -f *.app *.dmg *.msi *.tar.xz *.deb
