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

func TestChartMarkupsSuccess(t *testing.T) {
	server := jsonServer(chartMarkupsFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartMarkupsCommand(c, &buf)
	cmd.ExitErrHandler = func(_ context.Context, _ *cli.Command, _ error) {}

	err := cmd.Run(context.Background(), []string{"markups", "AAPL"})
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Contains(t, result, "data")
	assert.Contains(t, result, "metadata")

	meta, _ := result["metadata"].(map[string]any)
	assert.Equal(t, "AAPL", meta["symbol"])
}

func TestChartMarkupsWithFlags(t *testing.T) {
	server := jsonServer(chartMarkupsFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartMarkupsCommand(c, &buf)
	cmd.ExitErrHandler = func(_ context.Context, _ *cli.Command, _ error) {}

	err := cmd.Run(context.Background(), []string{
		"markups", "--frequency", "WEEKLY", "--sort-dir", "DESC", "AAPL",
	})
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Contains(t, result, "data")
}

func TestChartMarkupsMissingSymbol(t *testing.T) {
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartMarkupsCommand(c, &buf)
	cmd.ExitErrHandler = func(_ context.Context, _ *cli.Command, _ error) {}

	err := cmd.Run(context.Background(), []string{"markups"})
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)

	var result map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	errObj, _ := result["error"].(map[string]any)
	assert.Equal(t, "VALIDATION_ERROR", errObj["code"])
}
