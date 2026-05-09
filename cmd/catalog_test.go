package cmd

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/leodido/structcli"
	"github.com/spf13/cobra"

	"github.com/major/marketsurge-agent/internal/client"
	"github.com/major/marketsurge-agent/internal/constants"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalogListAllSourcesSucceed(t *testing.T) {
	t.Parallel()
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
	assertCatalogEntrySubset(t, entries, catalogReportEntryMap(constants.PredefinedReports[0]))
}

func TestCatalogListPartialFailure(t *testing.T) {
	t.Parallel()
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
	assertCatalogEntrySubset(t, entries, catalogReportEntryMap(constants.PredefinedReports[0]))
}

func TestCatalogListKindFilter(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	assert.Equal(t, catalogReportEntryMap(constants.PredefinedReports[0]), entries[0])
	for _, entry := range entries {
		assert.Equal(t, "report", entry["kind"])
	}
}

func TestCatalogListInvalidKind(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()

	c := testClient(t, server)
	_, err := executeCatalogList(t, c, "list", "--kind", "invalid")
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Equal(t, "invalid --kind \"invalid\": use one of watchlist, screen, report, coach_screen", err.Error())
}

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

	_, err := executeCatalogRun(t, testClient(t, server), "run")
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "missing --kind: use --kind watchlist --watchlist-id 12345, --kind report --report-id 67890, or --kind coach_screen --coach-screen-id ID")
}

func TestCatalogRunScreenKindValidation(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()

	_, err := executeCatalogRun(t, testClient(t, server), "run", "--kind", "screen")
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "invalid --kind screen for catalog run: screens are list-only; use catalog list --kind screen")
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
		{name: "report", args: []string{"run", "--kind", "report"}, message: "missing --report-id: --kind report requires --report-id 67890"},
		{name: "watchlist", args: []string{"run", "--kind", "watchlist"}, message: "missing --watchlist-id: --kind watchlist requires --watchlist-id 12345"},
		{name: "coach_screen", args: []string{"run", "--kind", "coach_screen"}, message: "missing --coach-screen-id: --kind coach_screen requires --coach-screen-id ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := executeCatalogRun(t, testClient(t, server), tt.args...)
			require.Error(t, err)

			var verr *mserrors.ValidationError
			assert.ErrorAs(t, err, &verr)
			assert.Contains(t, err.Error(), tt.message)
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

func TestClampCatalogOffset(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		offset   int
		length   int
		expected int
	}{
		{name: "zero offset", offset: 0, length: 10, expected: 0},
		{name: "positive offset within bounds", offset: 5, length: 10, expected: 5},
		{name: "negative offset clamps to zero", offset: -1, length: 10, expected: 0},
		{name: "large negative offset clamps to zero", offset: -100, length: 10, expected: 0},
		{name: "offset equals length", offset: 10, length: 10, expected: 10},
		{name: "offset exceeds length", offset: 15, length: 10, expected: 10},
		{name: "large offset clamps to length", offset: 1000, length: 10, expected: 10},
		{name: "zero length", offset: 0, length: 0, expected: 0},
		{name: "negative offset with zero length", offset: -5, length: 0, expected: 0},
		{name: "positive offset with zero length", offset: 5, length: 0, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := clampCatalogOffset(tt.offset, tt.length)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestPaginateSlice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		items    []string
		limit    int
		offset   int
		expected []string
	}{
		{name: "no pagination", items: []string{"a", "b", "c"}, limit: 0, offset: 0, expected: []string{"a", "b", "c"}},
		{name: "limit within bounds", items: []string{"a", "b", "c", "d"}, limit: 2, offset: 0, expected: []string{"a", "b"}},
		{name: "offset and limit", items: []string{"a", "b", "c", "d"}, limit: 2, offset: 1, expected: []string{"b", "c"}},
		{name: "negative offset clamps to zero", items: []string{"a", "b", "c"}, limit: 2, offset: -5, expected: []string{"a", "b"}},
		{name: "offset beyond length", items: []string{"a", "b", "c"}, limit: 2, offset: 10, expected: []string{}},
		{name: "negative limit returns full slice from offset", items: []string{"a", "b", "c"}, limit: -1, offset: 0, expected: []string{"a", "b", "c"}},
		{name: "empty slice", items: []string{}, limit: 10, offset: 0, expected: []string{}},
		{name: "offset at end of slice", items: []string{"a", "b", "c"}, limit: 10, offset: 3, expected: []string{}},
		{name: "limit exceeds remaining items", items: []string{"a", "b", "c"}, limit: 10, offset: 1, expected: []string{"b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := paginateSlice(tt.items, tt.limit, tt.offset)
			assert.Equal(t, tt.expected, got)
		})
	}
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

func TestCatalogRunFieldProjection(t *testing.T) {
	t.Parallel()
	server, _ := newCatalogRunServer(t)
	defer server.Close()

	envelope := runCatalogRunCommand(
		t,
		testClient(t, server),
		"--kind", "report",
		"--report-id", "124",
		"--fields", "symbol,composite-rating",
	)

	require.Len(t, envelope.Data.Entries, 2)
	assert.Equal(t, map[string]any{"symbol": "AAPL", "composite_rating": float64(99)}, envelope.Data.Entries[0])
	assert.Equal(t, map[string]any{"symbol": "MSFT", "composite_rating": float64(97)}, envelope.Data.Entries[1])
}

func TestCatalogRunBlankFieldsBehaveLikeUnset(t *testing.T) {
	t.Parallel()
	server, _ := newCatalogRunServer(t)
	defer server.Close()

	envelope := runCatalogRunCommand(
		t,
		testClient(t, server),
		"--kind", "report",
		"--report-id", "124",
		"--fields", "   ",
	)

	require.Len(t, envelope.Data.Entries, 2)
	assert.Contains(t, envelope.Data.Entries[0], "symbol")
	assert.Contains(t, envelope.Data.Entries[0], "composite_rating")
	assert.Contains(t, envelope.Data.Entries[0], "rs_rating")
}

func TestCatalogRunUnknownFieldsValidation(t *testing.T) {
	t.Parallel()
	server := jsonServer(`{}`)
	defer server.Close()

	_, err := executeCatalogRun(
		t,
		testClient(t, server),
		"run",
		"--kind", "report",
		"--report-id", "124",
		"--fields", "symbol,not_real",
	)
	require.Error(t, err)

	var verr *mserrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), `unknown --fields values "not_real"`)
	assert.Contains(t, err.Error(), "composite_rating")
}

