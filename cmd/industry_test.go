package cmd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/major/marketsurge-go/marketsurge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentcmd "github.com/major/marketsurge-agent/cmd"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

func TestIndustrySuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"marketData":[{"originRequest":{"symbol":"AAPL"},"industry":{"groupRS":[{"value":85}]}},{"originRequest":{"symbol":"NVDA"},"industry":{"groupRS":[{"value":92}]}}]}}`)
	}))
	t.Cleanup(server.Close)

	client := industryClient(t, server)
	output, err := runIndustry(t, client, agentcmd.IndustryCmd{Symbols: []string{"AAPL", "NVDA"}})
	require.NoError(t, err, "IndustryCmd.Run(success) error = %v, want nil", err)

	var items []struct {
		Ticker          string `json:"ticker"`
		IndustryGroupRS *int   `json:"industryGroupRS"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items), "json.Unmarshal(IndustryCmd.Run(success) output)")
	require.Len(t, items, 2, "IndustryCmd.Run(success) decoded items length = %d, want %d", len(items), 2)

	assert.Equal(t, "AAPL", items[0].Ticker)
	require.NotNil(t, items[0].IndustryGroupRS, "IndustryCmd.Run(success) items[0].industryGroupRS = nil, want non-nil")
	assert.Equal(t, 85, *items[0].IndustryGroupRS)

	assert.Equal(t, "NVDA", items[1].Ticker)
	require.NotNil(t, items[1].IndustryGroupRS, "IndustryCmd.Run(success) items[1].industryGroupRS = nil, want non-nil")
	assert.Equal(t, 92, *items[1].IndustryGroupRS)
}

func TestIndustrySingleSymbol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"marketData":[{"originRequest":{"symbol":"AAPL"},"industry":{"groupRS":[{"value":85}]}}]}}`)
	}))
	t.Cleanup(server.Close)

	client := industryClient(t, server)
	output, err := runIndustry(t, client, agentcmd.IndustryCmd{Symbols: []string{"aapl"}})
	require.NoError(t, err, "IndustryCmd.Run(single symbol) error = %v, want nil", err)

	var items []struct {
		Ticker          string `json:"ticker"`
		IndustryGroupRS *int   `json:"industryGroupRS"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items), "json.Unmarshal(IndustryCmd.Run(single symbol) output)")
	require.Len(t, items, 1, "IndustryCmd.Run(single symbol) decoded items length = %d, want %d", len(items), 1)
	assert.Equal(t, "AAPL", items[0].Ticker)
}

func TestIndustryNilGroupRS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"marketData":[{"originRequest":{"symbol":"AAPL"},"industry":{"groupRS":[]}}]}}`)
	}))
	t.Cleanup(server.Close)

	client := industryClient(t, server)
	output, err := runIndustry(t, client, agentcmd.IndustryCmd{Symbols: []string{"AAPL"}})
	require.NoError(t, err, "IndustryCmd.Run(nil RS) error = %v, want nil", err)

	var items []struct {
		Ticker          string `json:"ticker"`
		IndustryGroupRS *int   `json:"industryGroupRS"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items), "json.Unmarshal(IndustryCmd.Run(nil RS) output)")
	require.Len(t, items, 1, "IndustryCmd.Run(nil RS) decoded items length = %d, want %d", len(items), 1)
	assert.Equal(t, "AAPL", items[0].Ticker)
	assert.Nil(t, items[0].IndustryGroupRS, "IndustryCmd.Run(nil RS) items[0].industryGroupRS = %v, want nil", items[0].IndustryGroupRS)
}

func TestIndustryEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"marketData":[]}}`)
	}))
	t.Cleanup(server.Close)

	client := industryClient(t, server)
	output, err := runIndustry(t, client, agentcmd.IndustryCmd{Symbols: []string{"AAPL"}})
	require.NoError(t, err, "IndustryCmd.Run(empty response) error = %v, want nil", err)
	assert.Equal(t, "[]\n", output, "IndustryCmd.Run(empty response) output = %q, want %q", output, "[]\n")
}

func TestIndustryEmptySymbols(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("IndustryCmd.Run(empty symbols) sent unexpected HTTP request")
	}))
	t.Cleanup(server.Close)

	client := industryClient(t, server)
	_, err := runIndustry(t, client, agentcmd.IndustryCmd{Symbols: []string{"", "  "}})
	require.Error(t, err, "IndustryCmd.Run(empty symbols) error = nil, want non-nil")

	var validationErr *mserrors.ValidationError
	require.ErrorAs(t, err, &validationErr, "IndustryCmd.Run(empty symbols) error type = %T, want *mserrors.ValidationError", err)
}

func TestIndustryAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	client := industryClient(t, server)
	output, err := runIndustry(t, client, agentcmd.IndustryCmd{Symbols: []string{"AAPL"}})
	require.Error(t, err, "IndustryCmd.Run(auth error) error = nil, want non-nil")

	var authErr *mserrors.AuthenticationError
	require.ErrorAs(t, err, &authErr, "IndustryCmd.Run(auth error) error type = %T, want *mserrors.AuthenticationError", err)
	assert.Empty(t, output, "IndustryCmd.Run(auth error) stdout = %q, want empty", output)
}

func TestIndustryAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"errors":[{"message":"Service unavailable","path":["marketData"]}]}`)
	}))
	t.Cleanup(server.Close)

	client := industryClient(t, server)
	output, err := runIndustry(t, client, agentcmd.IndustryCmd{Symbols: []string{"AAPL"}})
	require.Error(t, err, "IndustryCmd.Run(GraphQL error) error = nil, want non-nil")

	var apiErr *mserrors.APIError
	require.ErrorAs(t, err, &apiErr, "IndustryCmd.Run(GraphQL error) error type = %T, want *mserrors.APIError", err)
	assert.Empty(t, output, "IndustryCmd.Run(GraphQL error) stdout = %q, want empty", output)
}

func industryClient(t *testing.T, server *httptest.Server) *marketsurge.Client {
	t.Helper()

	client, err := marketsurge.NewClient(
		marketsurge.WithJWT("test-token"),
		marketsurge.WithHTTPClient(server.Client()),
		marketsurge.WithGraphQLURL(server.URL),
	)
	require.NoError(t, err, "marketsurge.NewClient(test server %q) error = %v, want nil", server.URL, err)
	return client
}

func runIndustry(t *testing.T, client *marketsurge.Client, cmd agentcmd.IndustryCmd) (string, error) {
	t.Helper()

	var output bytes.Buffer
	runErr := cmd.RunForTest(client, &output)
	return output.String(), runErr
}
