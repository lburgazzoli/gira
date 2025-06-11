package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lburgazzoli/gira/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gira configuration",
	Long:  `Manage gira configuration settings including JIRA connection and AI settings.`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gira configuration",
	Long:  `Initialize gira configuration with an interactive setup wizard.`,
	RunE:  runInit,
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current gira configuration settings.`,
	RunE:  runShow,
}

var setCmd = &cobra.Command{
	Use:   "set KEY VALUE",
	Short: "Set a configuration value",
	Long: `Set a configuration value. Supported keys:
  jira.base_url    - JIRA instance URL
  jira.token       - JIRA Personal Access Token
  ai.provider      - AI provider (google)
  ai.api_key       - AI API key
  cli.output_format - Output format (table, json, yaml)
  cli.color        - Enable colored output (true, false)
  cli.verbose      - Enable verbose output (true, false)`,
	Args: cobra.ExactArgs(2),
	RunE: runSet,
}

func init() {
	Cmd.AddCommand(initCmd)
	Cmd.AddCommand(showCmd)
	Cmd.AddCommand(setCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("🔧 Gira Configuration Setup")
	fmt.Println("===========================")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	
	// JIRA Configuration
	fmt.Println("📋 JIRA Configuration")
	fmt.Println("---------------------")
	
	baseURL, err := promptString(reader, "JIRA Base URL (e.g., https://your-domain.atlassian.net)", "")
	if err != nil {
		return fmt.Errorf("failed to read JIRA base URL: %w", err)
	}
	
	token, err := promptString(reader, "JIRA Personal Access Token", "")
	if err != nil {
		return fmt.Errorf("failed to read JIRA token: %w", err)
	}
	
	// AI Configuration
	fmt.Println()
	fmt.Println("🤖 AI Configuration")
	fmt.Println("-------------------")
	
	provider, err := promptString(reader, "AI Provider", "google")
	if err != nil {
		return fmt.Errorf("failed to read AI provider: %w", err)
	}
	
	apiKey, err := promptString(reader, "AI API Key (Google AI)", "")
	if err != nil {
		return fmt.Errorf("failed to read AI API key: %w", err)
	}
	
	// CLI Configuration
	fmt.Println()
	fmt.Println("🖥️  CLI Configuration")
	fmt.Println("--------------------")
	
	outputFormat, err := promptString(reader, "Output Format", "table")
	if err != nil {
		return fmt.Errorf("failed to read output format: %w", err)
	}
	
	colorStr, err := promptString(reader, "Enable Colors", "true")
	if err != nil {
		return fmt.Errorf("failed to read color setting: %w", err)
	}
	color := strings.ToLower(colorStr) == "true"
	
	verboseStr, err := promptString(reader, "Enable Verbose Output", "false")
	if err != nil {
		return fmt.Errorf("failed to read verbose setting: %w", err)
	}
	verbose := strings.ToLower(verboseStr) == "true"
	
	// Create configuration
	cfg := config.Config{
		JIRA: config.JIRAConfig{
			BaseURL: baseURL,
			Token:   token,
		},
		AI: config.AIConfig{
			Provider: provider,
			APIKey:   apiKey,
			Models: map[string]string{
				"explain": "gemini-pro",
				"enhance": "gemini-pro",
				"chat":    "gemini-pro",
			},
		},
		CLI: config.CLIConfig{
			OutputFormat: outputFormat,
			Color:        color,
			Verbose:      verbose,
		},
	}
	
	// Save configuration
	return saveConfig(&cfg)
}

func runShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// Mask sensitive information
	maskedCfg := *cfg
	if cfg.JIRA.Token != "" {
		maskedCfg.JIRA.Token = "***masked***"
	}
	if cfg.AI.APIKey != "" {
		maskedCfg.AI.APIKey = "***masked***"
	}
	
	return outputResult(&maskedCfg)
}

func runSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]
	
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// Parse key and update configuration
	parts := strings.Split(key, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid key format. Use section.key (e.g., jira.base_url)")
	}
	
	section, field := parts[0], parts[1]
	
	switch section {
	case "jira":
		switch field {
		case "base_url":
			cfg.JIRA.BaseURL = value
		case "token":
			cfg.JIRA.Token = value
		default:
			return fmt.Errorf("unknown JIRA config field: %s", field)
		}
	case "ai":
		switch field {
		case "provider":
			cfg.AI.Provider = value
		case "api_key":
			cfg.AI.APIKey = value
		default:
			return fmt.Errorf("unknown AI config field: %s", field)
		}
	case "cli":
		switch field {
		case "output_format":
			cfg.CLI.OutputFormat = value
		case "color":
			cfg.CLI.Color = strings.ToLower(value) == "true"
		case "verbose":
			cfg.CLI.Verbose = strings.ToLower(value) == "true"
		default:
			return fmt.Errorf("unknown CLI config field: %s", field)
		}
	default:
		return fmt.Errorf("unknown config section: %s", section)
	}
	
	if err := saveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	
	fmt.Printf("✅ Configuration updated: %s = %s\n", key, value)
	return nil
}

func promptString(reader *bufio.Reader, prompt string, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	
	input = strings.TrimSpace(input)
	if input == "" && defaultValue != "" {
		return defaultValue, nil
	}
	
	return input, nil
}

func saveConfig(cfg *config.Config) error {
	// Get config directory
	configDir, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}
	
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Write configuration file
	configPath := filepath.Join(configDir, "config.yaml")
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}
	
	if err := os.WriteFile(configPath, yamlData, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	fmt.Printf("✅ Configuration saved to: %s\n", configPath)
	return nil
}

func getConfigDir() (string, error) {
	// Check for XDG_CONFIG_HOME first
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "gira"), nil
	}
	
	// Fallback to ~/.config/gira or ~/.gira depending on OS
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	// On Unix-like systems, prefer ~/.config/gira
	configDir := filepath.Join(homeDir, ".config", "gira")
	if _, err := os.Stat(filepath.Join(homeDir, ".config")); err == nil {
		return configDir, nil
	}
	
	// Fallback to ~/.gira
	return filepath.Join(homeDir, ".gira"), nil
}

// outputResult is a temporary function - should be provided by root package
func outputResult(result interface{}) error {
	// This will need to be imported from a shared package or passed as dependency
	fmt.Printf("%+v\n", result)
	return nil
}