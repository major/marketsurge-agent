package cmd

import (
	"encoding/json"
	"io"
	"os"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// ColumnsCmd lists available MarketSurge data columns from the local catalog.
type ColumnsCmd struct {
	Category string `help:"Filter columns by category (exact match)." default:""`
}

// Run writes the column catalog as a JSON array to stdout.
// This command uses the local catalog and does not require authentication.
func (c *ColumnsCmd) Run() error {
	return c.run(os.Stdout)
}

// run writes the column catalog as a JSON array to w.
func (c *ColumnsCmd) run(w io.Writer) error {
	var columns []marketsurge.ColumnInfo
	if c.Category != "" {
		columns = marketsurge.ColumnsByCategory(c.Category)
	} else {
		columns = marketsurge.Columns()
	}

	// Reshape into LLM-friendly output without Go-specific type names.
	items := make([]columnItem, 0, len(columns))
	for _, col := range columns {
		items = append(items, columnItem{
			Name:        string(col.Name),
			DisplayName: col.DisplayName,
			Description: col.Description,
			Category:    col.Category,
		})
	}

	if err := json.NewEncoder(w).Encode(items); err != nil {
		return mserrors.NewAPIError("failed to write columns output", err)
	}

	return nil
}

// columnItem is the JSON output shape for a single column entry.
type columnItem struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}
