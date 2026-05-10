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

func TestReportsListSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"user":{"screens":[{"id":"42","name":"IBD 50","type":"SCREEN","site":"marketsurge","description":"Top 50 stocks","createdAt":"2024-01-01T00:00:00Z","updatedAt":"2024-06-01T00:00:00Z","filterCriteria":null,"source":{"id":"1","type":"IBD","pub":"IBD"}},{"id":"99","name":"Growth 250","type":"REPORT","site":"marketsurge","description":"Growth report","createdAt":"2024-02-01T00:00:00Z","updatedAt":"2024-07-01T00:00:00Z","filterCriteria":{"type":"AND","terms":[{"left":{"name":"RSRating"},"operand":">","right":{"value":"80"}}]},"source":{"id":"2","type":"IBD","pub":"MarketSurge"}}]}}}`)
	}))
	t.Cleanup(server.Close)

	client := reportsListClient(t, server)
	output, err := runReportsList(t, client)
	require.NoError(t, err, "ReportsListCmd.Run(success) error = %v, want nil", err)

	var screens []marketsurge.ScreenEntry
	require.NoError(t, json.Unmarshal([]byte(output), &screens), "json.Unmarshal(ReportsListCmd.Run(success) output)")
	require.Len(t, screens, 2, "ReportsListCmd.Run(success) decoded screens length = %d, want %d", len(screens), 2)

	assert.Equal(t, "42", stringValue(t, screens[0].ID, "screens[0].id"))
	assert.Equal(t, "IBD 50", stringValue(t, screens[0].Name, "screens[0].name"))
	assert.Equal(t, "SCREEN", stringValue(t, screens[0].Type, "screens[0].type"))
	assert.Equal(t, "marketsurge", stringValue(t, screens[0].Site, "screens[0].site"))
	assert.Equal(t, "Top 50 stocks", stringValue(t, screens[0].Description, "screens[0].description"))
	assert.Equal(t, "1", stringValue(t, screens[0].Source.ID, "screens[0].source.id"))

	assert.Equal(t, "99", stringValue(t, screens[1].ID, "screens[1].id"))
	assert.Equal(t, "Growth 250", stringValue(t, screens[1].Name, "screens[1].name"))
	assert.Equal(t, "REPORT", stringValue(t, screens[1].Type, "screens[1].type"))
	assert.Equal(t, "MarketSurge", stringValue(t, screens[1].Source.Pub, "screens[1].source.pub"))
}

func TestReportsListEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"user":{"screens":[]}}}`)
	}))
	t.Cleanup(server.Close)

	client := reportsListClient(t, server)
	output, err := runReportsList(t, client)
	require.NoError(t, err, "ReportsListCmd.Run(empty) error = %v, want nil", err)
	assert.Equal(t, "[]\n", output, "ReportsListCmd.Run(empty) output = %q, want %q", output, "[]\n")
}

func TestReportsListAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	client := reportsListClient(t, server)
	output, err := runReportsList(t, client)
	require.Error(t, err, "ReportsListCmd.Run(auth error) error = nil, want non-nil")

	var authErr *mserrors.AuthenticationError
	require.ErrorAs(t, err, &authErr, "ReportsListCmd.Run(auth error) error type = %T, want *mserrors.AuthenticationError", err)
	assert.Empty(t, output, "ReportsListCmd.Run(auth error) stdout = %q, want empty", output)
}

func TestReportsListAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"errors":[{"message":"Not authorized","path":["user","screens"]}]}`)
	}))
	t.Cleanup(server.Close)

	client := reportsListClient(t, server)
	output, err := runReportsList(t, client)
	require.Error(t, err, "ReportsListCmd.Run(GraphQL error) error = nil, want non-nil")

	var apiErr *mserrors.APIError
	require.ErrorAs(t, err, &apiErr, "ReportsListCmd.Run(GraphQL error) error type = %T, want *mserrors.APIError", err)
	assert.Empty(t, output, "ReportsListCmd.Run(GraphQL error) stdout = %q, want empty", output)
}

func reportsListClient(t *testing.T, server *httptest.Server) *marketsurge.Client {
	t.Helper()

	client, err := marketsurge.NewClient(
		marketsurge.WithJWT("test-token"),
		marketsurge.WithHTTPClient(server.Client()),
		marketsurge.WithGraphQLURL(server.URL),
	)
	require.NoError(t, err, "marketsurge.NewClient(test server %q) error = %v, want nil", server.URL, err)
	return client
}

func runReportsList(t *testing.T, client *marketsurge.Client) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "os.Pipe() error = %v, want nil", err)
	t.Cleanup(func() {
		_ = r.Close()
	})

	os.Stdout = w
	runErr := (&agentcmd.ReportsListCmd{}).Run(client)
	closeErr := w.Close()
	os.Stdout = oldStdout
	require.NoError(t, closeErr, "stdout pipe Close() error = %v, want nil", closeErr)

	var output bytes.Buffer
	_, err = io.Copy(&output, r)
	require.NoError(t, err, "io.Copy(ReportsListCmd.Run stdout) error = %v, want nil", err)
	return output.String(), runErr
}

func stringValue(t *testing.T, got *string, field string) string {
	t.Helper()

	require.NotNil(t, got, "ReportsListCmd.Run(success) %s = nil, want non-nil", field)
	return *got
}
