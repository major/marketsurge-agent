// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/leodido/structcli"
	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/internal/output"
	"github.com/spf13/cobra"
)

const (
	antSignalExplanation       = "ANTS marks flag institutional accumulation: repeated upside price action with rising volume over a recent 15-day window."
	compactRecentSignalLimit   = 3
	compactRecentQuarterLimit  = 4
	compactRecentEstimateLimit = 2
)

var compactDroppedFields = map[string]bool{
	"address":                     true,
	"address2":                    true,
	"city":                        true,
	"code":                        true,
	"corporate_actions":           true,
	"country":                     true,
	"description":                 true,
	"estimate_type":               true,
	"group_rank_history":          true,
	"group_rs_history":            true,
	"historical_price_statistics": true,
	"id":                          true,
	"instrument_sub_type":         true,
	"pattern_id":                  true,
	"patterns":                    true,
	"period_offset":               true,
	"phone":                       true,
	"state_province":              true,
	"tight_areas":                 true,
	"volume_moving_averages":      true,
	"website":                     true,
}

func init() { rootCmd.AddCommand(newStockCmd()) }

func newStockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stock",
		Short: "Get stock data",
		Long: `Stock commands retrieve current MarketSurge ratings, price context,
base patterns, and signal flags.

  Need                             Command
  ----                             -------
  One-symbol quote/ratings/base    stock get AAPL
  Complete research packet         stock analyze AAPL
  Compare many candidates          stock analyze --summary AAPL NVDA TSLA
  Batch from a generated list      stock analyze --symbols AAPL,NVDA --compact
  Parser wants one-level keys      stock analyze AAPL --flat

Use "stock get" for targeted current stock data when fundamentals and
ownership are not needed. Output focus: ratings, price, valuation ratios,
risk metrics, short interest, base_pattern, and signals (blue dot, ant).

Use "stock analyze" for the complete research packet including stock,
fundamentals, and ownership data fetched concurrently.`,
	}
	cmd.AddCommand(newSymbolCmd("get", "Get stock data for a symbol", func(ctx context.Context, symbol string) (any, error) {
		return ClientFromContext(ctx).GetStock(ctx, symbol)
	}))
	cmd.AddCommand(newStockAnalyzeCmd())
	return cmd
}

// AnalysisResult holds the combined stock, fundamental, and ownership data for a single symbol.
type AnalysisResult struct {
	Symbol      string                  `json:"symbol"`
	Stock       *models.StockData       `json:"stock,omitempty"`
	Fundamental *models.FundamentalData `json:"fundamentals,omitempty"`
	Ownership   *models.OwnershipData   `json:"ownership,omitempty"`
	Errors      []string                `json:"errors,omitempty"`
}

// StockAnalyzeOptions holds the options for the stock analyze command.
type StockAnalyzeOptions struct {
	Symbols []string `flag:"symbols" flagshort:"s" flaggroup:"Input" flagdescr:"Stock symbols to analyze, for example AAPL,MSFT; accepts comma-separated or repeated values; positional symbols remain supported for shell use"`
	Tickers []string `flag:"tickers" flagshort:"t" flaggroup:"Input" flagdescr:"Additional stock symbols to analyze; accepts comma-separated or repeated values"`
	Compact bool     `flag:"compact" flaggroup:"Output Format" flagdescr:"Remove low-value fields such as formatted duplicates, profile metadata, internal IDs, empty fields, and stale arrays while keeping decision-relevant raw values; compatible with --flat. Example: stock analyze AAPL --compact"`
	Flat    bool     `flag:"flat" flaggroup:"Output Format" flagdescr:"Flatten nested analysis fields into single-level JSON keys, for example stock.pricing.market_cap becomes pricing_market_cap; compatible with --compact. Example: stock analyze AAPL --flat"`
	Summary bool     `flag:"summary" flaggroup:"Output Format" flagdescr:"Return compact screening objects for ranking many symbols. Example: stock analyze --summary AAPL MSFT NVDA"`
	Setup   bool     `flag:"setup" flaggroup:"Output Format" flagdescr:"Return --setup trade triage fields: summary fields plus base_length_weeks, volume_at_pivot_percent, ownership_funds_float_percent, and quarterly_funds. Example: stock analyze AAPL --setup"`
}

