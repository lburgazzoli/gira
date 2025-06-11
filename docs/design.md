# Gira - JIRA CLI Assistant Design Document

## Overview

Gira is a command-line interface tool written in Go that provides AI-enhanced interaction with Atlassian JIRA through REST APIs. It combines traditional JIRA operations with AI-powered features for issue analysis, explanation, and intelligent updates.

## Architecture

### Core Components

1. **CLI Framework**: Cobra for command structure, Viper for configuration
2. **JIRA Client**: Custom wrapper around vanilla Go HTTP client with hashicorp/go-cleanhttp
3. **AI Integration**: Google AI models for intelligent features
4. **Configuration Management**: YAML-based configuration with environment variable support

### Directory Structure

```
gira/
├── cmd/
│   ├── root.go
│   ├── get.go
│   ├── create.go
│   ├── update.go
│   ├── tree.go
│   ├── explain.go
│   ├── enhance.go
│   └── interactive.go
├── pkg/
│   ├── jira/
│   │   ├── client.go
│   │   ├── types.go
│   │   └── operations.go
│   ├── ai/
│   │   ├── provider.go
│   │   ├── google.go
│   │   └── prompts.go
│   ├── config/
│   │   └── config.go
│   └── utils/
│       ├── output.go
│       └── tree.go
├── internal/
│   └── version/
│       └── version.go
├── configs/
│   └── config.yaml
├── go.mod
├── go.sum
├── main.go
└── README.md
```

## Detailed Component Design

### 1. CLI Framework (cmd/)

#### Root Command (`cmd/root.go`)
- Initialize Cobra application
- Setup Viper configuration
- Global flags: `--config`, `--output`, `--verbose`
- Configuration validation and JIRA connectivity check

#### Core JIRA Commands

##### Get Command (`cmd/get.go`)
```bash
gira get issue ISSUE-123
gira get project PROJECT-KEY
gira get user username
```

**Implementation Requirements:**
- Support multiple output formats (table, json, yaml)
- Flexible field selection
- Pagination support for bulk operations
- Error handling with meaningful messages

##### Create Command (`cmd/create.go`)
```bash
gira create issue --project PROJECT --type Bug --summary "Issue summary"
gira create issue --template bug-template.json
gira create issue --interactive
```

**Implementation Requirements:**
- Template-based creation
- Interactive mode with prompts
- Field validation before submission
- Support for attachments

##### Update Command (`cmd/update.go`)
```bash
gira update ISSUE-123 --summary "New summary" --assignee user@example.com
gira update ISSUE-123 --transition "In Progress"
gira update ISSUE-123 --comment "Adding update comment"
```

**Implementation Requirements:**
- Field-specific updates
- Transition handling
- Comment addition
- Bulk update capabilities

#### Advanced Commands

##### Tree Command (`cmd/tree.go`)
```bash
gira tree ISSUE-123
gira tree ISSUE-123 --depth 3 --direction both
gira tree ISSUE-123 --format ascii
```

**Implementation Requirements:**
- Traverse issue hierarchies (Epic → Story → Subtask)
- Handle issue links (blocks, duplicates, relates to)
- ASCII art tree visualization
- Configurable depth and direction (up, down, both)
- Performance optimization for large hierarchies

##### AI-Powered Explain Command (`cmd/explain.go`)
```bash
gira explain ISSUE-123
gira explain ISSUE-123 --include-relations --model gemini-pro
gira explain ISSUE-123 --output detailed
```

**Implementation Requirements:**
- Comprehensive issue analysis
- Relationship context inclusion
- Multiple explanation formats (summary, detailed, technical)
- Configurable AI model selection

##### AI-Powered Enhance Command (`cmd/enhance.go`)
```bash
gira enhance ISSUE-123 --suggest-improvements
gira enhance ISSUE-123 --add-acceptance-criteria
gira enhance ISSUE-123 --estimate-effort
```

**Implementation Requirements:**
- Intelligent issue enhancement suggestions
- Acceptance criteria generation
- Effort estimation
- Risk assessment
- Improvement recommendations

