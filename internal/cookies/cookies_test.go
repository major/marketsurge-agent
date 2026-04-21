package cookies

import (
	"context"
	"testing"

	mserr "github.com/major/marketsurge-agent/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtractCookiesMissingDB verifies a CookieExtractionError for a nonexistent database.
func TestExtractCookiesMissingDB(t *testing.T) {
	_, err := ExtractCookies(context.Background(), "/tmp/does-not-exist-cookies.sqlite")
	require.Error(t, err)

	var cookieErr *mserr.CookieExtractionError
	require.ErrorAs(t, err, &cookieErr)
	assert.Equal(t, "Firefox", cookieErr.Browser)
}

// TestExtractCookiesAutoDiscoverNoProfiles verifies a CookieExtractionError when
// no Firefox profiles exist (empty HOME).
func TestExtractCookiesAutoDiscoverNoProfiles(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := ExtractCookies(context.Background(), "")
	require.Error(t, err)

	var cookieErr *mserr.CookieExtractionError
	require.ErrorAs(t, err, &cookieErr)
	assert.Equal(t, "Firefox", cookieErr.Browser)
}
