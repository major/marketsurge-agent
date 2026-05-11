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

func TestReportsGetSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"user":{"runScreen":{"numberOfMatchingInstruments":2,"responseValues":[[{"value":"AAPL","mdItem":{"name":"Symbol","mdItemID":1}},{"value":"213.50","mdItem":{"name":"Price","mdItemID":2}},{"value":"95","mdItem":{"name":"CompositeRating","mdItemID":3}}],[{"value":"NVDA","mdItem":{"name":"Symbol","mdItemID":1}},{"value":"950.00","mdItem":{"name":"Price","mdItemID":2}},{"value":"99","mdItem":{"name":"CompositeRating","mdItemID":3}}]]}}}}`)
	}))
	t.Cleanup(server.Close)

	client := reportsGetClient(t, server)
	output, err := runReportsGet(t, client, agentcmd.ReportsGetCmd{
		ScreenID: "screen-1",
		Columns:  []string{"Symbol", "Price", "CompositeRating"},
	})
	require.NoError(t, err, "ReportsGetCmd.Run(success) error = %v, want nil", err)

	var rows []map[string]string
	require.NoError(t, json.Unmarshal([]byte(output), &rows), "json.Unmarshal(ReportsGetCmd.Run(success) output)")
	require.Len(t, rows, 2, "ReportsGetCmd.Run(success) decoded rows length = %d, want %d", len(rows), 2)
	assert.Equal(t, "AAPL", rows[0]["Symbol"])
	assert.Equal(t, "213.50", rows[0]["Price"])
	assert.Equal(t, "95", rows[0]["CompositeRating"])
	assert.Equal(t, "NVDA", rows[1]["Symbol"])
	assert.Equal(t, "950.00", rows[1]["Price"])
	assert.Equal(t, "99", rows[1]["CompositeRating"])
}

func TestReportsGetCustomColumns(t *testing.T) {
	var requestScreenID string
	var requestColumns []string
	var requestErr error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Variables struct {
				Input struct {
					ScreenID        string `json:"screenID"`
					ResponseColumns []struct {
						Name string `json:"name"`
					} `json:"responseColumns"`
				} `json:"input"`
			} `json:"variables"`
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			requestErr = err
			http.Error(w, "read request body", http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(body, &req); err != nil {
			requestErr = err
			http.Error(w, "decode request body", http.StatusInternalServerError)
			return
		}
		requestScreenID = req.Variables.Input.ScreenID
		for _, col := range req.Variables.Input.ResponseColumns {
			requestColumns = append(requestColumns, col.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"user":{"runScreen":{"numberOfMatchingInstruments":1,"responseValues":[[{"value":"AAPL","mdItem":{"name":"Symbol","mdItemID":1}},{"value":"213.50","mdItem":{"name":"Price","mdItemID":2}}]]}}}}`)
	}))
	t.Cleanup(server.Close)

	client := reportsGetClient(t, server)
	_, err := runReportsGet(t, client, agentcmd.ReportsGetCmd{
		ScreenID: " screen-1 ",
		Columns:  []string{"Symbol", " Price "},
	})
	require.NoError(t, err, "ReportsGetCmd.Run(custom columns) error = %v, want nil", err)
	require.NoError(t, requestErr, "RunScreen request capture error = %v, want nil", requestErr)
	assert.Equal(t, "screen-1", requestScreenID)
	assert.Equal(t, []string{"Symbol", "Price"}, requestColumns)
}

func TestReportsGetValidationError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)

	client := reportsGetClient(t, server)
	output, err := runReportsGet(t, client, agentcmd.ReportsGetCmd{
		ScreenID: " ",
		Columns:  []string{"Symbol"},
	})
	require.Error(t, err, "ReportsGetCmd.Run(blank screen ID) error = nil, want non-nil")

	var validationErr *mserrors.ValidationError
	require.ErrorAs(t, err, &validationErr, "ReportsGetCmd.Run(blank screen ID) error type = %T, want *mserrors.ValidationError", err)
	assert.Empty(t, output, "ReportsGetCmd.Run(blank screen ID) stdout = %q, want empty", output)
}

func TestReportsGetBlankColumnValidationError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)

	client := reportsGetClient(t, server)
	output, err := runReportsGet(t, client, agentcmd.ReportsGetCmd{
		ScreenID: "screen-1",
		Columns:  []string{"Symbol", " "},
	})
	require.Error(t, err, "ReportsGetCmd.Run(blank column) error = nil, want non-nil")

	var validationErr *mserrors.ValidationError
	require.ErrorAs(t, err, &validationErr, "ReportsGetCmd.Run(blank column) error type = %T, want *mserrors.ValidationError", err)
	assert.Empty(t, output, "ReportsGetCmd.Run(blank column) stdout = %q, want empty", output)
}

func TestReportsGetEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"user":{"runScreen":{"numberOfMatchingInstruments":0,"responseValues":[]}}}}`)
	}))
	t.Cleanup(server.Close)

	client := reportsGetClient(t, server)
	output, err := runReportsGet(t, client, agentcmd.ReportsGetCmd{
		ScreenID: "screen-1",
		Columns:  []string{"Symbol", "Price"},
	})
	require.NoError(t, err, "ReportsGetCmd.Run(empty) error = %v, want nil", err)
	assert.Equal(t, "[]\n", output, "ReportsGetCmd.Run(empty) output = %q, want %q", output, "[]\n")
}

func TestReportsGetNilResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{}}`)
	}))
	t.Cleanup(server.Close)

	client := reportsGetClient(t, server)
	output, err := runReportsGet(t, client, agentcmd.ReportsGetCmd{
		ScreenID: "screen-1",
		Columns:  []string{"Symbol", "Price"},
	})
	require.NoError(t, err, "ReportsGetCmd.Run(nil response) error = %v, want nil", err)
	assert.Equal(t, "[]\n", output, "ReportsGetCmd.Run(nil response) output = %q, want %q", output, "[]\n")
}

func TestReportsGetAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	client := reportsGetClient(t, server)
	output, err := runReportsGet(t, client, agentcmd.ReportsGetCmd{
		ScreenID: "screen-1",
		Columns:  []string{"Symbol", "Price"},
	})
	require.Error(t, err, "ReportsGetCmd.Run(auth error) error = nil, want non-nil")

	var authErr *mserrors.AuthenticationError
	require.ErrorAs(t, err, &authErr, "ReportsGetCmd.Run(auth error) error type = %T, want *mserrors.AuthenticationError", err)
	assert.Empty(t, output, "ReportsGetCmd.Run(auth error) stdout = %q, want empty", output)
}

func reportsGetClient(t *testing.T, server *httptest.Server) *marketsurge.Client {
	t.Helper()

	client, err := marketsurge.NewClient(
		marketsurge.WithJWT("test-token"),
		marketsurge.WithHTTPClient(server.Client()),
		marketsurge.WithGraphQLURL(server.URL),
	)
	require.NoError(t, err, "marketsurge.NewClient(test server %q) error = %v, want nil", server.URL, err)
	return client
}

func runReportsGet(t *testing.T, client *marketsurge.Client, cmd agentcmd.ReportsGetCmd) (string, error) {
	t.Helper()

	var output bytes.Buffer
	runErr := cmd.RunForTest(client, &output)
	return output.String(), runErr
}
