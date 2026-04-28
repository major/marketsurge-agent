package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/internal/output"
)

// AnalysisResult holds the combined stock, fundamental, and ownership data
// for a single symbol.
type AnalysisResult struct {
	Symbol       string                  `json:"symbol"`
	Stock        *models.StockData       `json:"stock,omitempty"`
	Fundamentals *models.FundamentalData `json:"fundamentals,omitempty"`
	Ownership    *models.OwnershipData   `json:"ownership,omitempty"`
}

// StockAnalyzeCommand returns the CLI command for analyzing one or more symbols.
// For each symbol, it concurrently fetches stock, fundamental, and ownership data.
func StockAnalyzeCommand(c *client.Client, w io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "analyze",
		Usage:     "Analyze one or more stock symbols",
		ArgsUsage: "[symbol...]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "tickers", Usage: "Comma-separated ticker symbols to analyze"},
			&cli.BoolFlag{Name: "compact", Usage: "Remove formatted string fields from analysis data"},
			&cli.BoolFlag{Name: "flat", Usage: "Flatten each analysis result for token-efficient agent parsing"},
			&cli.BoolFlag{Name: "summary", Usage: "Return compact screening fields for ranking multi-symbol candidates"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			symbols := analyzeSymbols(cmd)
			if len(symbols) == 0 {
				verr := mserrors.NewValidationError("at least one symbol required", nil)
				return verr
			}

			results := make([]AnalysisResult, len(symbols))
			var allErrors []string
			var mu sync.Mutex

			// Process each symbol concurrently.
			var wg sync.WaitGroup
			for i, symbol := range symbols {
				wg.Go(func() {
					result, errs := analyzeSymbol(ctx, c, symbol)
					results[i] = result
					if len(errs) > 0 {
						mu.Lock()
						allErrors = append(allErrors, errs...)
						mu.Unlock()
					}
				})
			}
			wg.Wait()

			meta := analyzeMetadata(symbols)
			if cmd.Bool("summary") {
				meta["mode"] = "summary"
			}

			// Determine if we have any data at all.
			hasData := false
			for _, r := range results {
				if r.Stock != nil || r.Fundamentals != nil || r.Ownership != nil {
					hasData = true
					break
				}
			}

			if !hasData {
				// Total failure: no data for any symbol.
				err := fmt.Errorf("analysis failed for all symbols: %v", allErrors)
				return err
			}

			data, err := transformAnalysisOutput(results, cmd.Bool("compact"), cmd.Bool("flat"), cmd.Bool("summary"))
			if err != nil {
				return fmt.Errorf("transform analysis output: %w", err)
			}

			// Single symbol: unwrap from array.
			if len(symbols) == 1 {
				if len(allErrors) > 0 {
					return output.WritePartial(w, data, allErrors, meta)
				}
				return output.WriteSuccess(w, data, meta)
			}

			// Multi-symbol: output as array.
			if len(allErrors) > 0 {
				return output.WritePartial(w, data, allErrors, meta)
			}
			return output.WriteSuccess(w, data, meta)
		},
	}
}

// analyzeSymbols combines positional symbols with the --tickers comma-separated
// input used for larger batch analysis requests.
func analyzeSymbols(cmd *cli.Command) []string {
	symbols := append([]string{}, cmd.Args().Slice()...)
	for symbol := range strings.SplitSeq(cmd.String("tickers"), ",") {
		trimmed := strings.TrimSpace(symbol)
		if trimmed != "" {
			symbols = append(symbols, trimmed)
		}
	}
	return symbols
}

// analyzeSymbol fetches stock, fundamental, and ownership data concurrently
// for a single symbol. Returns the combined result and any errors encountered.
func analyzeSymbol(ctx context.Context, c *client.Client, symbol string) (result AnalysisResult, errs []string) {
	result = AnalysisResult{Symbol: symbol}
	var mu sync.Mutex

	var wg sync.WaitGroup

	wg.Go(func() {
		stock, err := c.GetStock(ctx, symbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Sprintf("%s: stock: %s", symbol, err))
			mu.Unlock()
			return
		}
		result.Stock = stock
	})

	wg.Go(func() {
		fund, err := c.GetFundamentals(ctx, symbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Sprintf("%s: fundamentals: %s", symbol, err))
			mu.Unlock()
			return
		}
		result.Fundamentals = fund
	})

	wg.Go(func() {
		own, err := c.GetOwnership(ctx, symbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Sprintf("%s: ownership: %s", symbol, err))
			mu.Unlock()
			return
		}
		result.Ownership = own
	})

	wg.Wait()
	return result, errs
}

// analyzeMetadata builds metadata for an analyze response.
func analyzeMetadata(symbols []string) map[string]any {
	meta := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	if len(symbols) == 1 {
		meta["symbol"] = symbols[0]
	} else {
		meta["symbols"] = symbols
	}
	return meta
}

// transformAnalysisOutput applies optional token-efficiency transforms while
// preserving the existing single-symbol object vs. multi-symbol array contract.
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
