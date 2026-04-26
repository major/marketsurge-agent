package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStockDataMarshalUnmarshal(t *testing.T) {
	// Create test data with various field types
	composite := 80
	eps := 85
	name := "Apple Inc."
	price := 150.25

	stock := StockData{
		Symbol: "AAPL",
		Ratings: &Ratings{
			Composite: &composite,
			EPS:       &eps,
		},
		Company: &Company{
			Name: &name,
		},
		Pricing: &Pricing{
			MarketCap: &price,
		},
		Patterns:   []Pattern{},
		TightAreas: []TightArea{},
	}

	// Marshal to JSON
	data, err := json.Marshal(stock)
	require.NoError(t, err)

	// Verify JSON contains expected fields
	var jsonMap map[string]any
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	assert.Equal(t, "AAPL", jsonMap["symbol"])
	assert.NotNil(t, jsonMap["ratings"])

	// Unmarshal back
	var unmarshalled StockData
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, stock.Symbol, unmarshalled.Symbol)
	assert.Equal(t, *stock.Ratings.Composite, *unmarshalled.Ratings.Composite)
	assert.Equal(t, *stock.Company.Name, *unmarshalled.Company.Name)
}

func TestNilPointerFieldsOmitted(t *testing.T) {
	// Create stock with only required fields
	stock := StockData{
		Symbol:     "TSLA",
		Patterns:   []Pattern{},
		TightAreas: []TightArea{},
	}

	data, err := json.Marshal(stock)
	require.NoError(t, err)

	// Verify nil fields are omitted
	var jsonMap map[string]any
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	assert.Equal(t, "TSLA", jsonMap["symbol"])
	assert.Nil(t, jsonMap["ratings"])
	assert.Nil(t, jsonMap["company"])
	assert.Nil(t, jsonMap["pricing"])
}

func TestCatalogKindConstants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		kind     CatalogKind
		expected string
	}{
		{CatalogKindWatchlist, "watchlist"},
		{CatalogKindScreen, "screen"},
		{CatalogKindReport, "report"},
		{CatalogKindCoachScreen, "coach_screen"},
	}

	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.kind))
		})
	}
}

func TestCatalogEntryMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	reportID := 124
	entry := CatalogEntry{
		Name:     "Bases Forming",
		Kind:     CatalogKindReport,
		ReportID: &reportID,
	}

	data, err := json.Marshal(entry)
	require.NoError(t, err)

	var unmarshalled CatalogEntry
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, entry.Name, unmarshalled.Name)
	assert.Equal(t, entry.Kind, unmarshalled.Kind)
	assert.Equal(t, *entry.ReportID, *unmarshalled.ReportID)
}

func TestCatalogKindInJSON(t *testing.T) {
	t.Parallel()
	jsonStr := `{"name":"Test","kind":"watchlist"}`
	var entry CatalogEntry
	err := json.Unmarshal([]byte(jsonStr), &entry)
	require.NoError(t, err)

	assert.Equal(t, CatalogKindWatchlist, entry.Kind)
}

func TestFundamentalDataMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	fundamental := FundamentalData{
		Symbol:           "NVDA",
		ReportedEarnings: []ReportedPeriod{},
		ReportedSales:    []ReportedPeriod{},
		EPSEstimates:     []EstimatePeriod{},
		SalesEstimates:   []EstimatePeriod{},
	}

	data, err := json.Marshal(fundamental)
	require.NoError(t, err)

	var unmarshalled FundamentalData
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, fundamental.Symbol, unmarshalled.Symbol)
}

func TestChartDataMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	chart := ChartData{
		Symbol:     "MSFT",
		TimeSeries: nil,
		Quote:      nil,
	}

	data, err := json.Marshal(chart)
	require.NoError(t, err)

	var unmarshalled ChartData
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, chart.Symbol, unmarshalled.Symbol)
}

func TestRSRatingHistoryMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	rsTrue := true
	rsHistory := RSRatingHistory{
		Symbol:        "GOOGL",
		Ratings:       []RSRatingSnapshot{},
		RSLineNewHigh: &rsTrue,
	}

	data, err := json.Marshal(rsHistory)
	require.NoError(t, err)

	var unmarshalled RSRatingHistory
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, rsHistory.Symbol, unmarshalled.Symbol)
	assert.Equal(t, *rsHistory.RSLineNewHigh, *unmarshalled.RSLineNewHigh)
}

func TestOwnershipDataMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	ownership := OwnershipData{
		Symbol:         "META",
		FundsFloatPct:  nil,
		QuarterlyFunds: []QuarterlyFundOwnership{},
	}

	data, err := json.Marshal(ownership)
	require.NoError(t, err)

	var unmarshalled OwnershipData
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, ownership.Symbol, unmarshalled.Symbol)
}

func TestCatalogMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	catalog := Catalog{
		Entries: []CatalogEntry{
			{
				Name: "Test Watchlist",
				Kind: CatalogKindWatchlist,
			},
		},
		Errors: []string{},
	}

	data, err := json.Marshal(catalog)
	require.NoError(t, err)

	var unmarshalled Catalog
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, len(catalog.Entries), len(unmarshalled.Entries))
	assert.Equal(t, catalog.Entries[0].Name, unmarshalled.Entries[0].Name)
}

func TestQuarterlyFinancialsMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	quarterly := QuarterlyFinancials{
		ReportedEarnings: []QuarterlyReportedPeriod{},
		ReportedSales:    []QuarterlyReportedPeriod{},
		EPSEstimates:     []QuarterlyEstimate{},
		SalesEstimates:   []QuarterlyEstimate{},
		ProfitMargins:    []QuarterlyProfitMargin{},
	}

	data, err := json.Marshal(quarterly)
	require.NoError(t, err)

	var unmarshalled QuarterlyFinancials
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, len(quarterly.ReportedEarnings), len(unmarshalled.ReportedEarnings))
}

func TestChartMarkupMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	markup := ChartMarkup{
		ID:   "markup-123",
		Name: nil,
		Data: `{"type":"line","points":[]}`,
	}

	data, err := json.Marshal(markup)
	require.NoError(t, err)

	var unmarshalled ChartMarkup
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, markup.ID, unmarshalled.ID)
	assert.Equal(t, markup.Data, unmarshalled.Data)
}

func TestPricingWithComplexFields(t *testing.T) {
	t.Parallel()
	pricing := Pricing{
		BlueDotDailyDates:         []string{"2024-01-01", "2024-01-02"},
		BlueDotWeeklyDates:        []string{},
		AntDates:                  []string{},
		HistoricalPriceStatistics: []HistoricalPriceStatistic{},
		VolumeMovingAverages:      []VolumeMovingAverage{},
	}

	data, err := json.Marshal(pricing)
	require.NoError(t, err)

	var unmarshalled Pricing
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, len(pricing.BlueDotDailyDates), len(unmarshalled.BlueDotDailyDates))
}

func TestCatalogResultMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	total := 42
	result := CatalogResult{
		Kind:             CatalogKindReport,
		Total:            &total,
		ScreenResult:     nil,
		AdhocResult:      nil,
		WatchlistEntries: []WatchlistEntry{},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var unmarshalled CatalogResult
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, result.Kind, unmarshalled.Kind)
	assert.Equal(t, *result.Total, *unmarshalled.Total)
}

func TestTimeSeriesMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	timeSeries := TimeSeries{
		Period:     "DAILY",
		DataPoints: []DataPoint{},
	}

	data, err := json.Marshal(timeSeries)
	require.NoError(t, err)

	var unmarshalled TimeSeries
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, timeSeries.Period, unmarshalled.Period)
}

func TestQuoteMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	last := 150.50
	volume := 1000000.0
	quote := Quote{
		Last:   &last,
		Volume: &volume,
	}

	data, err := json.Marshal(quote)
	require.NoError(t, err)

	var unmarshalled Quote
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, *quote.Last, *unmarshalled.Last)
	assert.Equal(t, *quote.Volume, *unmarshalled.Volume)
}
