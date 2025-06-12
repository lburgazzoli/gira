package get

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/lburgazzoli/gira/pkg/config"
	"github.com/lburgazzoli/gira/pkg/jira"
	stringutils "github.com/lburgazzoli/gira/pkg/utils/strings"
	tableutils "github.com/lburgazzoli/gira/pkg/utils/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var Cmd = &cobra.Command{
	Use:   "get",
	Short: "Get JIRA resources",
	Long:  `Get JIRA resources like issues, projects, and users.`,
}

var (
	treeFlag    bool
	treeDepth   int
	treeReverse bool
	treeShowAll bool
)

var issueCmd = &cobra.Command{
	Use:   "issue ISSUE-KEY",
	Short: "Get a JIRA issue",
	Long:  `Get details of a specific JIRA issue by its key. Use --tree to show issue hierarchy.`,
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
	issueCmd.Flags().BoolVar(&treeFlag, "tree", false, "Display issue hierarchy as a tree")
	issueCmd.Flags().IntVar(&treeDepth, "tree-depth", 3, "Maximum depth to traverse for tree view")
	issueCmd.Flags().BoolVar(&treeReverse, "tree-reverse", false, "Show children first, then parents in tree view")
	issueCmd.Flags().BoolVar(&treeShowAll, "tree-all", false, "Show all fields for each issue in tree view")

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

	if treeFlag {
		// Build the complete tree
		err = jira.BuildIssueTree(client, issue, treeDepth)
		if err != nil {
			return fmt.Errorf("failed to build issue tree: %w", err)
		}
		return outputTreeResult(cmd, issue)
	}

	return outputResult(cmd, issue)
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

	return outputResult(cmd, project)
}

func outputResult(cmd *cobra.Command, result interface{}) error {
	// Get output format from global flag
	outputFormat, _ := cmd.Root().PersistentFlags().GetString("output")

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
	case "":
		// Default to plain format for issues, table for other types
		return outputPlain(result)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

func outputTable(result interface{}) error {
	switch v := result.(type) {
	case *jira.Issue:
		renderer := tableutils.NewRenderer(
			tableutils.WithHeaders("Field", "Value"),
		)

		assignee := "Unassigned"
		if v.Fields.Assignee != nil {
			assignee = v.Fields.Assignee.DisplayName
		}

		rows := [][]any{
			{"Key", v.Key},
			{"Summary", v.Fields.Summary},
			{"Status", v.Fields.Status.Name},
			{"Type", v.Fields.IssueType.Name},
			{"Priority", v.Fields.Priority.Name},
			{"Project", v.Fields.Project.Name},
			{"Assignee", assignee},
			{"Reporter", v.Fields.Reporter.DisplayName},
			{"Created", v.Fields.Created.Format("2006-01-02 15:04:05")},
			{"Updated", v.Fields.Updated.Format("2006-01-02 15:04:05")},
			{"Description", stringutils.Truncate(v.Fields.Description, 100)},
		}

		if err := renderer.AppendAll(rows); err != nil {
			return err
		}

		return renderer.Render()

	case *jira.Project:
		renderer := tableutils.NewRenderer(
			tableutils.WithHeaders("Field", "Value"),
		)

		rows := [][]any{
			{"Key", v.Key},
			{"Name", v.Name},
			{"ID", v.ID},
		}

		if err := renderer.AppendAll(rows); err != nil {
			return err
		}

		return renderer.Render()

	case *config.Config:
		renderer := tableutils.NewRenderer(
			tableutils.WithHeaders("Configuration", "Value"),
		)

		rows := [][]any{
			{"JIRA Base URL", v.JIRA.BaseURL},
			{"JIRA Token", v.JIRA.Token},
			{"AI Provider", v.AI.Provider},
			{"AI API Key", v.AI.APIKey},
			{"Output Format", v.CLI.OutputFormat},
			{"Colors Enabled", fmt.Sprintf("%t", v.CLI.Color)},
			{"Verbose Mode", fmt.Sprintf("%t", v.CLI.Verbose)},
		}

		// Add AI model rows if they exist
		if len(v.AI.Models) > 0 {
			for key, model := range v.AI.Models {
				rows = append(rows, []any{fmt.Sprintf("AI Model (%s)", key), model})
			}
		}

		if err := renderer.AppendAll(rows); err != nil {
			return err
		}

		return renderer.Render()

	default:
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonBytes))
	}
	return nil
}

