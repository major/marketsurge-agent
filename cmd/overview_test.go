package cmd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/major/marketsurge-go/marketsurge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentcmd "github.com/major/marketsurge-agent/cmd"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

func TestOverviewSuccess(t *testing.T) {
	seenOperations := make(map[string]bool)
	var requestedSymbols []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := overviewRequest(t, r)
		seenOperations[req.OperationName] = true
		requestedSymbols = append(requestedSymbols, req.Variables.Symbols...)

		w.Header().Set("Content-Type", "application/json")
		switch req.OperationName {
		case "OtherMarketData":
			fmt.Fprint(w, `{"data":{"marketData":[{"id":"208144392","ratings":{"compRating":[{"value":96}],"epsRating":[{"value":83}],"rsRating":[{"value":91}],"smrRating":[{"letterValue":"A"}],"adRating":[{"letterValue":"B+"}]},"pricingStatistics":{"endOfDayStatistics":{"historicalPriceStatistics":[{"pricePercentChange":{"value":12.5,"formattedValue":"12.5%"}}],"avgDollarVolume50Day":{"value":2500000000,"formattedValue":"2.5B"},"marketCapitalization":{"value":300000000000,"formattedValue":"300B"},"averageTrueRangePercent":[{"value":4.1,"formattedValue":"4.1%"}],"antEvents":[{"value":"2026-05-01"},{"value":"2026-05-08"}],"upDownVolumeRatio":{"value":1.4,"formattedValue":"1.4"},"blueDotDailyEvents":[{"value":"2026-05-01"}],"blueDotWeeklyEvents":[{"formattedValue":"2026-04-24"}],"alpha":{"value":1.2,"formattedValue":"1.2"},"beta":{"value":1.4,"formattedValue":"1.4"},"shortInterest":{"daysToCover":{"value":1.5,"formattedValue":"1.5"},"daysToCoverPercentChange":{"value":-2.4,"formattedValue":"-2.4%"},"percentOfFloat":{"value":0.031,"formattedValue":"3.1%"},"volume":{"value":50000000,"formattedValue":"50M"}}},"intradayStatistics":{"isDailyBlueDotEvent":true,"isWeeklyBlueDotEvent":false}},"symbology":{"company":{"companyName":"Advanced Micro Devices"},"instrument":{"subType":"COMMON_STOCK"}},"patternInfo":{"patterns":[{"patternType":"CUP_WITH_HANDLE","baseStatus":"ACTIVE","baseStage":"1","baseLength":9,"baseDepth":{"value":18.2,"formattedValue":"18.2%"},"baseStartDate":{"value":"2026-03-01"},"baseEndDate":{"value":"2026-05-01"},"pivotPrice":{"value":187.55,"formattedValue":"187.55"},"pivotDate":{"value":"2026-05-02"},"handleDepth":{"value":7.5,"formattedValue":"7.5%"},"handleLength":3}],"tightAreas":[{"patternID":7,"startDate":{"value":"2026-04-22"},"endDate":{"value":"2026-04-26"},"length":5}]},"industry":{"name":"Electronics-Semiconductor Fabless","sector":"Technology","indCode":"123","numberOfStocksInGroup":42,"groupRanks":[{"value":5,"period":"P1M","periodOffset":"CURRENT"}],"groupRS":[{"value":92,"letterValue":"A","period":"P6M","periodOffset":"CURRENT"}]},"ownership":{"fundsFloatPercentHeld":{"value":49.5,"formattedValue":"49.5%"}},"fundamentals":{"debtPercent":{"value":20.1,"formattedValue":"20.1%"},"researchAndDevelopmentPercentLastQtr":{"value":24.2,"formattedValue":"24.2%"},"newCEODate":{"value":"2014-10-08"}}}]}}`)
		case "RSRatingRIPanel":
			fmt.Fprint(w, `{"data":{"marketData":[{"ratings":{"rsRating":[{"letterValue":"A","period":"P12M","periodOffset":"CURRENT","value":91},{"letterValue":"B","period":"P3M","periodOffset":"CURRENT","value":84}]},"pricingStatistics":{"intradayStatistics":{"rsLineNewHigh":true}}}]}}`)
		case "Ownership":
			fmt.Fprint(w, `{"data":{"marketData":[{"ownership":{"fundsFloatPercentHeld":{"formattedValue":"49.5%"},"fundOwnershipSummary":[{"date":{"value":"2026-03-31"},"numberOfFundsHeld":{"formattedValue":"3,210"}},{"date":{"value":"2025-12-31"},"numberOfFundsHeld":{"formattedValue":"3,100"}}]}}]}}`)
		case "FundermentalDataBox":
			fmt.Fprint(w, `{"data":{"marketData":[{"financials":{"consensusFinancials":{"eps":{"reportedEarnings":[{"value":{"value":1.65,"formattedValue":"$1.65"},"percentChangeYOY":{"value":12.5,"formattedValue":"+12.5%"},"periodOffset":"CURRENT","periodEndDate":{"value":"2026-03-31"}}]},"sales":{"reportedSales":[{"value":{"value":95200000000,"formattedValue":"$95.2B"},"percentChangeYOY":{"value":8.2,"formattedValue":"+8.2%"},"periodOffset":"CURRENT","periodEndDate":{"value":"2026-03-31"}}]}},"estimates":{"epsEstimates":[{"value":{"value":1.72,"formattedValue":"$1.72"},"percentChangeYOY":{"value":10.5,"formattedValue":"+10.5%"},"periodOffset":"P1Q_FUTURE","period":"P1Q","revisionDirection":"UP"}],"salesEstimates":[{"value":{"value":98500000000,"formattedValue":"$98.5B"},"percentChangeYOY":{"value":6.8,"formattedValue":"+6.8%"},"periodOffset":"P1Q_FUTURE","period":"P1Q"}]}}}]}}`)
		case "ChartMarketData":
			fmt.Fprint(w, `{"data":{"marketData":[{"pricing":{"timeSeries":{"period":"ONE_WEEK","dataPoints":[{"startDateTime":"2026-05-04T13:30:00.000Z","endDateTime":"2026-05-08T20:00:00.000Z","volume":{"value":185420300},"last":{"value":212.30},"low":{"value":205.50},"high":{"value":213.75},"open":{"value":206.20}}]},"quote":{"tradeDateTime":"2026-05-08T20:00:00.000Z","timeliness":"REAL_TIME","quoteType":"REGULAR","volume":{"value":35210400,"formattedValue":"35,210,400"},"percentChange":{"value":0.88,"formattedValue":"+0.88%"},"netChange":{"value":1.85,"formattedValue":"+1.85"},"last":{"value":212.30,"formattedValue":"212.30"}}}}]}}`)
		default:
			t.Fatalf("OverviewCmd.Run operationName = %q, want known operation", req.OperationName)
		}
	}))
	t.Cleanup(server.Close)

	client := overviewClient(t, server)
	output, err := runOverview(t, client, agentcmd.OverviewCmd{Symbol: " amd "})
	require.NoError(t, err, "OverviewCmd.Run(success) error = %v, want nil", err)

	var rows []map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &rows), "json.Unmarshal(OverviewCmd.Run(success) output) error = %v, want nil", err)
	require.Len(t, rows, 1, "OverviewCmd.Run(success) decoded rows length = %d, want %d", len(rows), 1)
	assert.Equal(t, []string{"AMD", "AMD", "AMD", "AMD", "AMD"}, requestedSymbols)
	assert.True(t, seenOperations["OtherMarketData"], "OverviewCmd.Run(success) called OtherMarketData = false, want true")
	assert.True(t, seenOperations["RSRatingRIPanel"], "OverviewCmd.Run(success) called RSRatingRIPanel = false, want true")
	assert.True(t, seenOperations["Ownership"], "OverviewCmd.Run(success) called Ownership = false, want true")
	assert.True(t, seenOperations["FundermentalDataBox"], "OverviewCmd.Run(success) called FundermentalDataBox = false, want true")
	assert.True(t, seenOperations["ChartMarketData"], "OverviewCmd.Run(success) called ChartMarketData = false, want true")

	row := rows[0]
	assert.Equal(t, "AMD", row["ticker"])
	assertNoOverviewInternalIDs(t, row)
	assert.Equal(t, "Advanced Micro Devices", row["name"])
	assert.Equal(t, "COMMON_STOCK", row["type"])
	ratings := row["ratings"].(map[string]any)
	assert.Equal(t, float64(96), ratings["composite"])
	assert.Equal(t, float64(83), ratings["epsRating"])
	assert.Equal(t, float64(91), ratings["relativeStrength"])
	assert.Equal(t, "A", ratings["salesMarginsROE"])
	assert.Equal(t, "B+", ratings["accumulationDistribution"])
	price := row["price"].(map[string]any)
	assert.Equal(t, float64(300000000000), price["marketCap"].(map[string]any)["v"])
	assert.Equal(t, float64(2500000000), price["avgDollarVolume50d"].(map[string]any)["v"])
	assert.Equal(t, float64(4.1), price["atrPercent"].(map[string]any)["v"])
	assert.Equal(t, float64(1.4), price["upDownVolumeRatio"].(map[string]any)["v"])
	assert.Equal(t, float64(12.5), price["quarterPercentChange"].(map[string]any)["v"])
	assert.Equal(t, true, price["blueDotDaily"])
	assert.Equal(t, false, price["blueDotWeekly"])
	assert.Equal(t, []any{"2026-05-01"}, price["blueDotDailyEvents"])
	assert.Equal(t, []any{"2026-04-24"}, price["blueDotWeeklyEvents"])
	assert.Equal(t, float64(2), row["ants"].(map[string]any)["count"])
	assert.Equal(t, "Electronics-Semiconductor Fabless", row["industry"].(map[string]any)["name"])
	relativeStrengthTrend := row["relativeStrengthTrend"].(map[string]any)
	assert.Equal(t, true, relativeStrengthTrend["newHigh"])
	assert.Len(t, relativeStrengthTrend["history"], 2)
	assert.Len(t, row["patterns"], 1)
	assert.Equal(t, float64(9), row["patterns"].([]any)[0].(map[string]any)["length"])
	assert.Len(t, row["tightAreas"], 1)
	assert.Equal(t, float64(5), row["tightAreas"].([]any)[0].(map[string]any)["length"])
	assert.NotContains(t, row["tightAreas"].([]any)[0].(map[string]any), "patternId")
	assert.Len(t, row["ownership"].(map[string]any)["funds"], 2)
	fundamentals := row["fundamentals"].(map[string]any)
	assert.Equal(t, "2014-10-08", fundamentals["ceoDate"])
	assert.Equal(t, float64(1.65), fundamentals["reportedEPS"].(map[string]any)["value"].(map[string]any)["v"])
	assert.Equal(t, "UP", fundamentals["estimatedEPS"].(map[string]any)["revisionDirection"])
	risk := row["risk"].(map[string]any)
	assert.Equal(t, float64(1.2), risk["alpha"].(map[string]any)["v"])
	assert.Equal(t, float64(0.031), risk["shortInterest"].(map[string]any)["percentOfFloat"].(map[string]any)["v"])
	weeklyTrend := row["weeklyTrend"].(map[string]any)
	assert.Equal(t, "ONE_WEEK", weeklyTrend["period"])
	assert.Equal(t, float64(212.30), weeklyTrend["latest"].(map[string]any)["last"].(map[string]any)["v"])
	assert.Equal(t, "REAL_TIME", weeklyTrend["quote"].(map[string]any)["timeliness"])
}

