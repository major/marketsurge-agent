package cookies

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	mserr "github.com/major/marketsurge-agent/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

// createTestCookieDB creates a temporary SQLite database mimicking Firefox's
// moz_cookies table and inserts the provided rows. Returns the database path.
func createTestCookieDB(t *testing.T, cookies []testCookie) string {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "cookies.sqlite")

	db, err := sql.Open("sqlite", "file:"+dbPath)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE moz_cookies (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		value TEXT NOT NULL,
		host TEXT NOT NULL,
		path TEXT NOT NULL DEFAULT '/',
		expiry INTEGER NOT NULL DEFAULT 0,
		isSecure INTEGER NOT NULL DEFAULT 0,
		isHttpOnly INTEGER NOT NULL DEFAULT 0,
		lastAccessed INTEGER NOT NULL DEFAULT 0,
		creationTime INTEGER NOT NULL DEFAULT 0,
		schemeMap INTEGER NOT NULL DEFAULT 0
	)`)
	require.NoError(t, err)

	for _, c := range cookies {
		_, err = db.Exec(
			`INSERT INTO moz_cookies (name, value, host, path, expiry, isSecure, isHttpOnly) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			c.name, c.value, c.host, c.path, c.expiry, c.isSecure, c.isHTTPOnly,
		)
		require.NoError(t, err)
	}

	return dbPath
}

// testCookie holds row data for moz_cookies test fixtures.
type testCookie struct {
	name       string
	value      string
	host       string
	path       string
	expiry     int64
	isSecure   int
	isHTTPOnly int
}

// TestExtractCookiesSuccess verifies cookies are extracted and mapped to http.Cookie correctly.
func TestExtractCookiesSuccess(t *testing.T) {
	expiry := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	dbPath := createTestCookieDB(t, []testCookie{
		{name: "session_id", value: "abc123", host: ".investors.com", path: "/", expiry: expiry, isSecure: 1, isHTTPOnly: 1},
		{name: "pref", value: "dark", host: "www.investors.com", path: "/app", expiry: expiry, isSecure: 0, isHTTPOnly: 0},
	})

	cookies, err := ExtractCookies(dbPath)
	require.NoError(t, err)
	require.Len(t, cookies, 2)

	// Verify first cookie mapping.
	assert.Equal(t, "session_id", cookies[0].Name)
	assert.Equal(t, "abc123", cookies[0].Value)
	assert.Equal(t, ".investors.com", cookies[0].Domain)
	assert.Equal(t, "/", cookies[0].Path)
	assert.Equal(t, time.Unix(expiry, 0), cookies[0].Expires)
	assert.True(t, cookies[0].Secure)
	assert.True(t, cookies[0].HttpOnly)

	// Verify second cookie mapping.
	assert.Equal(t, "pref", cookies[1].Name)
	assert.Equal(t, "dark", cookies[1].Value)
	assert.Equal(t, "www.investors.com", cookies[1].Domain)
	assert.Equal(t, "/app", cookies[1].Path)
	assert.False(t, cookies[1].Secure)
	assert.False(t, cookies[1].HttpOnly)
}

// TestExtractCookiesFiltersNonInvestors verifies that only investors.com cookies are returned.
func TestExtractCookiesFiltersNonInvestors(t *testing.T) {
	dbPath := createTestCookieDB(t, []testCookie{
		{name: "sid", value: "xyz", host: ".investors.com", path: "/", expiry: 0, isSecure: 0, isHTTPOnly: 0},
		{name: "other", value: "nope", host: ".example.com", path: "/", expiry: 0, isSecure: 0, isHTTPOnly: 0},
		{name: "track", value: "123", host: ".google.com", path: "/", expiry: 0, isSecure: 0, isHTTPOnly: 0},
	})

	cookies, err := ExtractCookies(dbPath)
	require.NoError(t, err)
	require.Len(t, cookies, 1)
	assert.Equal(t, "sid", cookies[0].Name)
}

// TestExtractCookiesEmptyResult verifies an empty slice is returned when no investors.com cookies exist.
func TestExtractCookiesEmptyResult(t *testing.T) {
	dbPath := createTestCookieDB(t, []testCookie{
		{name: "other", value: "val", host: ".example.com", path: "/", expiry: 0, isSecure: 0, isHTTPOnly: 0},
	})

	cookies, err := ExtractCookies(dbPath)
	require.NoError(t, err)
	assert.Empty(t, cookies)
}

// TestExtractCookiesMissingDB verifies a CookieExtractionError for a nonexistent database.
func TestExtractCookiesMissingDB(t *testing.T) {
	_, err := ExtractCookies("/tmp/does-not-exist-cookies.sqlite")
	require.Error(t, err)

	var cookieErr *mserr.CookieExtractionError
	require.ErrorAs(t, err, &cookieErr)
	assert.Equal(t, "Firefox", cookieErr.Browser)
}

// TestExtractCookiesEmptyDB verifies a CookieExtractionError when the database has no moz_cookies table.
func TestExtractCookiesEmptyDB(t *testing.T) {
	// Create an empty SQLite database without the moz_cookies table.
	dbPath := filepath.Join(t.TempDir(), "empty.sqlite")
	db, err := sql.Open("sqlite", "file:"+dbPath)
	require.NoError(t, err)
	// Create a dummy table so the file is a valid SQLite database.
	_, err = db.Exec("CREATE TABLE dummy (id INTEGER)")
	require.NoError(t, err)
	db.Close()

	_, err = ExtractCookies(dbPath)
	require.Error(t, err)

	var cookieErr *mserr.CookieExtractionError
	require.ErrorAs(t, err, &cookieErr)
}

