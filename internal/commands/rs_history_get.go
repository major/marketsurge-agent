package commands

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/client"
)

// RSHistoryGetCommand returns the CLI command for retrieving RS rating history.
func RSHistoryGetCommand(c *client.Client, w io.Writer) *cli.Command {
	return symbolGetCommand(w, "get", "Get RS rating history for a symbol", func(ctx context.Context, symbol string) (any, error) {
		return c.GetRSRatingHistory(ctx, symbol)
	})
}
