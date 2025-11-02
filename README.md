# Sanitize

A modern, robust command-line tool that recursively walks a folder tree and renames directories to be compatible with Windows naming conventions. Built with Go using clean architecture principles and comprehensive testing.

## 🚀 Features

- **Smart Processing**: Processes folders from lowest level to highest level (bottom-up traversal) to avoid path conflicts
- **Windows Compatible**: Removes invalid Windows characters: `< > : " | ? * \ /`
- **Unicode Support**: Converts Unicode/non-ASCII characters to closest ASCII equivalents (café → cafe)
- **Safety First**: Control characters (ASCII 0-31) removal and trailing spaces/periods cleanup
- **Reserved Names**: Handles Windows reserved names (CON, PRN, AUX, NUL, COM1-COM9, LPT1-LPT9)
- **Length Management**: Enforces 255-character length limit with smart truncation
- **Collision Detection**: Handles name conflicts by appending numbers (_1, _2, etc.)
- **Preview Mode**: Dry-run mode to preview changes without making them
- **Interactive UI**: Optional Terminal UI (TUI) with progress indicators using Bubble Tea
- **Verbose Logging**: Detailed progress reporting and error handling
- **Cross-Platform**: Builds for Linux, Windows, and macOS

## 📦 Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [GitHub Releases](https://github.com/punkscience/sanitize/releases) page.

### Install with Go

```bash
go install github.com/punkscience/sanitize@latest
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/punkscience/sanitize.git
cd sanitize

# Install dependencies
go mod tidy

# Build the application
go build -o sanitize

# Or build for specific platform
GOOS=windows GOARCH=amd64 go build -o sanitize.exe
```

## 🎯 Usage

### Basic Commands

```bash
# Sanitize current directory
sanitize

# Sanitize specific directory
sanitize --path "/path/to/directory"

# Preview changes without making them (recommended first step)
sanitize --path "/path/to/directory" --dry-run

# Enable verbose output for detailed progress
sanitize --path "/path/to/directory" --verbose

# Use interactive Terminal UI
sanitize --path "/path/to/directory" --tui

# Combine options
sanitize --path "/path/to/directory" --dry-run --verbose --tui
```

### Command-Line Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--path` | `-p` | Root path to sanitize | `.` (current directory) |
| `--dry-run` | `-d` | Show what would be renamed without making changes | `false` |
| `--verbose` | `-v` | Enable verbose output | `false` |
| `--tui` | `-t` | Use Terminal UI (Bubble Tea) for interactive progress | `false` |
| `--help` | `-h` | Show help information | - |

### Examples

```bash
# Quick dry run to see what would change
sanitize -p "/my/messy/folders" -d -v

# Interactive mode with progress bar
sanitize -p "/my/messy/folders" -t

# Quiet execution (no verbose output)
sanitize -p "/my/messy/folders"

# Cross-platform path examples
sanitize -p "C:\Users\Documents\Photos"    # Windows
sanitize -p "/home/user/documents"          # Linux
sanitize -p "/Users/user/Documents"         # macOS
```

## 🔄 Before & After Examples

### Directory Structure Transformation

**Before:**
```
project/
├── bad<chars>folder/
├── ending_with_period./
├── CON/
├── unicode_café_résumé/
├── file with spaces   /
└── very_long_folder_name_that_exceeds_the_255_character_limit_and_needs_truncation/
```

**After:**
```
project/
├── bad_chars_folder/
├── ending_with_period/
├── CON_/
├── unicode_cafe_resume/
├── file with spaces/
└── very_long_folder_name_that_exceeds_the_255_character_limit_and_needs_trunca.../
```

### Unicode Character Mapping

| Original | Sanitized | Type |
|----------|-----------|------|
| `café` | `cafe` | Accented characters |
| `naïve` | `naive` | Diacritics |
| `résumé` | `resume` | Mixed accents |
| `Москва` | `AAAAAA` | Cyrillic → Generic ASCII |

## 🏗️ Architecture

The application follows **SOLID principles** and clean architecture patterns:

- **Single Responsibility**: Each component has one clear purpose
- **Open/Closed**: Extensible through interfaces without modification
- **Liskov Substitution**: All implementations are interchangeable
- **Interface Segregation**: Focused, specific interfaces
- **Dependency Inversion**: Depends on abstractions, not concretions

### Key Components

- **🧹 Sanitizer**: Windows-compatible name sanitization logic
- **🚶 Walker**: Directory tree traversal and folder discovery
- **⚙️ Processor**: File system rename operations with collision handling  
- **📊 Reporter**: Progress reporting (CLI and TUI implementations)
- **🎼 Service**: Orchestrates all components together

## 🧪 Testing

The project includes comprehensive test coverage:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
go test -bench=. ./...
```

### CI/CD Pipeline

- ✅ **Automated Testing**: Unit tests, integration tests, and benchmarks
- 🔍 **Code Quality**: Linting, formatting, and security scanning
- 🏗️ **Multi-Platform Builds**: Linux, Windows, macOS (amd64, arm64)
- 📦 **Automated Releases**: GitHub Actions with binary artifacts
- 🛡️ **Security Scanning**: Vulnerability detection and SARIF reporting

## 🔒 Windows Folder Naming Rules

The tool enforces these Windows compatibility rules:

1. **Invalid Characters**: Cannot contain `< > : " | ? * \ /`
2. **Control Characters**: Removes ASCII 0-31 control characters
3. **Trailing Issues**: Cannot end with space or period
4. **Length Limits**: Cannot exceed 255 characters (truncated with `...`)
5. **Reserved Names**: Cannot use CON, PRN, AUX, NUL, COM1-COM9, LPT1-LPT9
6. **Empty Names**: Cannot be empty or contain only spaces
7. **Unicode Handling**: Converts to ASCII equivalents where possible

## 🛡️ Safety Features

- **🔍 Preview Mode**: Always test with `--dry-run` first
- **⬇️ Bottom-Up Processing**: Processes folders from deepest to shallowest
- **🔄 Collision Handling**: Automatic number appending for conflicts (_1, _2, etc.)
- **⚠️ Error Recovery**: Continues processing despite individual folder errors
- **📝 Comprehensive Logging**: Detailed error messages and warnings
- **🚫 Permission Handling**: Gracefully skips inaccessible directories

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Follow** the coding standards (run `go fmt` and `golangci-lint`)
4. **Write** tests for new functionality
5. **Commit** your changes (`git commit -m 'Add amazing feature'`)
6. **Push** to the branch (`git push origin feature/amazing-feature`)
7. **Open** a Pull Request

### Development Setup

```bash
# Clone and setup
git clone https://github.com/punkscience/sanitize.git
cd sanitize
go mod tidy

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Run quality checks
golangci-lint run
go vet ./...
staticcheck ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **[Cobra](https://github.com/spf13/cobra)**: Powerful CLI framework
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**: Terminal UI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)**: Style definitions for terminal applications

## 📞 Support

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/punkscience/sanitize/issues)
- 💡 **Feature Requests**: [GitHub Discussions](https://github.com/punkscience/sanitize/discussions)  
- 📖 **Documentation**: Check the [Wiki](https://github.com/punkscience/sanitize/wiki)
- 💬 **Community**: [GitHub Discussions](https://github.com/punkscience/sanitize/discussions)

---

**Made with ❤️ by the Sanitize team**