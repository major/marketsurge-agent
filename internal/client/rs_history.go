package client

import (
	"context"

	"github.com/major/marketsurge-agent/internal/constants"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/queries"
)

// GetRSRatingHistory returns RS rating history for a symbol.
func (c *Client) GetRSRatingHistory(ctx context.Context, symbol string) (*models.RSRatingHistory, error) {
	query, err := queries.Load("rs_rating_ri_panel.graphql")
	if err != nil {
		return nil, err
	}

	raw, err := c.Execute(ctx, Request{
		OperationName: "RSRatingRIPanel",
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

	ratings := getNestedSlice(item, "ratings", "rsRating")
	intraday := getNestedMap(item, "pricingStatistics", "intradayStatistics")
	result := &models.RSRatingHistory{
		Symbol: symbol,
		Ratings: buildSlice(ratings, func(item map[string]any) models.RSRatingSnapshot {
			return models.RSRatingSnapshot{
				LetterValue:  stringPtr(item["letterValue"]),
				Period:       stringPtr(item["period"]),
				PeriodOffset: stringPtr(item["periodOffset"]),
				Value:        intPtr(item["value"]),
			}
		}),
	}
	result.RSLineNewHigh = boolPtr(intraday["rsLineNewHigh"])

	return result, nil
}
