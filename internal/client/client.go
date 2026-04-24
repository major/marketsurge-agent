// Package client provides a GraphQL client for the MarketSurge API.
//
// This project is unofficial and is not affiliated with, approved by, or
// endorsed by MarketSurge or Investor's Business Daily.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/major/marketsurge-agent/internal/constants"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// Request is a GraphQL request payload.
type Request struct {
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
	Query         string         `json:"query"`
}

// Client executes MarketSurge GraphQL requests.
type Client struct {
	JWT        string
	Endpoint   string
	HTTPClient *http.Client
}

// NewClient constructs a GraphQL client with the default endpoint and timeout.
func NewClient(jwt string) *Client {
	return &Client{
		JWT:      jwt,
		Endpoint: constants.GraphQLEndpoint,
		HTTPClient: &http.Client{
			Timeout: constants.HTTPTimeout,
		},
	}
}

// Execute sends a GraphQL request and returns the decoded response body.
func (c *Client) Execute(ctx context.Context, payload Request) (map[string]any, error) {
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal graphql payload: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(), bytes.NewReader(encodedPayload))
	if err != nil {
		return nil, fmt.Errorf("build graphql request: %w", err)
	}

	request.Header = constants.GraphQLHeaders().Clone()
	request.Header.Set("Authorization", "Bearer "+c.JWT)

	response, err := c.httpClient().Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute graphql request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, constants.MaxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("read graphql response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, mapHTTPError(response.StatusCode, string(body), nil)
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode graphql response: %w", err)
	}

	if err := mapGraphQLError(raw); err != nil {
		return raw, err
	}

	return raw, nil
}

func (c *Client) endpoint() string {
	if c.Endpoint != "" {
		return c.Endpoint
	}
	return constants.GraphQLEndpoint
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: constants.HTTPTimeout}
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
