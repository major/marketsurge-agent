// Package auth handles JWT token exchange with the investors.com API.
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-go/marketsurge"
)

// jwtExchangeURL is the investors.com JWT exchange endpoint.
const jwtExchangeURL = "https://www.investors.com/client"

// exchangeURL is the JWT exchange endpoint URL. It can be overridden in tests.
var exchangeURL = jwtExchangeURL

// ExchangeJWT exchanges browser cookies for a JWT token through marketsurge-go
// while preserving this CLI's authentication error contract.
func ExchangeJWT(ctx context.Context, cookies []*http.Cookie) (string, error) {
	client, err := marketsurge.NewClient(
		marketsurge.WithInvestorsBaseURL(investorsBaseURL(exchangeURL)),
	)
	if err != nil {
		return "", mserrors.NewAuthenticationError("JWT exchange client setup failed", err)
	}

	info, err := client.ExchangeJWT(ctx, marketsurge.NewSession(cookies))
	if err != nil {
		return "", mapJWTExchangeError(err)
	}

	if !info.IsLoggedIn {
		return "", mserrors.NewAuthenticationError(
			"not logged in -- make sure you're signed into MarketSurge in the browser",
			nil,
		)
	}

	if info.JWT == "" {
		return "", mserrors.NewAuthenticationError(
			"no JWT found in response",
			nil,
		)
	}

	return info.JWT, nil
}

func investorsBaseURL(rawExchangeURL string) string {
	return strings.TrimSuffix(rawExchangeURL, "/client")
}

func mapJWTExchangeError(err error) error {
	statusErr, ok := errors.AsType[*marketsurge.StatusError](err)
	if ok {
		return mserrors.NewAuthenticationError(
			fmt.Sprintf("JWT exchange failed: HTTP %d", statusErr.StatusCode),
			err,
		)
	}

	if strings.Contains(err.Error(), "not logged in") {
		return mserrors.NewAuthenticationError(
			"not logged in -- make sure you're signed into MarketSurge in the browser",
			err,
		)
	}

	if strings.Contains(err.Error(), "JWT not found") {
		return mserrors.NewAuthenticationError("no JWT found in response", err)
	}

	return mserrors.NewAuthenticationError(
		"JWT exchange request failed",
		err,
	)
}
