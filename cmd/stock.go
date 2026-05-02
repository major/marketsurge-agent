// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

func init() { rootCmd.AddCommand(newStockCmd()) }

func newStockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stock",
		Short: "Get stock data",
	}
	cmd.AddCommand(newSymbolCmd("get", "Get stock data for a symbol", func(ctx context.Context, symbol string) (any, error) {
		return ClientFromContext(ctx).GetStock(ctx, symbol)
	}))
	return cmd
}
