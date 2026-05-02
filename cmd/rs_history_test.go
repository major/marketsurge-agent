package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

func TestRSHistoryGetSuccess(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeCommandWithClient(newRSHistoryCmd(), c, "get", "AAPL")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
	assertSymbolMeta(t, result, "AAPL")
}

func TestRSHistoryGetMultiSymbol(t *testing.T) {
	t.Parallel()
	server := jsonServer(rsHistoryMultiResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeCommandWithClient(newRSHistoryCmd(), c, "get", "AAPL", "MSFT")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
	data, ok := result["data"].(map[string]any)
	require.True(t, ok, "multi-symbol RS history should be keyed by symbol")
	assert.Contains(t, data, "AAPL")
	assert.Contains(t, data, "MSFT")

	meta, _ := result["metadata"].(map[string]any)
	symbols, _ := meta["symbols"].([]any)
	assert.Equal(t, []any{"AAPL", "MSFT"}, symbols)
}

func TestRSHistoryGetMultiSymbolPartial(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeCommandWithClient(newRSHistoryCmd(), c, "get", "AAPL", "MISSING")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
	data, ok := result["data"].(map[string]any)
	require.True(t, ok, "partial RS history data should be keyed by symbol")
	assert.Contains(t, data, "AAPL")
	assert.NotContains(t, data, "MISSING")

	errors, ok := result["errors"].([]any)
	require.True(t, ok, "partial envelope should include errors")
	assert.Equal(t, []any{"MISSING: symbol not found"}, errors)
}

func TestRSHistoryGetMultiSymbolPartialMissingMiddle(t *testing.T) {
	t.Parallel()
	server := jsonServer(rsHistoryMissingMiddleResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeCommandWithClient(newRSHistoryCmd(), c, "get", "AAPL", "MISSING", "MSFT")
	require.NoError(t, err)

	result := parseJSONEnvelope(t, output)
	data, ok := result["data"].(map[string]any)
	require.True(t, ok, "partial RS history data should be keyed by response symbol")
	assert.Contains(t, data, "AAPL")
	assert.Contains(t, data, "MSFT")
	assert.NotContains(t, data, "MISSING")

	aapl, ok := data["AAPL"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "AAPL", aapl["symbol"])
	msft, ok := data["MSFT"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "MSFT", msft["symbol"])

	errors, ok := result["errors"].([]any)
	require.True(t, ok, "partial envelope should include errors")
	assert.Equal(t, []any{"MISSING: symbol not found"}, errors)
}

func TestRSHistoryGetSymbolNotFound(t *testing.T) {
	t.Parallel()
	server := jsonServer(emptyMarketDataFixture())
	defer server.Close()
	c := testClient(t, server)

	_, err := executeCommandWithClient(newRSHistoryCmd(), c, "get", "MISSING")
	require.Error(t, err)

	var snf *mserrors.SymbolNotFoundError
	assert.ErrorAs(t, err, &snf)
}

func TestRSHistoryGetMissingSymbol(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	_, err := executeCommandWithClient(newRSHistoryCmd(), c, "get")
	require.Error(t, err)
}
