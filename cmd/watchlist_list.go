package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// WatchlistListCmd lists all saved watchlists.
type WatchlistListCmd struct{}

// Run fetches all watchlist names and writes them as a JSON array.
func (c *WatchlistListCmd) Run(ctx context.Context, client *marketsurge.Client) error {
	return c.run(ctx, client, os.Stdout)
}

func (c *WatchlistListCmd) run(ctx context.Context, client *marketsurge.Client, w io.Writer) error {
	resp, err := client.GetAllWatchlistNames(
		ctx,
		marketsurge.NewGetAllWatchlistNamesRequest(),
	)
	if err != nil {
		return clientError("watchlist list request failed", err)
	}

	watchlists := []marketsurge.WatchlistSummary{}
	if resp != nil && len(resp.Watchlists) > 0 {
		watchlists = resp.Watchlists
	}

	if err := json.NewEncoder(w).Encode(watchlists); err != nil {
		return mserrors.NewAPIError("failed to write watchlist list output", err)
	}
	return nil
}
