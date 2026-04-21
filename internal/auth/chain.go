package auth

import (
	"context"
	"log"
	"os"

	"github.com/major/marketsurge-agent/internal/cookies"
	"github.com/major/marketsurge-agent/internal/errors"
)

// ResolveJWT resolves a JWT token using the auth precedence chain:
//  1. flagJWT (from --jwt flag) - highest priority
//  2. MARKETSURGE_JWT env var
//  3. Explicit cookie DB path (--cookie-db flag)
//  4. Auto-discover Firefox profiles, try each until one succeeds
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

	// 3. Explicit cookie DB path.
	if cookieDBPath != "" {
		return resolveFromCookieDB(ctx, cookieDBPath)
	}

	// 4. Auto-discover: try each Firefox profile until one succeeds.
	return resolveFromFirefoxProfiles(ctx)
}

// resolveFromCookieDB extracts cookies from a specific database and exchanges
// them for a JWT.
func resolveFromCookieDB(ctx context.Context, cookieDBPath string) (string, error) {
	cookieJar, err := cookies.ExtractCookies(ctx, cookieDBPath)
	if err != nil {
		return "", errors.NewAuthenticationError(
			"no JWT available: try --jwt flag, MARKETSURGE_JWT env var, or sign into MarketSurge in Firefox",
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
			"no JWT available: could not discover Firefox profiles; try --jwt flag, MARKETSURGE_JWT env var, or --cookie-db",
			err,
		)
	}

	if len(dbPaths) == 0 {
		return "", errors.NewAuthenticationError(
			"no JWT available: no Firefox profiles found; try --jwt flag, MARKETSURGE_JWT env var, or sign into MarketSurge in Firefox",
			nil,
		)
	}

	var lastErr error
	for _, dbPath := range dbPaths {
		log.Printf("trying Firefox profile: %s", dbPath)

		cookieJar, err := cookies.ExtractCookies(ctx, dbPath)
		if err != nil {
			log.Printf("  cookie extraction failed: %v", err)
			lastErr = err
			continue
		}

		if len(cookieJar) == 0 {
			log.Printf("  no investors.com cookies found, skipping")
			continue
		}

		jwt, err := ExchangeJWT(ctx, cookieJar)
		if err != nil {
			log.Printf("  JWT exchange failed: %v", err)
			lastErr = err
			continue
		}

		log.Printf("  authentication successful")
		return jwt, nil
	}

	return "", errors.NewAuthenticationError(
		"no JWT available: tried all Firefox profiles but none produced a valid login; try --jwt flag, MARKETSURGE_JWT env var, or sign into MarketSurge in Firefox",
		lastErr,
	)
}