func TestCatalogRunOptionsValidateNormalizesFields(t *testing.T) {
	t.Parallel()

	opts := CatalogRunOptions{Kind: "report", ReportID: 42, Fields: []string{" ", " symbol ", "\t"}}

	assert.Nil(t, opts.Validate(context.Background()))
	assert.Equal(t, []string{"symbol"}, opts.Fields)
}

func TestCatalogListStructTags(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[CatalogListOptions]()
	field, ok := typ.FieldByName("Kind")
	require.True(t, ok, "Kind field should exist")
	assert.Equal(t, "kind", field.Tag.Get("flag"))
	assert.Equal(t, "Filtering", field.Tag.Get("flaggroup"))
	assert.Equal(t, "Filter by catalog kind (watchlist, report, coach_screen, screen); omit to list all sources", field.Tag.Get("flagdescr"))
}

func TestCatalogRunStructTags(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[CatalogRunOptions]()
	tests := []struct {
		field        string
		flag         string
		group        string
		descr        string
		defaultValue string
	}{
		{field: "Kind", flag: "kind", group: "Catalog Selection", descr: "Required catalog kind to run: watchlist uses --watchlist-id, report uses --report-id, coach_screen uses --coach-screen-id; screens are list-only. Example report: --kind report --report-id 124"},
		{field: "ReportID", flag: "report-id", group: "Kind-Specific IDs", descr: "Report ID; required when kind=report. Example report run: --kind report --report-id 124"},
		{field: "WatchlistID", flag: "watchlist-id", group: "Kind-Specific IDs", descr: "Watchlist ID; required when kind=watchlist. Example watchlist run: --kind watchlist --watchlist-id 99"},
		{field: "CoachScreenID", flag: "coach-screen-id", group: "Kind-Specific IDs", descr: "Coach screen ID; required when kind=coach_screen. Example coach screen run: --kind coach_screen --coach-screen-id screen-1"},
		{field: "Limit", flag: "limit", group: "Pagination", descr: "Maximum number of results to return", defaultValue: "50"},
		{field: "Offset", flag: "offset", group: "Pagination", descr: "Number of results to skip for pagination"},
		{field: "Fields", flag: "fields", group: "Filtering & Projection", descr: "Project specific result fields; accepts repeated --fields flags or comma-separated values. Examples: --fields symbol --fields price, or --fields symbol,group_rank,group_rs. Common fields: symbol, price, group_rank, group_rs, composite_rating, eps_rating, rs_rating, acc_dis_rating, smr_rating, industry_name, market_cap, volume_dollar_avg_50d"},
		{field: "ExcludeSPACs", flag: "exclude-spacs", group: "Filtering & Projection", descr: "Exclude SPAC/blank-check entries from results"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			f, ok := typ.FieldByName(tt.field)
			require.True(t, ok, "field %s should exist", tt.field)
			assert.Equal(t, tt.flag, f.Tag.Get("flag"))
			assert.Equal(t, tt.group, f.Tag.Get("flaggroup"))
			assert.Equal(t, tt.descr, f.Tag.Get("flagdescr"))
			assert.Equal(t, tt.defaultValue, f.Tag.Get("default"))
		})
	}

	for _, field := range []string{"MinComposite", "MinRS"} {
		t.Run(field, func(t *testing.T) {
			f, ok := typ.FieldByName(field)
			require.True(t, ok, "field %s should exist", field)
			assert.Empty(t, f.Tag.Get("flag"), "pointer field should stay untagged")
		})
	}
}

