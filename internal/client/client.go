// Package client provides a GraphQL client for the MarketSurge API.
//
// This project is unofficial and is not affiliated with, approved by, or
// endorsed by MarketSurge or Investor's Business Daily.
package client

import (
	"context"
	"fmt"
	"mime"
	"net/http"

	"github.com/major/marketsurge-agent/internal/constants"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"resty.dev/v3"
)

// Request is a GraphQL request payload.
type Request struct {
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
	Query         string         `json:"query"`
}

// Client executes MarketSurge GraphQL requests.
type Client struct {
	JWT         string
	BaseURL     string
	RestyClient *resty.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL sets the GraphQL endpoint URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.BaseURL = url
	}
}

// WithRestyClient sets the Resty client used for requests.
func WithRestyClient(restyClient *resty.Client) Option {
	return func(c *Client) {
		c.RestyClient = restyClient
	}
}

// NewClient constructs a GraphQL client with the default endpoint and timeout.
// Use With* options to override defaults.
func NewClient(jwt string, opts ...Option) *Client {
	c := &Client{
		JWT:     jwt,
		BaseURL: constants.GraphQLEndpoint,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.RestyClient == nil {
		c.RestyClient = resty.New().
			SetTimeout(constants.HTTPTimeout).
			SetResponseBodyLimit(constants.MaxResponseSize)
	}
	c.RestyClient.SetBaseURL(c.BaseURL)
	return c
}

// Close releases resources held by the underlying Resty client.
func (c *Client) Close() error {
	if c.RestyClient != nil {
		return c.RestyClient.Close()
	}
	return nil
}

// Execute sends a GraphQL request and returns the decoded response body.
func (c *Client) Execute(ctx context.Context, payload Request) (map[string]any, error) {
	var raw map[string]any
	req := c.RestyClient.R().SetContext(ctx).SetResult(&raw)

	headers := constants.GraphQLHeaders()
	for key, values := range headers {
		for _, v := range values {
			req.SetHeader(key, v)
		}
	}
	req.SetHeader("Authorization", "Bearer "+c.JWT)
	req.SetBody(payload)

	resp, err := req.Post("")
	if err != nil {
		return nil, fmt.Errorf("execute graphql request: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, mapHTTPError(resp.StatusCode(), string(resp.Bytes()), nil)
	}

	// Validate Content-Type before attempting JSON decode. Without this,
	// an HTML error page from a proxy or maintenance window produces a
	// cryptic json.Unmarshal error instead of a clear diagnostic.
	if ct := resp.Header().Get("Content-Type"); ct != "" {
		mediaType, _, err := mime.ParseMediaType(ct)
		if err == nil && mediaType != "application/json" {
			preview := string(resp.Bytes())
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			return nil, fmt.Errorf("unexpected Content-Type %q (expected application/json): %s", ct, preview)
		}
	}

	if err := mapGraphQLError(raw); err != nil {
		return raw, err
	}

	return raw, nil
}

func mapHTTPError(statusCode int, body string, cause error) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return mserrors.NewTokenExpiredError("JWT token has expired or is invalid (HTTP 401)", cause, statusCode)
	case http.StatusForbidden:
		return mserrors.NewAuthenticationError("access denied, token may lack required permissions (HTTP 403)", cause)
	default:
		message := fmt.Sprintf("HTTP error %d", statusCode)
		if statusCode == http.StatusTooManyRequests {
			message = "rate limited, retry after a delay"
		} else if statusCode >= http.StatusInternalServerError {
			message = "MarketSurge server error"
		}
		return mserrors.NewHTTPError(message, cause, statusCode, body)
	}
}
