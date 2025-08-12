# Technology Stack

## Language & Runtime
- **Go 1.23.4+** - Primary language
- Cross-platform compilation support

## Key Dependencies
- **Fyne v2.6.2** - GUI framework for cross-platform desktop applications
- Standard Go libraries for file system operations

## Build System
- **Makefile** - Primary build automation
- **Go modules** - Dependency management

## Common Commands

### Development
```bash
# Install dependencies
make deps

# Build for current platform
make build

# Run application
make run
./universal-cleaner

# Run CLI mode
./universal-cleaner --help
./universal-cleaner --path /some/path --dry-run
```

### Building & Distribution
```bash
# Build for all platforms
make build-all

# Build for specific platforms
make build-macos
make build-windows  
make build-linux

# Clean build artifacts
make clean
```

### Testing
```bash
make test
```

## Architecture Patterns
- **Single binary deployment** - No external dependencies
- **Dual interface pattern** - GUI and CLI modes in same binary
- **Resource embedding** - Icons and assets bundled into binary
- **Cross-platform file operations** - Using Go's standard library