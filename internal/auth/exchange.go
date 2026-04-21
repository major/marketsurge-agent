// Package auth handles JWT token exchange with the investors.com API.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/major/marketsurge-agent/internal/constants"
	"github.com/major/marketsurge-agent/internal/errors"
)

// exchangeURL is the JWT exchange endpoint URL. It defaults to
// constants.JWTExchangeURL and can be overridden in tests.
var exchangeURL = constants.JWTExchangeURL

// clientResponse represents the JSON response from the JWT exchange endpoint.
type clientResponse struct {
	IsLoggedIn bool   `json:"isLoggedIn"`
	JWT        string `json:"jwt"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
}

// ExchangeJWT exchanges browser cookies for a JWT token by calling the
// investors.com client endpoint. It sends a GET request with the provided
// cookies and extracts the JWT from the JSON response.
func ExchangeJWT(ctx context.Context, cookies []*http.Cookie) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, exchangeURL, nil)
	if err != nil {
		return "", errors.NewAuthenticationError(
			fmt.Sprintf("failed to build JWT exchange request: %s", err),
			err,
		)
	}

	// Set required headers from constants.
	for key, values := range constants.JWTExchangeHeaders() {
		for _, v := range values {
			req.Header.Set(key, v)
		}
	}

	// Forward all cookies to the request.
	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{Timeout: constants.HTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.NewAuthenticationError(
			fmt.Sprintf("JWT exchange request failed: %s", err),
			err,
		)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.NewAuthenticationError(
			fmt.Sprintf("failed to read JWT exchange response: %s", err),
			err,
		)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.NewAuthenticationError(
			fmt.Sprintf("JWT exchange failed: HTTP %d", resp.StatusCode),
			nil,
		)
	}

	var data clientResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return "", errors.NewAuthenticationError(
			fmt.Sprintf("failed to parse JWT exchange response: %s", err),
			err,
		)
	}

	if !data.IsLoggedIn {
		return "", errors.NewAuthenticationError(
			"not logged in -- make sure you're signed into MarketSurge in the browser",
			nil,
		)
	}

	if data.JWT == "" {
		return "", errors.NewAuthenticationError(
			"no JWT found in response",
			nil,
		)
	}

	return data.JWT, nil
}
