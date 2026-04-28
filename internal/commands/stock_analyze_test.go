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

func TestStockAnalyzeTechnicalSignals(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "analyze", "AAPL"))

	result := parseJSONEnvelope(t, &buf)
	data, _ := result["data"].(map[string]any)
	stock, _ := data["stock"].(map[string]any)

	signals, ok := stock["signals"].(map[string]any)
	require.True(t, ok, "stock should include technical signals")
	assert.Equal(t, true, signals["blue_dot"])
	assert.Equal(t, "2024-12-20", signals["blue_dot_date"])
	assert.Equal(t, true, signals["ant_signal"])

	basePattern, ok := stock["base_pattern"].(map[string]any)
	require.True(t, ok, "stock should include base pattern summary")
	assert.Equal(t, "Cup With Handle", basePattern["pattern_type"])
	assert.Equal(t, "STAGE_2", basePattern["base_stage"])
	assert.Equal(t, 199.99, basePattern["pivot_price"])
	assert.Equal(t, float64(7), basePattern["base_length_weeks"])
	assert.Equal(t, 18.5, basePattern["base_depth_percent"])
	assert.Equal(t, 42.3, basePattern["volume_at_pivot_pct"])

	pricing, _ := stock["pricing"].(map[string]any)
	assert.Equal(t, []any{"2024-12-20"}, pricing["blue_dot_daily_dates"])
	assert.Equal(t, []any{"2024-12-16"}, pricing["blue_dot_weekly_dates"])
	assert.Equal(t, []any{"2024-12-18"}, pricing["ant_dates"])

	patterns, ok := stock["patterns"].([]any)
	require.True(t, ok, "stock should include parsed patterns")
	assert.Len(t, patterns, 1)
	tightAreas, ok := stock["tight_areas"].([]any)
	require.True(t, ok, "stock should include parsed tight areas")
	assert.Len(t, tightAreas, 1)
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
	assert.Contains(t, data, "base_pattern_pivot_price")
	assert.Contains(t, data, "signals_blue_dot")
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

func TestStockAnalyzeSummaryOutput(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "analyze", "--summary", "AAPL", "MSFT"))

	result := parseJSONEnvelope(t, &buf)
	data, ok := result["data"].([]any)
	require.True(t, ok, "summary multi-symbol data should be an array")
	assert.Len(t, data, 2)

	first, ok := data[0].(map[string]any)
	require.True(t, ok, "summary item should be an object")
	assert.Equal(t, "AAPL", first["symbol"])
	assert.Equal(t, float64(99), first["composite"])
	assert.Equal(t, float64(95), first["eps"])
	assert.Equal(t, float64(90), first["rs"])
	assert.Equal(t, "B", first["ad"])
	assert.Equal(t, "A", first["smr"])
	assert.Equal(t, true, first["blue_dot"])
	assert.Equal(t, true, first["ant_signal"])
	assert.Equal(t, "Cup With Handle", first["base_type"])
	assert.Equal(t, "STAGE_2", first["base_stage"])
	assert.Equal(t, 199.99, first["pivot"])
	assert.Equal(t, 18.5, first["base_depth_percent"])
	assert.Equal(t, float64(95), first["industry_group_rs"])
	assert.Equal(t, 1.2, first["up_down_volume"])
	assert.Equal(t, float64(60), first["funds_float_percent"])
	assert.Equal(t, 2.3, first["atr_percent"])
	assert.Equal(t, float64(5000000), first["avg_dollar_volume"])
	assert.NotContains(t, first, "stock")
	assert.NotContains(t, first, "fundamentals")
	assert.NotContains(t, first, "ownership")

	meta, _ := result["metadata"].(map[string]any)
	assert.Equal(t, "summary", meta["mode"])
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
