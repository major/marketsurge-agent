package commands

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/client"
)

// OwnershipGetCommand returns the CLI command for retrieving ownership data.
func OwnershipGetCommand(c *client.Client, w io.Writer) *cli.Command {
	return symbolGetCommand(w, "get", "Get ownership data for a symbol", func(ctx context.Context, symbol string) (any, error) {
		return c.GetOwnership(ctx, symbol)
	})
}
