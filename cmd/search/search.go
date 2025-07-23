package search

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/lburgazzoli/gira/pkg/config"
	"github.com/lburgazzoli/gira/pkg/jira"
	stringutils "github.com/lburgazzoli/gira/pkg/utils/strings"
	tableutils "github.com/lburgazzoli/gira/pkg/utils/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	// pageSize defines the number of issues to fetch per page when using --all flag
	pageSize = 100
)

var (
	searchMaxResults int
	searchStartAt    int
	searchAll        bool
)

// searchFields defines the JIRA fields to retrieve for search results
var searchFields = []string{"summary", "status", "assignee", "reporter", "issuetype"}

// SearchCmd holds the configuration and client for search operations
type SearchCmd struct {
	cfg    *config.Config
	client *jira.Client
}

// execute performs the search operation
func (s *SearchCmd) execute(cmd *cobra.Command, jql string) error {
	var result *jira.SearchResult
	var err error

	if searchAll {
		result, err = s.searchAllIssues(jql)
		if err != nil {
			return fmt.Errorf("failed to search all issues: %w", err)
		}
	} else {
		result, err = s.client.SearchIssues(jql, searchStartAt, searchMaxResults, searchFields)
		if err != nil {
			return fmt.Errorf("failed to search issues: %w", err)
		}
	}

	return s.outputSearchResult(cmd, result)
}

var Cmd = &cobra.Command{
	Use:   "search JQL",
	Short: "Search JIRA issues using JQL",
	Long: `Search JIRA issues using JQL (JIRA Query Language).

Examples:
  gira search "project = PROJ"
  gira search "assignee = currentUser() AND status = 'In Progress'"
  gira search "created >= -7d" --max-results 50
  gira search "project = PROJ" --all
  gira search "project = PROJ" --output csv`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	Cmd.Flags().IntVar(&searchMaxResults, "max-results", pageSize, "Maximum number of results to return")
	Cmd.Flags().IntVar(&searchStartAt, "start-at", 0, "Starting index for pagination")
	Cmd.Flags().BoolVar(&searchAll, "all", false, "Retrieve all results by automatically handling pagination")

	// Override the global output flag to include CSV for search command
	Cmd.Flags().StringP("output", "o", "", "output format (table|json|yaml|csv)")
}

func runSearch(cmd *cobra.Command, args []string) error {
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

	searchCmd := &SearchCmd{
		cfg:    cfg,
		client: client,
	}

	return searchCmd.execute(cmd, args[0])
}

// searchAllIssues fetches all issues by automatically handling pagination
func (s *SearchCmd) searchAllIssues(jql string) (*jira.SearchResult, error) {
	var allIssues []jira.Issue
	startAt := 0
	totalCount := 0

	for {
		// Create a temporary search result to get a page of data
		tempResult, err := s.client.SearchIssues(jql, startAt, pageSize, searchFields)
		if err != nil {
			return nil, err
		}

		// Store total count from first request
		if startAt == 0 {
			totalCount = tempResult.Total
		}

		// Add issues from this page to our collection
		allIssues = append(allIssues, tempResult.Issues...)

		// Check if we've retrieved all issues
		if len(tempResult.Issues) < pageSize || startAt+len(tempResult.Issues) >= tempResult.Total {
			break
		}

		// Move to next page
		startAt += len(tempResult.Issues)
	}

	// Return a single SearchResult with all issues
	return &jira.SearchResult{
		Issues:     allIssues,
		StartAt:    0,
		MaxResults: len(allIssues),
		Total:      totalCount,
	}, nil
}

// buildTableRow creates a table row with hardcoded field order
func (s *SearchCmd) buildTableRow(issue *jira.Issue) []any {
	assignee := "Unassigned"
	if issue.Fields.Assignee != nil {
		assignee = issue.Fields.Assignee.DisplayName
	}

	return []any{
		issue.Key,
		issue.Fields.IssueType.Name,
		s.cfg.JIRA.BaseURL + "/browse/" + issue.Key,
		stringutils.Truncate(issue.Fields.Summary, 60),
		issue.Fields.Status.Name,
		assignee,
		issue.Fields.Reporter.DisplayName,
	}
}

