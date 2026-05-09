package cmd

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
)

func TestStockGetSuccess(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeCommandWithClient(t, newStockCmd(), c, "get", "AAPL")
	require.NoError(t, err)
	result := parseJSONEnvelope(t, output)
	assertSymbolMeta(t, result, "AAPL")
}

func TestStockGetSymbolFlag(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeCommandWithClient(t, newStockCmd(), c, "get", "--symbol", "AAPL")
	require.NoError(t, err)
	result := parseJSONEnvelope(t, output)
	assertSymbolMeta(t, result, "AAPL")
}

func TestStockGetSymbolNotFound(t *testing.T) {
	t.Parallel()
	server := jsonServer(emptyMarketDataFixture())
	defer server.Close()
	c := testClient(t, server)

	_, err := executeCommandWithClient(t, newStockCmd(), c, "get", "MISSING")
	require.Error(t, err)
	var snf *mserrors.SymbolNotFoundError
	assert.ErrorAs(t, err, &snf)
}

func TestStockGetMissingSymbol(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	_, err := executeCommandWithClient(t, newStockCmd(), c, "get")
	require.Error(t, err)
}

func TestStockAnalyzeSuccess(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeStockAnalyze(t, c, "analyze", "AAPL")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
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

	output, err := executeStockAnalyze(t, c, "analyze", "AAPL")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
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

	output, err := executeStockAnalyze(t, c, "analyze", "AAPL", "MSFT")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
	data, ok := result["data"].([]any)
	require.True(t, ok, "multi-symbol data should be an array")
	assert.Len(t, data, 2)

	meta, _ := result["metadata"].(map[string]any)
	symbols, _ := meta["symbols"].([]any)
	assert.Len(t, symbols, 2)
}

func TestStockAnalyzeConcurrentAllSuccess(t *testing.T) {
	t.Parallel()
	server := stockAnalyzeConcurrentServer(t, nil)
	defer server.Close()
	c := testClient(t, server)

	output, err := executeStockAnalyze(t, c, "analyze", "AAPL", "MSFT", "NVDA")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
	data, ok := result["data"].([]any)
	require.True(t, ok, "concurrent multi-symbol data should be an array")
	require.Len(t, data, 3)
	assert.NotContains(t, result, "errors")

	seen := make(map[string]bool, len(data))
	for _, item := range data {
		analysis, ok := item.(map[string]any)
		require.True(t, ok, "analysis item should be an object")
		seen[analysis["symbol"].(string)] = true
		assert.Contains(t, analysis, "stock")
		assert.Contains(t, analysis, "fundamentals")
		assert.Contains(t, analysis, "ownership")
	}
	assert.Equal(t, map[string]bool{"AAPL": true, "MSFT": true, "NVDA": true}, seen)
}

func TestStockAnalyzeConcurrentAllFailure(t *testing.T) {
	t.Parallel()
	server := stockAnalyzeConcurrentServer(t, map[string]bool{"AAPL": true, "MSFT": true, "NVDA": true})
	defer server.Close()
	c := testClient(t, server)

	output, err := executeStockAnalyze(t, c, "analyze", "AAPL", "MSFT", "NVDA")
	require.Error(t, err)
	assert.Empty(t, output)

	var apiErr *mserrors.APIError
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 33, apiErr.ExitCode())
	assert.Contains(t, apiErr.Error(), "analysis failed for all symbols")
}

func TestStockAnalyzeConcurrentMixed(t *testing.T) {
	t.Parallel()
	server := stockAnalyzeConcurrentServer(t, map[string]bool{"MSFT": true})
	defer server.Close()
	c := testClient(t, server)

	output, err := executeStockAnalyze(t, c, "analyze", "AAPL", "MSFT", "NVDA")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
	assert.Contains(t, result, "data")
	assert.Contains(t, result, "errors")

	data, ok := result["data"].([]any)
	require.True(t, ok, "partial concurrent data should be an array")
	assert.Len(t, data, 3)

	errors, ok := result["errors"].([]any)
	require.True(t, ok, "partial envelope should include top-level errors")
	assert.NotEmpty(t, errors)
	assert.Contains(t, output, "MSFT")
}

