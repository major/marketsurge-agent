package client

import (
	"context"

	"github.com/tidwall/gjson"

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

	result := &models.RSRatingHistory{
		Symbol: symbol,
		Ratings: buildSlice(item.Get("ratings.rsRating").Array(), func(entry gjson.Result) models.RSRatingSnapshot {
			return models.RSRatingSnapshot{
				LetterValue:  gStr(entry.Get("letterValue")),
				Period:       gStr(entry.Get("period")),
				PeriodOffset: gStr(entry.Get("periodOffset")),
				Value:        gInt(entry.Get("value")),
			}
		}),
	}
	result.RSLineNewHigh = gBool(item.Get("pricingStatistics.intradayStatistics.rsLineNewHigh"))

	return result, nil
}
