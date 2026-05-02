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
		Long: `Use "ownership get" for institutional sponsorship questions. Use
"stock analyze" instead when ownership is only one part of a broader
stock review.

Output focus:

  - Quarterly fund ownership counts
  - Funds as percentage of float
  - Ownership trend data

Rising fund count suggests growing institutional interest. High
funds-float percentage shows sponsorship but may imply crowded
ownership. For screening many symbols, prefer "stock analyze --summary"
because it includes funds_float_percent with other ranking fields.`,
	}
	cmd.AddCommand(newSymbolCmd("get", "Get ownership data for a symbol", func(ctx context.Context, symbol string) (any, error) {
		return ClientFromContext(ctx).GetOwnership(ctx, symbol)
	}))
	return cmd
}
