package commands

import (
	"context"
	"io"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/output"
)

// validLookbacks lists the accepted lookback period tokens.
var validLookbacks = map[string]bool{
	"1W": true, "1M": true, "3M": true, "6M": true, "1Y": true, "YTD": true,
}

// defaultExchangeName is used for daily chart queries.
const defaultExchangeName = "NYSE"

// ChartHistoryCommand returns the CLI command for retrieving chart history.
func ChartHistoryCommand(c *client.Client, w io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "history",
		Usage:     "Get chart history for a symbol",
		ArgsUsage: "<symbol>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "start-date", Usage: "Start date (YYYY-MM-DD)"},
			&cli.StringFlag{Name: "end-date", Usage: "End date (YYYY-MM-DD)"},
			&cli.StringFlag{Name: "lookback", Usage: "Relative lookback period: 1W, 1M, 3M, 6M, 1Y, YTD"},
			&cli.StringFlag{Name: "period", Value: "daily", Usage: "Chart period: daily or weekly"},
			&cli.StringFlag{Name: "benchmark", Usage: "Benchmark symbol for relative strength (e.g. 0S&P5)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			symbol, err := requireSymbol(cmd)
			if err != nil {
				return err
			}

			startDate, endDate, err := resolveChartDates(cmd, time.Now().UTC())
			if err != nil {
				return err
			}

			period := cmd.String("period")
			graphqlPeriod, daily := mapPeriod(period)

			exchangeName := ""
			if daily {
				exchangeName = defaultExchangeName
			}

			benchmark := cmd.String("benchmark")

			data, err := c.GetChartHistory(ctx, symbol, startDate, endDate, graphqlPeriod, daily, exchangeName, benchmark)
			if err != nil {
				return err
			}
			return output.WriteSuccess(w, data, output.SymbolMeta(symbol))
		},
	}
}

// resolveChartDates validates and resolves the date flags into start/end date strings.
// Either (--start-date AND --end-date) or --lookback must be provided, not both.
func resolveChartDates(cmd *cli.Command, now time.Time) (string, string, error) {
	startDate := cmd.String("start-date")
	endDate := cmd.String("end-date")
	lookback := cmd.String("lookback")

	hasExplicit := startDate != "" || endDate != ""
	hasLookback := lookback != ""

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
