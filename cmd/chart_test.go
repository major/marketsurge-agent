package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

func TestChartMarkupsSuccess(t *testing.T) {
	t.Parallel()
	server := jsonServer(chartMarkupsFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeCommandWithClient(newChartCmd(), c, "markups", "AAPL")
	require.NoError(t, err)
	result := parseJSONEnvelope(t, output)
	assertSymbolMeta(t, result, "AAPL")
}

func TestChartMarkupsWithFlags(t *testing.T) {
	t.Parallel()
	server := jsonServer(chartMarkupsFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeCommandWithClient(newChartCmd(), c, "markups", "--frequency", "WEEKLY", "--sort-dir", "DESC", "AAPL")
	require.NoError(t, err)
	parseJSONEnvelope(t, output)
}

func TestChartMarkupsMissingSymbol(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	_, err := executeCommandWithClient(newChartCmd(), c, "markups")
	require.Error(t, err)
}

func TestChartHistorySuccessWithExplicitDates(t *testing.T) {
	t.Parallel()
	server := jsonServer(chartResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeCommandWithClient(newChartCmd(), c,
		"history", "--start-date", "2024-01-01", "--end-date", "2024-06-30", "AAPL",
	)
	require.NoError(t, err)
	result := parseJSONEnvelope(t, output)
	assertSymbolMeta(t, result, "AAPL")
}

func TestChartHistorySuccessWithLookback(t *testing.T) {
	t.Parallel()
	server := jsonServer(chartResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeCommandWithClient(newChartCmd(), c,
		"history", "--lookback", "3M", "AAPL",
	)
	require.NoError(t, err)
	parseJSONEnvelope(t, output)
}

func TestChartHistorySymbolNotFound(t *testing.T) {
	t.Parallel()
	server := jsonServer(emptyMarketDataFixture())
	defer server.Close()
	c := testClient(t, server)

	_, err := executeCommandWithClient(newChartCmd(), c,
		"history", "--lookback", "1M", "MISSING",
	)
	require.Error(t, err)

	var snf *mserrors.SymbolNotFoundError
	assert.ErrorAs(t, err, &snf)
}

func TestChartHistoryMissingSymbol(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	_, err := executeCommandWithClient(newChartCmd(), c,
		"history", "--lookback", "1M",
	)
	require.Error(t, err)
}

func TestChartHistoryMutualExclusion(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	_, err := executeCommandWithClient(newChartCmd(), c,
		"history", "--start-date", "2024-01-01", "--end-date", "2024-06-30", "--lookback", "3M", "AAPL",
	)
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "cannot use both")
}

func TestChartHistoryNeitherDatesNorLookback(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	_, err := executeCommandWithClient(newChartCmd(), c, "history", "AAPL")
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "either")
}

func TestChartHistoryPartialExplicitDates(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	_, err := executeCommandWithClient(newChartCmd(), c,
		"history", "--start-date", "2024-01-01", "AAPL",
	)
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "both --start-date and --end-date")
}

func TestChartHistoryInvalidLookback(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	_, err := executeCommandWithClient(newChartCmd(), c,
		"history", "--lookback", "2W", "AAPL",
	)
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "invalid lookback")
}

func TestResolveLookback(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			result := resolveLookback(tt.lookback, now)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapPeriod(t *testing.T) {
	t.Parallel()
	period, daily := mapPeriod("daily")
	assert.Equal(t, "P1D", period)
	assert.True(t, daily)

	period, daily = mapPeriod("weekly")
	assert.Equal(t, "P1W", period)
	assert.False(t, daily)
}

func TestChartHistoryOptionsValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    ChartHistoryOptions
		wantErr string
	}{
		{
			name:    "both explicit dates and lookback",
			opts:    ChartHistoryOptions{StartDate: "2024-01-01", EndDate: "2024-06-30", Lookback: "3M"},
			wantErr: "cannot use both",
		},
		{
			name:    "neither dates nor lookback",
			opts:    ChartHistoryOptions{},
			wantErr: "either",
		},
		{
			name:    "only start-date without end-date",
			opts:    ChartHistoryOptions{StartDate: "2024-01-01"},
			wantErr: "both --start-date and --end-date",
		},
		{
			name: "valid explicit dates",
			opts: ChartHistoryOptions{StartDate: "2024-01-01", EndDate: "2024-06-30"},
		},
		{
			name: "valid lookback",
			opts: ChartHistoryOptions{Lookback: "3M"},
		},
		{
			name:    "invalid lookback value",
			opts:    ChartHistoryOptions{Lookback: "2W"},
			wantErr: "invalid lookback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.opts.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				var verr *mserrors.ValidationError
				assert.ErrorAs(t, err, &verr)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChartHistoryOptionsResolveDates(t *testing.T) {
	t.Parallel()
	// Fixed reference date: 2025-06-15
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	t.Run("explicit dates pass through", func(t *testing.T) {
		t.Parallel()
		opts := ChartHistoryOptions{StartDate: "2024-01-01", EndDate: "2024-06-30"}
		start, end := opts.ResolveDates(now)
		assert.Equal(t, "2024-01-01", start)
		assert.Equal(t, "2024-06-30", end)
	})

	t.Run("lookback computes dates", func(t *testing.T) {
		t.Parallel()
		opts := ChartHistoryOptions{Lookback: "3M"}
		start, end := opts.ResolveDates(now)
		assert.Equal(t, "2025-03-15", start)
		assert.Equal(t, "2025-06-15", end)
	})
}
