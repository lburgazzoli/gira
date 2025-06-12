package jira

import (
	"fmt"
	"strings"
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
	// Pre-allocate children slice with estimated capacity
	estimatedCapacity := len(parentIssue.Fields.Subtasks) + 5 // subtasks + some JQL results
	children := make([]*Issue, 0, estimatedCapacity)

	// Collect all found issues to avoid duplicates
	foundKeys := make(map[string]bool)

	// Batch fetch subtasks using JQL if we have any
	if len(parentIssue.Fields.Subtasks) > 0 {
		subtaskKeys := make([]string, len(parentIssue.Fields.Subtasks))
		for i, subtask := range parentIssue.Fields.Subtasks {
			subtaskKeys[i] = subtask.Key
			foundKeys[subtask.Key] = true
		}

		// Use JQL to batch fetch subtasks
		subtaskJQL := fmt.Sprintf("key IN (%s)", strings.Join(subtaskKeys, ","))
		result, err := client.SearchIssues(subtaskJQL, childrenSearchFields...)
		if err != nil {
			return nil, fmt.Errorf("failed to batch fetch subtasks: %w", err)
		}

		for _, issue := range result.Issues {
			children = append(children, &issue)
		}
	}

	// Build combined JQL query for parent-child and Epic Link relationships
	// All issues can potentially have Epic Link relationships, simplifying the logic
	combinedJQL := fmt.Sprintf("parent = %s OR \"Epic Link\" = %s", parentIssue.Key, parentIssue.Key)

	// Execute the combined JQL query
	result, err := client.SearchIssues(combinedJQL, childrenSearchFields...)
	if err != nil {
		return nil, fmt.Errorf("JQL search failed for '%s': %w", combinedJQL, err)
	}

	// Add issues found via JQL search (avoiding duplicates)
	for _, issue := range result.Issues {
		if foundKeys[issue.Key] {
			continue
		}

		children = append(children, &issue)
		foundKeys[issue.Key] = true
	}

	return children, nil
}

func BuildIssueTree(client *Client, issue *Issue, maxDepth int) error {
	// Get children (both subtasks and child issues) if we haven't reached max depth
	if maxDepth <= 0 {
		// Initialize empty children slice
		issue.Children = make([]*Issue, 0)
		return nil
	}

	children, err := GetChildIssues(client, issue)
	if err != nil {
		return fmt.Errorf("failed to get child issues for %s: %w", issue.Key, err)
	}

	// Pre-allocate children slice with exact capacity
	issue.Children = make([]*Issue, 0, len(children))

	for _, child := range children {
		err := BuildIssueTree(client, child, maxDepth-1)
		if err != nil {
			return err
		}

		issue.Children = append(issue.Children, child)
	}

	return nil
}
