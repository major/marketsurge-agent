// Package cookies extracts browser cookies for MarketSurge authentication.
package cookies

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"

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

	// Sort by modification time, most recent first. This makes the common
	// case fast: the actively-used profile is tried first.
	sort.Slice(matches, func(i, j int) bool {
		si, ei := os.Stat(matches[i])
		sj, ej := os.Stat(matches[j])
		if ei != nil || ej != nil {
			return ei == nil
		}
		return si.ModTime().After(sj.ModTime())
	})

	return matches, nil
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
