package client

import (
	"context"

	"github.com/tidwall/gjson"

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

	pricing := item.Get("pricing")
	result := &models.ChartData{
		Symbol:             symbol,
		TimeSeries:         buildTimeSeries(pricing.Get("timeSeries")),
		Quote:              buildQuote(pricing.Get("quote")),
		PremarketQuote:     buildQuote(pricing.Get("premarketQuote")),
		PostmarketQuote:    buildQuote(pricing.Get("postmarketQuote")),
		CurrentMarketState: gStr(pricing.Get("currentMarketState")),
	}

	exchangeData := gjson.GetBytes(raw, "data.exchangeData")
	if exchangeData.Exists() {
		if exchangeData.IsArray() {
			if first := exchangeData.Get("0"); first.Exists() {
				result.Exchange = buildExchange(first)
			}
		} else if exchangeData.IsObject() {
			result.Exchange = buildExchange(exchangeData)
		}
	}

	benchmarkItem := gjson.GetBytes(raw, "data.marketData.1")
	if benchmarkItem.Exists() {
		result.BenchmarkTimeSeries = buildTimeSeries(benchmarkItem.Get("pricing.timeSeries"))
	}

	return result, nil
}

func buildQuote(item gjson.Result) *models.Quote {
	if !item.Exists() {
		return nil
	}
	return &models.Quote{
		TradeDateTime:          gStr(item.Get("tradeDateTime")),
		Timeliness:             gStr(item.Get("timeliness")),
		QuoteType:              gStr(item.Get("quoteType")),
		Last:                   gFloat(item.Get("last")),
		Volume:                 gFloat(item.Get("volume")),
		PercentChange:          gFloat(item.Get("percentChange")),
		NetChange:              gFloat(item.Get("netChange")),
		LastFormatted:          gStr(item.Get("last.formattedValue")),
		VolumeFormatted:        gStr(item.Get("volume.formattedValue")),
		PercentChangeFormatted: gStr(item.Get("percentChange.formattedValue")),
		NetChangeFormatted:     gStr(item.Get("netChange.formattedValue")),
	}
}

func buildTimeSeries(item gjson.Result) *models.TimeSeries {
	if !item.Exists() {
		return nil
	}
	return &models.TimeSeries{
		Period: stringify(item.Get("period")),
		DataPoints: buildSlice(item.Get("dataPoints").Array(), func(point gjson.Result) models.DataPoint {
			return models.DataPoint{
				StartDateTime: gStr(point.Get("startDateTime")),
				EndDateTime:   gStr(point.Get("endDateTime")),
				Open:          gFloat(point.Get("open")),
				High:          gFloat(point.Get("high")),
				Low:           gFloat(point.Get("low")),
				Close:         gFloat(point.Get("last")),
				Volume:        gFloat(point.Get("volume")),
			}
		}),
	}
}

func buildExchange(item gjson.Result) *models.ExchangeInfo {
	if !item.Exists() {
		return nil
	}
	return &models.ExchangeInfo{
		City:        gStr(item.Get("city")),
		CountryCode: gStr(item.Get("countryCode")),
		ExchangeISO: gStr(item.Get("exchangeISO")),
		Holidays: buildSlice(item.Get("holidays").Array(), func(holiday gjson.Result) models.ExchangeHoliday {
			return models.ExchangeHoliday{
				Name:          stringify(holiday.Get("name")),
				HolidayType:   gStr(holiday.Get("holidayType")),
				Description:   gStr(holiday.Get("description")),
				StartDateTime: stringify(holiday.Get("startDateTime")),
				EndDateTime:   stringify(holiday.Get("endDateTime")),
			}
		}),
	}
}
