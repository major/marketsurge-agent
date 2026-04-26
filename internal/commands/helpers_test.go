package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/major/marketsurge-agent/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// runTestCommand configures the command to suppress os.Exit and runs it with the given args.
func runTestCommand(t *testing.T, cmd *cli.Command, args ...string) error {
	t.Helper()
	cmd.ExitErrHandler = func(_ context.Context, _ *cli.Command, _ error) {}
	return cmd.Run(t.Context(), args)
}

// parseJSONEnvelope unmarshals buf into a map and asserts it contains data and metadata keys.
func parseJSONEnvelope(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var result map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Contains(t, result, "data")
	assert.Contains(t, result, "metadata")
	return result
}

// assertSymbolMeta asserts that the envelope metadata contains the expected symbol.
func assertSymbolMeta(t *testing.T, envelope map[string]any, symbol string) {
	t.Helper()
	meta, _ := envelope["metadata"].(map[string]any)
	assert.Equal(t, symbol, meta["symbol"])
}

// testClient creates a *client.Client backed by the given httptest server.
func testClient(t *testing.T, server *httptest.Server) *client.Client {
	t.Helper()
	return client.NewClient("test-jwt", client.WithBaseURL(server.URL), client.WithHTTPClient(server.Client()))
}

// jsonServer returns an httptest.Server that always responds with the given JSON body.
func jsonServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
}

// stockResponseFixture returns a minimal but complete stock API response.
func stockResponseFixture() string {
	return `{
		"data": {
			"marketData": [{
				"ratings": {
					"compRating": [{"value": 99}],
					"epsRating": [{"value": 95}],
					"rsRating": [{"value": 90, "letterValue": "A", "period": "WEEK", "periodOffset": "0"}],
					"smrRating": [{"letterValue": "A"}],
					"adRating": [{"letterValue": "B"}]
				},
				"pricingStatistics": {
					"endOfDayStatistics": {
						"marketCapitalization": {"value": 3000000000000, "formattedValue": "$3T"},
						"avgDollarVolume50Day": {"value": 5000000, "formattedValue": "$5M"},
						"upDownVolumeRatio": {"value": 1.2, "formattedValue": "1.2"},
						"averageTrueRangePercent": [{"value": 2.3, "formattedValue": "2.3%"}],
						"antEvents": [{"value": "2024-12-18"}],
						"blueDotDailyEvents": [{"value": "2024-12-20", "formattedValue": "Dec 20, 2024"}],
						"blueDotWeeklyEvents": [{"value": "2024-12-16", "formattedValue": "Dec 16, 2024"}],
						"alpha": {"value": 1.1, "formattedValue": "1.1"},
						"beta": {"value": 0.9, "formattedValue": "0.9"},
						"pricingStartDate": {"value": "2024-01-01"},
						"pricingEndDate": {"value": "2024-12-31"}
					},
					"intradayStatistics": {
						"yield": {"value": 0.5, "formattedValue": "0.5%"},
						"priceToCashFlowRatio": {"value": 25, "formattedValue": "25"},
						"forwardPriceToEarningsRatio": {"value": 30, "formattedValue": "30"},
						"priceToSalesRatio": {"value": 8, "formattedValue": "8"},
						"priceToEarningsRatio": {"value": 32, "formattedValue": "32"},
						"priceToEarningsVsSP500": {"value": 1.5, "formattedValue": "1.5"},
						"cashFlowPerShareLastYear": {"value": 7, "formattedValue": "7"},
						"isDailyBlueDotEvent": true,
						"isWeeklyBlueDotEvent": false,
						"rsLineNewHigh": true
					}
				},
				"patternInfo": {
					"patterns": [{
						"id": "pattern-1",
						"patternType": "Cup With Handle",
						"periodicity": "DAILY",
						"baseStage": "STAGE_2",
						"baseNumber": 2,
						"baseStatus": "ACTIVE",
						"baseLength": 7,
						"baseDepth": {"value": 18.5, "formattedValue": "18.5%"},
						"baseStartDate": {"value": "2024-10-01"},
						"baseEndDate": {"value": "2024-12-15"},
						"baseBottomDate": {"value": "2024-11-04"},
						"leftSideHighDate": {"value": "2024-10-15"},
						"pivotPrice": {"value": 199.99, "formattedValue": "$199.99"},
						"pivotDate": {"value": "2024-12-16"},
						"pivotPriceDate": {"value": "2024-12-16"},
						"avgVolumeRatePctOnPivot": {"value": 42.3, "formattedValue": "42.3%"},
						"pricePctChangeOnPivot": {"value": 3.1, "formattedValue": "3.1%"}
					}],
					"tightAreas": [{
						"patternID": 123,
						"startDate": {"value": "2024-12-01"},
						"endDate": {"value": "2024-12-05"},
						"length": 5
					}]
				},
				"financials": {
					"epsDueDate": {"value": "2025-01-30"},
					"epsDueDateStatus": "CONFIRMED",
					"epsLastReportedDate": {"value": "2024-10-31"},
					"consensusFinancials": {
						"eps": {
							"growthRate": [{"value": 15.2}],
							"reportedEarnings": [
								{"value": {"value": 1.5, "formattedValue": "1.5"}, "percentChangeYOY": {"value": 12, "formattedValue": "12%"}, "periodOffset": "0", "periodEndDate": {"value": "2024-09-30"}}
							]
						},
						"sales": {
							"growthRate": [{"value": 8.5}],
							"reportedSales": [
								{"value": {"value": 94.9, "formattedValue": "94.9B"}, "percentChangeYOY": {"value": 6, "formattedValue": "6%"}, "periodOffset": "0", "periodEndDate": {"value": "2024-09-30"}}
							]
						}
					},
					"estimates": {
						"epsEstimates": [{"value": {"value": 2.0, "formattedValue": "2.0"}, "percentChangeYOY": {"value": 10, "formattedValue": "10%"}, "periodOffset": "F1", "period": "FY2025", "revisionDirection": "UP"}],
						"salesEstimates": [{"value": {"value": 100, "formattedValue": "100B"}, "percentChangeYOY": {"value": 5, "formattedValue": "5%"}, "periodOffset": "F1", "period": "FY2025", "revisionDirection": "UP"}]
					},
					"profitMarginValues": [{"periodOffset": "0", "periodEndDate": {"value": "2024-09-30"}, "preTaxMargin": {"value": 30}, "afterTaxMargin": {"value": 25}, "grossMargin": {"value": 46}, "returnOnEquity": {"value": 160}}]
				},
				"industry": {
					"name": "Comp-Minicomputers",
					"sector": "Technology",
					"indCode": "G3615",
					"numberOfStocksInGroup": 8,
					"groupRanks": [{"value": 5}],
					"groupRS": [{"value": 95, "letterValue": "A"}]
				},
				"ownership": {
					"fundsFloatPercentHeld": {"value": 60, "formattedValue": "60%"},
					"fundOwnershipSummary": [{"date": {"value": "2024-09-30"}, "numberOfFundsHeld": {"formattedValue": "5000"}}]
				},
				"fundamentals": {
					"researchAndDevelopmentPercentLastQtr": {"value": 7.5, "formattedValue": "7.5%"},
					"debtPercent": {"formattedValue": "100%"},
					"newCEODate": {"value": "2011-08-24"}
				},
				"corporateActions": {
					"dividendNextReportedExDate": {"value": "2025-02-07"}
				},
				"symbology": {
					"company": [{"companyName": "Apple Inc.", "businessDescription": "Consumer electronics", "url": "https://apple.com", "address": "One Apple Park Way", "city": "Cupertino", "country": "US", "stateProvince": "CA"}],
					"instrument": [{"ipoDate": {"value": "1980-12-12"}, "ipoPrice": {"value": 22, "formattedValue": "$22"}, "subType": "COMMON"}]
				}
			}]
		}
	}`
}

