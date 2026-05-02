// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

func init() { rootCmd.AddCommand(newFundamentalCmd()) }

func newFundamentalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fundamental",
		Short: "Get fundamental analysis data",
	}
	cmd.AddCommand(newSymbolCmd("get", "Get fundamental data for a symbol", func(ctx context.Context, symbol string) (any, error) {
		return ClientFromContext(ctx).GetFundamentals(ctx, symbol)
	}))
	return cmd
}
