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

func TestWatchlistGetSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"watchlist":{"id":"101","name":"My Watchlist","lastModifiedDateUtc":"2025-01-15T10:30:00Z","description":"Top picks","items":[{"key":"AAPL","dowJonesKey":"US:AAPL"},{"key":"NVDA","dowJonesKey":"US:NVDA"},{"key":"MSFT","dowJonesKey":"US:MSFT"}]}}}`)
	}))
	t.Cleanup(server.Close)

	client := watchlistGetClient(t, server)
	output, err := runWatchlistGet(t, client, agentcmd.WatchlistGetCmd{ID: "101"})
	require.NoError(t, err, "WatchlistGetCmd.Run(success) error = %v, want nil", err)

	var items []struct {
		ID                  string   `json:"id"`
		Name                string   `json:"name"`
		LastModifiedDateUtc string   `json:"lastModifiedDateUtc"`
		Description         string   `json:"description"`
		Symbols             []string `json:"symbols"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items), "json.Unmarshal(WatchlistGetCmd.Run(success) output) error = %v, want nil", err)
	require.Len(t, items, 1, "WatchlistGetCmd.Run(success) decoded items length = %d, want %d", len(items), 1)

	assert.Equal(t, "101", items[0].ID)
	assert.Equal(t, "My Watchlist", items[0].Name)
	assert.Equal(t, "2025-01-15T10:30:00Z", items[0].LastModifiedDateUtc)
	assert.Equal(t, "Top picks", items[0].Description)
	assert.Equal(t, []string{"AAPL", "NVDA", "MSFT"}, items[0].Symbols)
}

func TestWatchlistGetEmptyItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"watchlist":{"id":"101","name":"Empty List","lastModifiedDateUtc":"2025-01-15T10:30:00Z","description":"Nothing here","items":[]}}}`)
	}))
	t.Cleanup(server.Close)

	client := watchlistGetClient(t, server)
	output, err := runWatchlistGet(t, client, agentcmd.WatchlistGetCmd{ID: "101"})
	require.NoError(t, err, "WatchlistGetCmd.Run(empty items) error = %v, want nil", err)

	var items []struct {
		Symbols []string `json:"symbols"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items), "json.Unmarshal(WatchlistGetCmd.Run(empty items) output) error = %v, want nil", err)
	require.Len(t, items, 1, "WatchlistGetCmd.Run(empty items) decoded items length = %d, want %d", len(items), 1)
	assert.Equal(t, []string{}, items[0].Symbols, "WatchlistGetCmd.Run(empty items) symbols = %v, want empty", items[0].Symbols)
}

func TestWatchlistGetNullWatchlist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"watchlist":null}}`)
	}))
	t.Cleanup(server.Close)

	client := watchlistGetClient(t, server)
	output, err := runWatchlistGet(t, client, agentcmd.WatchlistGetCmd{ID: "404"})
	require.NoError(t, err, "WatchlistGetCmd.Run(null watchlist) error = %v, want nil", err)

	var items []struct {
		Symbols []string `json:"symbols"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items), "json.Unmarshal(WatchlistGetCmd.Run(null watchlist) output) error = %v, want nil", err)
	require.Len(t, items, 1, "WatchlistGetCmd.Run(null watchlist) decoded items length = %d, want %d", len(items), 1)
	assert.Equal(t, []string{}, items[0].Symbols, "WatchlistGetCmd.Run(null watchlist) symbols = %v, want empty", items[0].Symbols)
}

func TestWatchlistGetAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	client := watchlistGetClient(t, server)
	output, err := runWatchlistGet(t, client, agentcmd.WatchlistGetCmd{ID: "101"})
	require.Error(t, err, "WatchlistGetCmd.Run(auth error) error = nil, want non-nil")

	var authErr *mserrors.AuthenticationError
	require.ErrorAs(t, err, &authErr, "WatchlistGetCmd.Run(auth error) error type = %T, want *mserrors.AuthenticationError", err)
	assert.Empty(t, output, "WatchlistGetCmd.Run(auth error) stdout = %q, want empty", output)
}

func TestWatchlistGetAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"errors":[{"message":"Watchlist not found","path":["watchlist"]}]}`)
	}))
	t.Cleanup(server.Close)

	client := watchlistGetClient(t, server)
	output, err := runWatchlistGet(t, client, agentcmd.WatchlistGetCmd{ID: "999"})
	require.Error(t, err, "WatchlistGetCmd.Run(GraphQL error) error = nil, want non-nil")

	var apiErr *mserrors.APIError
	require.ErrorAs(t, err, &apiErr, "WatchlistGetCmd.Run(GraphQL error) error type = %T, want *mserrors.APIError", err)
	assert.Empty(t, output, "WatchlistGetCmd.Run(GraphQL error) stdout = %q, want empty", output)
}

func watchlistGetClient(t *testing.T, server *httptest.Server) *marketsurge.Client {
	t.Helper()

	client, err := marketsurge.NewClient(
		marketsurge.WithJWT("test-token"),
		marketsurge.WithHTTPClient(server.Client()),
		marketsurge.WithGraphQLURL(server.URL),
	)
	require.NoError(t, err, "marketsurge.NewClient(test server %q) error = %v, want nil", server.URL, err)
	return client
}

func runWatchlistGet(t *testing.T, client *marketsurge.Client, cmd agentcmd.WatchlistGetCmd) (string, error) {
	t.Helper()

	var output bytes.Buffer
	runErr := cmd.RunForTest(client, &output)
	return output.String(), runErr
}
