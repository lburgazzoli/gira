# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Gira is an AI-powered Jira CLI tool built in Go that combines traditional JIRA operations with AI-enhanced features for issue analysis, explanation, and intelligent updates using Google AI models.

## Technology Stack

- **Language**: Go 1.24.4
- **CLI Framework**: Cobra for command structure, Viper for configuration
- **HTTP Client**: hashicorp/go-cleanhttp for JIRA REST API calls
- **AI Integration**: Google AI models (Gemini) for intelligent features
- **Output**: tablewriter for formatted output, fatih/color for colored output

## Development Commands

```bash
# Using Makefile (recommended)
make build      # Build the binary to bin/gira
make test       # Run all tests with verbose output
make deps       # Install and update dependencies
make clean      # Clean build artifacts

# Direct Go commands
go build -o bin/gira .                    # Build binary
go test ./...                             # Run all tests
go test -v ./...                          # Run tests with verbose output
go test -run TestName ./path/to/package   # Run specific test
go mod tidy                               # Update dependencies
```

## Architecture Overview

### Directory Structure
```
gira/
├── cmd/                 # CLI command packages organized by functionality
│   ├── config/         # Configuration management commands (init, show, set)
│   ├── get/            # Resource retrieval commands (issue, project)
│   ├── version/        # Version information command
│   └── root.go         # Root command and CLI initialization
├── pkg/jira/           # JIRA client, types, and operations
├── pkg/ai/             # AI provider interface and Google implementation
├── pkg/config/         # Configuration management
├── pkg/utils/          # Output formatting and tree visualization
├── internal/version/   # Version information
└── configs/            # Configuration templates
```

### Core Components

1. **JIRA Client** (`pkg/jira/`): Custom HTTP client wrapper with authentication, issue operations, and tree traversal
2. **AI Integration** (`pkg/ai/`): Provider interface with Google AI implementation for issue explanation and enhancement
3. **CLI Commands** (`cmd/`): Cobra-based commands for JIRA operations and AI-powered features
4. **Configuration** (`pkg/config/`): Viper-based config with YAML files and environment variable support

## Implementation Phases

**Phase 1** ✅ (Complete): Core infrastructure with basic JIRA client, root command, and basic get/create/update commands
**Phase 2** 🚧 (In Progress): Advanced JIRA features including tree traversal, search, and bulk operations
**Phase 3** 📋 (Planned): AI integration with explain, enhance, and interactive commands
**Phase 4** 📋 (Planned): Polish, comprehensive testing, and performance optimization

## Configuration

Configuration sources (in order of precedence):
1. Command-line flags
2. Environment variables (GIRA_JIRA_BASE_URL, etc.)
3. Configuration file (follows XDG Base Directory specification)
4. Interactive setup wizard

### Configuration File Locations
- `$XDG_CONFIG_HOME/gira/config.yaml` (if XDG_CONFIG_HOME is set)
- `~/.config/gira/config.yaml` (preferred on Unix-like systems)
- `~/.gira/config.yaml` (fallback)

### Authentication
The tool uses JIRA Personal Access Tokens with Bearer authentication:
```yaml
jira:
  base_url: "https://your-domain.atlassian.net"
  token: "your-personal-access-token"  # JIRA Personal Access Token
```

For hosted JIRA instances, the tool uses:
- **API Version**: `/rest/api/2/` (JIRA Cloud/Server API v2)
- **Authentication**: `Authorization: Bearer <token>` header
- **Token Type**: JIRA Personal Access Token (not Basic Auth)
- **Rate Limiting**: Automatic retry with exponential backoff (3 retries max, 30s max delay)
- **Headers**: Supports `Retry-After` and `X-RateLimit-Reset` for intelligent backoff

## Key Features

### ✅ Implemented Features
- **Config Command**: Interactive setup wizard with `init`, `show`, and `set` subcommands
- **Get Commands**: Retrieve issues and projects with multiple output formats
  - Issue hierarchy visualization with `--tree` option supporting:
    - Epic Link relationships (`"Epic Link" = EPIC-KEY`)
    - Parent-child relationships (`parent = PARENT-KEY`)
    - Direct subtasks traversal
    - Multiple output formats (tree, table, JSON, YAML)
    - Depth control and reverse view options
- **Version Command**: Display build information with multiple output formats
- **Multiple Output Formats**: table, JSON, YAML with proper formatting
- **JIRA Authentication**: Bearer token support for personal access tokens
- **Rate Limiting**: Automatic retry with exponential backoff for rate-limited requests
- **Error Handling**: Clean error messages without CLI help display

