// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/major/marketsurge-agent/internal/output"
)

// symbolFetcher retrieves data for a single symbol from the MarketSurge API.
type symbolFetcher func(ctx context.Context, symbol string) (any, error)

// newSymbolCmd builds a cobra command that fetches data for a single symbol.
// It handles argument validation, calls the fetcher, and writes the JSON envelope.
func newSymbolCmd(use, short string, fetch symbolFetcher) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <symbol>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			symbol := args[0]
			data, err := fetch(cmd.Context(), symbol)
			if err != nil {
				return err
			}
			return output.WriteSuccess(cmd.OutOrStdout(), data, output.SymbolMeta(symbol))
		},
	}
}