##### Interactive Mode (`cmd/interactive.go`)
```bash
gira interactive
gira chat
```

**Implementation Requirements:**
- REPL-style interface
- Context-aware conversations
- Multi-turn dialogue support
- Command history
- Session persistence

### 2. JIRA Client (`pkg/jira/`)

#### Client Structure (`pkg/jira/client.go`)
```go
type Client struct {
    baseURL    string
    httpClient *http.Client
    auth       AuthConfig
    logger     *log.Logger
}

type AuthConfig struct {
    Username string
    APIToken string
    // Support for OAuth, Bearer tokens
}
```

**Key Methods:**
- `NewClient(baseURL string, auth AuthConfig) *Client`
- `Get(endpoint string, params url.Values) (*http.Response, error)`
- `Post(endpoint string, body interface{}) (*http.Response, error)`
- `Put(endpoint string, body interface{}) (*http.Response, error)`
- `Delete(endpoint string) (*http.Response, error)`

#### JIRA Types (`pkg/jira/types.go`)
```go
type Issue struct {
    Key         string            `json:"key"`
    ID          string            `json:"id"`
    Self        string            `json:"self"`
    Fields      IssueFields       `json:"fields"`
    Transitions []Transition      `json:"transitions,omitempty"`
}

type IssueFields struct {
    Summary     string      `json:"summary"`
    Description string      `json:"description"`
    IssueType   IssueType   `json:"issuetype"`
    Status      Status      `json:"status"`
    Priority    Priority    `json:"priority"`
    Assignee    *User       `json:"assignee"`
    Reporter    *User       `json:"reporter"`
    Project     Project     `json:"project"`
    Parent      *Issue      `json:"parent,omitempty"`
    Subtasks    []Issue     `json:"subtasks,omitempty"`
    IssueLinks  []IssueLink `json:"issuelinks,omitempty"`
    Created     time.Time   `json:"created"`
    Updated     time.Time   `json:"updated"`
}
```

#### Operations (`pkg/jira/operations.go`)
High-level operations wrapping REST API calls:
- `GetIssue(key string) (*Issue, error)`
- `CreateIssue(issue *Issue) (*Issue, error)`
- `UpdateIssue(key string, update IssueUpdate) (*Issue, error)`
- `GetIssueTree(key string, depth int, direction TreeDirection) (*IssueTree, error)`
- `SearchIssues(jql string, fields []string) (*SearchResult, error)`
- `GetProject(key string) (*Project, error)`

### 3. AI Integration (`pkg/ai/`)

#### Provider Interface (`pkg/ai/provider.go`)
```go
type Provider interface {
    ExplainIssue(issue *jira.Issue, relations []*jira.Issue, options ExplainOptions) (*Explanation, error)
    EnhanceIssue(issue *jira.Issue, enhancement EnhancementType) (*Enhancement, error)
    Chat(messages []Message) (*ChatResponse, error)
}

type ExplainOptions struct {
    Format        ExplanationFormat // summary, detailed, technical
    IncludeLinks  bool
    IncludeHistory bool
}
```

#### Google AI Implementation (`pkg/ai/google.go`)
```go
type GoogleProvider struct {
    client    *genai.Client
    model     string
    apiKey    string
}
```

**Key Features:**
- Context-aware prompting
- Issue relationship analysis
- Multi-turn conversations
- Response formatting
- Error handling and retry logic

#### Prompt Templates (`pkg/ai/prompts.go`)
Structured prompts for different AI operations:
- Issue explanation templates
- Enhancement suggestion templates
- Interactive chat system prompts
- Context building helpers

### 4. Configuration Management (`pkg/config/`)

#### Configuration Structure
```go
type Config struct {
    JIRA JIRAConfig `mapstructure:"jira"`
    AI   AIConfig   `mapstructure:"ai"`
    CLI  CLIConfig  `mapstructure:"cli"`
}

type JIRAConfig struct {
    BaseURL  string `mapstructure:"base_url"`
    Username string `mapstructure:"username"`
    APIToken string `mapstructure:"api_token"`
}

type AIConfig struct {
    Provider string            `mapstructure:"provider"`
    Models   map[string]string `mapstructure:"models"`
    APIKey   string            `mapstructure:"api_key"`
}
```

