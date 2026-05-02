package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
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
