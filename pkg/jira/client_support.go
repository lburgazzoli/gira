package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
)

// Parameter represents a URL query parameter
type Parameter struct {
	Key   string
	Value string
}

// HTTP helper methods

func (c *Client) get(endpoint string, params ...Parameter) (*http.Response, error) {
	return c.doRequest(http.MethodGet, endpoint, nil, params...)
}

func (c *Client) post(endpoint string, body interface{}) (*http.Response, error) {
	reqBody, err := marshalBody(body)
	if err != nil {
		return nil, err
	}

	return c.doRequest(http.MethodPost, endpoint, reqBody)
}

func (c *Client) put(endpoint string, body interface{}) (*http.Response, error) {
	reqBody, err := marshalBody(body)
	if err != nil {
		return nil, err
	}

	return c.doRequest(http.MethodPut, endpoint, reqBody)
}

func (c *Client) delete(endpoint string) (*http.Response, error) {
	return c.doRequest(http.MethodDelete, endpoint, nil)
}

// doRequest creates and executes an HTTP request with proper authentication and headers
func (c *Client) doRequest(method string, endpoint string, body io.Reader, params ...Parameter) (*http.Response, error) {
	requestURL, err := url.JoinPath(c.baseURL, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	if len(params) > 0 {
		urlParams := url.Values{}
		for _, param := range params {
			urlParams.Add(param.Key, param.Value)
		}
		requestURL += "?" + urlParams.Encode()
	}

	req, err := retryablehttp.NewRequest(method, requestURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication and headers
	req.Request.Header.Set(headerAuthorization, "Bearer "+c.auth.token)
	req.Request.Header.Set(headerContentType, contentTypeJSON)
	req.Request.Header.Set(headerAccept, contentTypeJSON)

	return c.retryableClient.Do(req)
}

// marshalBody converts an interface{} to an io.Reader for request body
func marshalBody(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	return bytes.NewBuffer(jsonBody), nil
}

func handleResponse(resp *http.Response, v interface{}) error {
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w. Raw response: %s", err, string(body))
		}
	}

	return nil
}
