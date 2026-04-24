package commands

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalogRunReportDispatch(t *testing.T) {
	t.Parallel()
	server, requests := newCatalogRunServer(t)
	defer server.Close()

	envelope := runCatalogRunCommand(t, testClient(t, server), "--kind", "report", "--report-id", "124")

	require.Len(t, *requests, 1)
	assert.Equal(t, "MarketDataAdhocScreen", (*requests)[0].OperationName)
	assert.Equal(t, float64(124), nestedMap(t, (*requests)[0].Variables, "includeSource", "screenId")["id"])
	assert.Equal(t, "report", envelope.Metadata["kind"])
	assert.Equal(t, float64(2), envelope.Metadata["total"])
	assert.Len(t, envelope.Data.Entries, 2)
	assert.Equal(t, "AAPL", envelope.Data.Entries[0]["symbol"])
	assert.Equal(t, float64(defaultCatalogRunLimit), envelope.Metadata["limit"])
	assert.Equal(t, float64(0), envelope.Metadata["offset"])
	assert.NotEmpty(t, envelope.Metadata["timestamp"])
	assert.Empty(t, envelope.Errors)
	assert.Nil(t, envelope.Error)
}

func TestCatalogRunWatchlistDispatch(t *testing.T) {
	t.Parallel()
	server, requests := newCatalogRunServer(t)
	defer server.Close()

	envelope := runCatalogRunCommand(t, testClient(t, server), "--kind", "watchlist", "--watchlist-id", "123")

	require.Len(t, *requests, 2)
	assert.Equal(t, "FlaggedSymbols", (*requests)[0].OperationName)
	assert.Equal(t, "123", (*requests)[0].Variables["watchlistId"])
	assert.Equal(t, "MarketDataAdhocScreen", (*requests)[1].OperationName)
	assert.Equal(t, []any{"DJ-AAPL", "DJ-MSFT"}, nestedMap(t, (*requests)[1].Variables, "includeSource", "instruments")["symbols"])
	assert.Equal(t, "watchlist", envelope.Metadata["kind"])
	assert.Equal(t, float64(2), envelope.Metadata["total"])
	assert.Len(t, envelope.Data.Entries, 2)
	assert.Equal(t, "AAPL", envelope.Data.Entries[0]["symbol"])
	assert.Empty(t, envelope.Errors)
	assert.Nil(t, envelope.Error)
}

func TestCatalogRunCoachScreenDispatch(t *testing.T) {
	t.Parallel()
	server, requests := newCatalogRunServer(t)
	defer server.Close()

	envelope := runCatalogRunCommand(t, testClient(t, server), "--kind", "coach_screen", "--coach-screen-id", "screen-1")

	require.Len(t, *requests, 1)
	assert.Equal(t, "RunScreen", (*requests)[0].OperationName)
	assert.Equal(t, "screen-1", nestedMap(t, (*requests)[0].Variables, "input")["screenId"])
	assert.Equal(t, "coach_screen", envelope.Metadata["kind"])
	assert.Equal(t, float64(2), envelope.Metadata["total"])
	assert.Len(t, envelope.Data.Entries, 2)
	assert.Equal(t, "AAPL", envelope.Data.Entries[0]["Symbol"])
	assert.Empty(t, envelope.Errors)
	assert.Nil(t, envelope.Error)
}

func TestCatalogRunMissingKind(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()

	var buf bytes.Buffer
	cmd := CatalogRunCommand(testClient(t, server), &buf)
	err := runTestCommand(t, cmd, "run")
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "kind is required")
	assert.Empty(t, buf.String())
}

func TestCatalogRunScreenKindValidation(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()

	var buf bytes.Buffer
	cmd := CatalogRunCommand(testClient(t, server), &buf)
	err := runTestCommand(t, cmd, "run", "--kind", "screen")
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "screens cannot be run directly")
	assert.Empty(t, buf.String())
}

func TestCatalogRunMissingIDForKind(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()

	tests := []struct {
		name    string
		args    []string
		message string
	}{
		{name: "report", args: []string{"run", "--kind", "report"}, message: "report-id is required when kind=report"},
		{name: "watchlist", args: []string{"run", "--kind", "watchlist"}, message: "watchlist-id is required when kind=watchlist"},
		{name: "coach_screen", args: []string{"run", "--kind", "coach_screen"}, message: "coach-screen-id is required when kind=coach_screen"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
			cmd := CatalogRunCommand(testClient(t, server), &buf)
			err := runTestCommand(t, cmd, tt.args...)
			require.Error(t, err)

			var verr *mserrors.ValidationError
			assert.ErrorAs(t, err, &verr)
			assert.Contains(t, err.Error(), tt.message)
			assert.Empty(t, buf.String())
})
	}
}

