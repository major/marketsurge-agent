package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// IndustryCmd shows industry group relative strength for stocks or ETFs.
type IndustryCmd struct {
	Symbols []string `arg:"" help:"Stock or ETF symbols to look up." sep:","`
}

// Run executes the industry group RS query and writes a JSON array.
func (c *IndustryCmd) Run(client *marketsurge.Client) error {
	return c.run(client, os.Stdout)
}

func (c *IndustryCmd) run(client *marketsurge.Client, w io.Writer) error {
	symbols := normalizeSymbols(c.Symbols)
	if len(symbols) == 0 {
		return mserrors.NewValidationError("at least one symbol is required", errors.New("empty symbols"))
	}

	resp, err := client.IndustryGroupRS(
		context.Background(),
		marketsurge.NewIndustryGroupRSRequest(symbols...),
	)
	if err != nil {
		return clientError("industry group RS request failed", err)
	}

	if err := json.NewEncoder(w).Encode(industryRows(resp)); err != nil {
		return mserrors.NewAPIError("failed to write industry output", err)
	}
	return nil
}

// industryItem is the LLM-friendly output shape for industry group RS.
type industryItem struct {
	Ticker          string `json:"ticker"`
	IndustryGroupRS *int   `json:"industryGroupRS"`
}

// industryRows reshapes an IndustryGroupRSResponse into LLM-friendly output.
func industryRows(resp *marketsurge.IndustryGroupRSResponse) []industryItem {
	if resp == nil || len(resp.MarketData) == 0 {
		return []industryItem{}
	}

	items := make([]industryItem, 0, len(resp.MarketData))
	for _, md := range resp.MarketData {
		item := industryItem{}

		if md.OriginRequest != nil && md.OriginRequest.Symbol != nil {
			item.Ticker = *md.OriginRequest.Symbol
		}

		if md.Industry != nil && len(md.Industry.GroupRS) > 0 {
			item.IndustryGroupRS = md.Industry.GroupRS[0].Value
		}

		items = append(items, item)
	}

	return items
}
