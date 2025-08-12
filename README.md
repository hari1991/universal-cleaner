# Universal Cleaner

[![CI](https://github.com/USERNAME/REPOSITORY/workflows/CI/badge.svg)](https://github.com/USERNAME/REPOSITORY/actions/workflows/ci.yml)
[![Release](https://github.com/USERNAME/REPOSITORY/workflows/Release/badge.svg)](https://github.com/USERNAME/REPOSITORY/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/USERNAME/REPOSITORY)](https://github.com/USERNAME/REPOSITORY/blob/main/go.mod)
[![Latest Release](https://img.shields.io/github/v/release/USERNAME/REPOSITORY)](https://github.com/USERNAME/REPOSITORY/releases/latest)

A cross-platform GUI application built in Go that helps you clean up build artifacts, dependencies, and temporary files from your development projects.

## Features

- **Cross-platform**: Works on Windows, macOS, and Linux
- **GUI Interface**: Easy-to-use graphical interface built with Fyne
- **Multiple Language Support**: Detects and cleans artifacts from various programming languages and tools:
  - **Node.js**: `node_modules`, `.npm`, `.yarn`, debug logs
  - **Java**: `target`, `build`, `.gradle`, `.m2/repository`
  - **Python**: `__pycache__`, `.pytest_cache`, `*.pyc`, virtual environments
  - **Rust**: `target`, `Cargo.lock`
  - **Go**: `vendor`, `bin`, `pkg`
  - **C/C++**: `build`, cmake directories, object files
  - **Docker**: `.docker`, override files
  - **IDE**: `.vscode`, `.idea`, swap files, system files
  - **Build**: `dist`, `out`, `output`, cache directories

## Installation

### Prerequisites
- Go 1.19 or later

### Building from Source
```bash
git clone <repository-url>
cd universal-cleaner
go mod download
go build -o universal-cleaner
```

### Running
```bash
./universal-cleaner
```

## Usage

1. **Select Folder**: Click "Select Folder" to choose the root directory you want to scan
2. **Scan**: Click "Scan for Cleanable Items" to recursively search for cleanable artifacts
3. **Review**: Review the found items in the list, showing path, type, and size
4. **Select**: Check the items you want to delete
5. **Clean**: Click "Clean Selected Items" and confirm the deletion

## Safety Features

- **Preview before deletion**: Always shows what will be deleted before taking action
- **Confirmation dialog**: Requires explicit confirmation before deleting files
- **Selective deletion**: Choose exactly which items to delete
- **Size calculation**: Shows how much space each item takes up

## Supported Platforms

- macOS
- Windows
- Linux

## CI/CD and Automated Builds

This project uses GitHub Actions for continuous integration and automated releases across multiple platforms.

### Continuous Integration

The CI workflow automatically runs on:
- **Push to main/develop branches**: Validates builds across all supported platforms
- **Pull requests**: Ensures code quality and cross-platform compatibility

**Build Matrix**: The CI system tests builds for:
- **Linux**: amd64, 386, arm64
- **Windows**: amd64, 386  
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)

**Build Status**: 
- ✅ All platforms are automatically tested on every push and pull request
- 🔄 Build status is visible via the CI badge above
- 📊 Detailed build logs available in the [Actions tab](https://github.com/USERNAME/REPOSITORY/actions)

**Quality Checks**:
- Code formatting verification (`go fmt`)
- Dependency verification (`go mod tidy`)
- Test execution with race detection
- Cross-platform build validation

### Automated Releases

Releases are automatically created when you push a git tag:

```bash
# Create and push a new release
git tag v1.0.0
git push origin v1.0.0
```

**Release Process**:
1. **Multi-platform builds**: Automatically builds binaries for all supported platforms
2. **Checksum generation**: Creates SHA256 checksums for security verification
3. **GitHub release**: Creates a release with all binaries and checksums
4. **Release notes**: Auto-generates changelog from commit messages

**Binary Naming Convention**:
- `universal-cleaner-linux-amd64`
- `universal-cleaner-windows-amd64.exe`
- `universal-cleaner-darwin-arm64`

### Manual Building for Different Platforms

For local development and testing:

### macOS
```bash
go build -o universal-cleaner-macos
```

### Windows (cross-compile from macOS/Linux)
```bash
GOOS=windows GOARCH=amd64 go build -o universal-cleaner-windows.exe
```

### Linux (cross-compile from macOS/Windows)
```bash
GOOS=linux GOARCH=amd64 go build -o universal-cleaner-linux
```

### Using the Makefile

The project includes a Makefile for common build tasks:

```bash
# Install dependencies
make deps

# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Clean build artifacts
make clean
```

### Troubleshooting CI/CD Issues

**Common Build Issues**:

1. **Go Version Mismatch**
   - Ensure your local Go version matches the CI version (1.23.4+)
   - Update go.mod if you need a different Go version

2. **Dependency Issues**
   - Run `go mod tidy` to clean up dependencies
   - Clear module cache: `go clean -modcache`

3. **Cross-compilation Failures**
   - Verify CGO is disabled for cross-compilation: `CGO_ENABLED=0`
   - Check for platform-specific code that might not compile

4. **Test Failures**
   - Run tests locally: `go test -v -race ./...`
   - Check for race conditions or platform-specific test issues

5. **Release Workflow Issues**
   - Ensure tag follows semantic versioning (v1.0.0, v1.2.3)
   - Check repository permissions for release creation
   - Verify GITHUB_TOKEN has sufficient permissions

**Debugging Steps**:
1. Check the Actions tab in GitHub for detailed logs
2. Run the same commands locally to reproduce issues
3. Verify all required files are committed and pushed
4. Check for any platform-specific dependencies or code

### Build Monitoring and Notifications

**Status Monitoring**:
- **CI Badge**: Shows current build status for the main branch
- **Release Badge**: Indicates the status of the latest release workflow
- **Go Version Badge**: Displays the Go version used in the project
- **Latest Release Badge**: Shows the most recent release version

**Workflow Notifications**:
- GitHub automatically sends email notifications for failed workflows to repository maintainers
- You can customize notification settings in your GitHub account preferences
- Consider setting up additional monitoring for production releases

**Supported Platforms Documentation**:

| Platform | Architecture | Status | Binary Name |
|----------|-------------|--------|-------------|
| Linux | amd64 | ✅ Supported | `universal-cleaner-linux-amd64` |
| Linux | 386 | ✅ Supported | `universal-cleaner-linux-386` |
| Linux | arm64 | ✅ Supported | `universal-cleaner-linux-arm64` |
| Windows | amd64 | ✅ Supported | `universal-cleaner-windows-amd64.exe` |
| Windows | 386 | ✅ Supported | `universal-cleaner-windows-386.exe` |
| macOS | amd64 (Intel) | ✅ Supported | `universal-cleaner-darwin-amd64` |
| macOS | arm64 (Apple Silicon) | ✅ Supported | `universal-cleaner-darwin-arm64` |

**Release Verification**:
- All releases include SHA256 checksums for security verification
- Download `checksums.txt` alongside binaries to verify integrity
- Example verification: `sha256sum -c checksums.txt`

## Contributing

Feel free to submit issues and enhancement requests!

## License

This project is open source and available under the MIT License.
