package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetChartHistoryDailyIncludesExchangeAndBenchmark(t *testing.T) {
	t.Parallel()
	var captured Request
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_, _ = w.Write([]byte(chartResponseJSON(true)))
	})

	data, err := client.GetChartHistory(context.Background(), "AAPL", "2024-01-01", "2024-02-01", "P1D", true, "NYSE", "0S&P5")
	require.NoError(t, err)
	assert.Equal(t, "NYSE", captured.Variables["exchangeName"])
	assert.Len(t, captured.Variables["symbols"], 2)
	assert.NotNil(t, data.Exchange)
	assert.NotNil(t, data.BenchmarkTimeSeries)
	assert.Equal(t, 1, len(data.TimeSeries.DataPoints))
}

func TestGetChartHistoryWeeklyOmitsExchange(t *testing.T) {
	t.Parallel()
	var captured Request
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_, _ = w.Write([]byte(chartResponseJSON(false)))
	})

	data, err := client.GetChartHistory(context.Background(), "AAPL", "2024-01-01", "2024-02-01", "P1W", false, "", "")
	require.NoError(t, err)
	assert.Nil(t, captured.Variables["exchangeName"])
	assert.Nil(t, data.BenchmarkTimeSeries)
	assert.Equal(t, "REGULAR", *data.CurrentMarketState)
	assert.Equal(t, 101.5, *data.Quote.Last)
}

func TestGetChartMarkupsSuccess(t *testing.T) {
	t.Parallel()
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"user":{"chartMarkups":{"cursorId":"cursor-1","chartMarkups":[{"id":"m1","name":"Base","data":"{}","frequency":"DAILY","site":"marketsurge"}]}}}}`))
	})

	data, err := client.GetChartMarkups(context.Background(), "DJ:1", "DAILY", "ASC")
	require.NoError(t, err)
	assert.Equal(t, "cursor-1", data.CursorID)
	assert.Len(t, data.Markups, 1)
	assert.Equal(t, "m1", data.Markups[0].ID)
}

func TestGetChartMarkupsPassesOperationName(t *testing.T) {
	t.Parallel()
	var captured Request
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_, _ = w.Write([]byte(`{"data":{"user":{"chartMarkups":{"cursorId":"","chartMarkups":[]}}}}`))
	})

	_, err := client.GetChartMarkups(context.Background(), "DJ:1", "WEEKLY", "DESC")
	require.NoError(t, err)
	assert.Equal(t, "FetchChartMarkups", captured.OperationName)
	assert.Equal(t, "WEEKLY", captured.Variables["frequency"])
	assert.Equal(t, "DESC", captured.Variables["sortDir"])
}

func TestGetChartHistoryReturnsSymbolNotFound(t *testing.T) {
	t.Parallel()
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"marketData":[]}}`))
	})

	_, err := client.GetChartHistory(context.Background(), "MISS", "2024-01-01", "2024-02-01", "P1D", true, "NYSE", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "symbol not found")
}

func chartResponseJSON(includeBenchmark bool) string {
	benchmark := ""
	if includeBenchmark {
		benchmark = `,{"pricing":{"timeSeries":{"period":"P1D","dataPoints":[{"startDateTime":"2024-01-01","endDateTime":"2024-01-02","open":{"value":200},"high":{"value":201},"low":{"value":199},"last":{"value":200.5},"volume":{"value":900}}]}}}`
	}
	return `{"data":{"marketData":[{"pricing":{"timeSeries":{"period":"P1D","dataPoints":[{"startDateTime":"2024-01-01","endDateTime":"2024-01-02","open":{"value":100},"high":{"value":102},"low":{"value":99},"last":{"value":101},"volume":{"value":1000}}]},"quote":{"tradeDateTime":"2024-01-02T10:00:00Z","timeliness":"REALTIME","quoteType":"LAST","last":{"value":101.5,"formattedValue":"101.5"},"volume":{"value":1000,"formattedValue":"1000"},"percentChange":{"value":1.1,"formattedValue":"1.1%"},"netChange":{"value":1.0,"formattedValue":"1.0"}},"currentMarketState":"REGULAR"}}` + benchmark + `],"exchangeData":[{"city":"New York","countryCode":"US","exchangeISO":"XNYS","holidays":[{"name":"Holiday","startDateTime":"2024-01-01","endDateTime":"2024-01-01"}]}]}}`
}
