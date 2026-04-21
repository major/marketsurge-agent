package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/major/marketsurge-agent/internal/client"
	"github.com/major/marketsurge-agent/internal/constants"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestCatalogListAllSourcesSucceed(t *testing.T) {
	server := newCatalogServer(t, func(op string, w http.ResponseWriter) {
		switch op {
		case "GetAllWatchlistNames":
			_, _ = w.Write([]byte(`{"data":{"watchlists":[{"id":"99","name":"My Watchlist","description":"desc"}]}}`))
		case "Screens":
			_, _ = w.Write([]byte(`{"data":{"user":{"screens":[{"name":"Saved Screen","description":"screen desc"}]}}}`))
		case "CoachTree":
			_, _ = w.Write([]byte(`{"data":{"user":{"screens":[{"name":"Coach Alpha","referenceId":"{\"screenId\":\"screen-1\"}"}],"watchlists":[{"name":"Coach WL","referenceId":"{\"watchlistId\":\"123\"}"}]}}}`))
		default:
			t.Fatalf("unexpected operation %s", op)
		}
	})
	defer server.Close()

	envelope := runCatalogListCommand(t, testClient(t, server))
	entries := catalogEntriesFromEnvelope(t, envelope)

	assert.Empty(t, envelope.Errors)
	assert.Equal(t, float64(len(entries)), envelope.Metadata["total"])
	assertCatalogEntrySubset(t, entries, map[string]any{"kind": "watchlist", "name": "My Watchlist", "watchlist_id": float64(99)})
	assertCatalogEntrySubset(t, entries, map[string]any{"kind": "screen", "name": "Saved Screen", "description": "screen desc"})
	assertCatalogEntrySubset(t, entries, map[string]any{"kind": "coach_screen", "name": "Coach Alpha", "coach_screen_id": "screen-1"})
	assertCatalogEntrySubset(t, entries, map[string]any{"kind": "report", "name": constants.PredefinedReports[0].Name, "report_id": float64(constants.PredefinedReports[0].ID)})
}

func TestCatalogListPartialFailure(t *testing.T) {
	server := newCatalogServer(t, func(op string, w http.ResponseWriter) {
		switch op {
		case "GetAllWatchlistNames":
			_, _ = w.Write([]byte(`{"data":{"watchlists":[{"id":"99","name":"My Watchlist"}]}}`))
		case "Screens":
			_, _ = w.Write([]byte(`{"errors":[{"message":"screens failed"}]}`))
		case "CoachTree":
			_, _ = w.Write([]byte(`{"data":{"user":{"screens":[{"name":"Coach Alpha","referenceId":"{\"screenId\":\"screen-1\"}"}],"watchlists":[]}}}`))
		default:
			t.Fatalf("unexpected operation %s", op)
		}
	})
	defer server.Close()

	envelope := runCatalogListCommand(t, testClient(t, server))
	entries := catalogEntriesFromEnvelope(t, envelope)

	require.NotEmpty(t, envelope.Errors)
	assert.Contains(t, envelope.Errors[0], "screens failed")
	assert.NotEmpty(t, entries)
	assert.Equal(t, float64(len(entries)), envelope.Metadata["total"])
	assertCatalogEntrySubset(t, entries, map[string]any{"kind": "watchlist", "name": "My Watchlist", "watchlist_id": float64(99)})
	assertCatalogEntrySubset(t, entries, map[string]any{"kind": "coach_screen", "name": "Coach Alpha", "coach_screen_id": "screen-1"})
	assertCatalogEntrySubset(t, entries, map[string]any{"kind": "report", "name": constants.PredefinedReports[0].Name, "report_id": float64(constants.PredefinedReports[0].ID)})
}

