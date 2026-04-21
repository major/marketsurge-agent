package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/major/marketsurge-agent/internal/constants"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunReportSuccess(t *testing.T) {
	var captured Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_, _ = w.Write([]byte(adhocResponseJSON()))
	}))
	defer server.Close()

	client := NewClient("jwt-token")
	client.Endpoint = server.URL
	client.HTTPClient = server.Client()

	data, err := client.RunReport(context.Background(), 124)
	require.NoError(t, err)
	assert.Equal(t, "MarketDataAdhocScreen", captured.OperationName)
	assert.Len(t, data.Entries, 1)
	assert.Equal(t, "AAPL", *data.Entries[0].Symbol)
}

func TestRunWatchlistTwoStepFlow(t *testing.T) {
	requests := []Request{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload Request
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		requests = append(requests, payload)
		if payload.OperationName == "FlaggedSymbols" {
			_, _ = w.Write([]byte(`{"data":{"watchlist":{"id":"123","items":[{"dowJonesKey":"DJ:1"},{"dowJonesKey":"DJ:2"}]}}}`))
			return
		}
		_, _ = w.Write([]byte(adhocResponseJSON()))
	}))
	defer server.Close()

	client := NewClient("jwt-token")
	client.Endpoint = server.URL
	client.HTTPClient = server.Client()

	data, err := client.RunWatchlist(context.Background(), 922337203685477580)
	require.NoError(t, err)
	require.Len(t, requests, 2)
	assert.Equal(t, "FlaggedSymbols", requests[0].OperationName)
	assert.Equal(t, "MarketDataAdhocScreen", requests[1].OperationName)
	includeSource := requests[1].Variables["includeSource"].(map[string]any)
	instruments := includeSource["instruments"].(map[string]any)
	assert.Equal(t, "DJ_KEY", instruments["dialect"])
	assert.Len(t, data.Entries, 1)
}

func TestRunWatchlistReturnsEmptyForNoSymbols(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"watchlist":{"id":"123","items":[]}}}`))
	}))
	defer server.Close()

	client := NewClient("jwt-token")
	client.Endpoint = server.URL
	client.HTTPClient = server.Client()

	data, err := client.RunWatchlist(context.Background(), 123)
	require.NoError(t, err)
	assert.Empty(t, data.Entries)
}

func TestRunWatchlistReturnsNotFoundWhenMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"watchlist":null}}`))
	}))
	defer server.Close()

	client := NewClient("jwt-token")
	client.Endpoint = server.URL
	client.HTTPClient = server.Client()

	_, err := client.RunWatchlist(context.Background(), 123)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "watchlist not found")
}

func TestRunCoachScreenSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"user":{"runScreen":{"numberOfMatchingInstruments":2,"responseValues":[[{"mdItem":{"name":"Symbol"},"value":"AAPL"}]]}}}}`))
	}))
	defer server.Close()

	client := NewClient("jwt-token")
	client.Endpoint = server.URL
	client.HTTPClient = server.Client()

	data, err := client.RunCoachScreen(context.Background(), "screen-1")
	require.NoError(t, err)
	assert.Equal(t, 2, *data.NumInstruments)
	assert.Equal(t, "AAPL", *data.Rows[0]["Symbol"])
}

func TestRunCoachScreenReturnsAPIErrorForMissingPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"user":{}}}`))
	}))
	defer server.Close()

	client := NewClient("jwt-token")
	client.Endpoint = server.URL
	client.HTTPClient = server.Client()

	_, err := client.RunCoachScreen(context.Background(), "screen-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no coach screen data returned")
}

func TestListCatalogAggregatesAllSources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload Request
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		switch payload.OperationName {
		case "GetAllWatchlistNames":
			_, _ = w.Write([]byte(`{"data":{"watchlists":[{"id":"99","name":"My Watchlist","description":"desc"}]}}`))
		case "Screens":
			_, _ = w.Write([]byte(`{"data":{"user":{"screens":[{"name":"Saved Screen","description":"screen desc"}]}}}`))
		case "CoachTree":
			_, _ = w.Write([]byte(`{"data":{"user":{"screens":[{"name":"Coach Alpha","referenceId":"{\"screenId\":\"screen-1\"}"}],"watchlists":[{"name":"Coach WL","referenceId":"{\"watchlistId\":\"123\"}"}]}}}`))
		default:
			t.Fatalf("unexpected operation %s", payload.OperationName)
		}
	}))
	defer server.Close()

	client := NewClient("jwt-token")
	client.Endpoint = server.URL
	client.HTTPClient = server.Client()

	catalog, err := client.ListCatalog(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, catalog.Errors)
	assert.GreaterOrEqual(t, len(catalog.Entries), 4)
	assert.Contains(t, catalog.Entries, models.CatalogEntry{Name: "Saved Screen", Kind: models.CatalogKindScreen, Description: strptr("screen desc")})
}

func TestListCatalogReportFilterSkipsRemoteSources(t *testing.T) {
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload Request
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		requests = append(requests, payload.OperationName)
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer server.Close()

	client := NewClient("jwt-token")
	client.Endpoint = server.URL
	client.HTTPClient = server.Client()
	kind := models.CatalogKindReport

	catalog, err := client.ListCatalog(context.Background(), &kind)
	require.NoError(t, err)
	assert.Empty(t, requests)
	assert.Len(t, catalog.Errors, 0)
	assert.Len(t, catalog.Entries, len(modelsToReportEntries()))
}

func TestListCatalogPartialFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload Request
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		if payload.OperationName == "Screens" {
			_, _ = w.Write([]byte(`{"errors":[{"message":"screens failed"}]}`))
			return
		}
		if payload.OperationName == "GetAllWatchlistNames" {
			_, _ = w.Write([]byte(`{"data":{"watchlists":[{"id":"99","name":"My Watchlist"}]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"user":{"screens":[],"watchlists":[]}}}`))
	}))
	defer server.Close()

	client := NewClient("jwt-token")
	client.Endpoint = server.URL
	client.HTTPClient = server.Client()

	catalog, err := client.ListCatalog(context.Background(), nil)
	require.NoError(t, err)
	require.NotEmpty(t, catalog.Errors)
	assert.Contains(t, catalog.Errors[0], "screens failed")
	assert.NotEmpty(t, catalog.Entries)
}

func TestListCatalogFiltersKind(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload Request
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		switch payload.OperationName {
		case "GetAllWatchlistNames":
			_, _ = w.Write([]byte(`{"data":{"watchlists":[]}}`))
		case "Screens":
			_, _ = w.Write([]byte(`{"data":{"user":{"screens":[]}}}`))
		case "CoachTree":
			_, _ = w.Write([]byte(`{"data":{"user":{"screens":[],"watchlists":[]}}}`))
		}
	}))
	defer server.Close()

	client := NewClient("jwt-token")
	client.Endpoint = server.URL
	client.HTTPClient = server.Client()
	kind := models.CatalogKindReport

	catalog, err := client.ListCatalog(context.Background(), &kind)
	require.NoError(t, err)
	require.NotEmpty(t, catalog.Entries)
	for _, entry := range catalog.Entries {
		assert.Equal(t, models.CatalogKindReport, entry.Kind)
	}
	assert.Empty(t, catalog.Errors)
}

func TestParseAdhocScreenResultIncludesErrorValues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"marketDataAdhocScreen":{"responseValues":[],"errorValues":["bad symbol"]}}}`))
	}))
	defer server.Close()

	client := NewClient("jwt-token")
	client.Endpoint = server.URL
	client.HTTPClient = server.Client()

	data, err := client.RunReport(context.Background(), 124)
	require.NoError(t, err)
	assert.Equal(t, []string{"bad symbol"}, data.ErrorValues)
}

func TestParseAdhocScreenResultMapsFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(adhocResponseJSON()))
	}))
	defer server.Close()

	client := NewClient("jwt-token")
	client.Endpoint = server.URL
	client.HTTPClient = server.Client()

	data, err := client.RunReport(context.Background(), 124)
	require.NoError(t, err)
	require.Len(t, data.Entries, 1)
	assert.Equal(t, 99, *data.Entries[0].CompositeRating)
	assert.Equal(t, "DJ:1", *data.Entries[0].DowJonesKey)
	assert.Equal(t, "COMMON", *data.Entries[0].InstrumentSubType)
	assert.Nil(t, data.Entries[0].Volume)
	assert.Nil(t, data.Entries[0].VolumePctChange)
}

func adhocResponseJSON() string {
	return `{"data":{"marketDataAdhocScreen":{"responseValues":[[{"mdItem":{"name":"Symbol"},"value":"AAPL"},{"mdItem":{"name":"CompanyName"},"value":"Apple Inc."},{"mdItem":{"name":"ListRank"},"value":"1"},{"mdItem":{"name":"Price"},"value":"101.5"},{"mdItem":{"name":"PriceNetChg"},"value":"1.0"},{"mdItem":{"name":"PricePctChg"},"value":"0.9"},{"mdItem":{"name":"PricePctOff52WHigh"},"value":"5.0"},{"mdItem":{"name":"VolumeAvg50Day"},"value":"1000"},{"mdItem":{"name":"VolumePctChgVs50DAvgVolume"},"value":"10.0"},{"mdItem":{"name":"CompositeRating"},"value":"99"},{"mdItem":{"name":"EPSRating"},"value":"95"},{"mdItem":{"name":"RSRating"},"value":"94"},{"mdItem":{"name":"AccDisRating"},"value":"A"},{"mdItem":{"name":"SMRRating"},"value":"A"},{"mdItem":{"name":"IndustryGroupRank"},"value":"3"},{"mdItem":{"name":"IndustryName"},"value":"Technology"},{"mdItem":{"name":"MarketCapIntraday"},"value":"1000"},{"mdItem":{"name":"VolumeDollarAvg50D"},"value":"500"},{"mdItem":{"name":"IPODate"},"value":"1980-12-12"},{"mdItem":{"name":"DowJonesKey"},"value":"DJ:1"},{"mdItem":{"name":"ChartingSymbol"},"value":"AAPL"},{"mdItem":{"name":"DowJonesInstrumentType"},"value":"EQUITY"},{"mdItem":{"name":"DowJonesInstrumentSubType"},"value":"COMMON"}]],"errorValues":[]}}}`
}

func strptr(value string) *string {
	return &value
}

func modelsToReportEntries() []models.CatalogEntry {
	entries := make([]models.CatalogEntry, 0)
	for _, report := range constants.PredefinedReports {
		reportID := report.ID
		entries = append(entries, models.CatalogEntry{Name: report.Name, Kind: models.CatalogKindReport, ReportID: &reportID})
	}
	return entries
}
