# Universal Cleaner

[![CI](https://github.com/hari1991/universal-cleaner/workflows/CI/badge.svg)](https://github.com/hari1991/universal-cleaner/actions/workflows/ci.yml)
[![Release](https://github.com/hari1991/universal-cleaner/workflows/Release/badge.svg)](https://github.com/hari1991/universal-cleaner/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/hari1991/universal-cleaner)](https://github.com/hari1991/universal-cleaner/blob/main/go.mod)
[![Latest Release](https://img.shields.io/github/v/release/hari1991/universal-cleaner)](https://github.com/hari1991/universal-cleaner/releases/latest)

A cross-platform GUI application built in Go that helps you clean up build artifacts, dependencies, and temporary files from your development projects.

## Features

- **Cross-platform**: Works on Windows, macOS, and Linux with native installers
- **GUI Interface**: Easy-to-use graphical interface built with Fyne
- **Disk Usage Display**: Shows total/used/free disk space for the selected volume
- **Smart Scanning**: Project-aware heuristics reduce false positives (e.g. `target/` only matched when `Cargo.toml` exists nearby)
- **Safe Deletion**: Move to Trash (recoverable) or permanent delete
- **Parallel Size Calculation**: Fast scan with 8-worker pool for size computation
- **Export Results**: Save scan results to CSV or JSON
- **Recent Folders**: Quick access to previously scanned directories
- **Drag & Drop**: Drag a folder onto the window to scan it
- **Persistent Settings**: Categories, exclusions, theme, and window geometry saved across sessions
- **Activity Log**: All actions logged to disk and displayed in the UI
- **Multiple Language Support**:
  - **Node.js**: `node_modules`, `.npm`, `.yarn`, debug logs
  - **Java**: `target`, `build`, `.gradle` (project-aware via `pom.xml`/`build.gradle`)
  - **Python**: `__pycache__`, `.pytest_cache`, `*.pyc`, virtual environments
  - **Rust**: `target` (project-aware via `Cargo.toml`)
  - **Go**: `bin`, `pkg` (project-aware via `go.mod`)
  - **C/C++**: cmake directories, object files
  - **Docker**: `.docker`, override files
  - **IDE**: `.vscode`, `.idea`, swap files, system files
  - **Build**: `dist`, `out`, `output`, cache directories

## Installation

### Download Pre-built Installers

Go to the [Releases page](https://github.com/hari1991/universal-cleaner/releases/latest) and download the installer for your platform:

| Platform | File | How to Install |
|----------|------|----------------|
| macOS (Apple Silicon) | `UniversalCleaner-macos-arm64.dmg` | Open `.dmg`, drag to Applications |
| macOS (Intel) | `UniversalCleaner-macos-x86_64.dmg` | Open `.dmg`, drag to Applications |
| Windows (64-bit) | `UniversalCleaner-windows-x86_64.msi` | Double-click `.msi` to install |
| Windows (32-bit) | `UniversalCleaner-windows-x86.msi` | Double-click `.msi` to install |
| Linux (64-bit) | `universal-cleaner-linux-x86_64.deb` | `sudo dpkg -i *.deb` |
| Linux (64-bit) | `universal-cleaner-linux-x86_64.tar.xz` | Extract and run |
| Linux (ARM64) | `universal-cleaner-linux-arm64.deb` | `sudo dpkg -i *.deb` |
| Linux (ARM64) | `universal-cleaner-linux-arm64.tar.xz` | Extract and run |

### Verify Download Integrity

Each release includes SHA256 checksums:
```bash
sha256sum -c checksums.txt
```

### Build from Source

**Prerequisites**: Go 1.23+ and platform build tools (see below)

```bash
git clone https://github.com/hari1991/universal-cleaner.git
cd universal-cleaner
go build -o universal-cleaner
./universal-cleaner
```

Or use the startup script:
```bash
./startup.sh          # build + run GUI
./startup.sh --cli    # run CLI mode
./startup.sh --build  # build only
```

### Quick Start Script

```bash
# Build and launch the GUI
./startup.sh

# Build the CLI and scan current directory (dry run)
./startup.sh --cli --path . --dry-run

# Stop any running instance
./stop.sh
```

## Packaging (Building Installers)

The project uses [Fyne's packaging tool](https://docs.fyne.io/started/packaging) to create native installers.

### Prerequisites by Platform

| Platform | Requirements |
|----------|-------------|
| **macOS** | Xcode Command Line Tools, `fyne` CLI |
| **Windows** | WiX Toolset 3.x, `fyne` CLI |
| **Linux** | `libgl1-mesa-dev`, `xorg-dev`, `gcc`, `pkg-config`, `fyne` CLI |

Install the fyne CLI:
```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
```

### Package for Current Platform

```bash
make package          # auto-detect host OS
```

### Package for Specific Platform (on native OS)

```bash
make package-macos    # produces .app + .dmg (run on macOS)
make package-windows  # produces .msi (run on Windows)
make package-linux    # produces .tar.xz + .deb (run on Linux)
```

### Cross-compile Raw Binaries (no installer)

```bash
make build-all        # builds for all OS/arch combos into dist/
```

### Docker Build (Linux packages)

```bash
# Build .tar.xz and .deb in a container, output to ./dist/
docker compose -f docker/docker-compose.yml run --rm builder
```

## Usage

### GUI

1. **Select Folder**: Click "Select Folder" or drag a folder onto the window
2. **Scan**: Click "Scan" to find cleanable items (progress bar shows scan activity)
3. **Review**: Items appear in a sortable table with path, type, size, file count, and modification date
4. **Filter**: Use the search box or category dropdown to narrow results
5. **Select**: Click the checkbox column to select items, or use Select All/None
6. **Dry Run**: Preview what would be deleted before committing
7. **Clean**: Click "Clean" and confirm — items go to Trash by default
8. **Export**: Save results to CSV or JSON for reporting

### CLI

```bash
universal-cleaner --path /path/to/project --dry-run
universal-cleaner --path /path/to/project --yes
universal-cleaner --path . --include-risky --no-trash --verbose
```

**CLI Flags**:
- `--path`: Directory to scan (default: current directory)
- `--dry-run`: Show what would be deleted without deleting
- `--yes`: Skip confirmation prompts
- `--verbose`: Show detailed output
- `--include-risky`: Include risky targets (Cargo.lock, vendor, IDE configs)
- `--no-trash`: Delete permanently instead of moving to Trash

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl + R` | Rescan current folder |
| `Cmd/Ctrl + F` | Focus the search/filter box |

## Project Structure

```
universal-cleaner/
├── main.go              # GUI entry point and layout
├── cli.go               # CLI entry point
├── theme.go             # Custom Fyne theme
├── ui_table.go          # Results table widget
├── ui_settings.go       # Settings dialog
├── resource.go          # Bundled app icon (auto-generated)
├── Icon.png             # App icon source
├── internal/
│   └── core/
│       ├── types.go     # Shared types (CleanableItem, Settings, etc.)
│       ├── targets.go   # Category definitions and target matching
│       ├── scanner.go   # Filesystem scanner with parallel sizing
│       ├── matcher.go   # Name/wildcard matching logic
│       ├── size.go      # Size calculation and formatting
│       ├── trash.go     # Cross-platform trash (macOS/Linux/Windows)
│       ├── disk.go      # Disk usage (cross-platform)
│       ├── git.go       # Git-awareness helpers
│       ├── open.go      # Reveal in Finder/Explorer
│       ├── export.go    # CSV/JSON export
│       └── settings.go  # Settings persistence
├── docker/
│   ├── Dockerfile       # Linux build environment
│   └── docker-compose.yml
├── .github/workflows/
│   ├── ci.yml           # CI: build + test on all platforms
│   └── release.yml      # Release: native installers on tag push
├── Makefile             # Build/package targets
├── startup.sh           # Build + run script
└── stop.sh              # Stop running instance
```

## Releasing

Releases are triggered by pushing a git tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The GitHub Actions release workflow will:
1. Build native installers on macOS, Windows, and Linux runners
2. Generate SHA256 checksums
3. Create a GitHub Release with all artifacts attached

## Contributing

Feel free to submit issues and enhancement requests!

## License

This project is open source and available under the MIT License.
