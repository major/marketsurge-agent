package cmd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/major/marketsurge-go/marketsurge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentcmd "github.com/major/marketsurge-agent/cmd"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

const chartSuccessResponse = `{"data":{"marketData":[{"id":"1","originRequest":{"fromDialect":"CHARTING","symbol":"AAPL"},"pricing":{"timeSeries":{"period":"ONE_DAY","dataPoints":[{"startDateTime":"2025-05-01T13:30:00.000Z","endDateTime":"2025-05-01T20:00:00.000Z","volume":{"value":45000000},"last":{"value":210.45},"low":{"value":208.10},"high":{"value":211.90},"open":{"value":209.50}},{"startDateTime":"2025-05-02T13:30:00.000Z","endDateTime":"2025-05-02T20:00:00.000Z","volume":{"value":38000000},"last":{"value":212.30},"low":{"value":209.80},"high":{"value":213.50},"open":{"value":210.00}}]},"quote":{"tradeDateTime":"2025-05-02T20:00:00.000Z","timeliness":"DELAYED","quoteType":"REGULAR","volume":{"value":38000000,"formattedValue":"38M"},"percentChange":{"value":0.88,"formattedValue":"0.88%"},"netChange":{"value":1.85,"formattedValue":"1.85"},"last":{"value":212.30,"formattedValue":"212.30"}},"premarketQuote":{"tradeDateTime":"2025-05-02T13:00:00.000Z","quoteType":"PREMARKET","last":{"value":211.00,"formattedValue":"211.00"}},"postmarketQuote":{"tradeDateTime":"2025-05-02T22:00:00.000Z","quoteType":"POSTMARKET","last":{"value":212.50,"formattedValue":"212.50"}},"currentMarketState":"REGULAR_MARKET"}}],"exchangeData":{"city":"New York","countryCode":"US","exchangeISO":"XNYS","id":"1"}}}`

func TestChartSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, chartSuccessResponse)
	}))
	t.Cleanup(server.Close)

	client := chartClient(t, server)
	output, err := runChart(t, client, agentcmd.ChartCmd{Symbol: "AAPL", Days: 90, Exchange: "XNYS"})
	require.NoError(t, err, "ChartCmd.Run(success) error = %v, want nil", err)

	var items []struct {
		Ticker             string `json:"ticker"`
		Days               int    `json:"days"`
		Exchange           string `json:"exchange"`
		CurrentMarketState string `json:"currentMarketState"`
		DataPoints         []struct {
			Date   string   `json:"date"`
			Open   *float64 `json:"open"`
			High   *float64 `json:"high"`
			Low    *float64 `json:"low"`
			Close  *float64 `json:"close"`
			Volume *float64 `json:"volume"`
		} `json:"dataPoints"`
		Quote *struct {
			Last *float64 `json:"last"`
		} `json:"quote"`
		PremarketQuote *struct {
			Last *float64 `json:"last"`
		} `json:"premarketQuote"`
		PostmarketQuote *struct {
			Last *float64 `json:"last"`
		} `json:"postmarketQuote"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items))
	require.Len(t, items, 1)

	item := items[0]
	assert.Equal(t, "AAPL", item.Ticker)
	assert.Equal(t, 2, item.Days)
	assert.Equal(t, "XNYS", item.Exchange)
	assert.Equal(t, "REGULAR_MARKET", item.CurrentMarketState)

	require.Len(t, item.DataPoints, 2)
	assert.Equal(t, "2025-05-01T13:30:00.000Z", item.DataPoints[0].Date)
	require.NotNil(t, item.DataPoints[0].Close)
	assert.InDelta(t, 210.45, *item.DataPoints[0].Close, 0.001)
	require.NotNil(t, item.DataPoints[0].Open)
	assert.InDelta(t, 209.50, *item.DataPoints[0].Open, 0.001)

	require.NotNil(t, item.Quote)
	require.NotNil(t, item.Quote.Last)
	assert.InDelta(t, 212.30, *item.Quote.Last, 0.001)

	require.NotNil(t, item.PremarketQuote)
	require.NotNil(t, item.PremarketQuote.Last)
	assert.InDelta(t, 211.00, *item.PremarketQuote.Last, 0.001)

	require.NotNil(t, item.PostmarketQuote)
	require.NotNil(t, item.PostmarketQuote.Last)
	assert.InDelta(t, 212.50, *item.PostmarketQuote.Last, 0.001)
}

func TestChartEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"marketData":[]}}`)
	}))
	t.Cleanup(server.Close)

	client := chartClient(t, server)
	output, err := runChart(t, client, agentcmd.ChartCmd{Symbol: "AAPL", Days: 90, Exchange: "XNYS"})
	require.NoError(t, err, "ChartCmd.Run(empty response) error = %v, want nil", err)

	var items []struct {
		Ticker     string `json:"ticker"`
		Days       int    `json:"days"`
		Exchange   string `json:"exchange"`
		DataPoints []any  `json:"dataPoints"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items))
	require.Len(t, items, 1)
	assert.Equal(t, "AAPL", items[0].Ticker)
	assert.Equal(t, 0, items[0].Days)
	assert.Equal(t, "XNYS", items[0].Exchange)
	assert.Empty(t, items[0].DataPoints)
}

func TestChartDefaultsExchange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"marketData":[]}}`)
	}))
	t.Cleanup(server.Close)

	client := chartClient(t, server)
	output, err := runChart(t, client, agentcmd.ChartCmd{Symbol: "AAPL", Days: 90})
	require.NoError(t, err, "ChartCmd.Run(default exchange) error = %v, want nil", err)

	var items []struct {
		Exchange string `json:"exchange"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items))
	require.Len(t, items, 1)
	assert.Equal(t, "XNYS", items[0].Exchange)
}

func TestChartEmptySymbol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("ChartCmd.Run(empty symbol) sent unexpected HTTP request")
	}))
	t.Cleanup(server.Close)

	client := chartClient(t, server)
	_, err := runChart(t, client, agentcmd.ChartCmd{Symbol: "  ", Days: 90, Exchange: "XNYS"})
	require.Error(t, err, "ChartCmd.Run(empty symbol) error = nil, want non-nil")

	var validationErr *mserrors.ValidationError
	require.ErrorAs(t, err, &validationErr, "ChartCmd.Run(empty symbol) error type = %T, want *mserrors.ValidationError", err)
}

func TestChartInvalidDays(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("ChartCmd.Run(invalid days) sent unexpected HTTP request")
	}))
	t.Cleanup(server.Close)

	client := chartClient(t, server)
	_, err := runChart(t, client, agentcmd.ChartCmd{Symbol: "AAPL", Days: 0, Exchange: "XNYS"})
	require.Error(t, err, "ChartCmd.Run(invalid days) error = nil, want non-nil")

	var validationErr *mserrors.ValidationError
	require.ErrorAs(t, err, &validationErr, "ChartCmd.Run(invalid days) error type = %T, want *mserrors.ValidationError", err)
}

func TestChartAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	client := chartClient(t, server)
	output, err := runChart(t, client, agentcmd.ChartCmd{Symbol: "AAPL", Days: 90, Exchange: "XNYS"})
	require.Error(t, err, "ChartCmd.Run(auth error) error = nil, want non-nil")

	var authErr *mserrors.AuthenticationError
	require.ErrorAs(t, err, &authErr, "ChartCmd.Run(auth error) error type = %T, want *mserrors.AuthenticationError", err)
	assert.Empty(t, output, "ChartCmd.Run(auth error) stdout = %q, want empty", output)
}

func TestChartAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"errors":[{"message":"Service unavailable","path":["marketData"]}]}`)
	}))
	t.Cleanup(server.Close)

	client := chartClient(t, server)
	output, err := runChart(t, client, agentcmd.ChartCmd{Symbol: "AAPL", Days: 90, Exchange: "XNYS"})
	require.Error(t, err, "ChartCmd.Run(API error) error = nil, want non-nil")

	var apiErr *mserrors.APIError
	require.ErrorAs(t, err, &apiErr, "ChartCmd.Run(API error) error type = %T, want *mserrors.APIError", err)
	assert.Empty(t, output, "ChartCmd.Run(API error) stdout = %q, want empty", output)
}

func chartClient(t *testing.T, server *httptest.Server) *marketsurge.Client {
	t.Helper()

	client, err := marketsurge.NewClient(
		marketsurge.WithJWT("test-token"),
		marketsurge.WithHTTPClient(server.Client()),
		marketsurge.WithGraphQLURL(server.URL),
	)
	require.NoError(t, err, "marketsurge.NewClient(test server %q) error = %v, want nil", server.URL, err)
	return client
}

func runChart(t *testing.T, client *marketsurge.Client, cmd agentcmd.ChartCmd) (string, error) {
	t.Helper()

	var buf bytes.Buffer
	runErr := cmd.RunForTest(client, &buf)

	return buf.String(), runErr
}
