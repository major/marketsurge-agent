// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/major/marketsurge-agent/internal/output"
)

func init() { rootCmd.AddCommand(newChartCmd()) }

func newChartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chart",
		Short: "Chart data commands",
	}
	cmd.AddCommand(newChartMarkupsCmd())
	// NOTE: Task 9 will add newChartHistoryCmd() here
	return cmd
}

func newChartMarkupsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "markups <symbol>",
		Short: "Get chart markup data for a symbol",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			symbol := args[0]
			frequency, _ := cmd.Flags().GetString("frequency")
			sortDir, _ := cmd.Flags().GetString("sort-dir")
			c := ClientFromContext(cmd.Context())
			data, err := c.GetChartMarkups(cmd.Context(), symbol, frequency, sortDir)
			if err != nil {
				return err
			}
			return output.WriteSuccess(cmd.OutOrStdout(), data, output.SymbolMeta(symbol))
		},
	}
	cmd.Flags().String("frequency", "DAILY", "Chart frequency: DAILY or WEEKLY")
	cmd.Flags().String("sort-dir", "ASC", "Sort direction: ASC or DESC")
	return cmd
}
