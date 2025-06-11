package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	JIRA JIRAConfig `mapstructure:"jira"`
	AI   AIConfig   `mapstructure:"ai"`
	CLI  CLIConfig  `mapstructure:"cli"`
}

type JIRAConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Token   string `mapstructure:"token"`
}

type AIConfig struct {
	Provider string            `mapstructure:"provider"`
	Models   map[string]string `mapstructure:"models"`
	APIKey   string            `mapstructure:"api_key"`
}

type CLIConfig struct {
	OutputFormat string `mapstructure:"output_format"`
	Color        bool   `mapstructure:"color"`
	Verbose      bool   `mapstructure:"verbose"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}
	v.AddConfigPath(configDir)
	v.AddConfigPath(".")

	v.SetEnvPrefix("GIRA")
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
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

func setDefaults(v *viper.Viper) {
	v.SetDefault("cli.output_format", "table")
	v.SetDefault("cli.color", true)
	v.SetDefault("cli.verbose", false)
	v.SetDefault("ai.provider", "google")
}
