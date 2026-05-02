package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

func TestOwnershipGetSuccess(t *testing.T) {
	t.Parallel()
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	output, err := executeCommandWithClient(t, newOwnershipCmd(), c, "get", "AAPL")
	require.NoError(t, err)
	result := parseJSONEnvelope(t, output)
	assertSymbolMeta(t, result, "AAPL")
}

func TestOwnershipGetSymbolNotFound(t *testing.T) {
	t.Parallel()
	server := jsonServer(emptyMarketDataFixture())
	defer server.Close()
	c := testClient(t, server)

	_, err := executeCommandWithClient(t, newOwnershipCmd(), c, "get", "MISSING")
	require.Error(t, err)
	var snf *mserrors.SymbolNotFoundError
	assert.ErrorAs(t, err, &snf)
}

func TestOwnershipGetMissingSymbol(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	_, err := executeCommandWithClient(t, newOwnershipCmd(), c, "get")
	require.Error(t, err)
}
