package client

import (
	"context"

	"github.com/major/marketsurge-agent/internal/constants"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/queries"
)

// GetChartHistory returns daily or weekly chart history for a symbol.
func (c *Client) GetChartHistory(ctx context.Context, symbol, startDate, endDate, period string, daily bool, exchangeName, benchmarkSymbol string) (*models.ChartData, error) {
	queryName := "chart_market_data_weekly.graphql"
	variables := map[string]any{
		"symbols":           []string{symbol},
		"symbolDialectType": constants.SymbolDialectType,
		"where": map[string]any{
			"startDateTime":       map[string]string{"eq": startDate},
			"endDateTime":         map[string]string{"eq": endDate},
			"timeSeriesType":      map[string]string{"eq": period},
			"includeIntradayData": true,
		},
	}
	if benchmarkSymbol != "" {
		variables["symbols"] = []string{symbol, benchmarkSymbol}
	}
	if daily {
		queryName = "chart_market_data.graphql"
		variables["exchangeName"] = exchangeName
	}

	query, err := queries.Load(queryName)
	if err != nil {
		return nil, err
	}

	raw, err := c.Execute(ctx, Request{OperationName: "ChartMarketData", Variables: variables, Query: query})
	if err != nil {
		return nil, err
	}

	item, err := firstMarketData(raw, symbol)
	if err != nil {
		return nil, err
	}

	pricing := getNestedMap(item, "pricing")
	result := &models.ChartData{
		Symbol:             symbol,
		TimeSeries:         buildTimeSeries(getNestedMap(pricing, "timeSeries")),
		Quote:              buildQuote(getNestedMap(pricing, "quote")),
		PremarketQuote:     buildQuote(getNestedMap(pricing, "premarketQuote")),
		PostmarketQuote:    buildQuote(getNestedMap(pricing, "postmarketQuote")),
		CurrentMarketState: stringPtr(pricing["currentMarketState"]),
	}

	exchangeData := raw["data"]
	if dataMap, ok := exchangeData.(map[string]any); ok {
		if rawExchange, ok := dataMap["exchangeData"]; ok {
			switch exchange := rawExchange.(type) {
			case []any:
				result.Exchange = buildExchange(firstMap(exchange))
			case map[string]any:
				result.Exchange = buildExchange(exchange)
			}
		}
	}

	marketData := getNestedSlice(raw, "data", "marketData")
	if len(marketData) > 1 {
		if benchmarkItem, ok := marketData[1].(map[string]any); ok {
			result.BenchmarkTimeSeries = buildTimeSeries(getNestedMap(benchmarkItem, "pricing", "timeSeries"))
		}
	}

	return result, nil
}

func buildQuote(item map[string]any) *models.Quote {
	if len(item) == 0 {
		return nil
	}
	return &models.Quote{
		TradeDateTime:          stringPtr(item["tradeDateTime"]),
		Timeliness:             stringPtr(item["timeliness"]),
		QuoteType:              stringPtr(item["quoteType"]),
		Last:                   floatPtr(item["last"]),
		Volume:                 floatPtr(item["volume"]),
		PercentChange:          floatPtr(item["percentChange"]),
		NetChange:              floatPtr(item["netChange"]),
		LastFormatted:          formattedValue(item["last"]),
		VolumeFormatted:        formattedValue(item["volume"]),
		PercentChangeFormatted: formattedValue(item["percentChange"]),
		NetChangeFormatted:     formattedValue(item["netChange"]),
	}
}

func buildTimeSeries(item map[string]any) *models.TimeSeries {
	if len(item) == 0 {
		return nil
	}
	result := &models.TimeSeries{Period: stringify(item["period"]), DataPoints: []models.DataPoint{}}
	for _, entry := range getNestedSlice(item, "dataPoints") {
		point, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		result.DataPoints = append(result.DataPoints, models.DataPoint{
			StartDateTime: stringPtr(point["startDateTime"]),
			EndDateTime:   stringPtr(point["endDateTime"]),
			Open:          floatPtr(point["open"]),
			High:          floatPtr(point["high"]),
			Low:           floatPtr(point["low"]),
			Close:         floatPtr(point["last"]),
			Volume:        floatPtr(point["volume"]),
		})
	}
	return result
}

func buildExchange(item map[string]any) *models.ExchangeInfo {
	if len(item) == 0 {
		return nil
	}
	result := &models.ExchangeInfo{
		City:        stringPtr(item["city"]),
		CountryCode: stringPtr(item["countryCode"]),
		ExchangeISO: stringPtr(item["exchangeISO"]),
		Holidays:    []models.ExchangeHoliday{},
	}
	for _, entry := range getNestedSlice(item, "holidays") {
		holiday, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		result.Holidays = append(result.Holidays, models.ExchangeHoliday{
			Name:          stringify(holiday["name"]),
			HolidayType:   stringPtr(holiday["holidayType"]),
			Description:   stringPtr(holiday["description"]),
			StartDateTime: stringify(holiday["startDateTime"]),
			EndDateTime:   stringify(holiday["endDateTime"]),
		})
	}
	return result
}
