package commands

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

func TestStockGetSuccess(t *testing.T) {
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockGetCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "get", "AAPL"))

	result := parseJSONEnvelope(t, &buf)
	assertSymbolMeta(t, result, "AAPL")
}

func TestStockGetSymbolNotFound(t *testing.T) {
	server := jsonServer(emptyMarketDataFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockGetCommand(c, &buf)
	err := runTestCommand(t, cmd, "get", "MISSING")
	require.Error(t, err)

	var snf *mserrors.SymbolNotFoundError
	assert.ErrorAs(t, err, &snf)
	assert.Empty(t, buf.String())
}

func TestStockGetMissingSymbol(t *testing.T) {
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := StockGetCommand(c, &buf)
	err := runTestCommand(t, cmd, "get")
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Empty(t, buf.String())
}
