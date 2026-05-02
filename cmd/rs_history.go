// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/internal/output"
)

func init() { rootCmd.AddCommand(newRSHistoryCmd()) }

func newRSHistoryCmd() *cobra.Command {
	parent := &cobra.Command{
		Use:   "rs-history",
		Short: "RS rating history commands",
	}
	parent.AddCommand(newRSHistoryGetCmd())
	return parent
}

func newRSHistoryGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <symbol> [symbol...]",
		Short: "Get RS rating history for one or more symbols",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c := ClientFromContext(ctx)

			if len(args) == 1 {
				data, err := c.GetRSRatingHistory(ctx, args[0])
				if err != nil {
					return err
				}
				return output.WriteSuccess(cmd.OutOrStdout(), data, output.SymbolMeta(args[0]))
			}

			histories, err := c.GetRSRatingHistories(ctx, args)
			if err != nil {
				return err
			}

			data, errs := orderedRSHistoryData(args, histories)
			meta := map[string]any{
				"symbols":   args,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}
			if len(data) == 0 {
				return fmt.Errorf("rs history failed for all symbols: %v", errs)
			}
			if len(errs) > 0 {
				return output.WritePartial(cmd.OutOrStdout(), data, errs, meta)
			}
			return output.WriteSuccess(cmd.OutOrStdout(), data, meta)
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
