package client

import (
	"context"
	"fmt"

	"github.com/tidwall/gjson"

	"github.com/major/marketsurge-agent/internal/constants"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/queries"
)

// ListCatalog aggregates screens, reports, watchlists, and coach screens.
func (c *Client) ListCatalog(ctx context.Context, kind *models.CatalogKind) (*models.Catalog, error) {
	entries := make([]models.CatalogEntry, 0, len(constants.PredefinedReports))
	errorsList := []string{}

	if kind == nil || *kind == models.CatalogKindWatchlist {
		watchlists, err := c.listWatchlists(ctx)
		if err != nil {
			errorsList = append(errorsList, err.Error())
		} else {
			entries = append(entries, watchlists...)
		}
	}

	if kind == nil || *kind == models.CatalogKindScreen {
		screens, err := c.listScreens(ctx)
		if err != nil {
			errorsList = append(errorsList, err.Error())
		} else {
			entries = append(entries, screens...)
		}
	}

	if kind == nil || *kind == models.CatalogKindReport {
		for _, report := range constants.PredefinedReports {
			reportID := report.ID
			entries = append(entries, models.CatalogEntry{Name: report.Name, Kind: models.CatalogKindReport, ReportID: &reportID})
		}
	}

	if kind == nil || *kind == models.CatalogKindCoachScreen {
		coachEntries, err := c.listCoachEntries(ctx)
		if err != nil {
			errorsList = append(errorsList, err.Error())
		} else {
			entries = append(entries, coachEntries...)
		}
	}

	return &models.Catalog{Entries: entries, Errors: errorsList}, nil
}

// RunReport runs a predefined report via MarketDataAdhocScreen.
func (c *Client) RunReport(ctx context.Context, reportID int) (*models.AdhocScreenResult, error) {
	query, err := queries.Load("adhoc_screen.graphql")
	if err != nil {
		return nil, err
	}

	raw, err := c.Execute(ctx, Request{
		OperationName: "MarketDataAdhocScreen",
		Variables: map[string]any{
			"correlationTag":  "marketsurge",
			"responseColumns": constants.WatchlistColumns,
			"adhocQuery":      nil,
			"includeSource": map[string]any{
				"screenId": map[string]any{"id": reportID, "dialect": "MS_LIST_ID"},
			},
			"pageSize":    1000,
			"resultLimit": 1000000,
			"pageSkip":    0,
			"resultType":  "RESULT_WITH_EXPRESSION_COUNTS",
		},
		Query: query,
	})
	if err != nil {
		return nil, err
	}

	return parseAdhocScreenResult(raw)
}

// RunWatchlist resolves a 64-bit watchlist ID through FlaggedSymbols, then screens the symbols.
func (c *Client) RunWatchlist(ctx context.Context, watchlistID int64) (*models.AdhocScreenResult, error) {
	flaggedQuery, err := queries.Load("flagged_symbols.graphql")
	if err != nil {
		return nil, err
	}

	flaggedRaw, err := c.Execute(ctx, Request{
		OperationName: "FlaggedSymbols",
		Variables: map[string]any{
			"pub":         "msr",
			"watchlistId": fmt.Sprintf("%d", watchlistID),
		},
		Query: flaggedQuery,
	})
	if err != nil {
		return nil, err
	}

	watchlist := gjson.GetBytes(flaggedRaw, "data.watchlist")
	if !watchlist.Exists() || watchlist.Type == gjson.Null {
		return nil, mserrors.NewAPIError(fmt.Sprintf("watchlist not found: %d", watchlistID), nil)
	}

	items := watchlist.Get("items").Array()
	symbols := make([]string, 0, len(items))
	for _, entry := range items {
		if key := gStr(entry.Get("dowJonesKey")); key != nil {
			symbols = append(symbols, *key)
		}
	}
	if len(symbols) == 0 {
		return &models.AdhocScreenResult{Entries: []models.WatchlistEntry{}}, nil
	}

	adhocQuery, err := queries.Load("adhoc_screen.graphql")
	if err != nil {
		return nil, err
	}

	raw, err := c.Execute(ctx, Request{
		OperationName: "MarketDataAdhocScreen",
		Variables: map[string]any{
			"correlationTag":  "marketsurge",
			"responseColumns": constants.WatchlistColumns,
			"adhocQuery":      nil,
			"includeSource": map[string]any{
				"instruments": map[string]any{"symbols": symbols, "dialect": "DJ_KEY"},
			},
			"pageSize":    1000,
			"resultLimit": 1000000,
			"pageSkip":    0,
			"resultType":  "RESULT_WITH_EXPRESSION_COUNTS",
		},
		Query: adhocQuery,
	})
	if err != nil {
		return nil, err
	}

	return parseAdhocScreenResult(raw)
}

