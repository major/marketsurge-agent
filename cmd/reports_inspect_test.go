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

func TestReportsInspectSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"user":{"screen":{"id":"screen-1","name":"Top RS Stocks","site":"marketsurge","description":"Stocks with top relative strength","filterCriteria":{"terms":[{"left":{"name":"RSRating"},"operand":"GREATER_THAN_OR_EQUAL","right":{"value":"90"}},{"left":{"name":"CompositeRating"},"operand":"GREATER_THAN_OR_EQUAL","right":{"value":"85"}}],"type":"AND"},"resultConfig":{"limit":250,"sortBy":{"field":"RSRating","direction":"DESC"}},"result":{"count":45,"description":"45 matches","updatedAt":"2025-05-08T12:00:00Z"},"type":"CUSTOM","createdAt":"2025-01-15T08:00:00Z","updatedAt":"2025-05-08T12:00:00Z"}}}}`)
	}))
	t.Cleanup(server.Close)

	client := reportsInspectClient(t, server)
	output, err := runReportsInspect(t, client, agentcmd.ReportsInspectCmd{ScreenID: "screen-1"})
	require.NoError(t, err, "ReportsInspectCmd.Run(success) error = %v, want nil", err)

	var items []struct {
		ID          *string `json:"id"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Type        *string `json:"type"`
		Filters     []struct {
			Field   *string `json:"field"`
			Operand *string `json:"operand"`
			Value   *string `json:"value"`
		} `json:"filters"`
		FilterType  *string `json:"filterType"`
		ResultLimit *int    `json:"resultLimit"`
		SortBy      *struct {
			Field     *string `json:"field"`
			Direction *string `json:"direction"`
		} `json:"sortBy"`
		LastResult *struct {
			Count       *int    `json:"count"`
			Description *string `json:"description"`
			UpdatedAt   *string `json:"updatedAt"`
		} `json:"lastResult"`
		CreatedAt *string `json:"createdAt"`
		UpdatedAt *string `json:"updatedAt"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items), "json.Unmarshal(ReportsInspectCmd.Run(success) output)")
	require.Len(t, items, 1, "ReportsInspectCmd.Run(success) decoded items length = %d, want %d", len(items), 1)

	item := items[0]
	require.NotNil(t, item.ID)
	assert.Equal(t, "screen-1", *item.ID)
	require.NotNil(t, item.Name)
	assert.Equal(t, "Top RS Stocks", *item.Name)
	require.NotNil(t, item.Description)
	assert.Equal(t, "Stocks with top relative strength", *item.Description)
	require.NotNil(t, item.Type)
	assert.Equal(t, "CUSTOM", *item.Type)
	require.Len(t, item.Filters, 2)
	require.NotNil(t, item.Filters[0].Field)
	assert.Equal(t, "RSRating", *item.Filters[0].Field)
	require.NotNil(t, item.Filters[0].Operand)
	assert.Equal(t, "GREATER_THAN_OR_EQUAL", *item.Filters[0].Operand)
	require.NotNil(t, item.Filters[0].Value)
	assert.Equal(t, "90", *item.Filters[0].Value)
	require.NotNil(t, item.Filters[1].Field)
	assert.Equal(t, "CompositeRating", *item.Filters[1].Field)
	require.NotNil(t, item.Filters[1].Operand)
	assert.Equal(t, "GREATER_THAN_OR_EQUAL", *item.Filters[1].Operand)
	require.NotNil(t, item.Filters[1].Value)
	assert.Equal(t, "85", *item.Filters[1].Value)
	require.NotNil(t, item.FilterType)
	assert.Equal(t, "AND", *item.FilterType)
	require.NotNil(t, item.ResultLimit)
	assert.Equal(t, 250, *item.ResultLimit)
	require.NotNil(t, item.SortBy)
	require.NotNil(t, item.SortBy.Field)
	assert.Equal(t, "RSRating", *item.SortBy.Field)
	require.NotNil(t, item.SortBy.Direction)
	assert.Equal(t, "DESC", *item.SortBy.Direction)
	require.NotNil(t, item.LastResult)
	require.NotNil(t, item.LastResult.Count)
	assert.Equal(t, 45, *item.LastResult.Count)
	require.NotNil(t, item.LastResult.Description)
	assert.Equal(t, "45 matches", *item.LastResult.Description)
	require.NotNil(t, item.LastResult.UpdatedAt)
	assert.Equal(t, "2025-05-08T12:00:00Z", *item.LastResult.UpdatedAt)
	require.NotNil(t, item.CreatedAt)
	assert.Equal(t, "2025-01-15T08:00:00Z", *item.CreatedAt)
	require.NotNil(t, item.UpdatedAt)
	assert.Equal(t, "2025-05-08T12:00:00Z", *item.UpdatedAt)
}

func TestReportsInspectNilResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{}}`)
	}))
	t.Cleanup(server.Close)

	client := reportsInspectClient(t, server)
	output, err := runReportsInspect(t, client, agentcmd.ReportsInspectCmd{ScreenID: "screen-1"})
	require.NoError(t, err, "ReportsInspectCmd.Run(nil response) error = %v, want nil", err)

	var items []struct {
		Filters []struct{} `json:"filters"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items), "json.Unmarshal(ReportsInspectCmd.Run(nil response) output)")
	require.Len(t, items, 1, "ReportsInspectCmd.Run(nil response) decoded items length = %d, want %d", len(items), 1)
	assert.Empty(t, items[0].Filters, "ReportsInspectCmd.Run(nil response) filters = %v, want empty", items[0].Filters)
}

func TestReportsInspectAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	client := reportsInspectClient(t, server)
	output, err := runReportsInspect(t, client, agentcmd.ReportsInspectCmd{ScreenID: "screen-1"})
	require.Error(t, err, "ReportsInspectCmd.Run(auth error) error = nil, want non-nil")

	var authErr *mserrors.AuthenticationError
	require.ErrorAs(t, err, &authErr, "ReportsInspectCmd.Run(auth error) error type = %T, want *mserrors.AuthenticationError", err)
	assert.Empty(t, output, "ReportsInspectCmd.Run(auth error) stdout = %q, want empty", output)
}

func TestReportsInspectAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"errors":[{"message":"Service unavailable","path":["screen"]}]}`)
	}))
	t.Cleanup(server.Close)

	client := reportsInspectClient(t, server)
	output, err := runReportsInspect(t, client, agentcmd.ReportsInspectCmd{ScreenID: "screen-1"})
	require.Error(t, err, "ReportsInspectCmd.Run(API error) error = nil, want non-nil")

	var apiErr *mserrors.APIError
	require.ErrorAs(t, err, &apiErr, "ReportsInspectCmd.Run(API error) error type = %T, want *mserrors.APIError", err)
	assert.Empty(t, output, "ReportsInspectCmd.Run(API error) stdout = %q, want empty", output)
}

