package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// ChartCmd retrieves daily OHLCV chart data and live quotes for a stock or ETF.
type ChartCmd struct {
	Symbol   string `arg:"" help:"Stock or ETF symbol to chart, such as AAPL."`
	Days     int    `help:"Number of calendar days to retrieve." default:"90"`
	Exchange string `help:"Exchange name for holiday calendar." default:"XNYS"`
}

// Run executes the chart query and writes a JSON array.
func (c *ChartCmd) Run(client *marketsurge.Client) error {
	return c.run(client, os.Stdout)
}

func (c *ChartCmd) run(client *marketsurge.Client, w io.Writer) error {
	symbol := strings.ToUpper(strings.TrimSpace(c.Symbol))
	if symbol == "" {
		return mserrors.NewValidationError("symbol is required", errors.New("empty symbol"))
	}

	if c.Days <= 0 {
		return mserrors.NewValidationError(
			"days must be positive",
			fmt.Errorf("invalid days: %d", c.Days),
		)
	}

	now := time.Now().UTC()
	end := now
	start := end.AddDate(0, 0, -c.Days)

	req := marketsurge.NewChartMarketDataRequest(
		[]string{symbol},
		start.Format(chartDateTimeLayout),
		end.Format(chartDateTimeLayout),
		"ONE_DAY",
		c.Exchange,
	)

	resp, err := client.ChartMarketData(context.Background(), req)
	if err != nil {
		return clientError("chart data request failed", err)
	}

	item := chartItemFrom(symbol, resp)

	if err := json.NewEncoder(w).Encode([]chartItem{item}); err != nil {
		return mserrors.NewAPIError("failed to write chart output", err)
	}
	return nil
}

// chartItem is the LLM-friendly output shape for chart data.
type chartItem struct {
	Ticker             string           `json:"ticker"`
	Days               int              `json:"days"`
	Exchange           *string          `json:"exchange,omitempty"`
	DataPoints         []chartDataPoint `json:"dataPoints"`
	Quote              *chartQuote      `json:"quote,omitempty"`
	PremarketQuote     *chartQuote      `json:"premarketQuote,omitempty"`
	PostmarketQuote    *chartQuote      `json:"postmarketQuote,omitempty"`
	CurrentMarketState *string          `json:"currentMarketState,omitempty"`
}

// chartDataPoint is a single daily OHLCV data point.
type chartDataPoint struct {
	Date   string   `json:"date"`
	Open   *float64 `json:"open,omitempty"`
	High   *float64 `json:"high,omitempty"`
	Low    *float64 `json:"low,omitempty"`
	Close  *float64 `json:"close,omitempty"`
	Volume *float64 `json:"volume,omitempty"`
}

// chartQuote represents a live, premarket, or postmarket quote.
type chartQuote struct {
	TradeDateTime *string  `json:"tradeDateTime,omitempty"`
	Timeliness    *string  `json:"timeliness,omitempty"`
	QuoteType     *string  `json:"quoteType,omitempty"`
	Last          *float64 `json:"last,omitempty"`
	PercentChange *float64 `json:"percentChange,omitempty"`
	NetChange     *float64 `json:"netChange,omitempty"`
	Volume        *float64 `json:"volume,omitempty"`
}

func chartItemFrom(symbol string, resp *marketsurge.ChartMarketDataResponse) chartItem {
	item := chartItem{
		Ticker:     symbol,
		DataPoints: []chartDataPoint{},
	}

	if resp == nil || len(resp.MarketData) == 0 {
		return item
	}

	md := &resp.MarketData[0]
	if md.Pricing != nil {
		if md.Pricing.TimeSeries != nil {
			item.DataPoints = chartDataPointsFrom(md.Pricing.TimeSeries.DataPoints)
		}
		item.Quote = chartQuoteFrom(md.Pricing.Quote)
		item.PremarketQuote = chartQuoteFrom(md.Pricing.PremarketQuote)
		item.PostmarketQuote = chartQuoteFrom(md.Pricing.PostmarketQuote)
		item.CurrentMarketState = md.Pricing.CurrentMarketState
	}

	item.Days = len(item.DataPoints)

	if resp.ExchangeData != nil && resp.ExchangeData.ExchangeISO != nil {
		item.Exchange = resp.ExchangeData.ExchangeISO
	}

	return item
}

func chartDataPointsFrom(points []marketsurge.ChartDataPoint) []chartDataPoint {
	if len(points) == 0 {
		return []chartDataPoint{}
	}

	result := make([]chartDataPoint, 0, len(points))
	for i := range points {
		p := &points[i]
		dp := chartDataPoint{
			Date: p.StartDateTime,
		}
		if p.Open != nil {
			dp.Open = p.Open.Value
		}
		if p.High != nil {
			dp.High = p.High.Value
		}
		if p.Low != nil {
			dp.Low = p.Low.Value
		}
		if p.Last != nil {
			dp.Close = p.Last.Value
		}
		if p.Volume != nil {
			dp.Volume = p.Volume.Value
		}
		result = append(result, dp)
	}
	return result
}

func chartQuoteFrom(quote *marketsurge.ChartQuote) *chartQuote {
	if quote == nil {
		return nil
	}
	q := &chartQuote{
		TradeDateTime: quote.TradeDateTime,
		Timeliness:    quote.Timeliness,
		QuoteType:     quote.QuoteType,
	}
	if quote.Last != nil {
		q.Last = quote.Last.Value
	}
	if quote.PercentChange != nil {
		q.PercentChange = quote.PercentChange.Value
	}
	if quote.NetChange != nil {
		q.NetChange = quote.NetChange.Value
	}
	if quote.Volume != nil {
		q.Volume = quote.Volume.Value
	}
	return q
}
