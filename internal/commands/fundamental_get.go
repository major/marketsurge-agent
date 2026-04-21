package commands

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/client"
)

// FundamentalGetCommand returns the CLI command for retrieving fundamental data.
func FundamentalGetCommand(c *client.Client, w io.Writer) *cli.Command {
	return symbolGetCommand(w, "get", "Get fundamental data for a symbol", func(ctx context.Context, symbol string) (any, error) {
		return c.GetFundamentals(ctx, symbol)
	})
}
