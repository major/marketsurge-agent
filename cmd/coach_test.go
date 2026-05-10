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

const coachSuccessResponse = `{"data":{"user":{"watchlists":[{"id":"w1","name":"IBD 50","parentId":null,"type":"FOLDER","children":[{"id":"w1c1","name":"Current Week","type":"LEAF"}],"contentType":"WATCHLIST","treeType":"MSR_NAV"},{"id":"w2","name":"Sector Leaders","parentId":null,"type":"LEAF","url":"/watchlist/w2","treeType":"MSR_NAV","referenceId":"ref-w2"}],"screens":[{"id":"s1","name":"Top Composite","parentId":null,"type":"FOLDER","children":[{"id":"s1c1","name":"Large Cap","type":"LEAF"},{"id":"s1c2","name":"Small Cap","type":"LEAF"}],"contentType":"SCREEN","treeType":"MSR_NAV"}]}}}`

func TestCoachSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, coachSuccessResponse)
	}))
	t.Cleanup(server.Close)

	client := coachClient(t, server)
	output, err := runCoach(t, client, agentcmd.CoachCmd{Type: "all"})
	require.NoError(t, err, "CoachCmd.Run(success) error = %v, want nil", err)

	var items []struct {
		ID          *string `json:"id"`
		Name        *string `json:"name"`
		Type        *string `json:"type"`
		ContentType *string `json:"contentType"`
		Category    string  `json:"category"`
		Children    []struct {
			ID   *string `json:"id"`
			Name *string `json:"name"`
			Type *string `json:"type"`
		} `json:"children"`
		URL         *string `json:"url"`
		ReferenceID *string `json:"referenceId"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items), "json.Unmarshal(CoachCmd.Run(success) output)")
	require.Len(t, items, 3, "CoachCmd.Run(success) decoded items length = %d, want %d", len(items), 3)

	require.NotNil(t, items[0].ID)
	require.NotNil(t, items[0].Name)
	assert.Equal(t, "w1", *items[0].ID)
	assert.Equal(t, "IBD 50", *items[0].Name)
	assert.Equal(t, "watchlist", items[0].Category)
	require.Len(t, items[0].Children, 1)

	require.NotNil(t, items[1].ID)
	require.NotNil(t, items[1].Name)
	assert.Equal(t, "w2", *items[1].ID)
	assert.Equal(t, "Sector Leaders", *items[1].Name)
	assert.Equal(t, "watchlist", items[1].Category)
	require.NotNil(t, items[1].URL)
	require.NotNil(t, items[1].ReferenceID)
	assert.Equal(t, "/watchlist/w2", *items[1].URL)
	assert.Equal(t, "ref-w2", *items[1].ReferenceID)

	require.NotNil(t, items[2].ID)
	require.NotNil(t, items[2].Name)
	assert.Equal(t, "s1", *items[2].ID)
	assert.Equal(t, "Top Composite", *items[2].Name)
	assert.Equal(t, "screen", items[2].Category)
	require.Len(t, items[2].Children, 2)
}

func TestCoachWatchlistOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, coachSuccessResponse)
	}))
	t.Cleanup(server.Close)

	client := coachClient(t, server)
	output, err := runCoach(t, client, agentcmd.CoachCmd{Type: "watchlist"})
	require.NoError(t, err, "CoachCmd.Run(watchlist only) error = %v, want nil", err)

	var items []struct {
		Category string `json:"category"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items), "json.Unmarshal(CoachCmd.Run(watchlist only) output)")
	require.Len(t, items, 2, "CoachCmd.Run(watchlist only) decoded items length = %d, want %d", len(items), 2)
	assert.Equal(t, "watchlist", items[0].Category)
	assert.Equal(t, "watchlist", items[1].Category)
}

func TestCoachScreenOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, coachSuccessResponse)
	}))
	t.Cleanup(server.Close)

	client := coachClient(t, server)
	output, err := runCoach(t, client, agentcmd.CoachCmd{Type: "screen"})
	require.NoError(t, err, "CoachCmd.Run(screen only) error = %v, want nil", err)

	var items []struct {
		Category string `json:"category"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items), "json.Unmarshal(CoachCmd.Run(screen only) output)")
	require.Len(t, items, 1, "CoachCmd.Run(screen only) decoded items length = %d, want %d", len(items), 1)
	assert.Equal(t, "screen", items[0].Category)
}

func TestCoachNilResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{}}`)
	}))
	t.Cleanup(server.Close)

	client := coachClient(t, server)
	output, err := runCoach(t, client, agentcmd.CoachCmd{Type: "all"})
	require.NoError(t, err, "CoachCmd.Run(nil response) error = %v, want nil", err)
	assert.Equal(t, "[]\n", output, "CoachCmd.Run(nil response) output = %q, want %q", output, "[]\n")
}

func TestCoachAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	client := coachClient(t, server)
	output, err := runCoach(t, client, agentcmd.CoachCmd{Type: "all"})
	require.Error(t, err, "CoachCmd.Run(auth error) error = nil, want non-nil")

	var authErr *mserrors.AuthenticationError
	require.ErrorAs(t, err, &authErr, "CoachCmd.Run(auth error) error type = %T, want *mserrors.AuthenticationError", err)
	assert.Empty(t, output, "CoachCmd.Run(auth error) stdout = %q, want empty", output)
}

func TestCoachAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"errors":[{"message":"Coach tree unavailable","path":["coachTree"]}]}`)
	}))
	t.Cleanup(server.Close)

	client := coachClient(t, server)
	output, err := runCoach(t, client, agentcmd.CoachCmd{Type: "all"})
	require.Error(t, err, "CoachCmd.Run(API error) error = nil, want non-nil")

	var apiErr *mserrors.APIError
	require.ErrorAs(t, err, &apiErr, "CoachCmd.Run(API error) error type = %T, want *mserrors.APIError", err)
	assert.Empty(t, output, "CoachCmd.Run(API error) stdout = %q, want empty", output)
}

func coachClient(t *testing.T, server *httptest.Server) *marketsurge.Client {
	t.Helper()

	client, err := marketsurge.NewClient(
		marketsurge.WithJWT("test-token"),
		marketsurge.WithHTTPClient(server.Client()),
		marketsurge.WithGraphQLURL(server.URL),
	)
	require.NoError(t, err, "marketsurge.NewClient(test server %q) error = %v, want nil", server.URL, err)
	return client
}

func runCoach(t *testing.T, client *marketsurge.Client, cmd agentcmd.CoachCmd) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "os.Pipe() error = %v, want nil", err)
	t.Cleanup(func() {
		_ = r.Close()
	})

	os.Stdout = w
	runErr := cmd.Run(client)
	closeErr := w.Close()
	os.Stdout = oldStdout
	require.NoError(t, closeErr, "stdout pipe Close() error = %v, want nil", closeErr)

	var output bytes.Buffer
	_, err = io.Copy(&output, r)
	require.NoError(t, err, "io.Copy(CoachCmd.Run stdout) error = %v, want nil", err)
	return output.String(), runErr
}
