package cmd

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
)

func TestChartMarkupsSuccess(t *testing.T) {
	t.Parallel()
	server := jsonServer(chartMarkupsFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeChartCmd(t, c, "markups", "AAPL")
	require.NoError(t, err)
	result := parseJSONEnvelope(t, output)
	assertSymbolMeta(t, result, "AAPL")
}

func TestChartMarkupsWithFlags(t *testing.T) {
	t.Parallel()
	server := jsonServer(chartMarkupsFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeChartCmd(t, c, "markups", "--frequency", "WEEKLY", "--sort-dir", "DESC", "AAPL")
	require.NoError(t, err)
	parseJSONEnvelope(t, output)
}

func TestChartMarkupsMissingSymbol(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	_, err := executeChartCmd(t, c, "markups")
	require.Error(t, err)
}

func TestChartHistorySuccessWithExplicitDates(t *testing.T) {
	t.Parallel()
	server := jsonServer(chartResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeChartCmd(t, c,
		"history", "--start-date", "2024-01-01", "--end-date", "2024-06-30", "AAPL")
	require.NoError(t, err)
	result := parseJSONEnvelope(t, output)
	assertSymbolMeta(t, result, "AAPL")
}

func TestChartHistorySuccessWithLookback(t *testing.T) {
	t.Parallel()
	server := jsonServer(chartResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeChartCmd(t, c,
		"history", "--lookback", "3M", "AAPL")
	require.NoError(t, err)
	parseJSONEnvelope(t, output)
}

func TestChartHistorySymbolNotFound(t *testing.T) {
	t.Parallel()
	server := jsonServer(emptyMarketDataFixture())
	defer server.Close()
	c := testClient(t, server)

	_, err := executeChartCmd(t, c,
		"history", "--lookback", "1M", "MISSING")
	require.Error(t, err)

	var snf *mserrors.SymbolNotFoundError
	assert.ErrorAs(t, err, &snf)
}

func TestChartHistoryMissingSymbol(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	_, err := executeChartCmd(t, c,
		"history", "--lookback", "1M")
	require.Error(t, err)
}

func TestChartHistoryMutualExclusion(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	_, err := executeChartCmd(t, c,
		"history", "--start-date", "2024-01-01", "--end-date", "2024-06-30", "--lookback", "3M", "AAPL")
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

	_, err := executeChartCmd(t, c, "history", "AAPL")
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

	_, err := executeChartCmd(t, c,
		"history", "--start-date", "2024-01-01", "AAPL")
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

	_, err := executeChartCmd(t, c,
		"history", "--lookback", "2W", "AAPL")
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
			errs := tt.opts.Validate(context.Background())
			if tt.wantErr != "" {
				require.Len(t, errs, 1)
				var verr *mserrors.ValidationError
				assert.ErrorAs(t, errs[0], &verr)
				assert.Contains(t, errs[0].Error(), tt.wantErr)
			} else {
				assert.Empty(t, errs)
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

func TestChartMarkupsStructTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		field    string
		tag      string
		value    string
		wantType reflect.Type
	}{
		{"Frequency", "flag", "frequency", reflect.TypeOf(models.Frequency(""))},
		{"Frequency", "default", "DAILY", reflect.TypeOf(models.Frequency(""))},
		{"SortDir", "flag", "sort-dir", reflect.TypeOf(models.SortDirection(""))},
		{"SortDir", "default", "ASC", reflect.TypeOf(models.SortDirection(""))},
	}

	rt := reflect.TypeOf(ChartMarkupsOptions{})
	for _, tt := range tests {
		t.Run(tt.field+"/"+tt.tag, func(t *testing.T) {
			t.Parallel()
			f, ok := rt.FieldByName(tt.field)
			require.True(t, ok, "field %s not found", tt.field)
			assert.Equal(t, tt.value, f.Tag.Get(tt.tag), "tag %q on field %s", tt.tag, tt.field)
			if tt.tag == "flag" {
				assert.Equal(t, tt.wantType, f.Type, "field %s type mismatch", tt.field)
			}
		})
	}
}

func TestChartHistoryStructTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		field    string
		tag      string
		value    string
		wantType reflect.Type
	}{
		{"StartDate", "flag", "start-date", reflect.TypeOf("")},
		{"EndDate", "flag", "end-date", reflect.TypeOf("")},
		{"Lookback", "flag", "lookback", reflect.TypeOf("")},
		{"Period", "flag", "period", reflect.TypeOf(models.Period(""))},
		{"Period", "default", "daily", reflect.TypeOf(models.Period(""))},
		{"Benchmark", "flag", "benchmark", reflect.TypeOf("")},
	}

	rt := reflect.TypeOf(ChartHistoryOptions{})
	for _, tt := range tests {
		t.Run(tt.field+"/"+tt.tag, func(t *testing.T) {
			t.Parallel()
			f, ok := rt.FieldByName(tt.field)
			require.True(t, ok, "field %s not found", tt.field)
			assert.Equal(t, tt.value, f.Tag.Get(tt.tag), "tag %q on field %s", tt.tag, tt.field)
			if tt.tag == "flag" {
				assert.Equal(t, tt.wantType, f.Type, "field %s type mismatch", tt.field)
			}
		})
	}
}

// executeChartCmd creates a chart command tree, injects the client into subcommand
// contexts, and executes with the given args. structcli.Bind sets a scope context on
// subcommands, which prevents cobra's parent-to-child context propagation. This helper
// layers the client onto each subcommand's existing context so both the structcli scope
// and the test client are available during RunE.
func executeChartCmd(t *testing.T, c *client.Client, args ...string) (string, error) {
	t.Helper()
	cmd := newChartCmd()
	ctx := ContextWithClient(context.Background(), c)
	cmd.SetContext(ctx)
	for _, child := range cmd.Commands() {
		childCtx := child.Context()
		if childCtx == nil {
			childCtx = ctx
		}
		child.SetContext(ContextWithClient(childCtx, c))
	}
	return executeCommand(t, cmd, args...)
}
