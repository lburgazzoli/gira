package tree

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lburgazzoli/gira/pkg/utils/table"

	"github.com/fatih/color"
	"github.com/lburgazzoli/gira/pkg/config"
	"github.com/lburgazzoli/gira/pkg/jira"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	depth   int
	showAll bool
	reverse bool
)

var Cmd = &cobra.Command{
	Use:   "tree ISSUE-KEY",
	Short: "Display issue hierarchy as a tree",
	Long: `Display the issue hierarchy starting from the specified issue key.
Shows parent issues, the specified issue, and all subtasks in a tree format.`,
	Args:         cobra.ExactArgs(1),
	RunE:         runTree,
	SilenceUsage: true,
}

func init() {
	Cmd.Flags().IntVarP(&depth, "depth", "d", 3, "Maximum depth to traverse")
	Cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all fields for each issue")
	Cmd.Flags().BoolVarP(&reverse, "reverse", "r", false, "Show children first, then parents")
}

func runTree(cmd *cobra.Command, args []string) error {
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

	// Get the root issue
	rootIssue, err := client.GetIssue(issueKey)
	if err != nil {
		return fmt.Errorf("failed to get issue %s: %w", issueKey, err)
	}

	// Build the complete tree
	err = jira.BuildIssueTree(client, rootIssue, depth)
	if err != nil {
		return fmt.Errorf("failed to build issue tree: %w", err)
	}

	// Get output format from global flag
	outputFormat, _ := cmd.Root().PersistentFlags().GetString("output")

	switch outputFormat {
	case "json", "yaml":
		return outputResult(cmd, rootIssue)
	case "table":
		return renderTable(rootIssue)
	default:
		// Render as ASCII tree (this is the default for tree command)
		if reverse {
			renderTreeReverse(rootIssue, "", 0, true)
		} else {
			renderTree(rootIssue, "", 0, true)
		}
		return nil
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
	if showAll {
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
	return "Unassigned"
}

func renderTable(rootIssue *jira.Issue) error {
	headers := []string{"Key", "Type", "Summary", "Status", "Assignee"}
	if showAll {
		headers = []string{"Key", "Type", "Summary", "Status", "Priority", "Assignee", "Created", "Updated"}
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	t := table.NewRenderer(
		table.WithHeaders(headers...),
		table.WithFormatter("STATUS", func(value interface{}) any {
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

	collectTableRowsRecursively(rootIssue, &rows, 0)

	// Add all collected rows to the table
	for i := range rows {
		if err := t.Append(rows[i]); err != nil {
			return err
		}

	}

	// Render the table
	if err := t.Render(); err != nil {
		return err
	}

	return nil
}

func collectTableRowsRecursively(issue *jira.Issue, rows *[][]any, depth int) {
	if issue == nil {
		return
	}

	// First add parent chain (if any)
	if issue.Parent != nil {
		collectTableRowsRecursively(issue.Parent, rows, depth-1)
	}

	// Add current issue row
	row := buildTableRow(issue, depth)
	*rows = append(*rows, row)

	// Add children recursively with incremented depth
	for _, child := range issue.Children {
		collectTableRowsRecursively(child, rows, depth+1)
	}
}

func buildTableRow(issue *jira.Issue, depth int) []any {
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
		// Child levels - use tree indicators
		// Build the prefix based on depth
		var prefix string
		if depth == 1 {
			prefix = "├── "
		} else {
			// For deeper levels, add proper indentation
			prefix = strings.Repeat("│   ", depth-1) + "├── "
		}
		idColumn = prefix + issue.Key
	}

	if showAll {
		return []any{
			idColumn,
			issue.Fields.IssueType.Name,
			truncateString(issue.Fields.Summary, 40),
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
		truncateString(issue.Fields.Summary, 50),
		issue.Fields.Status.Name,
		getAssigneeDisplay(issue),
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func outputResult(cmd *cobra.Command, result interface{}) error {
	// Get the actual output format to determine JSON vs YAML
	outputFormat, _ := cmd.Root().PersistentFlags().GetString("output")

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		return encoder.Encode(result)
	default:
		// Fallback to JSON if unclear
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}
}
