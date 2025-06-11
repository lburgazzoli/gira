package get

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lburgazzoli/gira/pkg/config"
	"github.com/lburgazzoli/gira/pkg/jira"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var Cmd = &cobra.Command{
	Use:   "get",
	Short: "Get JIRA resources",
	Long:  `Get JIRA resources like issues, projects, and users.`,
}

var issueCmd = &cobra.Command{
	Use:   "issue ISSUE-KEY",
	Short: "Get a JIRA issue",
	Long:  `Get details of a specific JIRA issue by its key.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGetIssue,
}

var projectCmd = &cobra.Command{
	Use:   "project PROJECT-KEY",
	Short: "Get a JIRA project",
	Long:  `Get details of a specific JIRA project by its key.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGetProject,
}

func init() {
	Cmd.AddCommand(issueCmd)
	Cmd.AddCommand(projectCmd)
}

func runGetIssue(cmd *cobra.Command, args []string) error {
	issueKey := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client, err := jira.NewClient(cfg.JIRA.BaseURL, jira.AuthConfig{
		Token: cfg.JIRA.Token,
	})
	if err != nil {
		return fmt.Errorf("failed to create JIRA client: %w", err)
	}

	issue, err := client.GetIssue(issueKey)
	if err != nil {
		return fmt.Errorf("failed to get issue %s: %w", issueKey, err)
	}

	return outputResult(issue)
}

func runGetProject(cmd *cobra.Command, args []string) error {
	projectKey := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client, err := jira.NewClient(cfg.JIRA.BaseURL, jira.AuthConfig{
		Token: cfg.JIRA.Token,
	})
	if err != nil {
		return fmt.Errorf("failed to create JIRA client: %w", err)
	}

	project, err := client.GetProject(projectKey)
	if err != nil {
		return fmt.Errorf("failed to get project %s: %w", projectKey, err)
	}

	return outputResult(project)
}

func outputResult(result interface{}) error {
	// Get output format from global flag
	outputFormat, _ := Cmd.Root().PersistentFlags().GetString("output")
	
	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		return encoder.Encode(result)
	case "table":
		return outputTable(result)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// createTable creates a new table with consistent formatting options
func createTable() *tablewriter.Table {
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
	return table
}

func outputTable(result interface{}) error {
	switch v := result.(type) {
	case *jira.Issue:
		table := createTable()
		table.Header("Field", "Value")

		table.Append([]string{"Key", v.Key})
		table.Append([]string{"Summary", v.Fields.Summary})
		table.Append([]string{"Status", v.Fields.Status.Name})
		table.Append([]string{"Type", v.Fields.IssueType.Name})
		table.Append([]string{"Priority", v.Fields.Priority.Name})
		table.Append([]string{"Project", v.Fields.Project.Name})

		assignee := "Unassigned"
		if v.Fields.Assignee != nil {
			assignee = v.Fields.Assignee.DisplayName
		}
		table.Append([]string{"Assignee", assignee})
		table.Append([]string{"Reporter", v.Fields.Reporter.DisplayName})
		table.Append([]string{"Created", v.Fields.Created.Format("2006-01-02 15:04:05")})
		table.Append([]string{"Updated", v.Fields.Updated.Format("2006-01-02 15:04:05")})

		if v.Fields.Description != "" {
			// Truncate description for table display
			desc := v.Fields.Description
			if len(desc) > 100 {
				desc = desc[:97] + "..."
			}
			table.Append([]string{"Description", desc})
		}

		return table.Render()

	case *jira.Project:
		table := createTable()
		table.Header("Field", "Value")

		table.Append([]string{"Key", v.Key})
		table.Append([]string{"Name", v.Name})
		table.Append([]string{"ID", v.ID})

		return table.Render()

	case *config.Config:
		table := createTable()
		table.Header("Configuration", "Value")

		table.Append([]string{"JIRA Base URL", v.JIRA.BaseURL})
		table.Append([]string{"JIRA Token", v.JIRA.Token})
		table.Append([]string{"AI Provider", v.AI.Provider})
		table.Append([]string{"AI API Key", v.AI.APIKey})
		table.Append([]string{"Output Format", v.CLI.OutputFormat})
		table.Append([]string{"Colors Enabled", fmt.Sprintf("%t", v.CLI.Color)})
		table.Append([]string{"Verbose Mode", fmt.Sprintf("%t", v.CLI.Verbose)})

		if len(v.AI.Models) > 0 {
			for key, model := range v.AI.Models {
				table.Append([]string{fmt.Sprintf("AI Model (%s)", key), model})
			}
		}

		return table.Render()

	default:
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonBytes))
	}
	return nil
}