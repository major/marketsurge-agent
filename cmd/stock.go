// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/internal/output"
	"github.com/spf13/cobra"
)

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
  Batch from a generated list      stock analyze --tickers AAPL,NVDA --compact
  Parser wants one-level keys      stock analyze AAPL --flat

Use "stock get" for targeted current stock data when fundamentals and
ownership are not needed. Output focus: ratings, price, valuation ratios,
risk metrics, short interest, base_pattern, and signals (blue dot, ant).

Use "stock analyze" for the complete research packet including stock,
fundamentals, and ownership data fetched concurrently.`,
		SilenceUsage:  true,
		SilenceErrors: true,
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
	Tickers []string
	Compact bool
	Flat    bool
	Summary bool
}

// BindFlags registers the stock analyze command flags and binds them to the options struct.
func (o *StockAnalyzeOptions) BindFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&o.Tickers, "tickers", []string{}, "Additional ticker symbols to analyze")
	cmd.Flags().BoolVar(&o.Compact, "compact", false, "Remove formatted string fields from analysis data")
	cmd.Flags().BoolVar(&o.Flat, "flat", false, "Flatten each analysis result for token-efficient agent parsing")
	cmd.Flags().BoolVar(&o.Summary, "summary", false, "Return compact screening fields for ranking multi-symbol candidates")
}

// MergeSymbols merges positional arguments with --tickers flag values, deduplicating and trimming whitespace.
func (o *StockAnalyzeOptions) MergeSymbols(args []string) []string {
	symbols := make([]string, 0, len(args)+len(o.Tickers))
	seen := make(map[string]struct{}, len(args)+len(o.Tickers))
	addSymbol := func(symbol string) {
		trimmed := strings.TrimSpace(symbol)
		if trimmed == "" {
			return
		}
		if _, ok := seen[trimmed]; ok {
			return
		}
		seen[trimmed] = struct{}{}
		symbols = append(symbols, trimmed)
	}

	for _, symbol := range args {
		addSymbol(symbol)
	}
	for _, symbol := range o.Tickers {
		addSymbol(symbol)
	}
	return symbols
}

func newStockAnalyzeCmd() *cobra.Command {
	opts := &StockAnalyzeOptions{}
	cmd := &cobra.Command{
		Use:   "analyze [symbol...]",
		Short: "Analyze one or more stock symbols",
		Long: `Fetches stock, fundamentals, and ownership concurrently for one or more
symbols. Accepts positional symbols and --tickers comma-separated symbols
together. Multi-symbol requests are concurrent and can return partial
results when only some symbols or subresources fail.

Flags:

  --summary   Compact screening keys: symbol, composite, eps, rs, ad, smr,
              blue_dot, ant_signal, base_type, base_stage, pivot,
              base_depth_percent, industry_group_rs, up_down_volume,
              atr_percent, avg_dollar_volume, funds_float_percent
  --compact   Removes duplicate formatted string fields, keeps raw values
  --flat      Flattens nested analysis fields inside the JSON envelope

Start with "stock analyze --summary" for candidate ranking, then rerun
interesting symbols without --summary for detail.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			symbols := opts.MergeSymbols(args)
			if len(symbols) == 0 {
				return mserrors.NewValidationError("at least one symbol required", nil)
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
				return fmt.Errorf("analysis failed for all symbols: %v", allErrors)
			}

			data, err := transformAnalysisOutput(results, opts.Compact, opts.Flat, opts.Summary)
			if err != nil {
				return fmt.Errorf("transform analysis output: %w", err)
			}

			metadata := analyzeMetadata(symbols)
			if opts.Summary {
				metadata["mode"] = "summary"
			}
			if len(allErrors) > 0 {
				return output.WritePartial(cmd.OutOrStdout(), data, allErrors, metadata)
			}
			return output.WriteSuccess(cmd.OutOrStdout(), data, metadata)
		},
	}
	opts.BindFlags(cmd)
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

func transformAnalysisOutput(results []AnalysisResult, compact, flat, summary bool) (any, error) {
	transformed := make([]any, 0, len(results))
	for _, result := range results {
		if summary {
			transformed = append(transformed, analysisSummaryMap(result))
			continue
		}

		data, err := analysisResultMap(result)
		if err != nil {
			return nil, err
		}
		if compact {
			data = removeFormattedFields(data).(map[string]any)
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
	addBaseSummary(data, result.Stock.BasePattern)
	addScreeningMetrics(data, result.Stock)
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

func addBaseSummary(data map[string]any, base *models.BasePattern) {
	if base == nil {
		return
	}
	addPtrValue(data, "base_type", base.PatternType)
	addPtrValue(data, "base_stage", base.BaseStage)
	addPtrValue(data, "pivot", base.PivotPrice)
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

func removeFormattedFields(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		cleaned := make(map[string]any, len(typed))
		for key, nested := range typed {
			if isFormattedField(key) {
				continue
			}
			cleaned[key] = removeFormattedFields(nested)
		}
		return cleaned
	case []any:
		cleaned := make([]any, 0, len(typed))
		for _, nested := range typed {
			cleaned = append(cleaned, removeFormattedFields(nested))
		}
		return cleaned
	default:
		return value
	}
}

func isFormattedField(key string) bool {
	return strings.HasSuffix(key, "_formatted") || strings.HasPrefix(key, "formatted_")
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
