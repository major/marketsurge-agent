package commands

import (
	"context"
	"fmt"
	"io"
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
		ArgsUsage: "<symbol> [symbol...]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				verr := mserrors.NewValidationError("at least one symbol argument required", nil)
				return verr
			}

			symbols := cmd.Args().Slice()
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

			// Single symbol: unwrap from array.
			if len(symbols) == 1 {
				if len(allErrors) > 0 {
					return output.WritePartial(w, results[0], allErrors, meta)
				}
				return output.WriteSuccess(w, results[0], meta)
			}

			// Multi-symbol: output as array.
			if len(allErrors) > 0 {
				return output.WritePartial(w, results, allErrors, meta)
			}
			return output.WriteSuccess(w, results, meta)
		},
	}
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
