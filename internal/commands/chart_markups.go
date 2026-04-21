package commands

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/output"
)

// ChartMarkupsCommand returns the CLI command for retrieving chart markups.
func ChartMarkupsCommand(c *client.Client, w io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "markups",
		Usage:     "Get chart markups for a symbol",
		ArgsUsage: "<symbol>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "frequency", Value: "DAILY", Usage: "Chart frequency: DAILY or WEEKLY"},
			&cli.StringFlag{Name: "sort-dir", Value: "ASC", Usage: "Sort direction: ASC or DESC"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				verr := mserrors.NewValidationError("symbol argument required", nil)
				_ = output.WriteError(w, verr)
				return verr
			}
			symbol := cmd.Args().First()
			frequency := cmd.String("frequency")
			sortDir := cmd.String("sort-dir")

			data, err := c.GetChartMarkups(ctx, symbol, frequency, sortDir)
			if err != nil {
				_ = output.WriteError(w, err)
				return err
			}
			return output.WriteSuccess(w, data, output.SymbolMeta(symbol))
		},
	}
}
