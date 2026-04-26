package cookies

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

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

// TestFindCookieDBPaths_MultipleProfiles verifies that FindCookieDBPaths
// discovers all cookies.sqlite files and sorts them by modification time.
func TestFindCookieDBPaths_MultipleProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mock Firefox profile directories with cookies.sqlite files.
	profiles := []string{"abc123.older-profile", "def456.newer-profile"}
	for _, p := range profiles {
		dir := filepath.Join(tmpDir, p)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "cookies.sqlite"), []byte{}, 0o644))
	}

	// Set the older profile's mtime to 1 hour ago so sorting is deterministic.
	olderDB := filepath.Join(tmpDir, profiles[0], "cookies.sqlite")
	hourAgo := time.Now().Add(-time.Hour)
	require.NoError(t, os.Chtimes(olderDB, hourAgo, hourAgo))

	// Override firefoxRoot for this test.
	origRoot := firefoxRoot
	firefoxRoot = func() (string, error) { return tmpDir, nil }
	t.Cleanup(func() { firefoxRoot = origRoot })

	paths, err := FindCookieDBPaths()
	require.NoError(t, err)
	require.Len(t, paths, 2)

	// Most recently modified should come first.
	assert.Contains(t, paths[0], "newer-profile")
	assert.Contains(t, paths[1], "older-profile")
}

// TestFindCookieDBPaths_NoProfiles verifies an empty result when no profiles
// exist in the Firefox root directory.
func TestFindCookieDBPaths_NoProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	origRoot := firefoxRoot
	firefoxRoot = func() (string, error) { return tmpDir, nil }
	t.Cleanup(func() { firefoxRoot = origRoot })

	paths, err := FindCookieDBPaths()
	require.NoError(t, err)
	assert.Empty(t, paths)
}

// TestFindCookieDBPaths_RootError verifies that an error from firefoxRoot
// propagates correctly.
func TestFindCookieDBPaths_RootError(t *testing.T) {
	origRoot := firefoxRoot
	firefoxRoot = func() (string, error) { return "", assert.AnError }
	t.Cleanup(func() { firefoxRoot = origRoot })

	paths, err := FindCookieDBPaths()
	require.Error(t, err)
	assert.Nil(t, paths)
}

// TestFindCookieDBPaths_StatFailureSortsLast verifies that cookie DB paths
// where os.Stat fails are preserved in the output but sort after paths
// with valid stats.
func TestFindCookieDBPaths_StatFailureSortsLast(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two valid profiles with different mtimes.
	profiles := []struct {
		dir string
		age time.Duration
	}{
		{"aaa111.oldest", 2 * time.Hour},
		{"bbb222.newest", 0},
	}
	for _, p := range profiles {
		dir := filepath.Join(tmpDir, p.dir)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		dbPath := filepath.Join(dir, "cookies.sqlite")
		require.NoError(t, os.WriteFile(dbPath, []byte{}, 0o644))
		if p.age > 0 {
			mtime := time.Now().Add(-p.age)
			require.NoError(t, os.Chtimes(dbPath, mtime, mtime))
		}
	}

	// Create a profile directory but remove the cookies.sqlite after glob
	// discovers it. We simulate this by creating a broken symlink instead:
	// glob matches it, but os.Stat fails.
	brokenDir := filepath.Join(tmpDir, "ccc333.broken")
	require.NoError(t, os.MkdirAll(brokenDir, 0o755))
	brokenDB := filepath.Join(brokenDir, "cookies.sqlite")
	require.NoError(t, os.Symlink("/nonexistent/path", brokenDB))

	origRoot := firefoxRoot
	firefoxRoot = func() (string, error) { return tmpDir, nil }
	t.Cleanup(func() { firefoxRoot = origRoot })

	paths, err := FindCookieDBPaths()
	require.NoError(t, err)
	require.Len(t, paths, 3, "stat failures must not be dropped")

	// Newest valid stat first, oldest valid stat second, broken last.
	assert.Contains(t, paths[0], "newest")
	assert.Contains(t, paths[1], "oldest")
	assert.Contains(t, paths[2], "broken")
}