func TestStockAnalyzePartialFailureWithCompactFlatOutput(t *testing.T) {
	server := stockAnalyzePartialServer(t)
	defer server.Close()
	c := testClient(t, server)

	output, err := executeStockAnalyze(t, c, "analyze", "--compact", "--flat", "AAPL", "MSFT")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
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

	output, err := executeStockAnalyze(t, c, "analyze", "--tickers", "AAPL, MSFT")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
	data, ok := result["data"].([]any)
	require.True(t, ok, "--tickers should return multi-symbol data as an array")
	assert.Len(t, data, 2)

	meta, _ := result["metadata"].(map[string]any)
	symbols, _ := meta["symbols"].([]any)
	assert.Equal(t, []any{"AAPL", "MSFT"}, symbols)
}

func TestStockAnalyzeSymbolsFlag(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeStockAnalyze(t, c, "analyze", "--symbols", "AAPL, MSFT")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
	data, ok := result["data"].([]any)
	require.True(t, ok, "--symbols should return multi-symbol data as an array")
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

	output, err := executeStockAnalyze(t, c, "analyze", "--compact", "AAPL")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
	data, _ := result["data"].(map[string]any)
	stock, _ := data["stock"].(map[string]any)
	pricing, _ := stock["pricing"].(map[string]any)
	assert.Contains(t, pricing, "market_cap")
	assert.NotContains(t, pricing, "market_cap_formatted")
	assert.NotContains(t, pricing, "forward_price_to_earnings_ratio_formatted")
	assert.NotContains(t, pricing, "volume_moving_averages")
	assert.NotContains(t, pricing, "historical_price_statistics")

	company, _ := stock["company"].(map[string]any)
	assert.Contains(t, company, "name")
	assert.NotContains(t, company, "website")
	assert.NotContains(t, company, "address")
	assert.NotContains(t, company, "city")
	assert.NotContains(t, company, "state_province")
	assert.NotContains(t, company, "instrument_sub_type")

	industry, _ := stock["industry"].(map[string]any)
	assert.Contains(t, industry, "name")
	assert.NotContains(t, industry, "code")
	assert.NotContains(t, stock, "corporate_actions")
	assert.NotContains(t, stock, "patterns")
	assert.NotContains(t, stock, "tight_areas")

	fundamentals, _ := data["fundamentals"].(map[string]any)
	assert.NotContains(t, fundamentals, "symbol")
	reportedEarnings, _ := fundamentals["reported_earnings"].([]any)
	firstEarnings, _ := reportedEarnings[0].(map[string]any)
	assert.NotContains(t, firstEarnings, "formatted_value")
	assert.NotContains(t, firstEarnings, "formatted_pct_change")
	assert.NotContains(t, firstEarnings, "period_offset")
}