// RunCoachScreen runs a coach screen by opaque screen ID.
func (c *Client) RunCoachScreen(ctx context.Context, screenID string) (*models.ScreenResult, error) {
	query, err := queries.Load("run_screen.graphql")
	if err != nil {
		return nil, err
	}

	raw, err := c.Execute(ctx, Request{
		OperationName: "RunScreen",
		Variables: map[string]any{
			"input": map[string]any{
				"correlationTag":  "marketsurge",
				"coachAccount":    true,
				"includeSource":   map[string]any{},
				"pageSize":        1000,
				"resultLimit":     1000000,
				"screenId":        screenID,
				"site":            "marketsurge",
				"skip":            0,
				"responseColumns": constants.WatchlistColumns,
			},
		},
		Query: query,
	})
	if err != nil {
		return nil, err
	}

	container := gjson.GetBytes(raw, "data.user.runScreen")
	if !container.Exists() || container.Type == gjson.Null {
		return nil, mserrors.NewAPIError("no coach screen data returned", nil)
	}

	return &models.ScreenResult{
		NumInstruments: gInt(container.Get("numberOfMatchingInstruments")),
		Rows:           parseRows(container.Get("responseValues").Array()),
	}, nil
}

func (c *Client) listWatchlists(ctx context.Context) ([]models.CatalogEntry, error) {
	query, err := queries.Load("watchlist_names.graphql")
	if err != nil {
		return nil, err
	}
	raw, err := c.Execute(ctx, Request{OperationName: "GetAllWatchlistNames", Variables: map[string]any{"pub": "msr"}, Query: query})
	if err != nil {
		return nil, err
	}
	items := gjson.GetBytes(raw, "data.watchlists").Array()
	entries := make([]models.CatalogEntry, 0, len(items))
	for _, item := range items {
		name := stringify(item.Get("name"))
		entries = append(entries, models.CatalogEntry{
			Name:        name,
			Kind:        models.CatalogKindWatchlist,
			Description: gStr(item.Get("description")),
			WatchlistID: gInt64(item.Get("id")),
		})
	}
	return entries, nil
}

func (c *Client) listScreens(ctx context.Context) ([]models.CatalogEntry, error) {
	query, err := queries.Load("screens.graphql")
	if err != nil {
		return nil, err
	}
	raw, err := c.Execute(ctx, Request{OperationName: "Screens", Variables: map[string]any{"site": "marketsurge"}, Query: query})
	if err != nil {
		return nil, err
	}
	entries := []models.CatalogEntry{}
	for _, item := range gjson.GetBytes(raw, "data.user.screens").Array() {
		name := gStr(item.Get("name"))
		if name == nil {
			continue
		}
		entries = append(entries, models.CatalogEntry{Name: *name, Kind: models.CatalogKindScreen, Description: gStr(item.Get("description"))})
	}
	return entries, nil
}

func (c *Client) listCoachEntries(ctx context.Context) ([]models.CatalogEntry, error) {
	query, err := queries.Load("coach_tree.graphql")
	if err != nil {
		return nil, err
	}
	raw, err := c.Execute(ctx, Request{OperationName: "CoachTree", Variables: map[string]any{"site": "marketsurge", "treeType": "MSR_NAV"}, Query: query})
	if err != nil {
		return nil, err
	}
	entries := []models.CatalogEntry{}
	entries = append(entries, coachTreeEntries(gjson.GetBytes(raw, "data.user.screens").Array(), true)...)
	return entries, nil
}

