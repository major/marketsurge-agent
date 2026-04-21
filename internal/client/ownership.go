package client

import (
	"context"

	"github.com/tidwall/gjson"

	"github.com/major/marketsurge-agent/internal/constants"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/queries"
)

// GetOwnership returns ownership history for a symbol.
func (c *Client) GetOwnership(ctx context.Context, symbol string) (*models.OwnershipData, error) {
	query, err := queries.Load("ownership.graphql")
	if err != nil {
		return nil, err
	}

	raw, err := c.Execute(ctx, Request{
		OperationName: "Ownership",
		Variables: map[string]any{
			"symbols":           []string{symbol},
			"symbolDialectType": constants.SymbolDialectType,
		},
		Query: query,
	})
	if err != nil {
		return nil, err
	}

	item, err := firstMarketData(raw, symbol)
	if err != nil {
		return nil, err
	}

	ownership := item.Get("ownership")
	return &models.OwnershipData{
		Symbol:        symbol,
		FundsFloatPct: gStr(ownership.Get("fundsFloatPercentHeld.formattedValue")),
		QuarterlyFunds: buildSlice(ownership.Get("fundOwnershipSummary").Array(), func(entry gjson.Result) models.QuarterlyFundOwnership {
			return models.QuarterlyFundOwnership{
				Date:  gStr(entry.Get("date")),
				Count: gStr(entry.Get("numberOfFundsHeld.formattedValue")),
			}
		}),
	}, nil
}
