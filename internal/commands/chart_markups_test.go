package commands

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

func TestChartMarkupsSuccess(t *testing.T) {
	t.Parallel()
	server := jsonServer(chartMarkupsFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartMarkupsCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd, "markups", "AAPL"))

	result := parseJSONEnvelope(t, &buf)
	assertSymbolMeta(t, result, "AAPL")
}

func TestChartMarkupsWithFlags(t *testing.T) {
	t.Parallel()
	server := jsonServer(chartMarkupsFixture())
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartMarkupsCommand(c, &buf)
	require.NoError(t, runTestCommand(t, cmd,
		"markups", "--frequency", "WEEKLY", "--sort-dir", "DESC", "AAPL",
	))

	parseJSONEnvelope(t, &buf)
}

func TestChartMarkupsMissingSymbol(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()
	c := testClient(t, server)

	var buf bytes.Buffer
	cmd := ChartMarkupsCommand(c, &buf)
	err := runTestCommand(t, cmd, "markups")
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Empty(t, buf.String())
}