func TestCompactAnalysisMapPrunesLowValueFields(t *testing.T) {
	t.Parallel()

	data := map[string]any{
		"symbol": "AAPL",
		"stock": map[string]any{
			"symbol": "AAPL",
			"pricing": map[string]any{
				"market_cap":                         3000000000000.0,
				"market_cap_formatted":               "$3T",
				"blue_dot_daily_dates":               []any{"2024-12-20", "2024-12-19", "2024-12-18", "2024-12-17"},
				"historical_price_statistics":        []any{map[string]any{"period_offset": "0"}},
				"price_to_cash_flow_ratio_formatted": "N/A",
			},
			"corporate_actions": map[string]any{},
			"patterns": []any{
				map[string]any{"id": "pattern-1", "pattern_type": "Cup With Handle"},
			},
		},
		"fundamentals": map[string]any{
			"symbol": "AAPL",
			"reported_earnings": []any{
				map[string]any{"period_offset": "0", "value": 1.5, "formatted_value": "1.5"},
				map[string]any{"period_offset": "-1", "value": 1.4},
				map[string]any{"period_offset": "-2", "value": 1.3},
				map[string]any{"period_offset": "-3", "value": 1.2},
				map[string]any{"period_offset": "-4", "value": 1.1},
			},
		},
	}

	compact := compactAnalysisMap(data)

	assert.Equal(t, "AAPL", compact["symbol"])
	stock, _ := compact["stock"].(map[string]any)
	assert.NotContains(t, stock, "symbol")
	assert.NotContains(t, stock, "corporate_actions")
	assert.NotContains(t, stock, "patterns")

	pricing, _ := stock["pricing"].(map[string]any)
	assert.Equal(t, 3000000000000.0, pricing["market_cap"])
	assert.NotContains(t, pricing, "market_cap_formatted")
	assert.NotContains(t, pricing, "price_to_cash_flow_ratio_formatted")
	assert.NotContains(t, pricing, "historical_price_statistics")
	assert.Equal(t, []any{"2024-12-20", "2024-12-19", "2024-12-18"}, pricing["blue_dot_daily_dates"])

	fundamentals, _ := compact["fundamentals"].(map[string]any)
	assert.NotContains(t, fundamentals, "symbol")
	reportedEarnings, _ := fundamentals["reported_earnings"].([]any)
	require.Len(t, reportedEarnings, 4)
	firstEarnings, _ := reportedEarnings[0].(map[string]any)
	assert.Equal(t, 1.5, firstEarnings["value"])
	assert.NotContains(t, firstEarnings, "period_offset")
	assert.NotContains(t, firstEarnings, "formatted_value")
}

func TestStockAnalyzeFlatOutput(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeStockAnalyze(t, c, "analyze", "--flat", "AAPL")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
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

	output, err := executeStockAnalyze(t, c, "analyze", "--compact", "--flat", "AAPL")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
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

	output, err := executeStockAnalyze(t, c, "analyze", "--summary", "AAPL", "MSFT")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
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
	assert.Equal(t, []any{"2024-12-18"}, first["ant_dates"])
	assert.Equal(t, antSignalExplanation, first["ant_explanation"])
	assert.Equal(t, "Cup With Handle", first["base_type"])
	assert.Equal(t, "STAGE_2", first["base_stage"])
	assert.Equal(t, 199.99, first["pivot"])
	assert.Equal(t, "2024-12-16", first["pivot_price_date"])
	assert.Equal(t, 18.5, first["base_depth_percent"])
	assert.Equal(t, "2024-01-01", first["pricing_start_date"])
	assert.Equal(t, "2024-12-31", first["pricing_end_date"])
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

func TestStockAnalyzeSetupOutput(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeStockAnalyze(t, c, "analyze", "--setup", "AAPL")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
	data, ok := result["data"].(map[string]any)
	require.True(t, ok, "setup single-symbol data should be an object")
	assert.Equal(t, "AAPL", data["symbol"])
	assert.Equal(t, float64(99), data["composite"])
	assert.Equal(t, "2024-12-31", data["pricing_end_date"])
	assert.Equal(t, float64(7), data["base_length_weeks"])
	assert.Equal(t, 42.3, data["volume_at_pivot_percent"])
	assert.Equal(t, "60%", data["ownership_funds_float_percent"])
	require.Contains(t, data, "quarterly_funds")
	assert.NotContains(t, data, "stock")
	assert.NotContains(t, data, "fundamentals")
	assert.NotContains(t, data, "ownership")

	meta, _ := result["metadata"].(map[string]any)
	assert.Equal(t, "setup", meta["mode"])
}

func TestStockAnalyzeSummaryOmitsAntDetailsWhenSignalFalse(t *testing.T) {
	t.Parallel()
	antSignal := false
	result := AnalysisResult{
		Symbol: "AAPL",
		Stock: &models.StockData{
			Signals: &models.Signals{AntSignal: &antSignal},
			Pricing: &models.Pricing{AntDates: []string{"2024-12-18"}},
		},
	}

	summary := analysisSummaryMap(result)

	assert.Equal(t, false, summary["ant_signal"])
	assert.NotContains(t, summary, "ant_dates")
	assert.NotContains(t, summary, "ant_explanation")
}

func TestStockAnalyzeStructTags(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[StockAnalyzeOptions]()
	tests := []struct {
		field    string
		flag     string
		short    string
		group    string
		descr    string
		hasShort bool
	}{
		{field: "Symbols", flag: "symbols", short: "s", group: "Input", descr: "Stock symbols to analyze, for example AAPL,MSFT; accepts comma-separated or repeated values; positional symbols remain supported for shell use", hasShort: true},
		{field: "Tickers", flag: "tickers", short: "t", group: "Input", descr: "Additional stock symbols to analyze; accepts comma-separated or repeated values", hasShort: true},
		{field: "Compact", flag: "compact", group: "Output Format", descr: "Remove low-value fields such as formatted duplicates, profile metadata, internal IDs, empty fields, and stale arrays while keeping decision-relevant raw values; compatible with --flat. Example: stock analyze AAPL --compact"},
		{field: "Flat", flag: "flat", group: "Output Format", descr: "Flatten nested analysis fields into single-level JSON keys, for example stock.pricing.market_cap becomes pricing_market_cap; compatible with --compact. Example: stock analyze AAPL --flat"},
		{field: "Summary", flag: "summary", group: "Output Format", descr: "Return compact screening objects for ranking many symbols. Example: stock analyze --summary AAPL MSFT NVDA"},
		{field: "Setup", flag: "setup", group: "Output Format", descr: "Return --setup trade triage fields: summary fields plus base_length_weeks, volume_at_pivot_percent, ownership_funds_float_percent, and quarterly_funds. Example: stock analyze AAPL --setup"},
	}
	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			f, ok := typ.FieldByName(tt.field)
			require.True(t, ok, "field %s should exist", tt.field)
			assert.Equal(t, tt.flag, f.Tag.Get("flag"), "flag tag")
			assert.Equal(t, tt.group, f.Tag.Get("flaggroup"), "flaggroup tag")
			assert.Equal(t, tt.descr, f.Tag.Get("flagdescr"), "flagdescr tag")
			if tt.hasShort {
				assert.Equal(t, tt.short, f.Tag.Get("flagshort"), "flagshort tag")
			}
		})
	}
}

