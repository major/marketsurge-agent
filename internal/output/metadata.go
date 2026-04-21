// Package output provides JSON envelope types and writers for marketsurge-agent responses.
package output

import (
	"time"
)

// utcTimestamp returns the current UTC time formatted as RFC3339.
func utcTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// SymbolMeta creates metadata for a symbol query response.
// It includes the symbol and a UTC timestamp in RFC3339 format.
func SymbolMeta(symbol string) map[string]any {
	return map[string]any{
		"symbol":    symbol,
		"timestamp": utcTimestamp(),
	}
}

// CatalogMeta creates metadata for a catalog response.
// It includes the catalog kind, total count, limit, offset, and a UTC timestamp in RFC3339 format.
func CatalogMeta(kind string, total, limit, offset int) map[string]any {
	return map[string]any{
		"kind":      kind,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
		"timestamp": utcTimestamp(),
	}
}
