package jira

import (
	"fmt"
	"time"
)

// JIRATime represents a timestamp in JIRA's format
type JIRATime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler for JIRATime
func (jt *JIRATime) UnmarshalJSON(data []byte) error {
	// Remove quotes from JSON string
	s := string(data)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	// JIRA format: "2025-05-12T06:54:41.542+0000"
	const jiraLayout = "2006-01-02T15:04:05.000-0700"

	t, err := time.Parse(jiraLayout, s)
	if err != nil {
		return fmt.Errorf("failed to parse JIRA timestamp %q: %w", s, err)
	}

	jt.Time = t
	return nil
}

// Format returns the time formatted for display
func (jt *JIRATime) Format(layout string) string {
	return jt.Time.Format(layout)
}

// Format returns the time formatted for display
func (jt *JIRATime) String() string {
	return jt.Format("2006-01-02")
}

type Issue struct {
	Key         string       `json:"key"`
	ID          string       `json:"id"`
	Self        string       `json:"self"`
	Fields      IssueFields  `json:"fields"`
	Transitions []Transition `json:"transitions,omitempty"`

	// Tree hierarchy fields (populated during tree traversal)
	Parent   *Issue   `json:"parent,omitempty"`
	Children []*Issue `json:"children,omitempty"`
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
	Created     JIRATime    `json:"created"`
	Updated     JIRATime    `json:"updated"`
}

type IssueType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
}

func (in IssueType) String() string {
	return in.Name
}

type Status struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (in Status) String() string {
	return in.Name
}

type Priority struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IconURL string `json:"iconUrl"`
}

func (in Priority) String() string {
	return in.Name
}

type User struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type IssueLink struct {
	ID           string   `json:"id"`
	Type         LinkType `json:"type"`
	InwardIssue  *Issue   `json:"inwardIssue,omitempty"`
	OutwardIssue *Issue   `json:"outwardIssue,omitempty"`
}

type LinkType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	To   Status `json:"to"`
}

type SearchResult struct {
	Issues     []Issue `json:"issues"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
}

type IssueUpdate struct {
	Fields map[string]interface{} `json:"fields,omitempty"`
	Update map[string]interface{} `json:"update,omitempty"`
}
