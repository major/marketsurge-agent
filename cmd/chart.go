// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"context"
	"time"

	"github.com/leodido/structcli"
	"github.com/spf13/cobra"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/internal/output"
)

func init() { rootCmd.AddCommand(newChartCmd()) }

func newChartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chart",
		Short: "Chart data commands",
		Long: `Chart commands retrieve OHLCV price candles, benchmark comparison
series, and user-saved chart annotations.

Use "chart history" for price history with daily or weekly candles
and optional benchmark series for relative strength calculations.

Use "chart markups" for user-saved annotations and drawings. Markup
data is opaque serialized chart data; do not parse it unless a
downstream renderer understands the format.`,
	}
	cmd.AddCommand(newChartMarkupsCmd())
	cmd.AddCommand(newChartHistoryCmd())
	return cmd
}

// ChartMarkupsOptions holds flags for the chart markups command.
type ChartMarkupsOptions struct {
	Frequency models.Frequency     `flag:"frequency" flaggroup:"Options" flagdescr:"Chart candle frequency for markup lookup (DAILY or WEEKLY)" default:"DAILY"`
	SortDir   models.SortDirection `flag:"sort-dir" flaggroup:"Options" flagdescr:"Sort direction for markup annotations (ASC or DESC)" default:"ASC"`
}

func newChartMarkupsCmd() *cobra.Command {
	opts := &ChartMarkupsOptions{}
	cmd := &cobra.Command{
		Use:   "markups <symbol>",
		Short: "Get chart markup data for a symbol",
		Long: `Fetches user-saved annotations and drawings for a symbol.

Flags:

  --frequency DAILY|WEEKLY   Default: DAILY
  --sort-dir ASC|DESC        Default: ASC

Markup data is opaque serialized chart data. Do not parse it unless
a downstream MarketSurge-specific renderer understands the format.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			symbol := args[0]
			c := ClientFromContext(cmd.Context())
			data, err := c.GetChartMarkups(cmd.Context(), symbol, string(opts.Frequency), string(opts.SortDir))
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

// validLookbacks lists the accepted lookback period tokens.
var validLookbacks = map[string]bool{
	"1W": true, "1M": true, "3M": true, "6M": true, "1Y": true, "YTD": true,
}

// defaultExchangeName is used for daily chart queries.
const defaultExchangeName = "NYSE"

// ChartHistoryOptions holds flags for the chart history command.
// Lookback stays string (not models.Lookback) because it is optional with no default;
// structcli's enum decoder rejects empty string for registered enums.
type ChartHistoryOptions struct {
	StartDate string        `flag:"start-date" flaggroup:"Date Range" flagdescr:"Start date in YYYY-MM-DD format; requires --end-date and is mutually exclusive with --lookback"`
	EndDate   string        `flag:"end-date" flaggroup:"Date Range" flagdescr:"End date in YYYY-MM-DD format; requires --start-date and is mutually exclusive with --lookback"`
	Lookback  string        `flag:"lookback" flaggroup:"Date Range" flagdescr:"Relative lookback period (1W, 1M, 3M, 6M, 1Y, YTD); mutually exclusive with --start-date/--end-date"`
	Period    models.Period `flag:"period" flaggroup:"Options" flagdescr:"Data period granularity (daily or weekly)" default:"daily"`
	Benchmark string        `flag:"benchmark" flaggroup:"Options" flagdescr:"Benchmark symbol for relative strength comparison"`
}

// Validate checks mutual exclusion and required-field constraints for chart history flags.
func (o *ChartHistoryOptions) Validate(_ context.Context) []error {
	hasExplicit := o.StartDate != "" || o.EndDate != ""
	hasLookback := o.Lookback != ""

	if hasExplicit && hasLookback {
		return []error{mserrors.NewValidationError(
			"cannot use both --start-date/--end-date and --lookback", nil,
		)}
	}

	if !hasExplicit && !hasLookback {
		return []error{mserrors.NewValidationError(
			"either --start-date and --end-date or --lookback is required", nil,
		)}
	}

	if hasExplicit {
		if o.StartDate == "" || o.EndDate == "" {
			return []error{mserrors.NewValidationError(
				"both --start-date and --end-date are required when using explicit dates", nil,
			)}
		}
		return nil
	}

	if !validLookbacks[o.Lookback] {
		return []error{mserrors.NewValidationError(
			"invalid lookback value: must be one of 1W, 1M, 3M, 6M, 1Y, YTD", nil,
		)}
	}

	return nil
}

// ResolveDates computes start and end date strings from validated options.
// Validate must be called before ResolveDates.
func (o *ChartHistoryOptions) ResolveDates(now time.Time) (startDate, endDate string) {
	if o.StartDate != "" {
		return o.StartDate, o.EndDate
	}
	return resolveLookback(o.Lookback, now), now.Format("2006-01-02")
}

func newChartHistoryCmd() *cobra.Command {
	opts := &ChartHistoryOptions{}
	cmd := &cobra.Command{
		Use:   "history <symbol>",
		Short: "Get chart history for a symbol",
		Long: `Fetches price history for a symbol. Exactly one date mode is required:

  Date mode           Example
  ---------           -------
  Relative lookback   chart history AAPL --lookback 3M
                      Valid: 1W, 1M, 3M, 6M, 1Y, YTD
  Explicit range      chart history AAPL --start-date 2024-01-01 --end-date 2024-04-21
                      Both dates required

Other flags:

  --period daily|weekly    Defaults to daily; weekly maps to P1W
  --benchmark 0S&P5       Includes benchmark_time_series for relative
                           strength calculations

Output: time_series.data_points with OHLCV fields, quote, exchange,
market state, and optional benchmark candles.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			symbol := args[0]

			if errs := opts.Validate(cmd.Context()); errs != nil {
				return errs[0]
			}

			startDate, endDate := opts.ResolveDates(time.Now().UTC())
			graphqlPeriod, daily := mapPeriod(string(opts.Period))

			exchangeName := ""
			if daily {
				exchangeName = defaultExchangeName
			}

			c := ClientFromContext(cmd.Context())
			data, err := c.GetChartHistory(cmd.Context(), symbol, startDate, endDate, graphqlPeriod, daily, exchangeName, opts.Benchmark)
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