func TestStockAnalyzeMergeSymbols(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		symbols []string
		tickers []string
		args    []string
		want    []string
	}{
		{name: "positional only", args: []string{"AAPL"}, want: []string{"AAPL"}},
		{name: "symbols only", symbols: []string{"AAPL,MSFT"}, want: []string{"AAPL", "MSFT"}},
		{name: "tickers only", tickers: []string{"MSFT"}, want: []string{"MSFT"}},
		{name: "all merged", symbols: []string{"MSFT"}, tickers: []string{"NVDA"}, args: []string{"AAPL"}, want: []string{"AAPL", "MSFT", "NVDA"}},
		{name: "deduplicates", tickers: []string{"AAPL"}, args: []string{"AAPL"}, want: []string{"AAPL"}},
		{name: "trims whitespace", tickers: []string{" MSFT "}, args: []string{" AAPL "}, want: []string{"AAPL", "MSFT"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &StockAnalyzeOptions{Symbols: tt.symbols, Tickers: tt.tickers}
			got := opts.MergeSymbols(tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveSingleSymbol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		symbolFlag string
		want       string
		wantErr    bool
	}{
		{name: "flag only", symbolFlag: "AAPL", want: "AAPL"},
		{name: "positional only", args: []string{"MSFT"}, want: "MSFT"},
		{name: "matching flag and positional", args: []string{"AAPL"}, symbolFlag: "AAPL", want: "AAPL"},
		{name: "trims whitespace", args: []string{" MSFT "}, symbolFlag: "  ", want: "MSFT"},
		{name: "missing symbol", wantErr: true},
		{name: "conflicting flag and positional", args: []string{"MSFT"}, symbolFlag: "AAPL", wantErr: true},
		{name: "too many positional", args: []string{"AAPL", "MSFT"}, wantErr: true},
		{name: "empty flag and empty positional", args: []string{"  "}, symbolFlag: " ", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveSingleSymbol(tt.args, tt.symbolFlag)
			if tt.wantErr {
				require.Error(t, err)
				var verr *mserrors.ValidationError
				assert.ErrorAs(t, err, &verr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMultiSymbolOptionsResolveSymbols(t *testing.T) {
	t.Parallel()

	opts := &MultiSymbolOptions{Symbols: []string{" MSFT,NVDA ", "AAPL"}}
	got := opts.ResolveSymbols([]string{"AAPL", "TSLA"}, []string{"NVDA,AMD"})

	assert.Equal(t, []string{"AAPL", "TSLA", "MSFT", "NVDA", "AMD"}, got)
}

func TestStockAnalyzeExamples(t *testing.T) {
	t.Parallel()

	cmd := newStockAnalyzeCmd()
	assert.Contains(t, cmd.Example, "marketsurge-agent stock analyze AAPL")
	assert.Contains(t, cmd.Example, "--tickers AAPL,MSFT,NVDA --compact")
	assert.Contains(t, cmd.Example, "--summary AAPL MSFT NVDA")
	assert.Contains(t, cmd.Example, "stock analyze AAPL --setup")
	assert.Contains(t, cmd.Example, "--flat --compact")

	typ := reflect.TypeFor[StockAnalyzeOptions]()
	for _, fieldName := range []string{"Compact", "Flat"} {
		t.Run(fieldName, func(t *testing.T) {
			field, ok := typ.FieldByName(fieldName)
			require.True(t, ok)
			descr := field.Tag.Get("flagdescr")
			assert.Contains(t, descr, "compatible with")
			assert.Contains(t, descr, "Example: stock analyze")
		})
	}

	field, ok := typ.FieldByName("Summary")
	require.True(t, ok)
	assert.Contains(t, field.Tag.Get("flagdescr"), "compact screening objects")
	assert.Contains(t, field.Tag.Get("flagdescr"), "Example: stock analyze")

	field, ok = typ.FieldByName("Setup")
	require.True(t, ok)
	descr := field.Tag.Get("flagdescr")
	assert.Contains(t, descr, "Return --setup trade triage fields")
	assert.Contains(t, descr, "base_length_weeks")
	assert.Contains(t, descr, "quarterly_funds")
	assert.Contains(t, descr, "Example: stock analyze")
}

func TestStockAnalyzeMissingSymbol(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	output, err := executeStockAnalyze(t, c, "analyze")
	require.Error(t, err)
	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Equal(t, "missing symbols: pass positional symbols like stock analyze AAPL MSFT or use --tickers AAPL,MSFT", err.Error())
	assert.Empty(t, output)
}

func TestStockAnalyzeTotalFailure(t *testing.T) {
	t.Parallel()
	server := jsonServer(emptyMarketDataFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeStockAnalyze(t, c, "analyze", "MISSING")
	require.Error(t, err)
	assert.Empty(t, output)
}

func TestStockAnalyzeSparseMarketDataTotalFailure(t *testing.T) {
	t.Parallel()
	server := jsonServer(sparseMarketDataFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeStockAnalyze(t, c, "analyze", "CYBR")
	require.Error(t, err)
	assert.Empty(t, output)

	var apiErr *mserrors.APIError
	assert.ErrorAs(t, err, &apiErr)
	assert.Contains(t, err.Error(), "analysis failed for all symbols")
}

// executeStockAnalyze creates a stock command tree, injects the client into subcommand
// contexts, and executes with the given args. structcli.Bind sets a scope context on the
// analyze subcommand, which prevents cobra's parent-to-child context propagation. This
// helper layers the client onto each subcommand's existing context so both the structcli
// scope and the test client are available during RunE.
func executeStockAnalyze(t *testing.T, c *client.Client, args ...string) (string, error) {
	t.Helper()
	cmd := newStockCmd()
	setClientContext(cmd, ContextWithClient(context.Background(), c), c)
	return executeCommand(t, cmd, args...)
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

func stockAnalyzeConcurrentServer(t *testing.T, failingSymbols map[string]bool) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := stockResponseFixture()
		bodyText := string(body)
		for symbol, shouldFail := range failingSymbols {
			if shouldFail && strings.Contains(bodyText, symbol) {
				response = emptyMarketDataFixture()
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(response))
		if err != nil {
			t.Errorf("write response body: %v", err)
		}
	}))
}
