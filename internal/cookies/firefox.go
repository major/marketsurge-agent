// Package cookies extracts browser cookies for MarketSurge authentication.
package cookies

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	mserr "github.com/major/marketsurge-agent/internal/errors"

	// Register the pure-Go SQLite driver.
	_ "modernc.org/sqlite"
)

// investorsDomainFilter is the SQL LIKE pattern for investors.com cookies.
const investorsDomainFilter = "%investors.com"

// cookieQuery selects all cookie fields needed for http.Cookie conversion.
const cookieQuery = `SELECT name, value, host, path, expiry, isSecure, isHttpOnly FROM moz_cookies WHERE host LIKE ?`

// ExtractCookies opens a Firefox cookies.sqlite database in read-only mode and
// returns all cookies matching the investors.com domain. The database is never
// modified: the DSN uses mode=ro, immutable=1, and query_only pragma.
func ExtractCookies(cookieDBPath string) ([]*http.Cookie, error) {
	dsn := fmt.Sprintf("file:%s?mode=ro&immutable=1&_pragma=query_only(1)&_pragma=busy_timeout(5000)", cookieDBPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, mserr.NewCookieExtractionError(
			fmt.Sprintf("failed to open cookie database: %s", err),
			err, "Firefox",
		)
	}
	defer db.Close()

	// Single connection avoids lock contention with a running browser.
	db.SetMaxOpenConns(1)

	// Verify the database is accessible by pinging it. This catches file-not-found
	// and permission errors before running queries.
	if err := db.Ping(); err != nil {
		return nil, mserr.NewCookieExtractionError(
			fmt.Sprintf("cookie database not accessible: %s", err),
			err, "Firefox",
		)
	}

	rows, err := db.Query(cookieQuery, investorsDomainFilter)
	if err != nil {
		return nil, mserr.NewCookieExtractionError(
			fmt.Sprintf("failed to query cookies: %s", err),
			err, "Firefox",
		)
	}
	defer rows.Close()

	var cookies []*http.Cookie
	for rows.Next() {
		var (
			name, value, host, path string
			expiry                  int64
			isSecure, isHTTPOnly    int
		)

		if err := rows.Scan(&name, &value, &host, &path, &expiry, &isSecure, &isHTTPOnly); err != nil {
			return nil, mserr.NewCookieExtractionError(
				fmt.Sprintf("failed to scan cookie row: %s", err),
				err, "Firefox",
			)
		}

		cookies = append(cookies, &http.Cookie{
			Name:     name,
			Value:    value,
			Domain:   host,
			Path:     path,
			Expires:  time.Unix(expiry, 0),
			Secure:   isSecure == 1,
			HttpOnly: isHTTPOnly == 1,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, mserr.NewCookieExtractionError(
			fmt.Sprintf("error iterating cookie rows: %s", err),
			err, "Firefox",
		)
	}

	return cookies, nil
}
