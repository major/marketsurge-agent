package client

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStockSuccess(t *testing.T) {
	t.Parallel()
	var captured Request
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(stockResponseJSON()))
	})

	stock, err := client.GetStock(t.Context(), "AAPL")
	require.NoError(t, err)
	require.NotNil(t, stock)
	assert.Equal(t, "OtherMarketData", captured.OperationName)
	assert.Equal(t, "CHARTING", captured.Variables["symbolDialectType"])
	assert.Contains(t, captured.Query, "202")
	assert.Equal(t, "AAPL", stock.Symbol)
	assert.Equal(t, 99, *stock.Ratings.Composite)
	assert.Equal(t, "Apple Inc.", *stock.Company.Name)
	assert.Equal(t, 3, len(stock.QuarterlyFinancials.ReportedEarnings))
}

func TestGetStockReturnsSymbolNotFoundForEmptyMarketData(t *testing.T) {
	t.Parallel()
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"marketData":[]}}`))
	})

	_, err := client.GetStock(t.Context(), "MISSING")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "symbol not found")
}

func TestGetFundamentalsSuccess(t *testing.T) {
	t.Parallel()
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(stockResponseJSON()))
	})

	data, err := client.GetFundamentals(t.Context(), "AAPL")
	require.NoError(t, err)
	assert.Equal(t, "AAPL", data.Symbol)
	assert.Equal(t, "Apple Inc.", *data.CompanyName)
	assert.Len(t, data.ReportedEarnings, 3)
	assert.Len(t, data.EPSEstimates, 1)
}

func TestGetOwnershipSuccess(t *testing.T) {
	t.Parallel()
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(stockResponseJSON()))
	})

	data, err := client.GetOwnership(t.Context(), "AAPL")
	require.NoError(t, err)
	assert.Equal(t, "65%", *data.FundsFloatPct)
	assert.Len(t, data.QuarterlyFunds, 1)
	assert.Equal(t, "100", *data.QuarterlyFunds[0].Count)
}

func TestGetRSRatingHistorySuccess(t *testing.T) {
	t.Parallel()
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(stockResponseJSON()))
	})

	data, err := client.GetRSRatingHistory(t.Context(), "AAPL")
	require.NoError(t, err)
	assert.Equal(t, "AAPL", data.Symbol)
	assert.True(t, *data.RSLineNewHigh)
	assert.Len(t, data.Ratings, 1)
	assert.Equal(t, 95, *data.Ratings[0].Value)
}

func stockResponseJSON() string {
	return `{
		"data": {
			"marketData": [{
				"ratings": {
					"compRating": [{"value": 99}],
					"epsRating": [{"value": 95}],
					"rsRating": [{"value": 95, "letterValue": "A", "period": "WEEK", "periodOffset": "0"}],
					"smrRating": [{"letterValue": "A"}],
					"adRating": [{"letterValue": "B"}]
				},
				"pricingStatistics": {
					"endOfDayStatistics": {
						"marketCapitalization": {"value": 1000, "formattedValue": "$1B"},
						"avgDollarVolume50Day": {"value": 5, "formattedValue": "$5M"},
						"upDownVolumeRatio": {"value": 1.2, "formattedValue": "1.2"},
						"averageTrueRangePercent": [{"value": 2.3, "formattedValue": "2.3%"}],
						"alpha": {"value": 1.1, "formattedValue": "1.1"},
						"beta": {"value": 0.9, "formattedValue": "0.9"},
						"pricingStartDate": {"value": "2024-01-01"},
						"pricingEndDate": {"value": "2024-02-01"}
					},
					"intradayStatistics": {
						"yield": {"value": 1.5, "formattedValue": "1.5%"},
						"priceToCashFlowRatio": {"value": 10, "formattedValue": "10"},
						"forwardPriceToEarningsRatio": {"value": 11, "formattedValue": "11"},
						"priceToSalesRatio": {"value": 12, "formattedValue": "12"},
						"priceToEarningsRatio": {"value": 13, "formattedValue": "13"},
						"priceToEarningsVsSP500": {"value": 1.4, "formattedValue": "1.4"},
						"cashFlowPerShareLastYear": {"value": 14, "formattedValue": "14"},
						"rsLineNewHigh": true
					}
				},
				"financials": {
					"epsDueDate": {"value": "2024-05-01"},
					"epsDueDateStatus": "CONFIRMED",
					"epsLastReportedDate": {"value": "2024-02-01"},
					"consensusFinancials": {
						"eps": {
							"growthRate": [{"value": 20.1}],
							"reportedEarnings": [
								{"value": {"value": 1.1, "formattedValue": "1.1"}, "percentChangeYOY": {"value": 10, "formattedValue": "10%"}, "periodOffset": "0", "periodEndDate": {"value": "2024-01-01"}},
								{"value": {"value": 1.2, "formattedValue": "1.2"}, "percentChangeYOY": {"value": 11, "formattedValue": "11%"}, "periodOffset": "1", "periodEndDate": {"value": "2023-10-01"}},
								{"value": {"value": 1.3, "formattedValue": "1.3"}, "percentChangeYOY": {"value": 12, "formattedValue": "12%"}, "periodOffset": "2", "periodEndDate": {"value": "2023-07-01"}}
							]
						},
						"sales": {
							"growthRate": [{"value": 15.1}],
							"reportedSales": [{"value": {"value": 5.5, "formattedValue": "5.5"}, "percentChangeYOY": {"value": 8, "formattedValue": "8%"}, "periodOffset": "0", "periodEndDate": {"value": "2024-01-01"}}]
						}
					},
					"estimates": {
						"epsEstimates": [{"value": {"value": 2.1, "formattedValue": "2.1"}, "percentChangeYOY": {"value": 9, "formattedValue": "9%"}, "periodOffset": "F1", "period": "FY2025", "revisionDirection": "UP"}],
						"salesEstimates": [{"value": {"value": 10.1, "formattedValue": "10.1"}, "percentChangeYOY": {"value": 7, "formattedValue": "7%"}, "periodOffset": "F1", "period": "FY2025", "revisionDirection": "UP"}]
					},
					"profitMarginValues": [{"periodOffset": "0", "periodEndDate": {"value": "2024-01-01"}, "preTaxMargin": {"value": 30}, "afterTaxMargin": {"value": 25}, "grossMargin": {"value": 40}, "returnOnEquity": {"value": 18}}],
					"earningsStability": 87
				},
				"industry": {
					"name": "Technology",
					"sector": "Tech",
					"indCode": "123",
					"numberOfStocksInGroup": 42,
					"groupRanks": [{"value": 3}],
					"groupRS": [{"value": 90, "letterValue": "A"}]
				},
				"ownership": {
					"fundsFloatPercentHeld": {"value": 65, "formattedValue": "65%"},
					"fundOwnershipSummary": [{"date": {"value": "2024-01-01"}, "numberOfFundsHeld": {"formattedValue": "100"}}]
				},
				"fundamentals": {
					"researchAndDevelopmentPercentLastQtr": {"value": 4.5, "formattedValue": "4.5%"},
					"debtPercent": {"formattedValue": "12%"},
					"newCEODate": {"value": "2023-01-01"}
				},
				"corporateActions": {
					"dividendNextReportedExDate": {"value": "2024-06-01"}
				},
				"symbology": {
					"company": [{"companyName": "Apple Inc.", "businessDescription": "Desc", "url": "https://apple.example", "address": "One Apple Park", "address2": "Suite 1", "phone": "123", "city": "Cupertino", "country": "USA", "stateProvince": "CA"}],
					"instrument": [{"ipoDate": {"value": "1980-12-12"}, "ipoPrice": {"value": 22, "formattedValue": "$22"}, "subType": "COMMON"}]
				}
			}]
		}
	}`
}
