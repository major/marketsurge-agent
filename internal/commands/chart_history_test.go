package commands

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

func TestChartHistorySuccessWithExplicitDates(t *testing.T) {
	server := jsonServer(chartResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartHistoryCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd,
		"history", "--start-date", "2024-01-01", "--end-date", "2024-06-30", "AAPL",
	))

	parseJSONEnvelope(t, &buf)
}

func TestChartHistorySuccessWithLookback(t *testing.T) {
	server := jsonServer(chartResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartHistoryCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd,
		"history", "--lookback", "3M", "AAPL",
	))

	parseJSONEnvelope(t, &buf)
}

func TestChartHistorySymbolNotFound(t *testing.T) {
	server := jsonServer(emptyMarketDataFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartHistoryCommand(c, &buf)
	err := runTestCommand(t, cmd,
		"history", "--lookback", "1M", "MISSING",
	)
	require.Error(t, err)

	var snf *mserrors.SymbolNotFoundError
	assert.ErrorAs(t, err, &snf)
}

func TestChartHistoryMissingSymbol(t *testing.T) {
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartHistoryCommand(c, &buf)
	err := runTestCommand(t, cmd,
		"history", "--lookback", "1M",
	)
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
}

func TestChartHistoryMutualExclusion(t *testing.T) {
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartHistoryCommand(c, &buf)
	err := runTestCommand(t, cmd,
		"history", "--start-date", "2024-01-01", "--end-date", "2024-06-30", "--lookback", "3M", "AAPL",
	)
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "cannot use both")
}

func TestChartHistoryNeitherDatesNorLookback(t *testing.T) {
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartHistoryCommand(c, &buf)
	err := runTestCommand(t, cmd, "history", "AAPL")
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "either")
}

func TestChartHistoryPartialExplicitDates(t *testing.T) {
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartHistoryCommand(c, &buf)
	err := runTestCommand(t, cmd,
		"history", "--start-date", "2024-01-01", "AAPL",
	)
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "both --start-date and --end-date")
}

func TestChartHistoryInvalidLookback(t *testing.T) {
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartHistoryCommand(c, &buf)
	err := runTestCommand(t, cmd,
		"history", "--lookback", "2W", "AAPL",
	)
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "invalid lookback")
}

func TestResolveLookback(t *testing.T) {
	// Fixed reference date: 2025-06-15
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		lookback string
		expected string
	}{
		{"1W", "2025-06-08"},
		{"1M", "2025-05-15"},
		{"3M", "2025-03-15"},
		{"6M", "2024-12-15"},
		{"1Y", "2024-06-15"},
		{"YTD", "2025-01-01"},
	}

	for _, tt := range tests {
		t.Run(tt.lookback, func(t *testing.T) {
			result := resolveLookback(tt.lookback, now)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapPeriod(t *testing.T) {
	period, daily := mapPeriod("daily")
	assert.Equal(t, "P1D", period)
	assert.True(t, daily)

	period, daily = mapPeriod("weekly")
	assert.Equal(t, "P1W", period)
	assert.False(t, daily)
}
