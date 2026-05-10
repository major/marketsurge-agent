package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// WatchlistGetCmd retrieves symbols in a watchlist by ID.
type WatchlistGetCmd struct {
	ID string `arg:"" help:"Watchlist ID from 'watchlist list' output."`
}

// Run fetches watchlist details and writes an LLM-friendly JSON array.
func (c *WatchlistGetCmd) Run(client *marketsurge.Client) error {
	return c.run(client, os.Stdout)
}

func (c *WatchlistGetCmd) run(client *marketsurge.Client, w io.Writer) error {
	resp, err := client.FlaggedSymbols(
		context.Background(),
		marketsurge.NewFlaggedSymbolsRequest(c.ID),
	)
	if err != nil {
		return clientError("watchlist get request failed", err)
	}

	item := watchlistGetItem(resp)

	if err := json.NewEncoder(w).Encode([]watchlistItem{item}); err != nil {
		return mserrors.NewAPIError("failed to write watchlist get output", err)
	}
	return nil
}

// watchlistItem is the LLM-friendly output shape for a watchlist.
type watchlistItem struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	LastModifiedDateUtc string   `json:"lastModifiedDateUtc"`
	Description        string   `json:"description"`
	Symbols            []string `json:"symbols"`
}

// watchlistGetItem reshapes a FlaggedSymbolsResponse into an LLM-friendly watchlistItem.
func watchlistGetItem(resp *marketsurge.FlaggedSymbolsResponse) watchlistItem {
	if resp == nil {
		return watchlistItem{Symbols: []string{}}
	}

	symbols := make([]string, 0, len(resp.Watchlist.Items))
	for _, item := range resp.Watchlist.Items {
		if item.Key != "" {
			symbols = append(symbols, item.Key)
		}
	}

	return watchlistItem{
		ID:                 resp.Watchlist.ID,
		Name:               resp.Watchlist.Name,
		LastModifiedDateUtc: resp.Watchlist.LastModifiedDateUtc,
		Description:        resp.Watchlist.Description,
		Symbols:            symbols,
	}
}
