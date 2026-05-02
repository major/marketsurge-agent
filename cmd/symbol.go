// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"context"
	"strings"

	"github.com/leodido/structcli"
	"github.com/spf13/cobra"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/output"
)

// symbolFetcher retrieves data for a single symbol from the MarketSurge API.
type symbolFetcher func(ctx context.Context, symbol string) (any, error)

// SymbolOptions holds schema-visible symbol input for single-symbol commands.
type SymbolOptions struct {
	Symbol string `flag:"symbol" flagshort:"s" flaggroup:"Input" flagdescr:"Stock symbol to fetch, for example AAPL; positional <symbol> remains supported for shell use"`
}

// ResolveSymbol returns the symbol from --symbol or one positional argument.
func (o *SymbolOptions) ResolveSymbol(args []string) (string, error) {
	return resolveSingleSymbol(args, o.Symbol)
}

// resolveSingleSymbol returns the symbol from a flag value or one positional argument.
func resolveSingleSymbol(args []string, symbolFlag string) (string, error) {
	if len(args) > 1 {
		return "", mserrors.NewValidationError("expected at most one positional symbol; pass --symbol AAPL or positional AAPL", nil)
	}

	flagSymbol := strings.TrimSpace(symbolFlag)
	positionalSymbol := ""
	if len(args) == 1 {
		positionalSymbol = strings.TrimSpace(args[0])
	}

	if flagSymbol != "" && positionalSymbol != "" && flagSymbol != positionalSymbol {
		return "", mserrors.NewValidationError("cannot use different values for --symbol and positional <symbol>", nil)
	}
	if flagSymbol != "" {
		return flagSymbol, nil
	}
	if positionalSymbol != "" {
		return positionalSymbol, nil
	}
	return "", mserrors.NewValidationError("symbol is required; pass --symbol AAPL or positional AAPL", nil)
}

// MultiSymbolOptions holds schema-visible symbol input for multi-symbol commands.
type MultiSymbolOptions struct {
	Symbols []string `flag:"symbols" flagshort:"s" flaggroup:"Input" flagdescr:"Stock symbols to fetch, for example AAPL,MSFT; accepts comma-separated or repeated values; positional symbols remain supported for shell use"`
}

// ResolveSymbols merges positional arguments with --symbols values, deduplicating and trimming whitespace.
func (o *MultiSymbolOptions) ResolveSymbols(args []string, extraGroups ...[]string) []string {
	groups := make([][]string, 0, 2+len(extraGroups))
	groups = append(groups, args, o.Symbols)
	groups = append(groups, extraGroups...)
	return mergeSymbolInputs(groups...)
}

// mergeSymbolInputs merges symbol groups while preserving first-seen order.
func mergeSymbolInputs(groups ...[]string) []string {
	capacity := 0
	for _, group := range groups {
		capacity += len(group)
	}

	symbols := make([]string, 0, capacity)
	seen := make(map[string]struct{}, capacity)
	for _, group := range groups {
		for _, value := range group {
			for symbol := range strings.SplitSeq(value, ",") {
				trimmed := strings.TrimSpace(symbol)
				if trimmed == "" {
					continue
				}
				if _, ok := seen[trimmed]; ok {
					continue
				}
				seen[trimmed] = struct{}{}
				symbols = append(symbols, trimmed)
			}
		}
	}
	return symbols
}

// newSymbolCmd builds a cobra command that fetches data for a single symbol.
// It handles argument validation, calls the fetcher, and writes the JSON envelope.
func newSymbolCmd(use, short string, fetch symbolFetcher) *cobra.Command {
	opts := &SymbolOptions{}
	cmd := &cobra.Command{
		Use:   use + " <symbol>",
		Short: short,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			symbol, err := opts.ResolveSymbol(args)
			if err != nil {
				return err
			}
			data, err := fetch(cmd.Context(), symbol)
			if err != nil {
				return err
			}
			return output.WriteSuccess(cmd.OutOrStdout(), data, output.SymbolMeta(symbol))
		},
	}
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(err)
	}
	return cmd
}
