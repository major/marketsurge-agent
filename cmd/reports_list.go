package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// ReportsListCmd lists all available screens and reports.
type ReportsListCmd struct{}

// Run fetches MarketSurge screens and writes them as a JSON array.
func (c *ReportsListCmd) Run(client *marketsurge.Client) error {
	resp, err := client.Screens(context.Background(), marketsurge.NewScreensRequest())
	if err != nil {
		if marketsurge.IsAuthError(err) {
			return mserrors.NewAuthenticationError("authentication failed", err)
		}
		if marketsurge.IsRateLimited(err) {
			return mserrors.NewHTTPError("rate limited", err, 429, "")
		}
		return mserrors.NewAPIError("API request failed", err)
	}

	screens := []marketsurge.ScreenEntry{}
	if resp != nil && resp.User != nil && len(resp.User.Screens) > 0 {
		screens = resp.User.Screens
	}

	if err := json.NewEncoder(os.Stdout).Encode(screens); err != nil {
		return mserrors.NewAPIError("failed to write reports list output", err)
	}
	return nil
}