// buildCSVRow creates a CSV row with hardcoded field order
func (s *SearchCmd) buildCSVRow(issue *jira.Issue) []string {
	assignee := "Unassigned"
	if issue.Fields.Assignee != nil {
		assignee = issue.Fields.Assignee.DisplayName
	}

	return []string{
		issue.Key,
		issue.Fields.IssueType.Name,
		s.cfg.JIRA.BaseURL + "/browse/" + issue.Key,
		issue.Fields.Summary,
		issue.Fields.Status.Name,
		assignee,
		issue.Fields.Reporter.DisplayName,
	}
}

func (s *SearchCmd) outputSearchResult(cmd *cobra.Command, result *jira.SearchResult) error {
	// Check local flag first, then fall back to global flag
	outputFormat, _ := cmd.Flags().GetString("output")
	if outputFormat == "" {
		outputFormat, _ = cmd.Root().PersistentFlags().GetString("output")
	}

	switch outputFormat {
	case "json":
		return outputResult(cmd, result)
	case "yaml":
		return outputResult(cmd, result)
	case "csv":
		return s.outputSearchCSV(result)
	case "table":
		return s.outputSearchTable(result)
	case "":
		// Default to table format for search results
		return s.outputSearchTable(result)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

func (s *SearchCmd) outputSearchTable(result *jira.SearchResult) error {
	if len(result.Issues) == 0 {
		fmt.Println("No issues found.")
		return nil
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	headers := []string{"KEY", "TYPE", "URL", "SUMMARY", "STATUS", "ASSIGNEE", "REPORTER"}
	renderer := tableutils.NewRenderer(
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
		}),
	)

	rows := make([][]any, 0, len(result.Issues))
	for _, issue := range result.Issues {
		row := s.buildTableRow(&issue)
		rows = append(rows, row)
	}

	if err := renderer.AppendAll(rows); err != nil {
		return err
	}

	if err := renderer.Render(); err != nil {
		return err
	}

	// Print pagination info
	fmt.Printf("\nShowing %d-%d of %d issues\n",
		result.StartAt+1,
		result.StartAt+len(result.Issues),
		result.Total)

	if result.StartAt+len(result.Issues) < result.Total {
		nextStart := result.StartAt + len(result.Issues)
		fmt.Printf("Use --start-at %d to see next page\n", nextStart)
	}

	return nil
}

func (s *SearchCmd) outputSearchCSV(result *jira.SearchResult) error {
	if len(result.Issues) == 0 {
		fmt.Println("No issues found.")
		return nil
	}

	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write CSV header
	headers := []string{"KEY", "TYPE", "URL", "SUMMARY", "STATUS", "ASSIGNEE", "REPORTER"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write CSV rows
	for _, issue := range result.Issues {
		row := s.buildCSVRow(&issue)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	// Print pagination info to stderr so it doesn't interfere with CSV output
	fmt.Printf("\nShowing %d-%d of %d issues\n",
		result.StartAt+1,
		result.StartAt+len(result.Issues),
		result.Total)

	if result.StartAt+len(result.Issues) < result.Total {
		nextStart := result.StartAt + len(result.Issues)
		fmt.Printf("Use --start-at %d to see next page\n", nextStart)
	}

	return nil
}

// outputResult handles JSON and YAML output formats
func outputResult(cmd *cobra.Command, result interface{}) error {
	// Check local flag first, then fall back to global flag
	outputFormat, _ := cmd.Flags().GetString("output")
	if outputFormat == "" {
		outputFormat, _ = cmd.Root().PersistentFlags().GetString("output")
	}

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		return encoder.Encode(result)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}