func outputPlain(result interface{}) error {
	switch v := result.(type) {
	case *jira.Issue:
		fmt.Printf("%-11s: %s\n", "Issue", v.Key)
		fmt.Printf("%-11s: %s\n", "Summary", v.Fields.Summary)
		fmt.Printf("%-11s: %s\n", "Status", v.Fields.Status.Name)
		fmt.Printf("%-11s: %s\n", "Type", v.Fields.IssueType.Name)
		fmt.Printf("%-11s: %s\n", "Priority", v.Fields.Priority.Name)
		fmt.Printf("%-11s: %s\n", "Project", v.Fields.Project.Name)

		if v.Fields.Assignee != nil {
			fmt.Printf("%-11s: %s\n", "Assignee", v.Fields.Assignee.DisplayName)
		} else {
			fmt.Printf("%-11s: Unassigned\n", "Assignee")
		}

		fmt.Printf("%-11s: %s\n", "Reporter", v.Fields.Reporter.DisplayName)
		fmt.Printf("%-11s: %s\n", "Created", v.Fields.Created.Format("2006-01-02 15:04:05"))
		fmt.Printf("%-11s: %s\n", "Updated", v.Fields.Updated.Format("2006-01-02 15:04:05"))

		if v.Fields.Description != "" {
			fmt.Printf("\n")
			if err := stringutils.PrintWrapped(os.Stdout, v.Fields.Description, 100); err != nil {
				return err
			}
		}

	case *jira.Project:
		fmt.Printf("Project: %s\n", v.Key)
		fmt.Printf("Name: %s\n", v.Name)
		fmt.Printf("ID: %s\n", v.ID)

	default:
		// For other types, fall back to table format
		return outputTable(result)
	}
	return nil
}

