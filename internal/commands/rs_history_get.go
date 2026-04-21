package commands

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/output"
)

// RSHistoryGetCommand returns the CLI command for retrieving RS rating history.
func RSHistoryGetCommand(c *client.Client, w io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get RS rating history for a symbol",
		ArgsUsage: "<symbol>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				verr := mserrors.NewValidationError("symbol argument required", nil)
				return verr
			}
			symbol := cmd.Args().First()
			data, err := c.GetRSRatingHistory(ctx, symbol)
			if err != nil {
				return err
			}
			return output.WriteSuccess(w, data, output.SymbolMeta(symbol))
		},
	}
}
