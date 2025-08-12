# Design Document

## Overview

The GitHub CI/CD system will consist of two main workflows: a continuous integration workflow for pull requests and pushes, and a release workflow for creating and distributing binaries. The design leverages GitHub Actions' matrix strategy for efficient multi-platform builds and follows Go best practices for cross-compilation.

## Architecture

### Workflow Structure
- **CI Workflow** (`.github/workflows/ci.yml`): Triggered on push/PR for build validation
- **Release Workflow** (`.github/workflows/release.yml`): Triggered on tag creation for distribution

### Build Matrix Strategy
The workflows will use GitHub Actions matrix builds to efficiently compile for multiple platforms:

```yaml
strategy:
  matrix:
    include:
      - os: ubuntu-latest
        goos: linux
        goarch: amd64
      - os: ubuntu-latest
        goos: linux
        goarch: 386
      - os: ubuntu-latest
        goos: linux
        goarch: arm64
      - os: ubuntu-latest
        goos: windows
        goarch: amd64
      - os: ubuntu-latest
        goos: windows
        goarch: 386
      - os: macos-latest
        goos: darwin
        goarch: amd64
      - os: macos-latest
        goos: darwin
        goarch: arm64
```

## Components and Interfaces

### CI Workflow Components
1. **Environment Setup**
   - Go installation (version from go.mod)
   - Dependency caching
   - Module download

2. **Quality Checks**
   - Code formatting verification (`go fmt`)
   - Linting (if golangci-lint is added)
   - Test execution (`go test`)

3. **Build Validation**
   - Cross-platform compilation test
   - Binary generation verification

### Release Workflow Components
1. **Build Matrix Execution**
   - Parallel builds for all target platforms
   - Binary naming with OS/architecture suffixes
   - Windows executable extension handling

2. **Artifact Management**
   - Binary collection and organization
   - Checksum generation (SHA256)
   - Archive creation for distribution

3. **Release Publishing**
   - GitHub release creation
   - Binary attachment to release
   - Checksum file publication

## Data Models

### Build Configuration
```yaml
# Platform matrix definition
platforms:
  - { os: "linux", arch: "amd64", runner: "ubuntu-latest" }
  - { os: "linux", arch: "386", runner: "ubuntu-latest" }
  - { os: "linux", arch: "arm64", runner: "ubuntu-latest" }
  - { os: "windows", arch: "amd64", runner: "ubuntu-latest" }
  - { os: "windows", arch: "386", runner: "ubuntu-latest" }
  - { os: "darwin", arch: "amd64", runner: "macos-latest" }
  - { os: "darwin", arch: "arm64", runner: "macos-latest" }
```

### Artifact Naming Convention
- Format: `universal-cleaner-{os}-{arch}[.exe]`
- Examples:
  - `universal-cleaner-linux-amd64`
  - `universal-cleaner-windows-amd64.exe`
  - `universal-cleaner-darwin-arm64`

### Checksum File Format
```
SHA256 checksums for Universal Cleaner binaries:

a1b2c3d4... universal-cleaner-linux-amd64
e5f6g7h8... universal-cleaner-windows-amd64.exe
i9j0k1l2... universal-cleaner-darwin-arm64
```

## Error Handling

### Build Failures
- Individual platform build failures won't stop other builds
- Failed builds will be reported in workflow summary
- Release workflow will fail if any critical platform fails

### Dependency Issues
- Go module download failures will retry with exponential backoff
- Cache misses will fall back to fresh dependency installation
- Version conflicts will be reported with clear error messages

### Release Process Errors
- Failed binary uploads will retry up to 3 times
- Checksum generation errors will fail the entire release
- Duplicate release attempts will be handled gracefully

## Testing Strategy

### CI Testing
1. **Unit Tests**: Run existing Go tests across all platforms
2. **Build Tests**: Verify successful compilation for all target platforms
3. **Integration Tests**: Test CLI functionality with sample directories

### Release Testing
1. **Binary Validation**: Verify each built binary can execute `--help`
2. **Checksum Verification**: Validate generated checksums match binaries
3. **Archive Integrity**: Test that uploaded artifacts are not corrupted

### Manual Testing
1. **Release Process**: Test complete release workflow on feature branch
2. **Download Verification**: Manually verify released binaries work on target platforms
3. **Checksum Validation**: Manually verify checksum file accuracy

## Security Considerations

### Build Environment
- Use official GitHub-hosted runners for consistent environment
- Pin action versions to prevent supply chain attacks
- Limit workflow permissions to minimum required scope

### Artifact Integrity
- Generate and publish SHA256 checksums for all binaries
- Use GitHub's built-in artifact signing when available
- Ensure reproducible builds where possible

### Access Control
- Restrict release workflow to repository maintainers
- Use GitHub's environment protection rules for release approval
- Implement branch protection rules for main branch