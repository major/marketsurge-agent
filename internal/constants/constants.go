package constants

import (
	"net/http"
	"time"
)

// API endpoints and authentication tokens
const (
	GraphQLEndpoint    = "https://shared-data.dowjones.io/gateway/graphql"
	JWTExchangeURL     = "https://www.investors.com/client"
	DylanToken         = "x4ckyhshg90pdq6bwf6n1voijs7r3fdk" //nolint:gosec // public API exchange token, not a secret
	UserAgent          = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:149.0) Gecko/20100101 Firefox/149.0"
	OriginURL          = "https://marketsurge.investors.com"
	RefererURL         = "https://marketsurge.investors.com/"
	OriginalHost       = "marketsurge.investors.com"
	ApolloClientName   = "marketsurge"
	CookieDomain       = "investors.com"
	SymbolDialectType  = "CHARTING"
	HTTPTimeout        = 30 * time.Second
	MaxResponseSize    = 10 * 1024 * 1024 // 10 MB
)

// GraphQLHeaders returns the HTTP headers required for GraphQL requests.
// Note: Authorization header (JWT) must be added per-request.
func GraphQLHeaders() http.Header {
	headers := http.Header{}
	headers.Set("User-Agent", UserAgent)
	headers.Set("Content-Type", "application/json")
	headers.Set("apollographql-client-name", ApolloClientName)
	headers.Set("dylan-entitlement-token", DylanToken)
	headers.Set("Referer", RefererURL)
	headers.Set("Origin", OriginURL)
	return headers
}

// JWTExchangeHeaders returns the HTTP headers required for JWT exchange requests.
// Note: Cookie header must be added per-request.
func JWTExchangeHeaders() http.Header {
	headers := http.Header{}
	headers.Set("User-Agent", UserAgent)
	headers.Set("x-encrypted-document-key", "")
	headers.Set("x-original-host", OriginalHost)
	headers.Set("x-original-referrer", "")
	headers.Set("x-original-url", "/mstool")
	headers.Set("Referer", RefererURL)
	headers.Set("Origin", OriginURL)
	return headers
}