func TestReportsInspectCoachFlag(t *testing.T) {
	var requestBody []byte
	var requestErr error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestBody, requestErr = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"user":{"screen":{}}}}`)
	}))
	t.Cleanup(server.Close)

	client := reportsInspectClient(t, server)
	output, err := runReportsInspect(t, client, agentcmd.ReportsInspectCmd{ScreenID: "coach-1", Coach: true})
	require.NoError(t, err, "ReportsInspectCmd.Run(coach flag) error = %v, want nil", err)
	require.NoError(t, requestErr, "ReportsInspectCmd.Run(coach flag) request body read error = %v, want nil", requestErr)
	assert.Contains(t, string(requestBody), `"coachScreen":true`, "ReportsInspectCmd.Run(coach flag) request body = %s, want coachScreen true", string(requestBody))
	assert.NotEmpty(t, output, "ReportsInspectCmd.Run(coach flag) stdout = %q, want output", output)
}

func reportsInspectClient(t *testing.T, server *httptest.Server) *marketsurge.Client {
	t.Helper()

	client, err := marketsurge.NewClient(
		marketsurge.WithJWT("test-token"),
		marketsurge.WithHTTPClient(server.Client()),
		marketsurge.WithGraphQLURL(server.URL),
	)
	require.NoError(t, err, "marketsurge.NewClient(test server %q) error = %v, want nil", server.URL, err)
	return client
}

func runReportsInspect(t *testing.T, client *marketsurge.Client, cmd agentcmd.ReportsInspectCmd) (string, error) {
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
	require.NoError(t, err, "io.Copy(ReportsInspectCmd.Run stdout) error = %v, want nil", err)
	return output.String(), runErr
}
