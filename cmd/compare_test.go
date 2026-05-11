package cmd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/major/marketsurge-go/marketsurge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentcmd "github.com/major/marketsurge-agent/cmd"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

func TestCompareSuccess(t *testing.T) {
	var request compareGraphQLRequest
	var requestErr error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, requestErr = compareRequest(r)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"marketDataAdhocScreen":{"numberOfInstrumentsInSource":2,"numberOfMatchingInstruments":2,"responseValues":[[{"value":"AMD","mdItem":{"name":"Symbol","mdItemID":1}},{"value":"Advanced Micro Devices","mdItem":{"name":"CompanyName","mdItemID":2}},{"value":"212.30","mdItem":{"name":"Price","mdItemID":3}},{"value":"96","mdItem":{"name":"CompositeRating","mdItemID":4}},{"value":"91","mdItem":{"name":"RSRating","mdItemID":5}},{"value":"A","mdItem":{"name":"IndustryGroupRSRatingLetter","mdItemID":6}},{"value":"4.1%","mdItem":{"name":"ATRPct21D","mdItemID":7}}],[{"value":"NVDA","mdItem":{"name":"Symbol","mdItemID":1}},{"value":"NVIDIA Corp","mdItem":{"name":"CompanyName","mdItemID":2}},{"value":"950.00","mdItem":{"name":"Price","mdItemID":3}},{"value":"99","mdItem":{"name":"CompositeRating","mdItemID":4}},{"value":"97","mdItem":{"name":"RSRating","mdItemID":5}},{"value":"A+","mdItem":{"name":"IndustryGroupRSRatingLetter","mdItemID":6}},{"value":"3.8%","mdItem":{"name":"ATRPct21D","mdItemID":7}}]]}}}`)
	}))
	t.Cleanup(server.Close)

	client := compareClient(t, server)
	output, err := runCompare(t, client, agentcmd.CompareCmd{Symbols: []string{" amd ", "NVDA", "amd"}})
	require.NoError(t, err, "CompareCmd.Run(success) error = %v, want nil", err)
	require.NoError(t, requestErr, "MarketDataAdhocScreen request capture error = %v, want nil", requestErr)
	assert.Equal(t, "MarketDataAdhocScreen", request.OperationName)
	assert.Equal(t, []string{"AMD", "NVDA"}, request.Variables.IncludeSource.Instruments.Symbols)
	assert.Equal(t, marketsurge.DefaultChartSymbolDialectType, request.Variables.IncludeSource.Instruments.Dialect)
	assert.Contains(t, requestColumnNames(&request), "CompositeRating")
	assert.Contains(t, requestColumnNames(&request), "ATRPct21D")

	var rows []map[string]any
	unmarshalErr := json.Unmarshal([]byte(output), &rows)
	require.NoError(t, unmarshalErr, "json.Unmarshal(CompareCmd.Run(success) output) error = %v, want nil", unmarshalErr)
	require.Len(t, rows, 2, "CompareCmd.Run(success) decoded rows length = %d, want %d", len(rows), 2)
	assert.Equal(t, "AMD", rows[0]["ticker"])
	assert.Equal(t, "Advanced Micro Devices", rows[0]["name"])
	assert.Equal(t, "96", rows[0]["ratings"].(map[string]any)["composite"])
	assert.Equal(t, "91", rows[0]["ratings"].(map[string]any)["relativeStrength"])
	assert.Equal(t, "212.30", rows[0]["price"].(map[string]any)["last"])
	assert.Equal(t, "4.1%", rows[0]["price"].(map[string]any)["atrPercent21d"])
	assert.Equal(t, "A", rows[0]["industry"].(map[string]any)["groupRSRating"])
	assert.Equal(t, "212.30", rows[0]["columns"].(map[string]any)["Price"])
	assert.Equal(t, "NVDA", rows[1]["ticker"])
}

func TestCompareCustomColumns(t *testing.T) {
	var request compareGraphQLRequest
	var requestErr error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, requestErr = compareRequest(r)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"marketDataAdhocScreen":{"numberOfMatchingInstruments":1,"responseValues":[[{"value":"AAPL","mdItem":{"name":"Symbol","mdItemID":1}},{"value":"213.50","mdItem":{"name":"Price","mdItemID":2}}]]}}}`)
	}))
	t.Cleanup(server.Close)

	client := compareClient(t, server)
	output, err := runCompare(t, client, agentcmd.CompareCmd{
		Symbols: []string{"AAPL"},
		Columns: []string{"Symbol", " Price ", ""},
	})
	require.NoError(t, err, "CompareCmd.Run(custom columns) error = %v, want nil", err)
	require.NoError(t, requestErr, "MarketDataAdhocScreen custom request capture error = %v, want nil", requestErr)
	assert.Equal(t, []string{"Symbol", "Price"}, requestColumnNames(&request))

	var rows []map[string]any
	unmarshalErr := json.Unmarshal([]byte(output), &rows)
	require.NoError(t, unmarshalErr, "json.Unmarshal(CompareCmd.Run(custom columns) output) error = %v, want nil", unmarshalErr)
	assert.Equal(t, "213.50", rows[0]["columns"].(map[string]any)["Price"])
}