func coachTreeEntries(nodes []gjson.Result, screen bool) []models.CatalogEntry {
	entries := []models.CatalogEntry{}
	for _, item := range nodes {
		if children := item.Get("children").Array(); len(children) > 0 {
			entries = append(entries, coachTreeEntries(children, screen)...)
			continue
		}
		name := stringify(item.Get("name"))
		if name == "" {
			continue
		}
		ref := getReferenceID(item)
		if screen {
			screenID := gStr(ref.Get("screenId"))
			if screenID == nil {
				continue
			}
			entries = append(entries, models.CatalogEntry{Name: name, Kind: models.CatalogKindCoachScreen, CoachScreenID: screenID})
			continue
		}
		watchlistID := gInt64(ref.Get("watchlistId"))
		if watchlistID == nil {
			continue
		}
		entries = append(entries, models.CatalogEntry{Name: name, Kind: models.CatalogKindWatchlist, WatchlistID: watchlistID})
	}
	return entries
}

func getReferenceID(node gjson.Result) gjson.Result {
	ref := node.Get("referenceId").String()
	if ref == "" {
		return gjson.Result{}
	}
	return gjson.Parse(ref)
}

func parseAdhocScreenResult(raw []byte) (*models.AdhocScreenResult, error) {
	container := gjson.GetBytes(raw, "data.marketDataAdhocScreen")
	if !container.Exists() || container.Type == gjson.Null {
		return nil, mserrors.NewAPIError("no adhoc screen data returned", nil)
	}
	return &models.AdhocScreenResult{
		Entries:     parseWatchlistEntries(container.Get("responseValues").Array()),
		ErrorValues: stringSlice(container.Get("errorValues").Array()),
	}, nil
}

func parseWatchlistEntries(rows []gjson.Result) []models.WatchlistEntry {
	result := make([]models.WatchlistEntry, 0, len(rows))
	for _, row := range rows {
		columns := row.Array()
		mapped := map[string]gjson.Result{}
		for _, col := range columns {
			name := col.Get("mdItem.name").String()
			mapped[name] = col.Get("value")
		}
		result = append(result, models.WatchlistEntry{
			Symbol:              gStr(mapped["Symbol"]),
			CompanyName:         gStr(mapped["CompanyName"]),
			ListRank:            gInt(mapped["ListRank"]),
			Price:               gFloat(mapped["Price"]),
			PriceNetChange:      gFloat(firstExisting(mapped["PriceNetChange"], mapped["PriceNetChg"])),
			PricePctChange:      gFloat(mapped["PricePctChg"]),
			PricePctOff52WHighs: gFloat(mapped["PricePctOff52WHigh"]),
			Volume:              gInt(mapped["Volume"]),
			VolumeChange:        gInt(mapped["VolumeChange"]),
			VolumePctChange:     gFloat(mapped["VolumePctChg"]),
			CompositeRating:     gInt(mapped["CompositeRating"]),
			EPSRating:           gInt(mapped["EPSRating"]),
			RSRating:            gInt(mapped["RSRating"]),
			AccDisRating:        gStr(mapped["AccDisRating"]),
			SMRRating:           gStr(mapped["SMRRating"]),
			IndustryGroupRank:   gInt(mapped["IndustryGroupRank"]),
			IndustryName:        gStr(mapped["IndustryName"]),
			MarketCap:           gFloat(mapped["MarketCapIntraday"]),
			VolumeDollarAvg50D:  gFloat(mapped["VolumeDollarAvg50D"]),
			IPODate:             gStr(mapped["IPODate"]),
			DowJonesKey:         gStr(mapped["DowJonesKey"]),
			ChartingSymbol:      gStr(mapped["ChartingSymbol"]),
			InstrumentType:      gStr(mapped["DowJonesInstrumentType"]),
			InstrumentSubType:   gStr(mapped["DowJonesInstrumentSubType"]),
		})
	}
	return result
}

func parseRows(rows []gjson.Result) []map[string]*string {
	result := make([]map[string]*string, 0, len(rows))
	for _, row := range rows {
		columns := row.Array()
		mapped := map[string]*string{}
		for _, col := range columns {
			mapped[col.Get("mdItem.name").String()] = gStr(col.Get("value"))
		}
		result = append(result, mapped)
	}
	return result
}