// MergeSymbols merges positional arguments with --symbols and --tickers flag values, deduplicating and trimming whitespace.
func (o *StockAnalyzeOptions) MergeSymbols(args []string) []string {
	return mergeSymbolInputs(args, o.Symbols, o.Tickers)
}

func newStockAnalyzeCmd() *cobra.Command {
	opts := &StockAnalyzeOptions{}
	cmd := &cobra.Command{
		Use:   "analyze [symbol...]",
		Short: "Analyze one or more stock symbols",
		Example: `  marketsurge-agent stock analyze AAPL
  marketsurge-agent stock analyze --tickers AAPL,MSFT,NVDA --compact
  marketsurge-agent stock analyze --summary AAPL MSFT NVDA
  marketsurge-agent stock analyze AAPL --setup
  marketsurge-agent stock analyze AAPL --flat --compact`,
		Long: `Fetches stock, fundamentals, and ownership concurrently for one or more
symbols. Accepts positional symbols, --symbols values, and backward-compatible
--tickers values together. Multi-symbol requests are concurrent and can return partial
results when only some symbols or subresources fail.

Flags:

  --summary   Compact screening keys: symbol, composite, eps, rs, ad, smr,
              blue_dot, ant_signal, ant_dates, ant_explanation,
              base_type, base_stage, pivot,
              pivot_price_date, pricing_start_date, pricing_end_date,
              base_depth_percent,
              industry_group_rs, up_down_volume, atr_percent,
              avg_dollar_volume, funds_float_percent
  --setup     Setup-focused trade triage keys: all --summary keys plus
              base_length_weeks, volume_at_pivot_percent,
              ownership_funds_float_percent, quarterly_funds
  --compact   Removes formatted duplicates, profile metadata, internal IDs,
              empty fields, and stale arrays while keeping raw decision fields
  --flat      Flattens nested analysis fields inside the JSON envelope

Start with "stock analyze --summary" for candidate ranking, then rerun
interesting symbols without --summary for detail.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			symbols := opts.MergeSymbols(args)
			if len(symbols) == 0 {
				return mserrors.NewValidationError("missing symbols: pass positional symbols like stock analyze AAPL MSFT or use --tickers AAPL,MSFT", nil)
			}

			ctx := cmd.Context()
			c := ClientFromContext(ctx)
			results := make([]AnalysisResult, len(symbols))
			allErrors := make([]string, 0)
			var mu sync.Mutex

			var wg sync.WaitGroup
			for i, symbol := range symbols {
				wg.Go(func() {
					result := analyzeSymbol(ctx, c, symbol)
					results[i] = result
					if len(result.Errors) > 0 {
						mu.Lock()
						allErrors = append(allErrors, result.Errors...)
						mu.Unlock()
					}
				})
			}
			wg.Wait()

			if !analysisHasData(results) {
				return mserrors.NewAPIError(fmt.Sprintf("analysis failed for all symbols: %v", allErrors), nil)
			}

			data, err := transformAnalysisOutput(results, opts.Compact, opts.Flat, opts.Summary, opts.Setup)
			if err != nil {
				return fmt.Errorf("transform analysis output: %w", err)
			}

			metadata := analyzeMetadata(symbols)
			if opts.Setup {
				metadata["mode"] = "setup"
			} else if opts.Summary {
				metadata["mode"] = "summary"
			}
			if len(allErrors) > 0 {
				return output.WritePartial(cmd.OutOrStdout(), data, allErrors, metadata)
			}
			return output.WriteSuccess(cmd.OutOrStdout(), data, metadata)
		},
	}
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(err)
	}
	return cmd
}

func analyzeSymbol(ctx context.Context, c *client.Client, symbol string) AnalysisResult {
	result := AnalysisResult{Symbol: symbol}
	var mu sync.Mutex

	var wg sync.WaitGroup
	wg.Go(func() {
		stock, err := c.GetStock(ctx, symbol)
		mu.Lock()
		defer mu.Unlock()
		result.Stock = stock
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: stock: %s", symbol, err))
		}
	})
	wg.Go(func() {
		fundamental, err := c.GetFundamentals(ctx, symbol)
		mu.Lock()
		defer mu.Unlock()
		result.Fundamental = fundamental
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: fundamentals: %s", symbol, err))
		}
	})
	wg.Go(func() {
		ownership, err := c.GetOwnership(ctx, symbol)
		mu.Lock()
		defer mu.Unlock()
		result.Ownership = ownership
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: ownership: %s", symbol, err))
		}
	})
	wg.Wait()
	return result
}

func analysisHasData(results []AnalysisResult) bool {
	for _, result := range results {
		if result.Stock != nil || result.Fundamental != nil || result.Ownership != nil {
			return true
		}
	}
	return false
}

func analyzeMetadata(symbols []string) map[string]any {
	if len(symbols) == 1 {
		return output.SymbolMeta(symbols[0])
	}
	return map[string]any{
		"symbols":   symbols,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
}

func transformAnalysisOutput(results []AnalysisResult, compact, flat, summary, setup bool) (any, error) {
	transformed := make([]any, 0, len(results))
	for _, result := range results {
		if setup {
			transformed = append(transformed, analysisSetupMap(result))
			continue
		}
		if summary {
			transformed = append(transformed, analysisSummaryMap(result))
			continue
		}

		data, err := analysisResultMap(result)
		if err != nil {
			return nil, err
		}
		if compact {
			data = compactAnalysisMap(data)
		}
		if flat {
			data = flattenAnalysisMap(data)
		}
		transformed = append(transformed, data)
	}

	if len(transformed) == 1 {
		return transformed[0], nil
	}
	return transformed, nil
}

func analysisSummaryMap(result AnalysisResult) map[string]any {
	data := map[string]any{"symbol": result.Symbol}
	if result.Stock == nil {
		return data
	}

	addRatingSummary(data, result.Stock.Ratings)
	addSignalSummary(data, result.Stock.Signals)
	addAntSummary(data, result.Stock)
	addBaseSummary(data, result.Stock.BasePattern)
	addScreeningMetrics(data, result.Stock)
	addPricingFreshness(data, result.Stock.Pricing)
	return data
}

func analysisSetupMap(result AnalysisResult) map[string]any {
	data := analysisSummaryMap(result)
	if result.Stock != nil {
		addSetupBaseContext(data, result.Stock.BasePattern)
	}

	addSetupOwnershipContext(data, result.Ownership)
	return data
}

func addRatingSummary(data map[string]any, ratings *models.Ratings) {
	if ratings == nil {
		return
	}
	addPtrValue(data, "composite", ratings.Composite)
	addPtrValue(data, "eps", ratings.EPS)
	addPtrValue(data, "rs", ratings.RS)
	addPtrValue(data, "ad", ratings.AD)
	addPtrValue(data, "smr", ratings.SMR)
}

func addSignalSummary(data map[string]any, signals *models.Signals) {
	if signals == nil {
		return
	}
	addPtrValue(data, "blue_dot", signals.BlueDot)
	addPtrValue(data, "ant_signal", signals.AntSignal)
}

func addAntSummary(data map[string]any, stock *models.StockData) {
	if stock.Signals == nil || stock.Signals.AntSignal == nil || !*stock.Signals.AntSignal {
		return
	}

	if stock.Pricing != nil && len(stock.Pricing.AntDates) > 0 {
		data["ant_dates"] = stock.Pricing.AntDates
	}
	data["ant_explanation"] = antSignalExplanation
}

func addBaseSummary(data map[string]any, base *models.BasePattern) {
	if base == nil {
		return
	}
	addPtrValue(data, "base_type", base.PatternType)
	addPtrValue(data, "base_stage", base.BaseStage)
	addPtrValue(data, "pivot", base.PivotPrice)
	addPtrValue(data, "pivot_price_date", base.PivotPriceDate)
	addPtrValue(data, "base_depth_percent", base.BaseDepthPercent)
}

func addScreeningMetrics(data map[string]any, stock *models.StockData) {
	if stock.Company != nil {
		addPtrValue(data, "industry_group_rs", stock.Company.IndustryGroupRS)
	}
	if stock.Pricing != nil {
		addPtrValue(data, "up_down_volume", stock.Pricing.UpDownVolumeRatio)
		addPtrValue(data, "atr_percent", stock.Pricing.ATRPercent21D)
		addPtrValue(data, "avg_dollar_volume", stock.Pricing.AvgDollarVolume50D)
	}
	if stock.Ownership != nil {
		addPtrValue(data, "funds_float_percent", stock.Ownership.FundsFloatPct)
	}
}

func addPricingFreshness(data map[string]any, pricing *models.Pricing) {
	if pricing == nil {
		return
	}
	addPtrValue(data, "pricing_start_date", pricing.PricingStartDate)
	addPtrValue(data, "pricing_end_date", pricing.PricingEndDate)
}

func addSetupBaseContext(data map[string]any, base *models.BasePattern) {
	if base == nil {
		return
	}
	addPtrValue(data, "base_length_weeks", base.BaseLengthWeeks)
	addPtrValue(data, "volume_at_pivot_percent", base.VolumeAtPivotPct)
}

func addSetupOwnershipContext(data map[string]any, ownership *models.OwnershipData) {
	if ownership == nil {
		return
	}
	addPtrValue(data, "ownership_funds_float_percent", ownership.FundsFloatPct)
	if len(ownership.QuarterlyFunds) > 0 {
		data["quarterly_funds"] = ownership.QuarterlyFunds
	}
}

func addPtrValue[T any](data map[string]any, key string, value *T) {
	if value != nil {
		data[key] = *value
	}
}

func analysisResultMap(result AnalysisResult) (map[string]any, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal analysis result: %w", err)
	}

	var resultMap map[string]any
	if err := json.Unmarshal(data, &resultMap); err != nil {
		return nil, fmt.Errorf("unmarshal analysis result: %w", err)
	}
	return resultMap, nil
}

func compactAnalysisMap(data map[string]any) map[string]any {
	cleaned, keep := compactAnalysisValue(data, 0, "")
	if !keep {
		return map[string]any{}
	}
	return cleaned.(map[string]any)
}

func compactAnalysisValue(value any, depth int, key string) (any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		cleaned := make(map[string]any, len(typed))
		for key, nested := range typed {
			if isCompactDroppedField(key, depth) {
				continue
			}
			cleanedValue, keep := compactAnalysisValue(nested, depth+1, key)
			if keep {
				cleaned[key] = cleanedValue
			}
		}
		if len(cleaned) == 0 {
			return nil, false
		}
		return cleaned, true
	case []any:
		limit := compactArrayLimit(key, len(typed))
		cleaned := make([]any, 0, limit)
		for _, nested := range typed[:limit] {
			cleanedValue, keep := compactAnalysisValue(nested, depth+1, key)
			if keep {
				cleaned = append(cleaned, cleanedValue)
			}
		}
		if len(cleaned) == 0 {
			return nil, false
		}
		return cleaned, true
	case string:
		if typed == "" || typed == "N/A" {
			return nil, false
		}
		return typed, true
	case nil:
		return nil, false
	default:
		return value, true
	}
}

func isCompactDroppedField(key string, parentDepth int) bool {
	if isFormattedField(key) {
		return true
	}
	if key == "symbol" && parentDepth > 0 {
		return true
	}
	return compactDroppedFields[key]
}

func isFormattedField(key string) bool {
	return strings.HasSuffix(key, "_formatted") || strings.HasPrefix(key, "formatted_")
}

func compactArrayLimit(key string, length int) int {
	limit := length
	switch key {
	case "ant_dates", "blue_dot_daily_dates", "blue_dot_weekly_dates":
		limit = min(length, compactRecentSignalLimit)
	case "eps_estimates", "sales_estimates":
		limit = min(length, compactRecentEstimateLimit)
	case "profit_margins", "quarterly_funds", "reported_earnings", "reported_sales":
		limit = min(length, compactRecentQuarterLimit)
	}
	return limit
}

func flattenAnalysisMap(data map[string]any) map[string]any {
	flat := map[string]any{}
	if symbol, ok := data["symbol"]; ok {
		flat["symbol"] = symbol
	}

	for key, value := range data {
		switch key {
		case "symbol":
			continue
		case "stock":
			flattenValue(flat, "", value)
		default:
			flattenValue(flat, key, value)
		}
	}
	return flat
}

func flattenValue(flat map[string]any, prefix string, value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			flattenValue(flat, joinFlatKey(prefix, key), nested)
		}
	case []any:
		if len(typed) > 0 {
			flat[prefix] = typed
		}
	case nil:
		return
	default:
		if prefix != "" {
			flat[prefix] = typed
		}
	}
}

func joinFlatKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "_" + key
}
