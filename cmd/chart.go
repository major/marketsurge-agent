// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"time"

	"github.com/spf13/cobra"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/output"
)

func init() { rootCmd.AddCommand(newChartCmd()) }

func newChartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chart",
		Short: "Chart data commands",
	}
	cmd.AddCommand(newChartMarkupsCmd())
	cmd.AddCommand(newChartHistoryCmd())
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

// validLookbacks lists the accepted lookback period tokens.
var validLookbacks = map[string]bool{
	"1W": true, "1M": true, "3M": true, "6M": true, "1Y": true, "YTD": true,
}

// defaultExchangeName is used for daily chart queries.
const defaultExchangeName = "NYSE"

func newChartHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history <symbol>",
		Short: "Get chart history for a symbol",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			symbol := args[0]

			startDate, endDate, err := resolveChartDates(cmd, time.Now().UTC())
			if err != nil {
				return err
			}

			period, _ := cmd.Flags().GetString("period")
			graphqlPeriod, daily := mapPeriod(period)

			exchangeName := ""
			if daily {
				exchangeName = defaultExchangeName
			}

			benchmark, _ := cmd.Flags().GetString("benchmark")

			c := ClientFromContext(cmd.Context())
			data, err := c.GetChartHistory(cmd.Context(), symbol, startDate, endDate, graphqlPeriod, daily, exchangeName, benchmark)
			if err != nil {
				return err
			}
			return output.WriteSuccess(cmd.OutOrStdout(), data, output.SymbolMeta(symbol))
		},
	}
	cmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().String("end-date", "", "End date (YYYY-MM-DD)")
	cmd.Flags().String("lookback", "", "Relative lookback period: 1W, 1M, 3M, 6M, 1Y, YTD")
	cmd.Flags().String("period", "daily", "Chart period: daily or weekly")
	cmd.Flags().String("benchmark", "", "Benchmark symbol for relative strength (e.g. 0S&P5)")
	return cmd
}

// resolveChartDates validates and resolves the date flags into start/end date strings.
// Either (--start-date AND --end-date) or --lookback must be provided, not both.
func resolveChartDates(cmd *cobra.Command, now time.Time) (startDate, endDate string, err error) {
	startDate, _ = cmd.Flags().GetString("start-date")
	endDate, _ = cmd.Flags().GetString("end-date")
	lookback, _ := cmd.Flags().GetString("lookback")

	hasExplicit := cmd.Flags().Changed("start-date") || cmd.Flags().Changed("end-date")
	hasLookback := cmd.Flags().Changed("lookback")

	if hasExplicit && hasLookback {
		return "", "", mserrors.NewValidationError(
			"cannot use both --start-date/--end-date and --lookback", nil,
		)
	}

	if !hasExplicit && !hasLookback {
		return "", "", mserrors.NewValidationError(
			"either --start-date and --end-date or --lookback is required", nil,
		)
	}

	if hasExplicit {
		if startDate == "" || endDate == "" {
			return "", "", mserrors.NewValidationError(
				"both --start-date and --end-date are required when using explicit dates", nil,
			)
		}
		return startDate, endDate, nil
	}

	// Lookback mode.
	if !validLookbacks[lookback] {
		return "", "", mserrors.NewValidationError(
			"invalid lookback value: must be one of 1W, 1M, 3M, 6M, 1Y, YTD", nil,
		)
	}

	return resolveLookback(lookback, now), now.Format("2006-01-02"), nil
}

// resolveLookback computes the start date for a given lookback token.
func resolveLookback(lookback string, now time.Time) string {
	switch lookback {
	case "1W":
		return now.AddDate(0, 0, -7).Format("2006-01-02")
	case "1M":
		return now.AddDate(0, -1, 0).Format("2006-01-02")
	case "3M":
		return now.AddDate(0, -3, 0).Format("2006-01-02")
	case "6M":
		return now.AddDate(0, -6, 0).Format("2006-01-02")
	case "1Y":
		return now.AddDate(-1, 0, 0).Format("2006-01-02")
	case "YTD":
		return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	default:
		return now.Format("2006-01-02")
	}
}

// mapPeriod converts a user-facing period string to the GraphQL period and daily flag.
func mapPeriod(period string) (string, bool) {
	if period == "weekly" {
		return "P1W", false
	}
	return "P1D", true
}
