package commands

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/output"
)

// OwnershipGetCommand returns the CLI command for retrieving ownership data.
func OwnershipGetCommand(c *client.Client, w io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get ownership data for a symbol",
		ArgsUsage: "<symbol>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				verr := mserrors.NewValidationError("symbol argument required", nil)
				_ = output.WriteError(w, verr)
				return verr
			}
			symbol := cmd.Args().First()
			data, err := c.GetOwnership(ctx, symbol)
			if err != nil {
				_ = output.WriteError(w, err)
				return err
			}
			return output.WriteSuccess(w, data, output.SymbolMeta(symbol))
		},
	}
}
