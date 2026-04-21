package client

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

func mapGraphQLError(raw map[string]any) error {
	entries, ok := raw["errors"]
	if !ok {
		return nil
	}

	errorList, ok := entries.([]any)
	if !ok || len(errorList) == 0 {
		return nil
	}

	messages := make([]string, 0, len(errorList))
	for _, entry := range errorList {
		if item, ok := entry.(map[string]any); ok {
			if message, ok := item["message"].(string); ok && message != "" {
				messages = append(messages, message)
				continue
			}
		}
		messages = append(messages, stringify(entry))
	}

	return mserrors.NewAPIError("graphql errors: "+strings.Join(messages, "; "), nil)
}

func firstMarketData(raw map[string]any, symbol string) (map[string]any, error) {
	marketData := getNestedSlice(raw, "data", "marketData")
	if len(marketData) == 0 {
		return nil, mserrors.NewSymbolNotFoundError(fmt.Sprintf("symbol not found: %q", symbol), nil, symbol)
	}

	item, ok := marketData[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid marketData item for %q", symbol)
	}

	return item, nil
}

func getNestedMap(root map[string]any, keys ...string) map[string]any {
	current := any(root)
	for _, key := range keys {
		mapping, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = mapping[key]
	}

	mapping, _ := current.(map[string]any)
	return mapping
}

func getNestedSlice(root map[string]any, keys ...string) []any {
	current := any(root)
	for _, key := range keys {
		mapping, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = mapping[key]
	}

	slice, _ := current.([]any)
	return slice
}

func firstMap(items []any) map[string]any {
	if len(items) == 0 {
		return nil
	}
	mapping, _ := items[0].(map[string]any)
	return mapping
}

func scalarValue(node any) any {
	if mapping, ok := node.(map[string]any); ok {
		if value, ok := mapping["value"]; ok {
			if stringValue, ok := value.(string); ok && stringValue == "0001-01-01" {
				return nil
			}
			return value
		}
	}
	return node
}

func formattedValue(node any) *string {
	mapping, ok := node.(map[string]any)
	if !ok {
		return nil
	}
	return stringPtr(mapping["formattedValue"])
}

func stringPtr(value any) *string {
	if value == nil {
		return nil
	}
	resolved := scalarValue(value)
	if resolved == nil {
		return nil
	}
	switch typed := resolved.(type) {
	case string:
		if typed == "" {
			return nil
		}
		return &typed
	default:
		text := stringify(typed)
		if text == "" {
			return nil
		}
		return &text
	}
}

func intPtr(value any) *int {
	if value == nil {
		return nil
	}
	resolved := scalarValue(value)
	switch typed := resolved.(type) {
	case float64:
		result := int(typed)
		return &result
	case int:
		result := typed
		return &result
	case int64:
		result := int(typed)
		return &result
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return nil
		}
		result := int(parsed)
		return &result
	case string:
		if typed == "" {
			return nil
		}
		parsed, err := strconv.Atoi(typed)
		if err != nil {
			return nil
		}
		return &parsed
	default:
		return nil
	}
}

func int64Ptr(value any) *int64 {
	if value == nil {
		return nil
	}
	resolved := scalarValue(value)
	switch typed := resolved.(type) {
	case float64:
		result := int64(typed)
		return &result
	case int64:
		result := typed
		return &result
	case int:
		result := int64(typed)
		return &result
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return nil
		}
		return &parsed
	case string:
		if typed == "" {
			return nil
		}
		parsed, err := strconv.ParseInt(typed, 10, 64)
		if err != nil {
			return nil
		}
		return &parsed
	default:
		return nil
	}
}

func floatPtr(value any) *float64 {
	if value == nil {
		return nil
	}
	resolved := scalarValue(value)
	switch typed := resolved.(type) {
	case float64:
		result := typed
		return &result
	case int:
		result := float64(typed)
		return &result
	case int64:
		result := float64(typed)
		return &result
	case json.Number:
		parsed, err := typed.Float64()
		if err != nil {
			return nil
		}
		return &parsed
	case string:
		if typed == "" {
			return nil
		}
		parsed, err := strconv.ParseFloat(typed, 64)
		if err != nil {
			return nil
		}
		return &parsed
	default:
		return nil
	}
}

func boolPtr(value any) *bool {
	if value == nil {
		return nil
	}
	resolved := scalarValue(value)
	switch typed := resolved.(type) {
	case bool:
		result := typed
		return &result
	case string:
		parsed, err := strconv.ParseBool(typed)
		if err != nil {
			return nil
		}
		return &parsed
	default:
		return nil
	}
}

func stringify(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprintf("%v", typed)
		}
		return string(encoded)
	}
}