func outputTreeResult(cmd *cobra.Command, issue *jira.Issue) error {
	// Get output format from global flag
	outputFormat, _ := cmd.Root().PersistentFlags().GetString("output")

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(issue)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		return encoder.Encode(issue)
	case "table":
		return renderTreeTable(issue)
	case "":
		// Default to ASCII tree format for tree view when no output format is specified
		if treeReverse {
			renderTreeReverse(issue, "", 0, true)
		} else {
			renderTree(issue, "", 0, true)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

func renderTree(issue *jira.Issue, prefix string, depth int, isLast bool) {
	if issue == nil {
		return
	}

	// Render current issue
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if depth <= 0 && issue.Parent == nil {
		connector = ""
	}

	issueInfo := formatIssueInfo(issue)
	fmt.Printf("%s%s%s\n", prefix, connector, issueInfo)

	// Prepare prefix for children
	childPrefix := prefix
	if depth <= 0 && issue.Parent == nil {
		childPrefix = ""
	} else if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	// Render children
	for i, child := range issue.Children {
		isLastChild := i == len(issue.Children)-1
		renderTree(child, childPrefix, depth+1, isLastChild)
	}
}

func renderTreeReverse(issue *jira.Issue, prefix string, depth int, isLast bool) {
	if issue == nil {
		return
	}

	// Render children first
	childPrefix := prefix
	if depth <= 0 {
		childPrefix = ""
	} else if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	for i, child := range issue.Children {
		isLastChild := i == len(issue.Children)-1
		renderTreeReverse(child, childPrefix, depth+1, isLastChild)
	}

	// Render current issue
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if depth <= 0 {
		connector = ""
	}

	issueInfo := formatIssueInfo(issue)
	fmt.Printf("%s%s%s\n", prefix, connector, issueInfo)

	// Render parent chain
	if issue.Parent != nil {
		renderTreeReverse(issue.Parent, "", depth-1, true)
	}
}

func formatIssueInfo(issue *jira.Issue) string {
	if treeShowAll {
		return fmt.Sprintf("%s: %s [%s] (%s) - %s",
			issue.Key,
			issue.Fields.Summary,
			issue.Fields.Status.Name,
			issue.Fields.IssueType.Name,
			getAssigneeDisplay(issue))
	}

	// Compact format
	status := ""
	if issue.Fields.Status.Name != "" {
		status = fmt.Sprintf(" [%s]", issue.Fields.Status.Name)
	}

	return fmt.Sprintf("%s: %s%s", issue.Key, issue.Fields.Summary, status)
}

func getAssigneeDisplay(issue *jira.Issue) string {
	if issue.Fields.Assignee != nil {
		return issue.Fields.Assignee.DisplayName
	}
	return ""
}

func renderTreeTable(rootIssue *jira.Issue) error {
	headers := []string{"Key", "Type", "Summary", "Status", "Assignee"}
	if treeShowAll {
		headers = []string{"Key", "Type", "Summary", "Status", "Priority", "Assignee", "Created", "Updated"}
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	t := tableutils.NewRenderer(
		tableutils.WithHeaders(headers...),
		tableutils.WithFormatter("STATUS", func(value interface{}) any {
			v := value.(string)

			switch v {
			case "Resolved":
				v = green(v)
			case "In Progress":
				v = blue(v)
			case "New":
				v = red(v)
			}

			return v
		}))

	// Collect all rows recursively, starting with root at depth 0
	rows := make([][]any, 0)

	collectTableRowsRecursively(rootIssue, &rows, 0, true)

	// Add all collected rows to the table using AppendAll
	if err := t.AppendAll(rows); err != nil {
		return err
	}

	// Render the table
	if err := t.Render(); err != nil {
		return err
	}

	return nil
}

func collectTableRowsRecursively(issue *jira.Issue, rows *[][]any, depth int, isLast bool) {
	if issue == nil {
		return
	}

	// First add parent chain (if any)
	if issue.Parent != nil {
		collectTableRowsRecursively(issue.Parent, rows, depth-1, true)
	}

	// Add current issue row
	row := buildTableRow(issue, depth, isLast)
	*rows = append(*rows, row)

	// Add children recursively with incremented depth
	for i, child := range issue.Children {
		isLastChild := i == len(issue.Children)-1
		collectTableRowsRecursively(child, rows, depth+1, isLastChild)
	}
}

func buildTableRow(issue *jira.Issue, depth int, isLast bool) []any {
	// Create merged ID column with the same hierarchy structure as ASCII tree
	var idColumn string

	// Use consistent logic with ASCII tree rendering
	if depth == 0 {
		// Root issue - no prefix
		idColumn = issue.Key
	} else if depth < 0 {
		// Parent levels - no tree indicators, just the key
		idColumn = issue.Key
	} else {
		// Child levels - use tree indicators matching ASCII tree
		// Build the prefix based on depth
		var prefix string
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		if depth == 1 {
			prefix = connector
		} else {
			// For deeper levels, add proper indentation
			prefix = strings.Repeat("│   ", depth-1) + connector
		}
		idColumn = prefix + issue.Key
	}

	if treeShowAll {
		return []any{
			idColumn,
			issue.Fields.IssueType.Name,
			stringutils.Truncate(issue.Fields.Summary, 40),
			issue.Fields.Status.Name,
			issue.Fields.Priority.Name,
			getAssigneeDisplay(issue),
			issue.Fields.Created.Format("2006-01-02"),
			issue.Fields.Updated.Format("2006-01-02"),
		}
	}

	return []any{
		idColumn,
		issue.Fields.IssueType.Name,
		stringutils.Truncate(issue.Fields.Summary, 50),
		issue.Fields.Status.Name,
		getAssigneeDisplay(issue),
	}
}
