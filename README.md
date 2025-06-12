# Gira

An AI-powered JIRA CLI tool built in Go that combines traditional JIRA operations with AI-enhanced features for issue analysis, explanation, and intelligent updates.

## Features

✅ **Complete JIRA Integration**
- Issue and project retrieval with multiple output formats
- Issue hierarchy visualization with `--tree` option
- Bearer token authentication with JIRA Personal Access Tokens
- Support for JIRA Cloud/Server API v2
- Automatic rate limiting with exponential backoff retry


✅ **Configuration Management**
- Interactive setup wizard
- XDG Base Directory specification compliance
- Environment variable support

🚧 **AI-Powered Features** (Planned)
- Issue analysis and explanation
- Intelligent suggestions for acceptance criteria
- Interactive AI-powered conversations

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/lburgazzoli/gira.git
cd gira

# Build the binary
make build

# Or use Go directly
go build -o bin/gira .
```

### Dependencies

- Go 1.24.4 or later
- A JIRA instance with Personal Access Token

## Quick Start

### 1. Initialize Configuration

```bash
# Interactive setup wizard
gira config init

# Manual configuration
gira config set jira.base_url "https://your-domain.atlassian.net"
gira config set jira.token "your-personal-access-token"
```

### 2. Basic Usage

```bash
# Get an issue
gira get issue PROJECT-123

# Get an issue with tree hierarchy
gira get issue EPIC-456 --tree

# Get a project
gira get project MYPROJECT

# Check version
gira version
```

## Commands

### Configuration Commands

```bash
# Initialize configuration interactively
gira config init

# Show current configuration
gira config show

# Set configuration values
gira config set jira.base_url "https://example.atlassian.net"
gira config set jira.token "your-token"
```

### Get Commands

```bash
# Get issue details
gira get issue PROJECT-123
gira get issue PROJECT-123 --output json
gira get issue PROJECT-123 --output yaml

# Get issue with hierarchy tree view
gira get issue EPIC-123 --tree
gira get issue EPIC-123 --tree --tree-depth 5
gira get issue EPIC-123 --tree --tree-reverse
gira get issue EPIC-123 --tree --tree-all

# Tree view with different output formats
gira get issue EPIC-123 --tree --output table
gira get issue EPIC-123 --tree --output json
gira get issue EPIC-123 --tree --output yaml

# Get project information
gira get project MYPROJECT
```

### Version Command

Display build information including version, commit, and build date:

```bash
# Default plain text output
gira version

# Table format
gira version --output table

# JSON/YAML output
gira version --output json
gira version --output yaml
```


### Global Options

```bash
# Verbose output
gira --verbose get issue PROJECT-123

# Custom config file
gira --config /path/to/config.yaml get issue PROJECT-123

# Output formats
gira --output table get issue PROJECT-123
gira --output json get project MYPROJECT
gira --output yaml version
```

## Configuration

### Configuration File

Gira follows the XDG Base Directory specification:

- `$XDG_CONFIG_HOME/gira/config.yaml` (if XDG_CONFIG_HOME is set)
- `~/.config/gira/config.yaml` (preferred on Unix-like systems)  
- `~/.gira/config.yaml` (fallback)

### Configuration Format

```yaml
jira:
  base_url: "https://your-domain.atlassian.net"
  token: "your-personal-access-token"

cli:
  output_format: "table"
  color: true
  verbose: false

ai:
  provider: "google"
  api_key: "your-google-ai-api-key"
  models:
    default: "gemini-pro"
```

### Authentication

Gira uses JIRA Personal Access Tokens with Bearer authentication:

1. **Create a Personal Access Token** in your JIRA instance
2. **Configure the token** using `gira config init` or `gira config set jira.token "your-token"`
3. **Set your JIRA base URL** using `gira config set jira.base_url "https://your-domain.atlassian.net"`


## Development

### Build Commands

```bash
# Using Makefile (recommended)
make build      # Build the binary to bin/gira
make test       # Run all tests with verbose output  
make deps       # Install and update dependencies
make clean      # Clean build artifacts

# Direct Go commands
go build -o bin/gira .
go test ./...
go mod tidy
```

### Project Structure

```
gira/
├── cmd/                 # CLI commands (get, config, version)
├── pkg/jira/           # JIRA client, types, and operations
├── pkg/ai/             # AI provider interface (planned)
├── pkg/config/         # Configuration management
├── pkg/utils/          # Output formatting utilities  
├── internal/version/   # Version information
└── configs/            # Configuration templates
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.