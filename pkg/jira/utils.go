package jira

import (
	"fmt"
)

var (
	childrenSearchFields = []string{
		"summary",
		"status",
		"issuetype",
		"priority",
		"assignee",
		"reporter",
		"created",
		"updated",
		"parent",
	}
)

func GetChildIssues(client *Client, parentIssue *Issue) ([]*Issue, error) {
	var children []*Issue

	// Add subtasks (direct children)
	for _, subtask := range parentIssue.Fields.Subtasks {
		child, err := client.GetIssue(subtask.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to get subtask %s: %w", subtask.Key, err)
		}
		children = append(children, child)
	}

	// Search for child issues using different JQL queries based on issue type and relationships
	var jqlQueries []string

	// For epics, search for issues that have this epic as their Epic Link
	if parentIssue.Fields.IssueType.Name == "Epic" {
		jqlQueries = append(jqlQueries, fmt.Sprintf("\"Epic Link\" = %s", parentIssue.Key))
	}

	// Always search for direct parent-child relationships
	jqlQueries = append(jqlQueries, fmt.Sprintf("parent = %s", parentIssue.Key))

	// For non-epic issues that might also have epic relationships
	// (e.g., a Story that has both subtasks and is linked to an epic)
	// We should also search for issues that have this issue as their Epic Link
	// This handles cases where any issue type can act as an "epic" in some workflows
	if parentIssue.Fields.IssueType.Name != "Epic" {
		jqlQueries = append(jqlQueries, fmt.Sprintf("\"Epic Link\" = %s", parentIssue.Key))
	}

	// Collect all found issues
	foundKeys := make(map[string]bool)

	// Add subtasks to found keys to avoid duplicates
	for _, subtask := range parentIssue.Fields.Subtasks {
		foundKeys[subtask.Key] = true
	}

	// Execute each JQL query
	for _, jql := range jqlQueries {
		result, err := client.SearchIssues(jql, childrenSearchFields)
		if err != nil {
			return nil, fmt.Errorf("JQL search failed for '%s': %v\n", jql, err)
		}

		// Add issues found via JQL search (avoiding duplicates)
		for _, issue := range result.Issues {
			if !foundKeys[issue.Key] {
				children = append(children, &issue)
				foundKeys[issue.Key] = true
			}
		}
	}

	return children, nil
}

func BuildIssueTree(client *Client, issue *Issue, maxDepth int) error {
	// Initialize issue tree fields
	issue.Children = make([]*Issue, 0)

	// Get children (both subtasks and child issues) if we haven't reached max depth
	if maxDepth <= 0 {
		return nil
	}

	children, err := GetChildIssues(client, issue)
	if err != nil {
		return fmt.Errorf("failed to get child issues for %s: %w", issue.Key, err)
	}

	for _, child := range children {
		err := BuildIssueTree(client, child, maxDepth-1)
		if err != nil {
			return err
		}

		issue.Children = append(issue.Children, child)
	}

	return nil
}
