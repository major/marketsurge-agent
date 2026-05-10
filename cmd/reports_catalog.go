package cmd

import (
	"encoding/json"
	"io"
	"os"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// ReportsCatalogCmd lists built-in MarketSurge report screens from the local catalog.
type ReportsCatalogCmd struct{}

// Run writes the report screen catalog as a JSON array to stdout.
// This command uses the local catalog and does not require authentication.
func (c *ReportsCatalogCmd) Run() error {
	return c.run(os.Stdout)
}

// run writes the report screen catalog as a JSON array to w.
func (c *ReportsCatalogCmd) run(w io.Writer) error {
	screens := marketsurge.ReportScreens()

	items := make([]reportScreenItem, 0, len(screens))
	for _, s := range screens {
		items = append(items, reportScreenItem{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
		})
	}

	if err := json.NewEncoder(w).Encode(items); err != nil {
		return mserrors.NewAPIError("failed to write report screens catalog output", err)
	}

	return nil
}

// reportScreenItem is the JSON output shape for a single built-in report screen.
type reportScreenItem struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
