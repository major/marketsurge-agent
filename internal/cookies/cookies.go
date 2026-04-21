// Package cookies extracts browser cookies for MarketSurge authentication.
package cookies

import (
	"context"
	"fmt"
	"net/http"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/browser/firefox"

	mserr "github.com/major/marketsurge-agent/internal/errors"
)

// investorsDomain is the domain suffix used to filter MarketSurge cookies.
const investorsDomain = "investors.com"

// ExtractCookies retrieves investors.com cookies from Firefox.
// If cookieDBPath is provided, reads from that specific cookies.sqlite;
// otherwise kooky auto-discovers Firefox profiles via profiles.ini.
func ExtractCookies(ctx context.Context, cookieDBPath string) ([]*http.Cookie, error) {
	filters := []kooky.Filter{
		kooky.Valid,
		kooky.DomainHasSuffix(investorsDomain),
	}

	var kookyCookies []*kooky.Cookie
	var err error

	if cookieDBPath != "" {
		kookyCookies, err = firefox.ReadCookies(ctx, cookieDBPath, filters...)
	} else {
		// Importing browser/firefox registers its profile finder,
		// so TraverseCookies discovers Firefox profiles via profiles.ini.
		kookyCookies, err = kooky.TraverseCookies(ctx, filters...).ReadAllCookies(ctx)
	}

	if err != nil {
		return nil, mserr.NewCookieExtractionError(
			fmt.Sprintf("failed to extract cookies: %s", err),
			err, "Firefox",
		)
	}

	// Auto-discovery returning no cookies means either no Firefox profiles
	// exist or none contain investors.com cookies. Surface this as an error
	// so the auth chain can suggest alternatives (--jwt flag, env var).
	if len(kookyCookies) == 0 && cookieDBPath == "" {
		return nil, mserr.NewCookieExtractionError(
			"no investors.com cookies found in any Firefox profile",
			nil, "Firefox",
		)
	}

	httpCookies := make([]*http.Cookie, len(kookyCookies))
	for i, c := range kookyCookies {
		httpCookies[i] = &c.Cookie
	}

	return httpCookies, nil
}
