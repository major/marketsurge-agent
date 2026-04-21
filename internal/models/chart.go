package models

// DataPoint represents a single OHLCV data point from a time series.
type DataPoint struct {
	StartDateTime *string  `json:"start_date_time,omitempty"`
	EndDateTime   *string  `json:"end_date_time,omitempty"`
	Open          *float64 `json:"open,omitempty"`
	High          *float64 `json:"high,omitempty"`
	Low           *float64 `json:"low,omitempty"`
	Close         *float64 `json:"close,omitempty"`
	Volume        *float64 `json:"volume,omitempty"`
}

// TimeSeries represents a time series container with period and data points.
type TimeSeries struct {
	Period     string      `json:"period"`
	DataPoints []DataPoint `json:"data_points"`
}

// Quote represents real-time or extended-hours quote data.
type Quote struct {
	TradeDateTime           *string  `json:"trade_date_time,omitempty"`
	Timeliness              *string  `json:"timeliness,omitempty"`
	QuoteType               *string  `json:"quote_type,omitempty"`
	Last                    *float64 `json:"last,omitempty"`
	Volume                  *float64 `json:"volume,omitempty"`
	PercentChange           *float64 `json:"percent_change,omitempty"`
	NetChange               *float64 `json:"net_change,omitempty"`
	LastFormatted           *string  `json:"last_formatted,omitempty"`
	VolumeFormatted         *string  `json:"volume_formatted,omitempty"`
	PercentChangeFormatted  *string  `json:"percent_change_formatted,omitempty"`
	NetChangeFormatted      *string  `json:"net_change_formatted,omitempty"`
}

// ExchangeHoliday represents an exchange holiday entry.
type ExchangeHoliday struct {
	Name          string  `json:"name"`
	HolidayType   *string `json:"holiday_type,omitempty"`
	Description   *string `json:"description,omitempty"`
	StartDateTime string  `json:"start_date_time"`
	EndDateTime   string  `json:"end_date_time"`
}

// ExchangeInfo represents exchange metadata and holiday schedule.
type ExchangeInfo struct {
	City          *string             `json:"city,omitempty"`
	CountryCode   *string             `json:"country_code,omitempty"`
	ExchangeISO   *string             `json:"exchange_iso,omitempty"`
	Holidays      []ExchangeHoliday   `json:"holidays"`
}

// ChartData represents complete chart data response from the ChartMarketData query.
type ChartData struct {
	Symbol                string       `json:"symbol"`
	TimeSeries            *TimeSeries  `json:"time_series,omitempty"`
	BenchmarkTimeSeries   *TimeSeries  `json:"benchmark_time_series,omitempty"`
	Quote                 *Quote       `json:"quote,omitempty"`
	PremarketQuote        *Quote       `json:"premarket_quote,omitempty"`
	PostmarketQuote       *Quote       `json:"postmarket_quote,omitempty"`
	CurrentMarketState    *string      `json:"current_market_state,omitempty"`
	Exchange              *ExchangeInfo `json:"exchange,omitempty"`
}
