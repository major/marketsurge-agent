package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

func TestFundamentalGetSuccess(t *testing.T) {
	server := jsonServer(stockResponseFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := FundamentalGetCommand(c, &buf)
	cmd.ExitErrHandler = func(_ context.Context, _ *cli.Command, _ error) {}

	err := cmd.Run(context.Background(), []string{"get", "AAPL"})
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Contains(t, result, "data")
	assert.Contains(t, result, "metadata")

	meta, _ := result["metadata"].(map[string]any)
	assert.Equal(t, "AAPL", meta["symbol"])
}

func TestFundamentalGetSymbolNotFound(t *testing.T) {
	server := jsonServer(emptyMarketDataFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := FundamentalGetCommand(c, &buf)
	cmd.ExitErrHandler = func(_ context.Context, _ *cli.Command, _ error) {}

	err := cmd.Run(context.Background(), []string{"get", "MISSING"})
	require.Error(t, err)

	var snf *mserrors.SymbolNotFoundError
	assert.ErrorAs(t, err, &snf)

	var result map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	errObj, _ := result["error"].(map[string]any)
	assert.Equal(t, "SYMBOL_NOT_FOUND", errObj["code"])
}

func TestFundamentalGetMissingSymbol(t *testing.T) {
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := FundamentalGetCommand(c, &buf)
	cmd.ExitErrHandler = func(_ context.Context, _ *cli.Command, _ error) {}

	err := cmd.Run(context.Background(), []string{"get"})
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
}
