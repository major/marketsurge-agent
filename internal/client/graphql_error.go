package client

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// mapGraphQLError checks the raw response body for GraphQL-level errors.
func mapGraphQLError(body []byte) error {
	errs := gjson.GetBytes(body, "errors")
	if !errs.Exists() {
		return nil
	}

	items := errs.Array()
	if len(items) == 0 {
		return nil
	}

	messages := make([]string, 0, len(items))
	for _, item := range items {
		if msg := item.Get("message").String(); msg != "" {
			messages = append(messages, msg)
		} else {
			messages = append(messages, item.Raw)
		}
	}

	return mserrors.NewAPIError("graphql errors: "+strings.Join(messages, "; "), nil)
}

// firstMarketData returns the first element of data.marketData from a raw response.
func firstMarketData(body []byte, symbol string) (gjson.Result, error) {
	item := gjson.GetBytes(body, "data.marketData.0")
	if !item.Exists() {
		return gjson.Result{}, mserrors.NewSymbolNotFoundError(
			fmt.Sprintf("symbol not found: %q", symbol), nil, symbol,
		)
	}
	return item, nil
}

// gScalar extracts the scalar value from a gjson result, unwrapping {value: X}
// wrappers used by the MarketSurge API. Returns the result unchanged if it is
// not a wrapper object. Filters the sentinel date "0001-01-01".
func gScalar(r gjson.Result) gjson.Result {
	if !r.Exists() || r.Type == gjson.Null {
		return r
	}
	if r.IsObject() {
		if v := r.Get("value"); v.Exists() {
			if v.Type == gjson.String && v.String() == "0001-01-01" {
				return gjson.Result{}
			}
			return v
		}
	}
	return r
}

// gStr extracts a string pointer, unwrapping {value: X} wrappers.
func gStr(r gjson.Result) *string {
	v := gScalar(r)
	if !v.Exists() || v.Type == gjson.Null {
		return nil
	}
	s := v.String()
	if s == "" {
		return nil
	}
	return &s
}

// gInt extracts an int pointer, unwrapping {value: X} wrappers.
func gInt(r gjson.Result) *int {
	v := gScalar(r)
	if !v.Exists() || v.Type == gjson.Null {
		return nil
	}
	n := int(v.Int())
	return &n
}

// gInt64 extracts an int64 pointer, unwrapping {value: X} wrappers.
func gInt64(r gjson.Result) *int64 {
	v := gScalar(r)
	if !v.Exists() || v.Type == gjson.Null {
		return nil
	}
	n := v.Int()
	return &n
}

// gFloat extracts a float64 pointer, unwrapping {value: X} wrappers.
func gFloat(r gjson.Result) *float64 {
	v := gScalar(r)
	if !v.Exists() || v.Type == gjson.Null {
		return nil
	}
	f := v.Float()
	return &f
}

// gBool extracts a bool pointer, unwrapping {value: X} wrappers.
func gBool(r gjson.Result) *bool {
	v := gScalar(r)
	if !v.Exists() || v.Type == gjson.Null {
		return nil
	}
	b := v.Bool()
	return &b
}

// buildSlice maps a gjson array to a typed slice.
func buildSlice[T any](items []gjson.Result, fn func(gjson.Result) T) []T {
	result := make([]T, 0, len(items))
	for _, item := range items {
		result = append(result, fn(item))
	}
	return result
}

// stringify converts a gjson result to its string representation.
func stringify(r gjson.Result) string {
	if !r.Exists() {
		return ""
	}
	if r.Type == gjson.String {
		return r.String()
	}
	return r.Raw
}

// firstExisting returns the first result that exists and is not null.
func firstExisting(results ...gjson.Result) gjson.Result {
	for _, r := range results {
		if r.Exists() && r.Type != gjson.Null {
			return r
		}
	}
	return gjson.Result{}
}

// stringSlice converts a gjson array to a string slice, skipping empty values.
func stringSlice(items []gjson.Result) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if s := item.String(); s != "" {
			result = append(result, s)
		}
	}
	return result
}