func TestCatalogListKindFilter(t *testing.T) {
	server := newCatalogServer(t, func(op string, w http.ResponseWriter) {
		switch op {
		case "GetAllWatchlistNames":
			_, _ = w.Write([]byte(`{"data":{"watchlists":[{"id":"99","name":"My Watchlist"}]}}`))
		case "Screens":
			_, _ = w.Write([]byte(`{"data":{"user":{"screens":[{"name":"Saved Screen"}]}}}`))
		case "CoachTree":
			_, _ = w.Write([]byte(`{"data":{"user":{"screens":[{"name":"Coach Alpha","referenceId":"{\"screenId\":\"screen-1\"}"}],"watchlists":[]}}}`))
		default:
			t.Fatalf("unexpected operation %s", op)
		}
	})
	defer server.Close()

	envelope := runCatalogListCommand(t, testClient(t, server), "--kind", string(models.CatalogKindCoachScreen))
	entries := catalogEntriesFromEnvelope(t, envelope)

	assert.Empty(t, envelope.Errors)
	assert.Equal(t, string(models.CatalogKindCoachScreen), envelope.Metadata["kind"])
	require.Len(t, entries, 1)
	assert.Equal(t, map[string]any{"kind": "coach_screen", "name": "Coach Alpha", "coach_screen_id": "screen-1"}, entries[0])
}

func TestCatalogListAllAPISourcesFailStillReturnsReports(t *testing.T) {
	server := newCatalogServer(t, func(op string, w http.ResponseWriter) {
		switch op {
		case "GetAllWatchlistNames":
			_, _ = w.Write([]byte(`{"errors":[{"message":"watchlists failed"}]}`))
		case "Screens":
			_, _ = w.Write([]byte(`{"errors":[{"message":"screens failed"}]}`))
		case "CoachTree":
			_, _ = w.Write([]byte(`{"errors":[{"message":"coach tree failed"}]}`))
		default:
			t.Fatalf("unexpected operation %s", op)
		}
	})
	defer server.Close()

	envelope := runCatalogListCommand(t, testClient(t, server))
	entries := catalogEntriesFromEnvelope(t, envelope)

	require.Len(t, envelope.Errors, 3)
	assert.Len(t, entries, len(constants.PredefinedReports))
	assert.Equal(t, float64(len(constants.PredefinedReports)), envelope.Metadata["total"])
	assert.Equal(t, map[string]any{"kind": "report", "name": constants.PredefinedReports[0].Name, "report_id": float64(constants.PredefinedReports[0].ID)}, entries[0])
	for _, entry := range entries {
		assert.Equal(t, "report", entry["kind"])
	}
}

func TestCatalogListInvalidKind(t *testing.T) {
	server := jsonServer(`{}`)
	defer server.Close()

	var buf bytes.Buffer
	cmd := CatalogListCommand(testClient(t, server), &buf)
	cmd.ExitErrHandler = func(_ context.Context, _ *cli.Command, _ error) {}

	err := cmd.Run(context.Background(), []string{"list", "--kind", "invalid"})
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "kind must be one of")
	assert.Empty(t, buf.String())
}

type catalogListEnvelope struct {
	Data struct {
		Entries []map[string]any `json:"entries"`
	} `json:"data"`
	Errors   []string       `json:"errors"`
	Metadata map[string]any `json:"metadata"`
}

func runCatalogListCommand(t *testing.T, c *client.Client, args ...string) catalogListEnvelope {
	t.Helper()

	var buf bytes.Buffer
	cmd := CatalogListCommand(c, &buf)
	cmd.ExitErrHandler = func(_ context.Context, _ *cli.Command, _ error) {}

	argv := append([]string{"list"}, args...)
	require.NoError(t, cmd.Run(context.Background(), argv))

	var envelope catalogListEnvelope
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope))
	return envelope
}

func catalogEntriesFromEnvelope(t *testing.T, envelope catalogListEnvelope) []map[string]any {
	t.Helper()
	require.NotNil(t, envelope.Metadata)
	assert.Equal(t, float64(0), envelope.Metadata["limit"])
	assert.Equal(t, float64(0), envelope.Metadata["offset"])
	assert.NotEmpty(t, envelope.Metadata["timestamp"])
	return envelope.Data.Entries
}

func newCatalogServer(t *testing.T, handler func(op string, w http.ResponseWriter)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload struct {
			OperationName string `json:"operationName"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.Header().Set("Content-Type", "application/json")
		handler(payload.OperationName, w)
	}))
}

func assertCatalogEntrySubset(t *testing.T, entries []map[string]any, expected map[string]any) {
	t.Helper()
	for _, entry := range entries {
		matched := true
		for key, want := range expected {
			if entry[key] != want {
				matched = false
				break
			}
		}
		if matched {
			return
		}
	}
	t.Fatalf("no catalog entry matched subset %#v", expected)
}