#### Configuration Sources
1. Configuration file (`~/.gira/config.yaml`)
2. Environment variables (`GIRA_JIRA_BASE_URL`, etc.)
3. Command-line flags
4. Interactive setup wizard

### 5. Utilities (`pkg/utils/`)

#### Output Formatting (`pkg/utils/output.go`)
- Table formatting for issue lists
- JSON/YAML output options
- Progress bars for long operations
- Color-coded status indicators

#### Tree Visualization (`pkg/utils/tree.go`)
- ASCII tree generation
- Unicode box drawing characters
- Configurable tree styles
- Large tree handling

## Implementation Guidelines

### Phase 1: Core Infrastructure
1. Setup project structure and dependencies
2. Implement basic JIRA client with authentication
3. Create root command with configuration loading
4. Implement basic get/create/update commands

### Phase 2: Advanced JIRA Features
1. Implement tree traversal and visualization
2. Add comprehensive error handling
3. Implement search and filtering capabilities
4. Add bulk operations support

### Phase 3: AI Integration
1. Implement Google AI provider
2. Create explain command with basic functionality
3. Add enhance command
4. Implement interactive mode

### Phase 4: Polish and Testing
1. Add comprehensive tests
2. Improve error messages and user experience
3. Add configuration validation
4. Performance optimization

## Dependencies

### Core Dependencies
```go
// CLI and Configuration
"github.com/spf13/cobra"
"github.com/spf13/viper"

// HTTP Client
"github.com/hashicorp/go-cleanhttp"

// AI Integration
"github.com/google/generative-ai-go/genai"
"google.golang.org/api/option"

// Utilities
"github.com/olekukonko/tablewriter"
"github.com/fatih/color"
"gopkg.in/yaml.v3"
```

## Configuration Examples

### Basic Configuration (`~/.gira/config.yaml`)
```yaml
jira:
  base_url: "https://your-domain.atlassian.net"
  username: "your-email@example.com"
  api_token: "your-api-token"

ai:
  provider: "google"
  api_key: "your-google-ai-api-key"
  models:
    explain: "gemini-pro"
    enhance: "gemini-pro"
    chat: "gemini-pro"

cli:
  output_format: "table"
  color: true
  verbose: false
```

## Usage Examples

### Basic Operations
```bash
# Setup
gira config init

# Get issue details
gira get issue PROJ-123

# Create new issue
gira create issue --project PROJ --type Story --summary "New feature request"

# Update issue
gira update PROJ-123 --assignee john.doe@example.com --status "In Progress"
```

### Advanced Operations
```bash
# View issue hierarchy
gira tree PROJ-123 --depth 3

# AI-powered issue explanation
gira explain PROJ-123 --include-relations

# AI-powered issue enhancement
gira enhance PROJ-123 --add-acceptance-criteria

# Interactive AI mode
gira interactive
```

## Error Handling Strategy

1. **Network Errors**: Retry with exponential backoff
2. **Authentication Errors**: Clear instructions for token setup
3. **JIRA API Errors**: Meaningful error translation
4. **AI Provider Errors**: Fallback mechanisms and retry logic
5. **Configuration Errors**: Validation with helpful suggestions

## Testing Strategy

1. **Unit Tests**: All core functions and utilities
2. **Integration Tests**: JIRA API interactions
3. **E2E Tests**: Complete command workflows
4. **Mock Testing**: AI provider interactions
5. **Performance Tests**: Large dataset handling

## Security Considerations

1. **API Token Storage**: Secure credential management
2. **AI Data Privacy**: Option to exclude sensitive fields
3. **Network Security**: TLS verification, proxy support
4. **Input Validation**: Prevent injection attacks
5. **Rate Limiting**: Respect API limits

This design document provides a comprehensive blueprint for implementing Gira. Each component is designed to be modular, testable, and extensible. The phased implementation approach ensures steady progress while maintaining code quality.