package commands

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

func TestFundamentalGetSuccess(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := FundamentalGetCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "get", "AAPL"))

	result := parseJSONEnvelope(t, &buf)
	assertSymbolMeta(t, result, "AAPL")
}

func TestFundamentalGetSymbolNotFound(t *testing.T) {
	t.Parallel()
	server := jsonServer(emptyMarketDataFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := FundamentalGetCommand(c, &buf)
	err := runTestCommand(t, cmd, "get", "MISSING")
	require.Error(t, err)

	var snf *mserrors.SymbolNotFoundError
	assert.ErrorAs(t, err, &snf)
	assert.Empty(t, buf.String())
}

func TestFundamentalGetMissingSymbol(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := FundamentalGetCommand(c, &buf)
	err := runTestCommand(t, cmd, "get")
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Empty(t, buf.String())
}
