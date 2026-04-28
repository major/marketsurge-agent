package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/major/marketsurge-agent/internal/constants"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/queries"
)

// GetRSRatingHistory returns RS rating history for a symbol.
func (c *Client) GetRSRatingHistory(ctx context.Context, symbol string) (*models.RSRatingHistory, error) {
	histories, err := c.GetRSRatingHistories(ctx, []string{symbol})
	if err != nil {
		return nil, err
	}

	history, ok := histories[symbol]
	if !ok {
		return nil, firstMissingRSHistoryError([]string{symbol})
	}
	return history, nil
}

// GetRSRatingHistories returns RS rating history for each symbol returned by the API.
func (c *Client) GetRSRatingHistories(ctx context.Context, symbols []string) (map[string]*models.RSRatingHistory, error) {
	query, err := queries.Load("rs_rating_ri_panel.graphql")
	if err != nil {
		return nil, err
	}

	raw, err := c.Execute(ctx, Request{
		OperationName: "RSRatingRIPanel",
		Variables: map[string]any{
			"symbols":           symbols,
			"symbolDialectType": constants.SymbolDialectType,
		},
		Query: query,
	})
	if err != nil {
		return nil, err
	}

	marketData := getNestedSlice(raw, "data", "marketData")
	if len(marketData) == 0 {
		return nil, firstMissingRSHistoryError(symbols)
	}

	requestedSymbols := make(map[string]string, len(symbols))
	for _, symbol := range symbols {
		requestedSymbols[strings.ToUpper(symbol)] = symbol
	}

	histories := make(map[string]*models.RSRatingHistory, len(marketData))
	for _, item := range marketData {
		mapping, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid marketData item")
		}
		responseSymbol := rsHistoryResponseSymbol(mapping)
		if responseSymbol == "" {
			return nil, fmt.Errorf("missing originRequest symbol in marketData item")
		}
		symbol, ok := requestedSymbols[strings.ToUpper(responseSymbol)]
		if !ok {
			symbol = responseSymbol
		}
		histories[symbol] = rsRatingHistoryFromMarketData(symbol, mapping)
	}

	return histories, nil
}

func rsHistoryResponseSymbol(item map[string]any) string {
	originRequest := getNestedMap(item, "originRequest")
	return stringify(originRequest["symbol"])
}

func rsRatingHistoryFromMarketData(symbol string, item map[string]any) *models.RSRatingHistory {
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

	return result
}

func firstMissingRSHistoryError(symbols []string) error {
	if len(symbols) == 1 {
		return firstMarketDataMissingError(symbols[0])
	}
	return firstMarketDataMissingError(strings.Join(symbols, ","))
}

func firstMarketDataMissingError(symbol string) error {
	return mserrors.NewSymbolNotFoundError(fmt.Sprintf("symbol not found: %q", symbol), nil, symbol)
}
