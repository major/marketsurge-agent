package cookies

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	mserr "github.com/major/marketsurge-agent/internal/errors"
)

// profileGlobPatterns returns the glob patterns for Firefox cookie databases,
// ordered by preference (default-release first, then default fallback).
func profileGlobPatterns(home string) []string {
	var base string
	switch runtime.GOOS {
	case "darwin":
		base = filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles")
	default:
		// Linux and other Unix-like systems.
		base = filepath.Join(home, ".mozilla", "firefox")
	}

	return []string{
		filepath.Join(base, "*.default-release", "cookies.sqlite"),
		filepath.Join(base, "*.default", "cookies.sqlite"),
	}
}

// FindFirefoxCookieDB locates the Firefox cookies.sqlite database by searching
// known profile directory patterns. If multiple profiles match, the most
// recently modified cookies.sqlite is returned.
func FindFirefoxCookieDB() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", mserr.NewCookieExtractionError(
			fmt.Sprintf("cannot determine home directory: %s", err),
			err, "Firefox",
		)
	}

	patterns := profileGlobPatterns(home)

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			// Glob only returns an error for malformed patterns, which
			// should never happen with our hardcoded patterns.
			continue
		}

		if len(matches) == 0 {
			continue
		}

		return newestFile(matches)
	}

	return "", mserr.NewCookieExtractionError(
		"no Firefox cookies.sqlite found in any profile directory",
		nil, "Firefox",
	)
}

// newestFile returns the path with the most recent modification time.
// The input slice must be non-empty.
func newestFile(paths []string) (string, error) {
	type fileEntry struct {
		path    string
		modTime int64
	}

	entries := make([]fileEntry, 0, len(paths))
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			// Skip files we cannot stat (e.g., permission denied).
			continue
		}
		entries = append(entries, fileEntry{path: p, modTime: info.ModTime().UnixNano()})
	}

	if len(entries) == 0 {
		return "", mserr.NewCookieExtractionError(
			"Firefox cookie databases found but none are accessible",
			nil, "Firefox",
		)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].modTime > entries[j].modTime
	})

	return entries[0].path, nil
}
