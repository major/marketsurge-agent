package auth

import (
	"context"
	"testing"

	"github.com/major/marketsurge-agent/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResolveJWT_NoSources verifies that AuthenticationError is returned when
// no Firefox profile is available.
func TestResolveJWT_NoSources(t *testing.T) {
	// Point HOME to a temp dir so FindCookieDBPaths finds no profiles.
	t.Setenv("HOME", t.TempDir())

	jwt, err := ResolveJWT(context.Background(), "")
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
	jwt, err := ResolveJWT(context.Background(), "/nonexistent/cookies.sqlite")
	assert.Empty(t, jwt)
	require.Error(t, err)

	var authErr *errors.AuthenticationError
	assert.ErrorAs(t, err, &authErr)
}