// emptyMarketDataFixture returns a response with empty marketData (symbol not found).
func emptyMarketDataFixture() string {
	return `{"data":{"marketData":[]}}`
}

// chartResponseFixture returns a minimal chart API response.
func chartResponseFixture() string {
	return `{"data":{"marketData":[{"pricing":{"timeSeries":{"period":"P1D","dataPoints":[{"startDateTime":"2024-01-01","endDateTime":"2024-01-02","open":{"value":100},"high":{"value":102},"low":{"value":99},"last":{"value":101},"volume":{"value":50000}}]},"quote":{"tradeDateTime":"2024-01-02T16:00:00Z","timeliness":"REALTIME","quoteType":"LAST","last":{"value":101.5,"formattedValue":"101.5"},"volume":{"value":50000,"formattedValue":"50000"},"percentChange":{"value":1.5,"formattedValue":"1.5%"},"netChange":{"value":1.5,"formattedValue":"1.5"}},"currentMarketState":"REGULAR"}}],"exchangeData":[{"city":"New York","countryCode":"US","exchangeISO":"XNYS"}]}}`
}

// chartMarkupsFixture returns a minimal chart markups API response.
func chartMarkupsFixture() string {
	return `{"data":{"user":{"chartMarkups":{"cursorId":"cursor-abc","chartMarkups":[{"id":"m1","name":"My Markup","data":"{}","frequency":"DAILY","site":"marketsurge","createdAt":"2024-01-01","updatedAt":"2024-01-02"}]}}}}`
}
