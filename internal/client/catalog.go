package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/major/marketsurge-agent/internal/constants"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/queries"
)

const industryGroupsReportID = 46

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
			"responseColumns": adhocResponseColumns(),
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

	result, err := parseAdhocScreenResult(raw)
	if err != nil {
		return nil, err
	}
	if reportID == industryGroupsReportID {
		if err := c.enrichIndustryGroupReport(ctx, result); err != nil {
			return nil, err
		}
	}
	return result, nil
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

	watchlist := getNestedMap(flaggedRaw, "data", "watchlist")
	if len(watchlist) == 0 {
		return nil, mserrors.NewAPIError(fmt.Sprintf("watchlist not found: %d", watchlistID), nil)
	}

	symbols := make([]string, 0, len(getNestedSlice(watchlist, "items")))
	for _, entry := range getNestedSlice(watchlist, "items") {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if key := stringPtr(item["dowJonesKey"]); key != nil {
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
			"responseColumns": adhocResponseColumns(),
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

	container := getNestedMap(raw, "data", "user", "runScreen")
	if len(container) == 0 {
		return nil, mserrors.NewAPIError("no coach screen data returned", nil)
	}

	return &models.ScreenResult{
		NumInstruments: intPtr(container["numberOfMatchingInstruments"]),
		Rows:           parseRows(getNestedSlice(container, "responseValues")),
	}, nil
}

func adhocResponseColumns() []map[string]string {
	columns := make([]map[string]string, 0, len(constants.WatchlistColumns))
	for _, column := range constants.WatchlistColumns {
		columns = append(columns, map[string]string{"name": column})
	}
	return columns
}

type industryGroupInfo struct {
	Rank int
	RS   int
}

// enrichIndustryGroupReport adds verified stock-level industry rank data to the
// static 197 Industry Groups report. MarketSurge exposes the report rows as
// group index instruments, but the group rank and RS values only appear on
// stock-level industry data.
func (c *Client) enrichIndustryGroupReport(ctx context.Context, result *models.AdhocScreenResult) error {
	groups, err := c.industryGroupsByName(ctx)
	if err != nil {
		return err
	}
	for i := range result.Entries {
		name := result.Entries[i].CompanyName
		if name == nil {
			continue
		}
		info, ok := groups[normalizeIndustryGroupName(*name)]
		if !ok {
			continue
		}
		if info.Rank != 0 {
			result.Entries[i].GroupRank = &info.Rank
		}
		if info.RS != 0 {
			result.Entries[i].GroupRS = &info.RS
		}
	}
	return nil
}

// industryGroupsByName builds a lookup from current stock universe industry
// names to their published group rank and group RS values.
func (c *Client) industryGroupsByName(ctx context.Context) (map[string]industryGroupInfo, error) {
	query, err := queries.Load("adhoc_screen.graphql")
	if err != nil {
		return nil, err
	}
	raw, err := c.Execute(ctx, Request{
		OperationName: "MarketDataAdhocScreen",
		Variables: map[string]any{
			"correlationTag": "marketsurge-industry-groups",
			"responseColumns": []map[string]string{
				{"name": "Symbol"},
				{"name": "IndustryName"},
				{"name": "IndustryGroupRank"},
			},
			"adhocQuery":    nil,
			"includeSource": map[string]any{},
			"pageSize":      20000,
			"resultLimit":   20000,
			"pageSkip":      0,
			"resultType":    "RESULT_WITH_EXPRESSION_COUNTS",
		},
		Query: query,
	})
	if err != nil {
		return nil, err
	}
	stocks, err := parseAdhocScreenResult(raw)
	if err != nil {
		return nil, err
	}
	representatives := map[string]string{}
	groups := map[string]industryGroupInfo{}
	for i := range stocks.Entries {
		stock := &stocks.Entries[i]
		if stock.Symbol == nil || stock.IndustryName == nil || stock.IndustryGroupRank == nil {
			continue
		}
		key := normalizeIndustryGroupName(*stock.IndustryName)
		groups[key] = industryGroupInfo{Rank: *stock.IndustryGroupRank}
		if _, exists := representatives[key]; !exists {
			representatives[key] = *stock.Symbol
		}
	}
	if len(representatives) == 0 {
		return groups, nil
	}
	symbols := make([]string, 0, len(representatives))
	nameBySymbol := make(map[string]string, len(representatives))
	for name, symbol := range representatives {
		symbols = append(symbols, symbol)
		nameBySymbol[symbol] = name
	}
	sort.Strings(symbols)
	rsBySymbol, err := c.industryGroupRSByRepresentative(ctx, symbols)
	if err != nil {
		return nil, err
	}
	for symbol, rs := range rsBySymbol {
		name := nameBySymbol[symbol]
		info := groups[name]
		info.RS = rs
		groups[name] = info
	}
	return groups, nil
}

// industryGroupRSByRepresentative fetches group RS from one stock per industry.
// The adhoc screener exposes the rank, while the marketData industry object is
// the verified source for the matching RS value.
func (c *Client) industryGroupRSByRepresentative(ctx context.Context, symbols []string) (map[string]int, error) {
	query, err := queries.Load("industry_group_rs.graphql")
	if err != nil {
		return nil, fmt.Errorf("loading industry group RS query: %w", err)
	}

	result := map[string]int{}
	for start := 0; start < len(symbols); start += 250 {
		end := min(start+250, len(symbols))
		raw, err := c.Execute(ctx, Request{
			OperationName: "IndustryGroupRS",
			Variables: map[string]any{
				"symbols":           symbols[start:end],
				"symbolDialectType": "CHARTING",
			},
			Query: query,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range getNestedSlice(raw, "data", "marketData") {
			mapped, ok := item.(map[string]any)
			if !ok {
				continue
			}
			rs := firstMap(getNestedSlice(mapped, "industry", "groupRS"))
			if value := intPtr(rs["value"]); value != nil {
				result[stringify(getNestedMap(mapped, "originRequest")["symbol"])] = *value
			}
		}
	}
	return result, nil
}

// normalizeIndustryGroupName removes report-only markers and spacing so report
// 46 rows can be matched to stock universe industry names when MarketSurge uses
// the same display name in both sources.
func normalizeIndustryGroupName(name string) string {
	name = strings.TrimPrefix(name, "(L)")
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "")
	return name
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
	entries := make([]models.CatalogEntry, 0, len(getNestedSlice(raw, "data", "watchlists")))
	for _, entry := range getNestedSlice(raw, "data", "watchlists") {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		name := stringify(item["name"])
		entries = append(entries, models.CatalogEntry{
			Name:        name,
			Kind:        models.CatalogKindWatchlist,
			Description: stringPtr(item["description"]),
			WatchlistID: int64Ptr(item["id"]),
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
	for _, entry := range getNestedSlice(raw, "data", "user", "screens") {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		name := stringPtr(item["name"])
		if name == nil {
			continue
		}
		entries = append(entries, models.CatalogEntry{Name: *name, Kind: models.CatalogKindScreen, Description: stringPtr(item["description"])})
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
	entries = append(entries, coachTreeEntries(getNestedSlice(raw, "data", "user", "screens"), true)...)
	return entries, nil
}

func coachTreeEntries(nodes []any, screen bool) []models.CatalogEntry {
	entries := []models.CatalogEntry{}
	for _, entry := range nodes {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if children := getNestedSlice(item, "children"); len(children) > 0 {
			entries = append(entries, coachTreeEntries(children, screen)...)
			continue
		}
		name := stringify(item["name"])
		if name == "" {
			continue
		}
		ref := getReferenceID(item)
		if screen {
			screenID := stringPtr(ref["screenId"])
			if screenID == nil {
				continue
			}
			entries = append(entries, models.CatalogEntry{Name: name, Kind: models.CatalogKindCoachScreen, CoachScreenID: screenID})
			continue
		}
		watchlistID := int64Ptr(ref["watchlistId"])
		if watchlistID == nil {
			continue
		}
		entries = append(entries, models.CatalogEntry{Name: name, Kind: models.CatalogKindWatchlist, WatchlistID: watchlistID})
	}
	return entries
}

func getReferenceID(node map[string]any) map[string]any {
	reference, ok := node["referenceId"].(string)
	if !ok || reference == "" {
		return nil
	}
	parsed := map[string]any{}
	_ = json.Unmarshal([]byte(reference), &parsed)
	return parsed
}

func parseAdhocScreenResult(raw map[string]any) (*models.AdhocScreenResult, error) {
	container := getNestedMap(raw, "data", "marketDataAdhocScreen")
	if len(container) == 0 {
		return nil, mserrors.NewAPIError("no adhoc screen data returned", nil)
	}
	return &models.AdhocScreenResult{Entries: parseWatchlistEntries(getNestedSlice(container, "responseValues")), ErrorValues: stringSlice(getNestedSlice(container, "errorValues"))}, nil
}

// extractColumns converts a row's column slice into a name-to-raw-value map.
// Each column is expected to be a map[string]any with an "mdItem" sub-map
// containing a "name" key and a top-level "value" key.
func extractColumns(columns []any) map[string]any {
	mapped := map[string]any{}
	for _, column := range columns {
		item, ok := column.(map[string]any)
		if !ok {
			continue
		}
		name := stringify(getNestedMap(item, "mdItem")["name"])
		mapped[name] = item["value"]
	}
	return mapped
}

func parseWatchlistEntries(rows []any) []models.WatchlistEntry {
	result := make([]models.WatchlistEntry, 0, len(rows))
	for _, row := range rows {
		columns, ok := row.([]any)
		if !ok {
			continue
		}
		mapped := extractColumns(columns)
		result = append(result, models.WatchlistEntry{
			Symbol:              stringPtr(mapped["Symbol"]),
			CompanyName:         stringPtr(mapped["CompanyName"]),
			ListRank:            intPtr(mapped["ListRank"]),
			GroupRank:           intPtr(mapped["GroupRank"]),
			GroupRS:             intPtr(mapped["GroupRS"]),
			Price:               floatPtr(mapped["Price"]),
			PriceNetChange:      floatPtr(firstNonNil(mapped["PriceNetChange"], mapped["PriceNetChg"])),
			PricePctChange:      floatPtr(mapped["PricePctChg"]),
			PricePctOff52WHighs: floatPtr(mapped["PricePctOff52WHigh"]),
			Volume:              intPtr(mapped["Volume"]),
			VolumeChange:        intPtr(mapped["VolumeChange"]),
			VolumePctChange:     floatPtr(mapped["VolumePctChg"]),
			CompositeRating:     intPtr(mapped["CompositeRating"]),
			EPSRating:           intPtr(mapped["EPSRating"]),
			RSRating:            intPtr(mapped["RSRating"]),
			AccDisRating:        stringPtr(mapped["AccDisRating"]),
			SMRRating:           stringPtr(mapped["SMRRating"]),
			IndustryGroupRank:   intPtr(mapped["IndustryGroupRank"]),
			IndustryName:        stringPtr(mapped["IndustryName"]),
			MarketCap:           floatPtr(mapped["MarketCapIntraday"]),
			VolumeDollarAvg50D:  floatPtr(mapped["VolumeDollarAvg50D"]),
			IPODate:             stringPtr(mapped["IPODate"]),
			DowJonesKey:         stringPtr(mapped["DowJonesKey"]),
			ChartingSymbol:      stringPtr(mapped["ChartingSymbol"]),
			InstrumentType:      stringPtr(mapped["DowJonesInstrumentType"]),
			InstrumentSubType:   stringPtr(mapped["DowJonesInstrumentSubType"]),
		})
	}
	return result
}

func parseRows(rows []any) []map[string]*string {
	result := make([]map[string]*string, 0, len(rows))
	for _, row := range rows {
		columns, ok := row.([]any)
		if !ok {
			continue
		}
		raw := extractColumns(columns)
		mapped := make(map[string]*string, len(raw))
		for k, v := range raw {
			mapped[k] = stringPtr(v)
		}
		result = append(result, mapped)
	}
	return result
}

func firstNonNil(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func stringSlice(values []any) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if text := stringify(value); text != "" {
			result = append(result, text)
		}
	}
	return result
}
