// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

func init() { rootCmd.AddCommand(newOwnershipCmd()) }

func newOwnershipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ownership",
		Short: "Get ownership data",
	}
	cmd.AddCommand(newSymbolCmd("get", "Get ownership data for a symbol", func(ctx context.Context, symbol string) (any, error) {
		return ClientFromContext(ctx).GetOwnership(ctx, symbol)
	}))
	return cmd
}
