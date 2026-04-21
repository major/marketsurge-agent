package commands

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/output"
)

// symbolFetcher retrieves data for a single symbol from the MarketSurge API.
type symbolFetcher func(ctx context.Context, symbol string) (any, error)

// requireSymbol validates that a symbol argument was provided and returns it.
func requireSymbol(cmd *cli.Command) (string, error) {
	if cmd.Args().Len() == 0 {
		return "", mserrors.NewValidationError("symbol argument required", nil)
	}
	return cmd.Args().First(), nil
}

// symbolGetCommand builds a CLI command that fetches data for a single symbol.
// It handles argument validation, calls the fetcher, and writes the JSON envelope.
//
//nolint:unparam // name is always "get" today but kept as a parameter for future subcommands
func symbolGetCommand(w io.Writer, name, usage string, fetch symbolFetcher) *cli.Command {
	return &cli.Command{
		Name:      name,
		Usage:     usage,
		ArgsUsage: "<symbol>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			symbol, err := requireSymbol(cmd)
			if err != nil {
				return err
			}
			data, err := fetch(ctx, symbol)
			if err != nil {
				return err
			}
			return output.WriteSuccess(w, data, output.SymbolMeta(symbol))
		},
	}
}