func TestCompareCustomColumnsPrependsSymbol(t *testing.T) {
	var request compareGraphQLRequest
	var requestErr error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, requestErr = compareRequest(r)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"marketDataAdhocScreen":{"numberOfMatchingInstruments":1,"responseValues":[[{"value":"AAPL","mdItem":{"name":"Symbol","mdItemID":1}},{"value":"213.50","mdItem":{"name":"Price","mdItemID":2}}]]}}}`)
	}))
	t.Cleanup(server.Close)

	client := compareClient(t, server)
	output, err := runCompare(t, client, agentcmd.CompareCmd{
		Symbols: []string{"AAPL"},
		Columns: []string{"Price"},
	})
	require.NoError(t, err, "CompareCmd.Run(custom columns without symbol) error = %v, want nil", err)
	require.NoError(t, requestErr, "MarketDataAdhocScreen custom request capture error = %v, want nil", requestErr)
	assert.Equal(t, []string{"Symbol", "Price"}, requestColumnNames(&request))

	var rows []map[string]any
	unmarshalErr := json.Unmarshal([]byte(output), &rows)
	require.NoError(t, unmarshalErr, "json.Unmarshal(CompareCmd.Run(custom columns without symbol) output) error = %v, want nil", unmarshalErr)
	assert.Equal(t, "AAPL", rows[0]["ticker"])
}

func TestCompareBlankCustomColumnsValidationError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)

	client := compareClient(t, server)
	output, err := runCompare(t, client, agentcmd.CompareCmd{
		Symbols: []string{"AAPL"},
		Columns: []string{" ", ""},
	})
	require.Error(t, err, "CompareCmd.Run(blank custom columns) error = nil, want non-nil")

	var validationErr *mserrors.ValidationError
	require.ErrorAs(t, err, &validationErr, "CompareCmd.Run(blank custom columns) error type = %T, want *mserrors.ValidationError", err)
	assert.Empty(t, output, "CompareCmd.Run(blank custom columns) stdout = %q, want empty", output)
}

func TestCompareMixedBlankSymbols(t *testing.T) {
	var request compareGraphQLRequest
	var requestErr error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, requestErr = compareRequest(r)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"marketDataAdhocScreen":{"numberOfMatchingInstruments":1,"responseValues":[[{"value":"MSFT","mdItem":{"name":"Symbol","mdItemID":1}}]]}}}`)
	}))
	t.Cleanup(server.Close)

	client := compareClient(t, server)
	_, err := runCompare(t, client, agentcmd.CompareCmd{Symbols: []string{" ", "MSFT", ""}})
	require.NoError(t, err, "CompareCmd.Run(mixed blank symbols) error = %v, want nil", err)
	require.NoError(t, requestErr, "MarketDataAdhocScreen mixed blank request capture error = %v, want nil", requestErr)
	assert.Equal(t, []string{"MSFT"}, request.Variables.IncludeSource.Instruments.Symbols)
}

func TestCompareEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"marketDataAdhocScreen":{"numberOfMatchingInstruments":0,"responseValues":[]}}}`)
	}))
	t.Cleanup(server.Close)

	client := compareClient(t, server)
	output, err := runCompare(t, client, agentcmd.CompareCmd{Symbols: []string{"AAPL"}})
	require.NoError(t, err, "CompareCmd.Run(empty) error = %v, want nil", err)
	assert.Equal(t, "[]\n", output, "CompareCmd.Run(empty) output = %q, want %q", output, "[]\n")
}

func TestCompareValidationError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)

	client := compareClient(t, server)
	output, err := runCompare(t, client, agentcmd.CompareCmd{Symbols: []string{" ", ""}})
	require.Error(t, err, "CompareCmd.Run(blank symbols) error = nil, want non-nil")

	var validationErr *mserrors.ValidationError
	require.ErrorAs(t, err, &validationErr, "CompareCmd.Run(blank symbols) error type = %T, want *mserrors.ValidationError", err)
	assert.Empty(t, output, "CompareCmd.Run(blank symbols) stdout = %q, want empty", output)
}

func TestCompareAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	client := compareClient(t, server)
	output, err := runCompare(t, client, agentcmd.CompareCmd{Symbols: []string{"AMD"}})
	require.Error(t, err, "CompareCmd.Run(auth error) error = nil, want non-nil")

	var authErr *mserrors.AuthenticationError
	require.ErrorAs(t, err, &authErr, "CompareCmd.Run(auth error) error type = %T, want *mserrors.AuthenticationError", err)
	assert.Empty(t, output, "CompareCmd.Run(auth error) stdout = %q, want empty", output)
}

type compareGraphQLRequest struct {
	OperationName string `json:"operationName"`
	Variables     struct {
		ResponseColumns []struct {
			Name string `json:"name"`
		} `json:"responseColumns"`
		IncludeSource struct {
			Instruments struct {
				Symbols []string `json:"symbols"`
				Dialect string   `json:"dialect"`
			} `json:"instruments"`
		} `json:"includeSource"`
	} `json:"variables"`
}

func compareRequest(r *http.Request) (compareGraphQLRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return compareGraphQLRequest{}, err
	}

	var req compareGraphQLRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return compareGraphQLRequest{}, err
	}
	return req, nil
}

func requestColumnNames(req *compareGraphQLRequest) []string {
	names := make([]string, 0, len(req.Variables.ResponseColumns))
	for _, column := range req.Variables.ResponseColumns {
		names = append(names, column.Name)
	}

	return names
}

func compareClient(t *testing.T, server *httptest.Server) *marketsurge.Client {
	t.Helper()

	client, err := marketsurge.NewClient(
		marketsurge.WithJWT("test-token"),
		marketsurge.WithHTTPClient(server.Client()),
		marketsurge.WithGraphQLURL(server.URL),
	)
	require.NoError(t, err, "marketsurge.NewClient(test server %q) error = %v, want nil", server.URL, err)
	return client
}

func runCompare(t *testing.T, client *marketsurge.Client, cmd agentcmd.CompareCmd) (string, error) {
	t.Helper()

	var output bytes.Buffer
	runErr := cmd.RunForTest(client, &output)
	return output.String(), runErr
}
