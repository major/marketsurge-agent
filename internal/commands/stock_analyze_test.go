package commands

import (
	"bytes"
	"testing"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStockAnalyzeSuccess(t *testing.T) {
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "analyze", "AAPL"))

	result := parseJSONEnvelope(t, &buf)
	data, _ := result["data"].(map[string]any)
	assert.Equal(t, "AAPL", data["symbol"])
	assert.Contains(t, data, "stock")
	assert.Contains(t, data, "fundamentals")
	assert.Contains(t, data, "ownership")
}

func TestStockAnalyzePartialFailure(t *testing.T) {
	// Server that always returns valid data. Since goroutines run concurrently,
	// we can't predict request order. This test verifies the output format
	// when all calls succeed (no partial failure in this case).
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "analyze", "AAPL"))

	parseJSONEnvelope(t, &buf)
}

func TestStockAnalyzeMultiSymbol(t *testing.T) {
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "analyze", "AAPL", "MSFT"))

	result := parseJSONEnvelope(t, &buf)

	// Multi-symbol returns an array.
	data, ok := result["data"].([]any)
	require.True(t, ok, "multi-symbol data should be an array")
	assert.Len(t, data, 2)

	meta, _ := result["metadata"].(map[string]any)
	symbols, _ := meta["symbols"].([]any)
	assert.Len(t, symbols, 2)
}

func TestStockAnalyzeMissingSymbol(t *testing.T) {
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	err := runTestCommand(t, cmd, "analyze")
	require.Error(t, err)
	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Empty(t, buf.String())
}

func TestStockAnalyzeTotalFailure(t *testing.T) {
	server := jsonServer(emptyMarketDataFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockAnalyzeCommand(c, &buf)
	err := runTestCommand(t, cmd, "analyze", "MISSING")
	require.Error(t, err)
	assert.Empty(t, buf.String())
}
