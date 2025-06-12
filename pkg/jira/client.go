package jira

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	// Content type for JSON requests
	contentTypeJSON = "application/json"
	// Header names
	headerAuthorization = "Authorization"
	headerContentType   = "Content-Type"
	headerAccept        = "Accept"
	
	// JIRA API endpoints
	apiIssueEndpoint   = "/rest/api/2/issue/%s"
	apiCreateEndpoint  = "/rest/api/2/issue"
	apiSearchEndpoint  = "/rest/api/2/search"
	apiProjectEndpoint = "/rest/api/2/project/%s"
	
	// URL prefixes
	httpPrefix  = "http://"
	httpsPrefix = "https://"
)

type Client struct {
	baseURL         string
	retryableClient *retryablehttp.Client
	auth            authConfig
}

type authConfig struct {
	token string
}

type AuthConfig struct {
	Token string
}

func NewClient(baseURL string, auth AuthConfig) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL cannot be empty")
	}
	if auth.Token == "" {
		return nil, fmt.Errorf("API token cannot be empty")
	}

	baseURL = strings.TrimSuffix(baseURL, "/")
	if !strings.HasPrefix(baseURL, httpPrefix) && !strings.HasPrefix(baseURL, httpsPrefix) {
		baseURL = httpsPrefix + baseURL
	}

	// Create retryable HTTP client with rate limiting
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.Backoff = retryablehttp.DefaultBackoff
	retryClient.Logger = nil // Disable debug logging

	// Configure which HTTP status codes to retry
	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		// Default retry logic for network errors
		if err != nil {
			return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
		}

		// Retry on rate limiting
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			return true, nil
		}

		// Use default retry policy for other cases
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	return &Client{
		baseURL:         baseURL,
		retryableClient: retryClient,
		auth: authConfig{
			token: auth.Token,
		},
	}, nil
}


// JIRA Operations

func (c *Client) GetIssue(key string) (*Issue, error) {
	resp, err := c.get(fmt.Sprintf(apiIssueEndpoint, key))
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	var issue Issue
	if err := handleResponse(resp, &issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

func (c *Client) CreateIssue(issue *Issue) (*Issue, error) {
	resp, err := c.post(apiCreateEndpoint, issue)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	var createdIssue Issue
	if err := handleResponse(resp, &createdIssue); err != nil {
		return nil, err
	}

	return &createdIssue, nil
}

func (c *Client) UpdateIssue(key string, update IssueUpdate) (*Issue, error) {
	resp, err := c.put(fmt.Sprintf(apiIssueEndpoint, key), update)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}

	if err := handleResponse(resp, nil); err != nil {
		return nil, err
	}

	return c.GetIssue(key)
}

func (c *Client) SearchIssues(jql string, fields ...string) (*SearchResult, error) {
	var params []Parameter
	params = append(params, Parameter{Key: "jql", Value: jql})
	
	if len(fields) > 0 {
		for _, field := range fields {
			params = append(params, Parameter{Key: "fields", Value: field})
		}
	}

	resp, err := c.get(apiSearchEndpoint, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	var result SearchResult
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetProject(key string) (*Project, error) {
	resp, err := c.get(fmt.Sprintf(apiProjectEndpoint, key))
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	var project Project
	if err := handleResponse(resp, &project); err != nil {
		return nil, err
	}

	return &project, nil
}