func assertNoOverviewInternalIDs(t *testing.T, value any) {
	t.Helper()

	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			assert.NotContains(t, []string{"sym", "id", "marketSurgeID", "requestedSymbol", "patternId"}, key)
			assertNoOverviewInternalIDs(t, nested)
		}
	case []any:
		for _, nested := range typed {
			assertNoOverviewInternalIDs(t, nested)
		}
	}
}

func TestOverviewSymbolNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"marketData":[]}}`)
	}))
	t.Cleanup(server.Close)

	client := overviewClient(t, server)
	output, err := runOverview(t, client, agentcmd.OverviewCmd{Symbol: "ZZZZ"})
	require.Error(t, err, "OverviewCmd.Run(symbol not found) error = nil, want non-nil")

	var symbolErr *mserrors.SymbolNotFoundError
	require.ErrorAs(t, err, &symbolErr, "OverviewCmd.Run(symbol not found) error type = %T, want *mserrors.SymbolNotFoundError", err)
	assert.Empty(t, output, "OverviewCmd.Run(symbol not found) stdout = %q, want empty", output)
}

func TestOverviewValidationError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)

	client := overviewClient(t, server)
	output, err := runOverview(t, client, agentcmd.OverviewCmd{Symbol: " "})
	require.Error(t, err, "OverviewCmd.Run(blank symbol) error = nil, want non-nil")

	var validationErr *mserrors.ValidationError
	require.ErrorAs(t, err, &validationErr, "OverviewCmd.Run(blank symbol) error type = %T, want *mserrors.ValidationError", err)
	assert.Empty(t, output, "OverviewCmd.Run(blank symbol) stdout = %q, want empty", output)
}

func TestOverviewAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	client := overviewClient(t, server)
	output, err := runOverview(t, client, agentcmd.OverviewCmd{Symbol: "AMD"})
	require.Error(t, err, "OverviewCmd.Run(auth error) error = nil, want non-nil")

	var authErr *mserrors.AuthenticationError
	require.ErrorAs(t, err, &authErr, "OverviewCmd.Run(auth error) error type = %T, want *mserrors.AuthenticationError", err)
	assert.Empty(t, output, "OverviewCmd.Run(auth error) stdout = %q, want empty", output)
}

type overviewGraphQLRequest struct {
	OperationName string `json:"operationName"`
	Variables     struct {
		Symbols []string `json:"symbols"`
	} `json:"variables"`
}

func overviewRequest(t *testing.T, r *http.Request) overviewGraphQLRequest {
	t.Helper()

	body, err := io.ReadAll(r.Body)
	require.NoError(t, err, "io.ReadAll(OverviewCmd.Run request body) error = %v, want nil", err)

	var req overviewGraphQLRequest
	require.NoError(t, json.Unmarshal(body, &req), "json.Unmarshal(OverviewCmd.Run request body) error = %v, want nil", err)
	return req
}

func overviewClient(t *testing.T, server *httptest.Server) *marketsurge.Client {
	t.Helper()

	client, err := marketsurge.NewClient(
		marketsurge.WithJWT("test-token"),
		marketsurge.WithHTTPClient(server.Client()),
		marketsurge.WithGraphQLURL(server.URL),
	)
	require.NoError(t, err, "marketsurge.NewClient(test server %q) error = %v, want nil", server.URL, err)
	return client
}

func runOverview(t *testing.T, client *marketsurge.Client, cmd agentcmd.OverviewCmd) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "os.Pipe() error = %v, want nil", err)
	t.Cleanup(func() {
		_ = r.Close()
	})

	os.Stdout = w
	runErr := cmd.Run(client)
	closeErr := w.Close()
	os.Stdout = oldStdout
	require.NoError(t, closeErr, "stdout pipe Close() error = %v, want nil", closeErr)

	var output bytes.Buffer
	_, err = io.Copy(&output, r)
	require.NoError(t, err, "io.Copy(OverviewCmd.Run stdout) error = %v, want nil", err)
	return output.String(), runErr
}
