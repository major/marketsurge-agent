package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// ReportsListCmd lists all available screens and reports.
type ReportsListCmd struct{}

// Run fetches MarketSurge screens and writes them as a JSON array to stdout.
func (c *ReportsListCmd) Run(client *marketsurge.Client) error {
	return c.run(client, os.Stdout)
}

// run fetches MarketSurge screens and writes them as a JSON array to w.
func (c *ReportsListCmd) run(client *marketsurge.Client, w io.Writer) error {
	resp, err := client.Screens(context.Background(), marketsurge.NewScreensRequest())
	if err != nil {
		return clientError("API request failed", err)
	}

	screens := []marketsurge.ScreenEntry{}
	if resp != nil && resp.User != nil && len(resp.User.Screens) > 0 {
		screens = resp.User.Screens
	}

	if err := json.NewEncoder(w).Encode(screens); err != nil {
		return mserrors.NewAPIError("failed to write reports list output", err)
	}

	return nil
}
