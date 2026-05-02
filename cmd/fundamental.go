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
		Long: `Use "fundamental get" when the question is about earnings, sales,
margins, estimates, or cash flow. Use "stock analyze" instead when
price, ratings, or ownership are also needed.

Output focus:

  - Historical EPS and sales with year-over-year changes
  - Future EPS and sales estimates
  - Quarterly earnings, sales, and margin breakdowns
  - Cash flow per share

Do not call this after "stock analyze" unless the prior analysis used
--summary mode and omitted fundamentals. Pair with "rs-history get"
when separating business improvement from price-relative strength.`,
	}
	cmd.AddCommand(newSymbolCmd("get", "Get fundamental data for a symbol", func(ctx context.Context, symbol string) (any, error) {
		return ClientFromContext(ctx).GetFundamentals(ctx, symbol)
	}))
	return cmd
}
