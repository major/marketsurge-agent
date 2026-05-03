package client

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type graphQLErrorCode interface {
	ErrorCode() string
}

func TestGraphQLStringPtr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{name: "non-empty string", input: "AAPL", expected: "AAPL"},
		{name: "empty string", input: "", expected: nil},
		{name: "nil", input: nil, expected: nil},
		{name: "non-string fallback", input: 42, expected: "42"},
		{name: "scalar value wrapper", input: map[string]any{"value": "MSFT"}, expected: "MSFT"},
		{name: "scalar date zero value", input: map[string]any{"value": "0001-01-01"}, expected: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertPointerValue(t, tt.expected, stringPtr(tt.input))
		})
	}
}

func TestGraphQLIntPtr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{name: "float64 truncates", input: float64(42.9), expected: 42},
		{name: "int", input: 43, expected: 43},
		{name: "int64", input: int64(44), expected: 44},
		{name: "json number valid", input: json.Number("45"), expected: 45},
		{name: "json number invalid", input: json.Number("45.2"), expected: nil},
		{name: "string valid", input: "46", expected: 46},
		{name: "string invalid", input: "abc", expected: nil},
		{name: "string empty", input: "", expected: nil},
		{name: "nil", input: nil, expected: nil},
		{name: "unknown type", input: true, expected: nil},
		{name: "scalar wrapper", input: map[string]any{"value": json.Number("47")}, expected: 47},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertPointerValue(t, tt.expected, intPtr(tt.input))
		})
	}
}

func TestGraphQLInt64Ptr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{name: "float64 truncates", input: float64(52.9), expected: int64(52)},
		{name: "int", input: 53, expected: int64(53)},
		{name: "int64", input: int64(54), expected: int64(54)},
		{name: "json number valid", input: json.Number("55"), expected: int64(55)},
		{name: "json number invalid", input: json.Number("55.2"), expected: nil},
		{name: "string valid", input: "56", expected: int64(56)},
		{name: "string invalid", input: "abc", expected: nil},
		{name: "string empty", input: "", expected: nil},
		{name: "nil", input: nil, expected: nil},
		{name: "unknown type", input: true, expected: nil},
		{name: "scalar wrapper", input: map[string]any{"value": int64(57)}, expected: int64(57)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertPointerValue(t, tt.expected, int64Ptr(tt.input))
		})
	}
}

func TestGraphQLFloatPtr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{name: "float64", input: float64(1.25), expected: 1.25},
		{name: "int", input: 2, expected: 2.0},
		{name: "int64", input: int64(3), expected: 3.0},
		{name: "json number valid", input: json.Number("4.5"), expected: 4.5},
		{name: "json number invalid", input: json.Number("bad"), expected: nil},
		{name: "string valid", input: "5.5", expected: 5.5},
		{name: "string invalid", input: "abc", expected: nil},
		{name: "string empty", input: "", expected: nil},
		{name: "nil", input: nil, expected: nil},
		{name: "unknown type", input: true, expected: nil},
		{name: "scalar wrapper", input: map[string]any{"value": json.Number("6.5")}, expected: 6.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertPointerValue(t, tt.expected, floatPtr(tt.input))
		})
	}
}

func TestGraphQLBoolPtr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{name: "bool true", input: true, expected: true},
		{name: "bool false", input: false, expected: false},
		{name: "string true", input: "true", expected: true},
		{name: "string false", input: "false", expected: false},
		{name: "string one", input: "1", expected: true},
		{name: "string zero", input: "0", expected: false},
		{name: "string invalid", input: "maybe", expected: nil},
		{name: "nil", input: nil, expected: nil},
		{name: "unknown type", input: 1, expected: nil},
		{name: "scalar wrapper", input: map[string]any{"value": "true"}, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertPointerValue(t, tt.expected, boolPtr(tt.input))
		})
	}
}

func TestGraphQLScalarValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{name: "plain value", input: "AAPL", expected: "AAPL"},
		{name: "map with value key", input: map[string]any{"value": 99}, expected: 99},
		{name: "map with zero date", input: map[string]any{"value": "0001-01-01"}, expected: nil},
		{name: "map without value key", input: map[string]any{"formattedValue": "99"}, expected: map[string]any{"formattedValue": "99"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, scalarValue(tt.input))
		})
	}
}

func TestGraphQLFormattedValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{name: "map with formattedValue key", input: map[string]any{"formattedValue": "$1B"}, expected: "$1B"},
		{name: "map without formattedValue key", input: map[string]any{"value": 1}, expected: nil},
		{name: "non-map input", input: "$1B", expected: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertPointerValue(t, tt.expected, formattedValue(tt.input))
		})
	}
}

func TestGraphQLNestedMap(t *testing.T) {
	t.Parallel()

	root := map[string]any{"data": map[string]any{"marketData": map[string]any{"symbol": "AAPL"}}, "flat": "value"}
	tests := []struct {
		name     string
		keys     []string
		expected map[string]any
	}{
		{name: "valid path", keys: []string{"data", "marketData"}, expected: map[string]any{"symbol": "AAPL"}},
		{name: "missing key", keys: []string{"data", "missing"}, expected: nil},
		{name: "non-map intermediate", keys: []string{"flat", "symbol"}, expected: nil},
		{name: "empty keys", keys: nil, expected: root},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, getNestedMap(root, tt.keys...))
		})
	}
}

