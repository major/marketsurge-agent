package commands

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/internal/output"
)

// RSHistoryGetCommand returns the CLI command for retrieving RS rating history.
func RSHistoryGetCommand(c *client.Client, w io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get RS rating history for one or more symbols",
		ArgsUsage: "[symbol...]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			symbols := cmd.Args().Slice()
			if len(symbols) == 0 {
				return mserrors.NewValidationError("symbol is required", nil)
			}

			if len(symbols) == 1 {
				data, err := c.GetRSRatingHistory(ctx, symbols[0])
				if err != nil {
					return err
				}
				return output.WriteSuccess(w, data, output.SymbolMeta(symbols[0]))
			}

			histories, err := c.GetRSRatingHistories(ctx, symbols)
			if err != nil {
				return err
			}

			data, errs := orderedRSHistoryData(symbols, histories)
			meta := map[string]any{
				"symbols":   symbols,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}
			if len(data) == 0 {
				return fmt.Errorf("rs history failed for all symbols: %v", errs)
			}
			if len(errs) > 0 {
				return output.WritePartial(w, data, errs, meta)
			}
			return output.WriteSuccess(w, data, meta)
		},
	}
}

func orderedRSHistoryData(symbols []string, histories map[string]*models.RSRatingHistory) (data map[string]*models.RSRatingHistory, errs []string) {
	data = make(map[string]*models.RSRatingHistory, len(histories))
	errs = make([]string, 0)
	for _, symbol := range symbols {
		history, ok := histories[symbol]
		if !ok {
			errs = append(errs, fmt.Sprintf("%s: symbol not found", symbol))
			continue
		}
		data[symbol] = history
	}
	return data, errs
}
