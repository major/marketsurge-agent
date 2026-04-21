package auth

import (
	"context"
	"os"

	"github.com/major/marketsurge-agent/internal/cookies"
	"github.com/major/marketsurge-agent/internal/errors"
)

// ResolveJWT resolves a JWT token using the auth precedence chain:
//  1. flagJWT (from --jwt flag) - highest priority
//  2. MARKETSURGE_JWT env var
//  3. Firefox cookie extraction + JWT exchange
//
// Returns AuthenticationError if all sources fail.
func ResolveJWT(ctx context.Context, flagJWT, cookieDBPath string) (string, error) {
	// 1. CLI flag takes highest precedence.
	if flagJWT != "" {
		return flagJWT, nil
	}

	// 2. MARKETSURGE_JWT env var.
	if jwt := os.Getenv("MARKETSURGE_JWT"); jwt != "" {
		return jwt, nil
	}

	// 3. Firefox cookie flow: extract cookies, exchange for JWT.
	cookieJar, err := cookies.ExtractCookies(ctx, cookieDBPath)
	if err != nil {
		return "", errors.NewAuthenticationError(
			"no JWT available: try --jwt flag, MARKETSURGE_JWT env var, or sign into MarketSurge in Firefox",
			err,
		)
	}

	return ExchangeJWT(ctx, cookieJar)
}