func TestGraphQLNestedSlice(t *testing.T) {
	t.Parallel()

	marketData := []any{map[string]any{"symbol": "AAPL"}}
	root := map[string]any{"data": map[string]any{"marketData": marketData, "marketDataMap": map[string]any{}}, "flat": "value"}
	tests := []struct {
		name     string
		keys     []string
		expected []any
	}{
		{name: "valid path", keys: []string{"data", "marketData"}, expected: marketData},
		{name: "missing key", keys: []string{"data", "missing"}, expected: nil},
		{name: "non-slice value", keys: []string{"data", "marketDataMap"}, expected: nil},
		{name: "non-map intermediate", keys: []string{"flat", "marketData"}, expected: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, getNestedSlice(root, tt.keys...))
		})
	}
}

func TestGraphQLFirstMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []any
		expected map[string]any
	}{
		{name: "non-empty slice with map", input: []any{map[string]any{"symbol": "AAPL"}}, expected: map[string]any{"symbol": "AAPL"}},
		{name: "non-empty slice with non-map", input: []any{"AAPL"}, expected: nil},
		{name: "empty slice", input: []any{}, expected: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, firstMap(tt.input))
		})
	}
}

func TestGraphQLMapGraphQLError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       map[string]any
		expectsCode bool
		contains    string
	}{
		{name: "no errors key", input: map[string]any{}, expectsCode: false},
		{name: "non-list errors", input: map[string]any{"errors": "bad"}, expectsCode: false},
		{name: "empty list", input: map[string]any{"errors": []any{}}, expectsCode: false},
		{name: "list with message", input: map[string]any{"errors": []any{map[string]any{"message": "failed"}}}, expectsCode: true, contains: "failed"},
		{name: "list with non-map entry", input: map[string]any{"errors": []any{"failed"}}, expectsCode: true, contains: "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := mapGraphQLError(tt.input)
			if !tt.expectsCode {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.contains)

			var coded graphQLErrorCode
			require.ErrorAs(t, err, &coded)
			assert.Equal(t, "API_ERROR", coded.ErrorCode())
		})
	}
}

func TestGraphQLFirstMarketData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        map[string]any
		expected     map[string]any
		expectedCode string
	}{
		{name: "empty marketData", input: map[string]any{"data": map[string]any{"marketData": []any{}}}, expectedCode: "SYMBOL_NOT_FOUND"},
		{name: "non-map item", input: map[string]any{"data": map[string]any{"marketData": []any{"bad"}}}, expectedCode: "API_ERROR"},
		{name: "valid item", input: map[string]any{"data": map[string]any{"marketData": []any{map[string]any{"symbol": "AAPL"}}}}, expected: map[string]any{"symbol": "AAPL"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := firstMarketData(tt.input, "AAPL")
			if tt.expectedCode == "" {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
				return
			}

			require.Error(t, err)
			assert.Nil(t, result)

			var coded graphQLErrorCode
			require.ErrorAs(t, err, &coded)
			assert.Equal(t, tt.expectedCode, coded.ErrorCode())
		})
	}
}

func TestGraphQLBuildSlice(t *testing.T) {
	t.Parallel()

	mapper := func(item map[string]any) string {
		value, ok := item["symbol"].(string)
		if !ok {
			return ""
		}
		return value
	}

	tests := []struct {
		name     string
		input    []any
		expected []string
	}{
		{name: "valid items", input: []any{map[string]any{"symbol": "AAPL"}, map[string]any{"symbol": "MSFT"}}, expected: []string{"AAPL", "MSFT"}},
		{name: "items with non-map entries", input: []any{map[string]any{"symbol": "AAPL"}, "bad", map[string]any{"symbol": "MSFT"}}, expected: []string{"AAPL", "MSFT"}},
		{name: "empty slice", input: []any{}, expected: []string{}},
		{name: "nil slice", input: nil, expected: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := buildSlice(tt.input, mapper)
			require.NotNil(t, result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGraphQLStringify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		input            any
		expected         string
		expectedContains string
	}{
		{name: "string", input: "AAPL", expected: "AAPL"},
		{name: "json number", input: json.Number("99"), expected: "99"},
		{name: "int", input: 42, expected: "42"},
		{name: "bool", input: true, expected: "true"},
		{name: "nil", input: nil, expected: "null"},
		{name: "struct", input: struct{ Symbol string }{Symbol: "AAPL"}, expected: `{"Symbol":"AAPL"}`},
		{name: "marshal fallback", input: func() {}, expectedContains: "0x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := stringify(tt.input)
			if tt.expectedContains != "" {
				assert.Contains(t, result, tt.expectedContains)
				return
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func assertPointerValue[T any](t *testing.T, expected any, actual *T) {
	t.Helper()
	if expected == nil {
		assert.Nil(t, actual)
		return
	}
	require.NotNil(t, actual)
	assert.Equal(t, expected, *actual)
}
