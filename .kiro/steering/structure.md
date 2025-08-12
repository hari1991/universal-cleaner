# Project Structure

## Root Files
- `main.go` - GUI application entry point and core logic
- `cli.go` - Command-line interface implementation
- `resource.go` - Embedded resources (icons, assets)
- `go.mod` / `go.sum` - Go module dependencies
- `Makefile` - Build automation and common tasks
- `README.md` - Project documentation

## Key Directories
- `dist/` - Build output directory for cross-platform binaries
- `test-cleanup/` - Test directories for development/testing cleanup functionality
- `.kiro/` - Kiro IDE configuration and steering rules
- `.vscode/` - VS Code configuration

## Code Organization

### Main Application (`main.go`)
- GUI setup and event handling
- Core cleaning logic and file scanning
- Fyne UI components and layout

### CLI Interface (`cli.go`) 
- Command-line argument parsing
- Non-interactive cleaning operations
- Shared logic with GUI version

### Resource Management (`resource.go`)
- Embedded static resources
- Icon and asset definitions

## Naming Conventions
- Go standard naming (camelCase for private, PascalCase for public)
- Descriptive function names (`findCleanableItems`, `performCleanup`)
- Clear struct names (`CleanableItem`, `UniversalCleaner`)

## File Patterns
- Single-file modules for focused functionality
- Shared data structures between GUI and CLI
- Resource embedding for self-contained deployment