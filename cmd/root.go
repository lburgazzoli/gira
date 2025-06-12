package cmd

import (
	"fmt"
	"os"

	"github.com/lburgazzoli/gira/cmd/config"
	"github.com/lburgazzoli/gira/cmd/get"
	versionCmd "github.com/lburgazzoli/gira/cmd/version"
	"github.com/lburgazzoli/gira/internal/version"
	pkgConfig "github.com/lburgazzoli/gira/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *pkgConfig.Config
)

var rootCmd = &cobra.Command{
	Use:   "gira",
	Short: "An AI-powered Jira CLI tool",
	Long: `Gira is a command-line interface tool that provides AI-enhanced 
interaction with Atlassian JIRA through REST APIs. It combines traditional 
JIRA operations with AI-powered features for issue analysis, explanation, 
and intelligent updates.`,
	Version: version.GetVersion(),
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gira/config.yaml)")
	rootCmd.PersistentFlags().StringP("output", "o", "", "output format (table|json|yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(config.Cmd)
	rootCmd.AddCommand(get.Cmd)
	rootCmd.AddCommand(versionCmd.Cmd)
}

func initConfig() {
	var err error
	cfg, err = pkgConfig.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
}
