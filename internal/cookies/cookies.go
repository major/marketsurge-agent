// Package cookies extracts browser cookies for MarketSurge authentication.
package cookies

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/browser/firefox"

	mserr "github.com/major/marketsurge-agent/internal/errors"
)

// investorsDomain is the domain suffix used to filter MarketSurge cookies.
const investorsDomain = "investors.com"

// firefoxRoot returns the platform-specific Firefox profile root directory.
// Overridable in tests.
var firefoxRoot = defaultFirefoxRoot

func defaultFirefoxRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home dir: %w", err)
	}

	switch runtime.GOOS {
	case "linux":
		return filepath.Join(home, ".mozilla", "firefox"), nil
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Firefox"), nil
	default:
		return "", fmt.Errorf("unsupported platform for Firefox cookie discovery: %s", runtime.GOOS)
	}
}

// cookiePath pairs a file path with its precomputed stat result.
// info is nil when os.Stat fails, keeping the path in output but
// sorting it after paths with valid stats.
type cookiePath struct {
	path string
	info os.FileInfo
}

// FindCookieDBPaths returns paths to all Firefox profile cookies.sqlite files,
// sorted by modification time (most recent first). Profiles that were used
// most recently are tried first during authentication.
func FindCookieDBPaths() ([]string, error) {
	root, err := firefoxRoot()
	if err != nil {
		return nil, err
	}

	pattern := filepath.Join(root, "*", "cookies.sqlite")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob firefox profiles: %w", err)
	}

	// Precompute stat results once per path instead of repeatedly
	// inside the comparator. Stat failures get nil info.
	entries := make([]cookiePath, len(matches))
	for i, p := range matches {
		info, statErr := os.Stat(p)
		if statErr != nil {
			entries[i] = cookiePath{path: p}
		} else {
			entries[i] = cookiePath{path: p, info: info}
		}
	}

	// Sort by modification time, most recent first. Paths where
	// stat failed sort last so they are still tried, just not first.
	slices.SortFunc(entries, func(a, b cookiePath) int {
		switch {
		case a.info == nil && b.info == nil:
			return 0
		case a.info == nil:
			return 1 // a sorts after b
		case b.info == nil:
			return -1 // a sorts before b
		default:
			return b.info.ModTime().Compare(a.info.ModTime())
		}
	})

	sorted := make([]string, len(entries))
	for i, e := range entries {
		sorted[i] = e.path
	}

	return sorted, nil
}

// ExtractCookies retrieves investors.com cookies from a specific Firefox
// cookies.sqlite database. The cookieDBPath must point to a valid
// cookies.sqlite file.
func ExtractCookies(ctx context.Context, cookieDBPath string) ([]*http.Cookie, error) {
	filters := []kooky.Filter{
		kooky.Valid,
		kooky.DomainHasSuffix(investorsDomain),
	}

	kookyCookies, err := firefox.ReadCookies(ctx, cookieDBPath, filters...)
	if err != nil {
		return nil, mserr.NewCookieExtractionError(
			fmt.Sprintf("failed to extract cookies from %s: %s", cookieDBPath, err),
			err, "Firefox",
		)
	}

	httpCookies := make([]*http.Cookie, len(kookyCookies))
	for i, c := range kookyCookies {
		httpCookies[i] = &c.Cookie
	}

	return httpCookies, nil
}