func TestCatalogRunManualFilterFlags(t *testing.T) {
	t.Parallel()

	cmd := newCatalogRunCmd()
	tests := []struct {
		name  string
		descr string
	}{
		{name: "min-composite", descr: "Minimum composite rating for report/watchlist rows (0-99); omitted when unset. Example: --min-composite 90"},
		{name: "min-rs", descr: "Minimum RS rating for report/watchlist rows (0-99); omitted when unset. Example: --min-rs 80"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.name)
			require.NotNil(t, flag)
			assert.Equal(t, tt.descr, flag.Usage)
			assert.Equal(t, []string{"Filtering & Projection"}, flag.Annotations[structcliFlagGroupAnnotation])
		})
	}
}

func TestCatalogRunExamples(t *testing.T) {
	t.Parallel()

	cmd := newCatalogRunCmd()
	assert.Contains(t, cmd.Example, "marketsurge-agent catalog run --kind report --report-id 124")
	assert.Contains(t, cmd.Example, "--kind watchlist --watchlist-id 99")
	assert.Contains(t, cmd.Example, "--kind coach_screen --coach-screen-id screen-1")

	typ := reflect.TypeFor[CatalogRunOptions]()
	tests := []struct {
		field string
		want  string
	}{
		{field: "Kind", want: "watchlist uses --watchlist-id"},
		{field: "Kind", want: "report uses --report-id"},
		{field: "Kind", want: "coach_screen uses --coach-screen-id"},
		{field: "ReportID", want: "--kind report --report-id 124"},
		{field: "WatchlistID", want: "--kind watchlist --watchlist-id 99"},
		{field: "CoachScreenID", want: "--kind coach_screen --coach-screen-id screen-1"},
		{field: "Fields", want: "--fields symbol,group_rank,group_rs"},
	}
	for _, tt := range tests {
		t.Run(tt.field+"/"+tt.want, func(t *testing.T) {
			field, ok := typ.FieldByName(tt.field)
			require.True(t, ok)
			assert.Contains(t, field.Tag.Get("flagdescr"), tt.want)
		})
	}
}

