# Gira

An AI-powered JIRA CLI tool built in Go that combines traditional JIRA operations with AI-enhanced features for issue analysis, explanation, and intelligent updates.

## Features

✅ **Complete JIRA Integration**
- Issue and project retrieval with multiple output formats
- Bearer token authentication with JIRA Personal Access Tokens
- Support for JIRA Cloud/Server API v2
- Automatic rate limiting with exponential backoff retry

✅ **Advanced Issue Hierarchy Visualization**
- Tree view of issue relationships with ASCII art
- Support for Epic Links, parent-child relationships, and subtasks
- Multiple output formats: tree view, table, JSON, YAML
- Depth control and reverse view options

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

# Get a project
gira get project MYPROJECT

# View issue hierarchy as a tree
gira tree EPIC-456

# View with more details
gira tree EPIC-456 --all --depth 5
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

# Get project information
gira get project MYPROJECT
```

### Tree Command

Visualize issue hierarchies including Epic Links, parent-child relationships, and subtasks:

```bash
# Basic tree view
gira tree EPIC-123

# Table format
gira tree EPIC-123 --table
gira tree EPIC-123 --output table

# Control depth and show all fields
gira tree EPIC-123 --depth 5 --all

# Reverse view (children first)
gira tree EPIC-123 --reverse

# JSON/YAML output
gira tree EPIC-123 --output json
gira tree EPIC-123 --output yaml
```

### Global Options

```bash
# Verbose output
gira --verbose tree EPIC-123

# Custom config file
gira --config /path/to/config.yaml get issue PROJECT-123

# Output formats
gira --output table get issue PROJECT-123
gira --output json get project MYPROJECT
gira --output yaml tree EPIC-123
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

## Examples

### Tree Visualization

The tree command provides comprehensive issue hierarchy visualization:

```bash
# Example output for an Epic with multiple child issues
$ gira tree RHOAIENG-25251 --depth 2

RHOAIENG-25251: [ODH Operator] Support Kueue for Enhanced Workload Management in OpenShift AI [In Progress]
├── RHOAIENG-27390: [DOC] Add documentation for the new HardwareProfile API [New]
├── RHOAIENG-27389: [DOC] Add documentation for Kueue migration and supported configurations [New]
├── RHOAIENG-26727: Add the new HardwareProfile API to the API tiers for OpenShift AI doc [Backlog]
├── RHOAIENG-26595: [QE] Include new Kueue operator and its dependencies in the QE repositories [Backlog]
├── RHOAIENG-26500: [ODH Operator] Remove VAP Config and Bindings from RHOAI Operator Kueue Manifests [Backlog]
├── RHOAIENG-26410: [ODH Operator] Set controller deployment to 3 replicas for webhook high availability [Backlog]
├── RHOAIENG-26336: [ODH Operator] Ensure the admin-rolebinding only include valid groups [Backlog]
├── RHOAIENG-26316: [SPIKE] Research and define default Kueue Operator configuration values [Backlog]
├── RHOAIENG-26315: [SPIKE] Research and define default ClusterQueue configuration values [Backlog]
├── RHOAIENG-26135: [QE] Include new Kueue operator and its dependencies in the QE repositories [In Progress]
├── RHOAIENG-25404: [ODH Operator] Create OpenShift Kueue Configuration [Backlog]
├── RHOAIENG-25258: [ODH Operator] Implement Validating Webhook for Kueue label Enforcement [Review]
├── RHOAIENG-25255: [ODH Operator] Migrate Hardware Profile Management to RHOAI Operator [Backlog]
├── RHOAIENG-25253: [ODH Operator] Create Default Kueue Resources for labeled Namespaces [Backlog]
├── RHOAIENG-25252: [ODH Operator] Support "Unmanaged" Kueue in OpenShift AI [In Progress]
└── RHOAIENG-24289: Move Kueue manifests into RHOAI operator [Resolved]
```

### Multiple Output Formats

```bash
# Table format with detailed information
$ gira tree EPIC-123 --table --all

# JSON output for programmatic processing
$ gira tree EPIC-123 --output json | jq '.children[].issue.key'

# YAML output for configuration management
$ gira tree EPIC-123 --output yaml
```

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
├── cmd/                 # CLI commands (get, config, tree)
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