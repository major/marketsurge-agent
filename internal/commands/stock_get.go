// Package commands implements CLI command handlers for marketsurge-agent.
package commands

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/output"
)

// StockGetCommand returns the CLI command for retrieving stock data.
func StockGetCommand(c *client.Client, w io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get stock data for a symbol",
		ArgsUsage: "<symbol>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				verr := mserrors.NewValidationError("symbol argument required", nil)
				return verr
			}
			symbol := cmd.Args().First()
			data, err := c.GetStock(ctx, symbol)
			if err != nil {
				return err
			}
			return output.WriteSuccess(w, data, output.SymbolMeta(symbol))
		},
	}
}
