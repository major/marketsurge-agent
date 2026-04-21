package client

import (
	"context"

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

	ownership := getNestedMap(item, "ownership")
	quarterly := getNestedSlice(ownership, "fundOwnershipSummary")
	result := &models.OwnershipData{
		Symbol:         symbol,
		FundsFloatPct:  formattedValue(ownership["fundsFloatPercentHeld"]),
		QuarterlyFunds: make([]models.QuarterlyFundOwnership, 0, len(quarterly)),
	}

	for _, entry := range quarterly {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		result.QuarterlyFunds = append(result.QuarterlyFunds, models.QuarterlyFundOwnership{
			Date:  stringPtr(item["date"]),
			Count: formattedValue(item["numberOfFundsHeld"]),
		})
	}

	return result, nil
}
