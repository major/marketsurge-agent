package cmd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/major/marketsurge-go/marketsurge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentcmd "github.com/major/marketsurge-agent/cmd"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

func TestWatchlistListSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"watchlists":[{"id":"101","name":"My Watchlist","lastModifiedDateUtc":"2025-01-15T10:30:00Z","description":"Top picks"},{"id":"202","name":"Growth Stocks","lastModifiedDateUtc":"2025-02-20T14:00:00Z","description":"High growth"}]}}`)
	}))
	t.Cleanup(server.Close)

	client := watchlistListClient(t, server)
	output, err := runWatchlistList(t, client)
	require.NoError(t, err, "WatchlistListCmd.Run(success) error = %v, want nil", err)

	var watchlists []marketsurge.WatchlistSummary
	require.NoError(t, json.Unmarshal([]byte(output), &watchlists), "json.Unmarshal(WatchlistListCmd.Run(success) output) error = %v, want nil", err)
	require.Len(t, watchlists, 2, "WatchlistListCmd.Run(success) decoded watchlists length = %d, want %d", len(watchlists), 2)

	assert.Equal(t, "101", watchlists[0].ID)
	assert.Equal(t, "My Watchlist", watchlists[0].Name)
	assert.Equal(t, "2025-01-15T10:30:00Z", watchlists[0].LastModifiedDateUtc)
	assert.Equal(t, "Top picks", watchlists[0].Description)

	assert.Equal(t, "202", watchlists[1].ID)
	assert.Equal(t, "Growth Stocks", watchlists[1].Name)
	assert.Equal(t, "2025-02-20T14:00:00Z", watchlists[1].LastModifiedDateUtc)
	assert.Equal(t, "High growth", watchlists[1].Description)
}

func TestWatchlistListEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"watchlists":[]}}`)
	}))
	t.Cleanup(server.Close)

	client := watchlistListClient(t, server)
	output, err := runWatchlistList(t, client)
	require.NoError(t, err, "WatchlistListCmd.Run(empty) error = %v, want nil", err)
	assert.Equal(t, "[]\n", output, "WatchlistListCmd.Run(empty) output = %q, want %q", output, "[]\n")
}

func TestWatchlistListAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	client := watchlistListClient(t, server)
	output, err := runWatchlistList(t, client)
	require.Error(t, err, "WatchlistListCmd.Run(auth error) error = nil, want non-nil")

	var authErr *mserrors.AuthenticationError
	require.ErrorAs(t, err, &authErr, "WatchlistListCmd.Run(auth error) error type = %T, want *mserrors.AuthenticationError", err)
	assert.Empty(t, output, "WatchlistListCmd.Run(auth error) stdout = %q, want empty", output)
}

func TestWatchlistListAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"errors":[{"message":"Internal error","path":["watchlists"]}]}`)
	}))
	t.Cleanup(server.Close)

	client := watchlistListClient(t, server)
	output, err := runWatchlistList(t, client)
	require.Error(t, err, "WatchlistListCmd.Run(GraphQL error) error = nil, want non-nil")

	var apiErr *mserrors.APIError
	require.ErrorAs(t, err, &apiErr, "WatchlistListCmd.Run(GraphQL error) error type = %T, want *mserrors.APIError", err)
	assert.Empty(t, output, "WatchlistListCmd.Run(GraphQL error) stdout = %q, want empty", output)
}

func watchlistListClient(t *testing.T, server *httptest.Server) *marketsurge.Client {
	t.Helper()

	client, err := marketsurge.NewClient(
		marketsurge.WithJWT("test-token"),
		marketsurge.WithHTTPClient(server.Client()),
		marketsurge.WithGraphQLURL(server.URL),
	)
	require.NoError(t, err, "marketsurge.NewClient(test server %q) error = %v, want nil", server.URL, err)
	return client
}

func runWatchlistList(t *testing.T, client *marketsurge.Client) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "os.Pipe() error = %v, want nil", err)
	t.Cleanup(func() {
		_ = r.Close()
	})

	os.Stdout = w
	runErr := (&agentcmd.WatchlistListCmd{}).Run(client)
	closeErr := w.Close()
	os.Stdout = oldStdout
	require.NoError(t, closeErr, "stdout pipe Close() error = %v, want nil", closeErr)

	var output bytes.Buffer
	_, err = io.Copy(&output, r)
	require.NoError(t, err, "io.Copy(WatchlistListCmd.Run stdout) error = %v, want nil", err)
	return output.String(), runErr
}