func TestCatalogRunPagination(t *testing.T) {
	t.Parallel()
	server, _ := newCatalogRunServer(t)
	defer server.Close()

	envelope := runCatalogRunCommand(
		t,
		testClient(t, server),
		"--kind", "report",
		"--report-id", "124",
		"--limit", "1",
		"--offset", "1",
	)

	assert.Equal(t, float64(2), envelope.Metadata["total"])
	assert.Equal(t, float64(1), envelope.Metadata["limit"])
	assert.Equal(t, float64(1), envelope.Metadata["offset"])
	require.Len(t, envelope.Data.Entries, 1)
	assert.Equal(t, "MSFT", envelope.Data.Entries[0]["symbol"])
}

func TestCatalogRunWatchlistID64Bit(t *testing.T) {
	t.Parallel()
	server, requests := newCatalogRunServer(t)
	defer server.Close()

	maxID := int64(math.MaxInt64)
	envelope := runCatalogRunCommand(t, testClient(t, server), "--kind", "watchlist", "--watchlist-id", "9223372036854775807")

	require.Len(t, *requests, 2)
	assert.Equal(t, "FlaggedSymbols", (*requests)[0].OperationName)
	assert.Equal(t, "9223372036854775807", (*requests)[0].Variables["watchlistId"])
	assert.NotZero(t, maxID)
	assert.Equal(t, "watchlist", envelope.Metadata["kind"])
	assert.Len(t, envelope.Data.Entries, 2)
}

type catalogRunEnvelope struct {
	Data struct {
		Entries []map[string]any `json:"entries"`
	} `json:"data"`
	Errors   []string       `json:"errors"`
	Metadata map[string]any `json:"metadata"`
	Error    map[string]any `json:"error"`
}

type catalogRunRequest struct {
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
}

// runCatalogRunCommand executes the catalog run command and decodes its JSON response.
func runCatalogRunCommand(t *testing.T, c *client.Client, args ...string) catalogRunEnvelope {
	t.Helper()

	var buf bytes.Buffer
	cmd := CatalogRunCommand(c, &buf)
	argv := append([]string{"run"}, args...)
	require.NoError(t, runTestCommand(t, cmd, argv...))

	var envelope catalogRunEnvelope
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope))
	return envelope
}

// newCatalogRunServer builds an httptest server for report, watchlist, and coach screen flows.
func newCatalogRunServer(t *testing.T) (*httptest.Server, *[]catalogRunRequest) {
	t.Helper()

	requests := []catalogRunRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var payload catalogRunRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		requests = append(requests, payload)

		w.Header().Set("Content-Type", "application/json")
		switch payload.OperationName {
		case "FlaggedSymbols":
			_, _ = w.Write([]byte(`{"data":{"watchlist":{"items":[{"dowJonesKey":"DJ-AAPL"},{"dowJonesKey":"DJ-MSFT"}]}}}`))
		case "MarketDataAdhocScreen":
			_, _ = w.Write([]byte(catalogRunAdhocFixture()))
		case "RunScreen":
			_, _ = w.Write([]byte(catalogRunScreenFixture()))
		default:
			t.Fatalf("unexpected operation %s", payload.OperationName)
		}
	}))

	return server, &requests
}

// nestedMap walks a decoded JSON object through successive keys.
func nestedMap(t *testing.T, value map[string]any, keys ...string) map[string]any {
	t.Helper()

	current := value
	for _, key := range keys {
		next, ok := current[key].(map[string]any)
		require.Truef(t, ok, "key %s was not a JSON object", key)
		current = next
	}
	return current
}

// catalogRunAdhocFixture returns a minimal adhoc screen response with two rows.
func catalogRunAdhocFixture() string {
	return `{
		"data": {
			"marketDataAdhocScreen": {
				"responseValues": [
					[
						{"mdItem":{"name":"Symbol"},"value":"AAPL"},
						{"mdItem":{"name":"CompositeRating"},"value":99},
						{"mdItem":{"name":"RSRating"},"value":95},
						{"mdItem":{"name":"DowJonesInstrumentSubType"},"value":"COMMON"}
					],
					[
						{"mdItem":{"name":"Symbol"},"value":"MSFT"},
						{"mdItem":{"name":"CompositeRating"},"value":97},
						{"mdItem":{"name":"RSRating"},"value":90},
						{"mdItem":{"name":"DowJonesInstrumentSubType"},"value":"COMMON"}
					]
				],
				"errorValues": []
			}
		}
	}`
}

// catalogRunScreenFixture returns a minimal coach screen response with two rows.
func catalogRunScreenFixture() string {
	return `{
		"data": {
			"user": {
				"runScreen": {
					"numberOfMatchingInstruments": 2,
					"responseValues": [
						[
							{"mdItem":{"name":"Symbol"},"value":"AAPL"},
							{"mdItem":{"name":"CompositeRating"},"value":"99"}
						],
						[
							{"mdItem":{"name":"Symbol"},"value":"MSFT"},
							{"mdItem":{"name":"CompositeRating"},"value":"97"}
						]
					]
				}
			}
		}
	}`
}
