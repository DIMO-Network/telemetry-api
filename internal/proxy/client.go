package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// AuthHeaderKey is the context key for the raw Authorization header value.
type AuthHeaderKey struct{}

type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type gqlError struct {
	Message string `json:"message"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []gqlError      `json:"errors"`
}

// Client is a thin GraphQL HTTP client for forwarding requests to the dq service.
type Client struct {
	endpoint string
	http     *http.Client
}

// NewClient creates a new proxy client targeting the given GraphQL endpoint URL.
func NewClient(endpoint string) *Client {
	return &Client{endpoint: endpoint, http: &http.Client{}}
}

// Execute posts a GraphQL query to dq, forwarding the Authorization header from ctx.
// It returns the raw JSON "data" object on success.
func (c *Client) Execute(ctx context.Context, query string, variables map[string]any) (json.RawMessage, error) {
	body, err := json.Marshal(gqlRequest{Query: query, Variables: variables})
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if auth, _ := ctx.Value(AuthHeaderKey{}).(string); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	var gqlResp gqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("dq error: %s", gqlResp.Errors[0].Message)
	}
	return gqlResp.Data, nil
}
