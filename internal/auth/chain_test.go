package auth

import (
	"context"
	"testing"

	"github.com/major/marketsurge-agent/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clearAuthEnv unsets all JWT env vars so tests start from a clean state.
func clearAuthEnv(t *testing.T) {
	t.Helper()
	t.Setenv("MARKETSURGE_JWT", "")
}

// TestResolveJWT_FlagPrecedence verifies that the --jwt flag takes highest
// priority, even when env vars are also set.
func TestResolveJWT_FlagPrecedence(t *testing.T) {
	t.Setenv("MARKETSURGE_JWT", "env-jwt")

	jwt, err := ResolveJWT(context.Background(), "flag-jwt", "")
	require.NoError(t, err)
	assert.Equal(t, "flag-jwt", jwt)
}

// TestResolveJWT_MarketSurgeEnv verifies MARKETSURGE_JWT is used when no flag
// is provided.
func TestResolveJWT_MarketSurgeEnv(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv("MARKETSURGE_JWT", "ms-jwt")

	jwt, err := ResolveJWT(context.Background(), "", "")
	require.NoError(t, err)
	assert.Equal(t, "ms-jwt", jwt)
}

// TestResolveJWT_NoSources verifies that AuthenticationError is returned when
// no JWT flag, env vars, or Firefox profile are available.
func TestResolveJWT_NoSources(t *testing.T) {
	clearAuthEnv(t)

	// Point HOME to a temp dir so FindCookieDBPaths finds no profiles.
	t.Setenv("HOME", t.TempDir())

	jwt, err := ResolveJWT(context.Background(), "", "")
	assert.Empty(t, jwt)
	require.Error(t, err)

	var authErr *errors.AuthenticationError
	assert.ErrorAs(t, err, &authErr)
	assert.Contains(t, authErr.Message, "no JWT available")
}

// TestResolveJWT_CookieDBPath verifies that an explicit cookieDBPath pointing
// to a nonexistent file produces an AuthenticationError wrapping a
// CookieExtractionError.
func TestResolveJWT_CookieDBPath(t *testing.T) {
	clearAuthEnv(t)

	jwt, err := ResolveJWT(context.Background(), "", "/nonexistent/cookies.sqlite")
	assert.Empty(t, jwt)
	require.Error(t, err)

	var authErr *errors.AuthenticationError
	assert.ErrorAs(t, err, &authErr)
}
