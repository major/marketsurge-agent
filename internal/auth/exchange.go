// Package auth handles JWT token exchange with the investors.com API.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/major/marketsurge-agent/internal/constants"
	"github.com/major/marketsurge-agent/internal/errors"
	"resty.dev/v3"
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
	rc := resty.New()
	defer rc.Close()

	rc.SetTimeout(constants.HTTPTimeout).SetResponseBodyLimit(constants.MaxResponseSize)

	req := rc.R().SetContext(ctx)

	// Set required headers from constants.
	for key, values := range constants.JWTExchangeHeaders() {
		for _, v := range values {
			req.SetHeader(key, v)
		}
	}

	// Forward all cookies to the request.
	req.SetCookies(cookies)

	resp, err := req.Get(exchangeURL)
	if err != nil {
		return "", errors.NewAuthenticationError(
			fmt.Sprintf("JWT exchange request failed: %s", err),
			err,
		)
	}

	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return "", errors.NewAuthenticationError(
			fmt.Sprintf("JWT exchange failed: HTTP %d", resp.StatusCode()),
			nil,
		)
	}

	body := resp.Bytes()

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