type catalogListEnvelope struct {
	Data struct {
		Entries []map[string]any `json:"entries"`
	} `json:"data"`
	Errors   []string       `json:"errors"`
	Metadata map[string]any `json:"metadata"`
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

func runCatalogListCommand(t *testing.T, c *client.Client, args ...string) catalogListEnvelope {
	t.Helper()

	argv := append([]string{"list"}, args...)
	output, err := executeCatalogList(t, c, argv...)
	require.NoError(t, err)

	var envelope catalogListEnvelope
	require.NoError(t, json.Unmarshal([]byte(output), &envelope))
	return envelope
}

func runCatalogRunCommand(t *testing.T, c *client.Client, args ...string) catalogRunEnvelope {
	t.Helper()

	argv := append([]string{"run"}, args...)
	output, err := executeCatalogRun(t, c, argv...)
	require.NoError(t, err)

	var envelope catalogRunEnvelope
	require.NoError(t, json.Unmarshal([]byte(output), &envelope))
	return envelope
}

// executeCatalogList layers the test client onto the list subcommand's context.
// structcli.Bind sets subcommand scope context before execution, which prevents
// cobra from propagating the parent command's client context automatically.
func executeCatalogList(t *testing.T, c *client.Client, args ...string) (string, error) {
	t.Helper()
	return executeCatalogSubcommand(t, c, "list", args...)
}

// executeCatalogRun layers the test client onto the run subcommand's context.
// structcli.Bind sets subcommand scope context before execution, which prevents
// cobra from propagating the parent command's client context automatically.
func executeCatalogRun(t *testing.T, c *client.Client, args ...string) (string, error) {
	t.Helper()
	return executeCatalogSubcommand(t, c, "run", args...)
}

func executeCatalogSubcommand(t *testing.T, c *client.Client, subcommand string, args ...string) (string, error) {
	t.Helper()
	cmd := newCatalogCmd()
	ctx := ContextWithClient(context.Background(), c)
	cmd.SetContext(ctx)
	for _, child := range cmd.Commands() {
		if child.Name() != subcommand {
			continue
		}
		childCtx := child.Context()
		if childCtx == nil {
			childCtx = ctx
		}
		child.SetContext(ContextWithClient(childCtx, c))
	}
	return executeCommand(t, cmd, args...)
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

func catalogRunAdhocFixture() string {
	return `{
		"data": {
			"marketDataAdhocScreen": {
				"responseValues": [
					[
						{"mdItem":{"name":"Symbol"},"value":"AAPL"},
						{"mdItem":{"name":"GroupRank"},"value":5},
						{"mdItem":{"name":"GroupRS"},"value":95},
						{"mdItem":{"name":"CompositeRating"},"value":99},
						{"mdItem":{"name":"RSRating"},"value":95},
						{"mdItem":{"name":"DowJonesInstrumentSubType"},"value":"COMMON"}
					],
					[
						{"mdItem":{"name":"Symbol"},"value":"MSFT"},
						{"mdItem":{"name":"GroupRank"},"value":7},
						{"mdItem":{"name":"GroupRS"},"value":90},
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

func TestCatalogRunOptionsValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    CatalogRunOptions
		wantErr string
	}{
		{
			name:    "missing kind",
			opts:    CatalogRunOptions{},
			wantErr: "missing --kind: use --kind watchlist --watchlist-id 12345, --kind report --report-id 67890, or --kind coach_screen --coach-screen-id ID",
		},
		{
			name:    "invalid kind",
			opts:    CatalogRunOptions{Kind: "bogus"},
			wantErr: "invalid --kind \"bogus\" for catalog run: use one of watchlist, report, coach_screen; screen is list-only",
		},
		{
			name:    "screen kind rejected",
			opts:    CatalogRunOptions{Kind: "screen"},
			wantErr: "invalid --kind screen for catalog run: screens are list-only; use catalog list --kind screen",
		},
		{
			name:    "report without report-id",
			opts:    CatalogRunOptions{Kind: "report"},
			wantErr: "missing --report-id: --kind report requires --report-id 67890",
		},
		{
			name: "report with report-id",
			opts: CatalogRunOptions{Kind: "report", ReportID: 42},
		},
		{
			name:    "watchlist without watchlist-id",
			opts:    CatalogRunOptions{Kind: "watchlist"},
			wantErr: "missing --watchlist-id: --kind watchlist requires --watchlist-id 12345",
		},
		{
			name: "watchlist with watchlist-id",
			opts: CatalogRunOptions{Kind: "watchlist", WatchlistID: 99},
		},
		{
			name:    "coach_screen without coach-screen-id",
			opts:    CatalogRunOptions{Kind: "coach_screen"},
			wantErr: "missing --coach-screen-id: --kind coach_screen requires --coach-screen-id ID",
		},
		{
			name: "coach_screen with coach-screen-id",
			opts: CatalogRunOptions{Kind: "coach_screen", CoachScreenID: "s-1"},
		},
		{
			name:    "report with unknown fields",
			opts:    CatalogRunOptions{Kind: "report", ReportID: 42, Fields: []string{"symbol", "not_real"}},
			wantErr: `unknown --fields values "not_real": use one or more of acc_dis_rating, charting_symbol, company_name, composite_rating, dow_jones_key, eps_rating, group_rank, group_rs, industry_group_rank, industry_name, instrument_sub_type, instrument_type, ipo_date, list_rank, market_cap, price, price_net_change, price_pct_change, price_pct_off_52w_high, rs_rating, smr_rating, symbol, volume, volume_change, volume_dollar_avg_50d, volume_pct_change`,
		},
		{
			name:    "coach_screen with fields",
			opts:    CatalogRunOptions{Kind: "coach_screen", CoachScreenID: "s-1", Fields: []string{"symbol"}},
			wantErr: "unsupported --fields with --kind coach_screen: field projection is only supported for report and watchlist rows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs := tt.opts.Validate(context.Background())
			if tt.wantErr == "" {
				assert.Nil(t, errs)
			} else {
				require.NotEmpty(t, errs)
				err := errs[0]
				require.Error(t, err)
				var verr *mserrors.ValidationError
				assert.ErrorAs(t, err, &verr)
				assert.Equal(t, tt.wantErr, err.Error())
			}
		})
	}
}

func TestCatalogRunOptionsFromCommand(t *testing.T) {
	t.Parallel()

	t.Run("explicit zero is distinguishable from unset", func(t *testing.T) {
		t.Parallel()
		opts := &CatalogRunOptions{}
		cmd := &cobra.Command{Use: "test"}
		require.NoError(t, structcli.Bind(cmd, opts))
		cmd.Flags().Int("min-composite", 0, "Minimum composite rating for report/watchlist rows (0-99); omitted when unset")
		cmd.Flags().Int("min-rs", 0, "Minimum RS rating for report/watchlist rows (0-99); omitted when unset")

		require.NoError(t, cmd.ParseFlags([]string{"--min-composite", "0"}))
		opts.FromCommand(cmd)

		require.NotNil(t, opts.MinComposite, "explicitly set --min-composite 0 should produce non-nil pointer")
		assert.Equal(t, 0, *opts.MinComposite)
		assert.Nil(t, opts.MinRS, "unset --min-rs should remain nil")
	})

	t.Run("unset flags remain nil", func(t *testing.T) {
		t.Parallel()
		opts := &CatalogRunOptions{}
		cmd := &cobra.Command{Use: "test"}
		require.NoError(t, structcli.Bind(cmd, opts))
		cmd.Flags().Int("min-composite", 0, "Minimum composite rating for report/watchlist rows (0-99); omitted when unset")
		cmd.Flags().Int("min-rs", 0, "Minimum RS rating for report/watchlist rows (0-99); omitted when unset")

		require.NoError(t, cmd.ParseFlags([]string{}))
		opts.FromCommand(cmd)

		assert.Nil(t, opts.MinComposite)
		assert.Nil(t, opts.MinRS)
	})
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

func catalogReportEntryMap(report constants.ReportInfo) map[string]any {
	entry := map[string]any{
		"kind":      "report",
		"name":      report.Name,
		"report_id": float64(report.ID),
	}
	if report.Description != "" {
		entry["description"] = report.Description
	}
	return entry
}
