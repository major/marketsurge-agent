package commands

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStockAnalyzeSuccess(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "analyze", "AAPL"))

	result := parseJSONEnvelope(t, &buf)
	data, _ := result["data"].(map[string]any)
	assert.Equal(t, "AAPL", data["symbol"])
	assert.Contains(t, data, "stock")
	assert.Contains(t, data, "fundamentals")
	assert.Contains(t, data, "ownership")
}

func TestStockAnalyzeMultiSymbol(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "analyze", "AAPL", "MSFT"))

	result := parseJSONEnvelope(t, &buf)

	// Multi-symbol returns an array.
	data, ok := result["data"].([]any)
	require.True(t, ok, "multi-symbol data should be an array")
	assert.Len(t, data, 2)

	meta, _ := result["metadata"].(map[string]any)
	symbols, _ := meta["symbols"].([]any)
	assert.Len(t, symbols, 2)
}

func TestStockAnalyzePartialFailureWithCompactFlatOutput(t *testing.T) {
	server := stockAnalyzePartialServer(t)
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "analyze", "--compact", "--flat", "AAPL", "MSFT"))

	result := parseJSONEnvelope(t, &buf)
	data, ok := result["data"].([]any)
	require.True(t, ok, "partial multi-symbol data should remain an array")
	assert.Len(t, data, 2)
	assert.Contains(t, result, "errors")

	first, ok := data[0].(map[string]any)
	require.True(t, ok, "flattened successful item should be an object")
	assert.Contains(t, first, "pricing_market_cap")
	assert.NotContains(t, first, "pricing_market_cap_formatted")
	assert.NotContains(t, first, "stock")

	second, ok := data[1].(map[string]any)
	require.True(t, ok, "flattened failed item should be an object")
	assert.Equal(t, "MSFT", second["symbol"])
	assert.NotContains(t, second, "stock")

	errors, ok := result["errors"].([]any)
	require.True(t, ok, "partial envelope should include top-level errors")
	assert.NotEmpty(t, errors)
}

func TestStockAnalyzeTickersFlag(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "analyze", "--tickers", "AAPL, MSFT"))

	result := parseJSONEnvelope(t, &buf)
	data, ok := result["data"].([]any)
	require.True(t, ok, "--tickers should return multi-symbol data as an array")
	assert.Len(t, data, 2)

	meta, _ := result["metadata"].(map[string]any)
	symbols, _ := meta["symbols"].([]any)
	assert.Equal(t, []any{"AAPL", "MSFT"}, symbols)
}

func TestStockAnalyzeCompactOutput(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "analyze", "--compact", "AAPL"))

	result := parseJSONEnvelope(t, &buf)
	data, _ := result["data"].(map[string]any)
	stock, _ := data["stock"].(map[string]any)
	pricing, _ := stock["pricing"].(map[string]any)
	assert.Contains(t, pricing, "market_cap")
	assert.NotContains(t, pricing, "market_cap_formatted")
	assert.NotContains(t, pricing, "forward_price_to_earnings_ratio_formatted")
}

func TestStockAnalyzeFlatOutput(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "analyze", "--flat", "AAPL"))

	result := parseJSONEnvelope(t, &buf)
	data, _ := result["data"].(map[string]any)
	assert.Equal(t, "AAPL", data["symbol"])
	assert.NotContains(t, data, "stock")
	assert.Contains(t, data, "ratings_composite")
	assert.Contains(t, data, "pricing_market_cap")
	assert.Contains(t, data, "fundamentals_reported_earnings")
}

func TestStockAnalyzeCompactFlatOutput(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "analyze", "--compact", "--flat", "AAPL"))

	result := parseJSONEnvelope(t, &buf)
	data, _ := result["data"].(map[string]any)
	assert.Contains(t, data, "pricing_market_cap")
	assert.NotContains(t, data, "pricing_market_cap_formatted")
	assert.NotContains(t, data, "pricing_forward_price_to_earnings_ratio_formatted")
}

func TestStockAnalyzeMissingSymbol(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	err := runTestCommand(t, cmd, "analyze")
	require.Error(t, err)
	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Empty(t, buf.String())
}

func TestStockAnalyzeTotalFailure(t *testing.T) {
	t.Parallel()
	server := jsonServer(emptyMarketDataFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	err := runTestCommand(t, cmd, "analyze", "MISSING")
	require.Error(t, err)
	assert.Empty(t, buf.String())
}

func stockAnalyzePartialServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(string(body), "MSFT") {
			_, err = w.Write([]byte(emptyMarketDataFixture()))
		} else {
			_, err = w.Write([]byte(stockResponseFixture()))
		}
		if err != nil {
			t.Errorf("write response body: %v", err)
		}
	}))
}
