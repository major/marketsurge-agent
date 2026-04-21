// Package commands implements CLI command handlers for marketsurge-agent.
package commands

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/client"
)

// StockGetCommand returns the CLI command for retrieving stock data.
func StockGetCommand(c *client.Client, w io.Writer) *cli.Command {
	return symbolGetCommand(w, "get", "Get stock data for a symbol", func(ctx context.Context, symbol string) (any, error) {
		return c.GetStock(ctx, symbol)
	})
}