// TestFindFirefoxCookieDBWithFakeProfile creates a temp directory structure mimicking
// a Firefox profile and verifies the database is discovered.
func TestFindFirefoxCookieDBWithFakeProfile(t *testing.T) {
	// Save and restore HOME to isolate this test.
	origHome := os.Getenv("HOME")
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)

	// Build the profile directory structure.
	var profileDir string
	if runtime.GOOS == "darwin" {
		profileDir = filepath.Join(tmpHome, "Library", "Application Support", "Firefox", "Profiles", "abc123.default-release")
	} else {
		profileDir = filepath.Join(tmpHome, ".mozilla", "firefox", "abc123.default-release")
	}
	require.NoError(t, os.MkdirAll(profileDir, 0o755))

	cookiePath := filepath.Join(profileDir, "cookies.sqlite")
	require.NoError(t, os.WriteFile(cookiePath, []byte("fake"), 0o644))

	found, err := FindFirefoxCookieDB()
	require.NoError(t, err)
	assert.Equal(t, cookiePath, found)
}

// TestFindFirefoxCookieDBFallbackToDefault verifies the *.default fallback pattern
// works when no *.default-release profile exists.
func TestFindFirefoxCookieDBFallbackToDefault(t *testing.T) {
	origHome := os.Getenv("HOME")
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)

	var profileDir string
	if runtime.GOOS == "darwin" {
		profileDir = filepath.Join(tmpHome, "Library", "Application Support", "Firefox", "Profiles", "xyz789.default")
	} else {
		profileDir = filepath.Join(tmpHome, ".mozilla", "firefox", "xyz789.default")
	}
	require.NoError(t, os.MkdirAll(profileDir, 0o755))

	cookiePath := filepath.Join(profileDir, "cookies.sqlite")
	require.NoError(t, os.WriteFile(cookiePath, []byte("fake"), 0o644))

	found, err := FindFirefoxCookieDB()
	require.NoError(t, err)
	assert.Equal(t, cookiePath, found)
}

// TestFindFirefoxCookieDBNoProfile verifies a CookieExtractionError when no
// Firefox profile exists.
func TestFindFirefoxCookieDBNoProfile(t *testing.T) {
	origHome := os.Getenv("HOME")
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)

	_, err := FindFirefoxCookieDB()
	require.Error(t, err)

	var cookieErr *mserr.CookieExtractionError
	require.ErrorAs(t, err, &cookieErr)
	assert.Equal(t, "Firefox", cookieErr.Browser)
}

// TestFindFirefoxCookieDBPicksNewest verifies that when multiple profiles match,
// the most recently modified cookies.sqlite is selected.
func TestFindFirefoxCookieDBPicksNewest(t *testing.T) {
	origHome := os.Getenv("HOME")
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)

	var base string
	if runtime.GOOS == "darwin" {
		base = filepath.Join(tmpHome, "Library", "Application Support", "Firefox", "Profiles")
	} else {
		base = filepath.Join(tmpHome, ".mozilla", "firefox")
	}

	// Create two default-release profiles.
	oldDir := filepath.Join(base, "old111.default-release")
	newDir := filepath.Join(base, "new222.default-release")
	require.NoError(t, os.MkdirAll(oldDir, 0o755))
	require.NoError(t, os.MkdirAll(newDir, 0o755))

	oldPath := filepath.Join(oldDir, "cookies.sqlite")
	newPath := filepath.Join(newDir, "cookies.sqlite")

	require.NoError(t, os.WriteFile(oldPath, []byte("old"), 0o644))
	// Set the old file's mtime to the past.
	oldTime := time.Now().Add(-24 * time.Hour)
	require.NoError(t, os.Chtimes(oldPath, oldTime, oldTime))

	require.NoError(t, os.WriteFile(newPath, []byte("new"), 0o644))

	found, err := FindFirefoxCookieDB()
	require.NoError(t, err)
	assert.Equal(t, newPath, found)
}

// TestNewestFileAllInaccessible verifies newestFile returns an error when all
// paths are inaccessible.
func TestNewestFileAllInaccessible(t *testing.T) {
	_, err := newestFile([]string{"/nonexistent/a", "/nonexistent/b"})
	require.Error(t, err)

	var cookieErr *mserr.CookieExtractionError
	require.ErrorAs(t, err, &cookieErr)
}

// TestCookieExpiresConversion verifies that Unix timestamps from moz_cookies
// are correctly converted to time.Time in http.Cookie.Expires.
func TestCookieExpiresConversion(t *testing.T) {
	// 2030-06-15 12:00:00 UTC
	expiry := int64(1907928000)
	dbPath := createTestCookieDB(t, []testCookie{
		{name: "ts", value: "check", host: ".investors.com", path: "/", expiry: expiry, isSecure: 0, isHTTPOnly: 0},
	})

	cookies, err := ExtractCookies(dbPath)
	require.NoError(t, err)
	require.Len(t, cookies, 1)

	expected := time.Unix(expiry, 0)
	assert.Equal(t, expected, cookies[0].Expires)
}

// TestExtractCookiesReturnType verifies that an empty result returns nil slice, not an error.
func TestExtractCookiesReturnType(t *testing.T) {
	dbPath := createTestCookieDB(t, nil)

	cookies, err := ExtractCookies(dbPath)
	require.NoError(t, err)

	// Empty table should return nil slice (no allocations).
	var expected []*http.Cookie
	assert.Equal(t, expected, cookies)
}
