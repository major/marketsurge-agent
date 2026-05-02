// Package auth resolves JWT credentials for the MarketSurge API.
//
// This project is unofficial and is not affiliated with, approved by, or
// endorsed by MarketSurge or Investor's Business Daily.
package auth

import (
	"context"
	"log/slog"

	"github.com/major/marketsurge-agent/internal/cookies"
	"github.com/major/marketsurge-agent/internal/errors"
)

// ResolveJWT resolves a JWT token from Firefox browser cookies:
//  1. Explicit cookie DB path (--cookie-db flag)
//  2. Auto-discover Firefox profiles, try each until one succeeds
//
// Returns AuthenticationError if all sources fail.
func ResolveJWT(ctx context.Context, cookieDBPath string) (string, error) {
	// 1. Explicit cookie DB path.
	if cookieDBPath != "" {
		return resolveFromCookieDB(ctx, cookieDBPath)
	}

	// 2. Auto-discover: try each Firefox profile until one succeeds.
	return resolveFromFirefoxProfiles(ctx)
}

// resolveFromCookieDB extracts cookies from a specific database and exchanges
// them for a JWT.
func resolveFromCookieDB(ctx context.Context, cookieDBPath string) (string, error) {
	cookieJar, err := cookies.ExtractCookies(ctx, cookieDBPath)
	if err != nil {
		return "", errors.NewAuthenticationError(
			"no JWT available: sign into MarketSurge in Firefox or check the --cookie-db path",
			err,
		)
	}

	return ExchangeJWT(ctx, cookieJar)
}

// resolveFromFirefoxProfiles discovers all Firefox profiles and tries each
// one's cookies until a valid JWT exchange succeeds. Profiles are tried in
// order of most recently modified cookies.sqlite first.
func resolveFromFirefoxProfiles(ctx context.Context) (string, error) {
	dbPaths, err := cookies.FindCookieDBPaths()
	if err != nil {
		return "", errors.NewAuthenticationError(
			"no JWT available: could not discover Firefox profiles; sign into MarketSurge in Firefox or pass --cookie-db",
			err,
		)
	}

	if len(dbPaths) == 0 {
		return "", errors.NewAuthenticationError(
			"no JWT available: no Firefox profiles found; sign into MarketSurge in Firefox or pass --cookie-db",
			nil,
		)
	}

	var lastErr error
	for _, dbPath := range dbPaths {
		slog.Debug("trying Firefox profile", "path", dbPath)

		cookieJar, err := cookies.ExtractCookies(ctx, dbPath)
		if err != nil {
			slog.Debug("cookie extraction failed", "path", dbPath, "error", err)
			lastErr = err
			continue
		}

		if len(cookieJar) == 0 {
			slog.Debug("no investors.com cookies found, skipping", "path", dbPath)
			continue
		}

		jwt, err := ExchangeJWT(ctx, cookieJar)
		if err != nil {
			slog.Debug("JWT exchange failed", "path", dbPath, "error", err)
			lastErr = err
			continue
		}

		slog.Debug("authentication successful", "path", dbPath)
		return jwt, nil
	}

	return "", errors.NewAuthenticationError(
		"no JWT available: tried all Firefox profiles but none produced a valid login; sign into MarketSurge in Firefox or pass --cookie-db",
		lastErr,
	)
}