### 📋 Planned Features
- **AI Explain**: Comprehensive issue analysis with relationship context
- **AI Enhance**: Intelligent suggestions for acceptance criteria, effort estimation
- **Interactive Mode**: REPL-style AI-powered conversations
- **Search Command**: Advanced JQL query interface
- **Bulk Operations**: Multi-issue operations and updates

## Code Style Guidelines

### Go Best Practices
- **Function Parameters**: Always declare each parameter with its own explicit type, avoid grouping parameters with shared types
  ```go
  // ✅ Good - explicit types for each parameter
  func doRequest(method string, endpoint string, body io.Reader, params url.Values) (*http.Response, error)
  
  // ❌ Avoid - grouped parameter types
  func doRequest(method, endpoint string, body io.Reader, params url.Values) (*http.Response, error)
  ```

- **HTTP Client Structure**: Use separation of concerns for HTTP client architecture:
  - `client.go`: Public domain operations (GetIssue, CreateIssue, etc.)
  - `client_support.go`: Private HTTP transport helpers (get, post, put, delete, doRequest)
  - Use standard library constants (`http.MethodGet`, `http.StatusOK`, etc.)
  - Make HTTP transport methods private, expose only domain-specific operations

- **Table Formatting**: Standardize table creation across all commands using consistent options:
  ```go
  table := tablewriter.NewTable(os.Stdout)
  table.Options(tablewriter.WithRendition(
      tw.Rendition{
          Settings: tw.Settings{
              Separators: tw.Separators{
                  BetweenColumns: tw.Off,
              },
          },
      },
  ))
  ```

- **Rate Limiting**: Use established libraries like `github.com/hashicorp/go-retryablehttp` instead of custom implementations
- **URL Construction**: Use `net/url.JoinPath` for idiomatic URL building instead of string concatenation

### CLI Command Structure
- **Package Organization**: Each major command group should have its own package under `cmd/`
- **Command Export**: Export the main command as `Cmd` variable (e.g., `var Cmd = &cobra.Command{...}`)
- **Subcommand Registration**: Register subcommands in package `init()` functions, not in root
- **Flag Scoping**: Keep command-specific flags local to their packages
- **Dependency Injection**: Pass shared dependencies (config, output formatting) rather than importing root package

Example package structure:
```go
// cmd/get/get.go
package get

var Cmd = &cobra.Command{
    Use:   "get",
    Short: "Get JIRA resources",
}

var issueCmd = &cobra.Command{
    Use:  "issue ISSUE-KEY",
    RunE: runGetIssue,
}

func init() {
    Cmd.AddCommand(issueCmd)
}

// cmd/root.go
func init() {
    rootCmd.AddCommand(get.Cmd)
}
```

- **Error Handling**: Use modern Go error handling with `errors.As` and `errors.Is` instead of type assertions
- **Constants**: Extract hardcoded strings to constants when they represent API endpoints, configuration keys, or repeated values

### Output Format Handling
- **Global Flag Access**: Commands should respect the global `--output` flag from the root command
- **Circular Dependency Prevention**: When accessing global flags in helper functions, pass the command as a parameter to avoid circular dependencies:
  ```go
  // ✅ Good - pass command to avoid circular references
  func outputResult(cmd *cobra.Command, result interface{}) error {
      outputFormat, _ := cmd.Root().PersistentFlags().GetString("output")
      // ...
  }
  
  // ❌ Avoid - circular reference through global Cmd variable
  func outputResult(result interface{}) error {
      outputFormat, _ := Cmd.Root().PersistentFlags().GetString("output")
      // ...
  }
  ```
- **Format Support**: All commands should support table, JSON, and YAML output formats consistently
- **Format Priority**: Respect command-specific flags (like `--table`) before checking global format
- **Encoder Configuration**: Use consistent JSON/YAML encoder settings (indentation, etc.) across commands
- **Default Output Format**: Commands should use appropriate defaults:
  - Version command: Plain text format (not table) for better CLI UX
  - Get issue command: Plain text format for better readability  
  - Other commands: Table format for structured data display

### Version Command Implementation
The version command (`cmd/version/`) demonstrates proper output format handling:

- **Build-time Information**: Uses Go's `-ldflags` to inject version, commit, and build date
- **Makefile Integration**: Automatically extracts git information during build:
  ```makefile
  VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
  COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
  DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
  
  LDFLAGS = -X 'github.com/lburgazzoli/gira/internal/version.Version=$(VERSION)' \
            -X 'github.com/lburgazzoli/gira/internal/version.Commit=$(COMMIT)' \
            -X 'github.com/lburgazzoli/gira/internal/version.Date=$(DATE)'
  ```
- **Output Formats**: Supports plain text (default), table, JSON, and YAML
- **Plain Text Format**: Uses simple `key : value` format for version information
- **Version Package**: Separates version storage (`internal/version/`) from command logic (`cmd/version/`)
