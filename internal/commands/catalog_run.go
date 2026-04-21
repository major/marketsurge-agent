package commands

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/internal/output"
)

const defaultCatalogRunLimit = 50

var watchlistFieldAliases = map[string]string{
	"symbol":                      "symbol",
	"companyname":                 "company_name",
	"company_name":                "company_name",
	"listrank":                    "list_rank",
	"list_rank":                   "list_rank",
	"price":                       "price",
	"pricenetchange":              "price_net_change",
	"price_net_change":            "price_net_change",
	"pricenetchg":                 "price_net_change",
	"pricepctchange":              "price_pct_change",
	"price_pct_change":            "price_pct_change",
	"pricepctchg":                 "price_pct_change",
	"pricepctoff52whighs":         "price_pct_off_52w_high",
	"pricepctoff52whigh":          "price_pct_off_52w_high",
	"price_pct_off_52w_high":      "price_pct_off_52w_high",
	"volume":                      "volume",
	"volumechange":                "volume_change",
	"volume_change":               "volume_change",
	"volumepctchange":             "volume_pct_change",
	"volume_pct_change":           "volume_pct_change",
	"compositerating":             "composite_rating",
	"composite_rating":            "composite_rating",
	"epsrating":                   "eps_rating",
	"eps_rating":                  "eps_rating",
	"rsrating":                    "rs_rating",
	"rs_rating":                   "rs_rating",
	"accdisrating":                "acc_dis_rating",
	"acc_dis_rating":              "acc_dis_rating",
	"smrrating":                   "smr_rating",
	"smr_rating":                  "smr_rating",
	"industrygrouprank":           "industry_group_rank",
	"industry_group_rank":         "industry_group_rank",
	"industryname":                "industry_name",
	"industry_name":               "industry_name",
	"marketcap":                   "market_cap",
	"market_cap":                  "market_cap",
	"marketcapintraday":           "market_cap",
	"volumedollaravg50d":          "volume_dollar_avg_50d",
	"volume_dollar_avg_50d":       "volume_dollar_avg_50d",
	"ipodate":                     "ipo_date",
	"ipo_date":                    "ipo_date",
	"dowjoneskey":                 "dow_jones_key",
	"dow_jones_key":               "dow_jones_key",
	"chartingsymbol":              "charting_symbol",
	"charting_symbol":             "charting_symbol",
	"instrumenttype":              "instrument_type",
	"instrument_type":             "instrument_type",
	"dowjonesinstrumenttype":      "instrument_type",
	"instrumentsubtype":           "instrument_sub_type",
	"instrument_sub_type":         "instrument_sub_type",
	"dowjonesinstrumentsubtype":   "instrument_sub_type",
	"volumepctchgvs50davgvolume":  "volume_pct_change",
	"volumeavg50day":              "volume",
	"excludeinstrumentsubtype":    "instrument_sub_type",
	"exclude_instrument_sub_type": "instrument_sub_type",
}

// CatalogRunCommand returns the CLI command for running catalog entries.
func CatalogRunCommand(c *client.Client, w io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Run a catalog report, watchlist, or coach screen",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "kind", Usage: "Catalog kind: report, watchlist, or coach_screen"},
			&cli.IntFlag{Name: "report-id", Usage: "Report ID (required when --kind=report)"},
			&cli.Int64Flag{Name: "watchlist-id", Usage: "Watchlist ID (required when --kind=watchlist)"},
			&cli.StringFlag{Name: "coach-screen-id", Usage: "Coach screen ID (required when --kind=coach_screen)"},
			&cli.IntFlag{Name: "limit", Value: defaultCatalogRunLimit, Usage: "Maximum number of entries to return"},
			&cli.IntFlag{Name: "offset", Value: 0, Usage: "Starting offset for pagination"},
			&cli.StringSliceFlag{Name: "fields", Usage: "Optional fields to include in each entry"},
			&cli.IntFlag{Name: "min-composite", Usage: "Minimum composite rating filter"},
			&cli.IntFlag{Name: "min-rs", Usage: "Minimum RS rating filter"},
			&cli.BoolFlag{Name: "exclude-spacs", Value: false, Usage: "Exclude SPAC/blank-check entries"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			kind, err := validateCatalogRunKind(cmd.String("kind"))
			if err != nil {
				_ = output.WriteError(w, err)
				return err
			}

			filters := catalogRunFilters{
				MinComposite: optionalIntFlag(cmd, "min-composite"),
				MinRS:        optionalIntFlag(cmd, "min-rs"),
				ExcludeSPACs: cmd.Bool("exclude-spacs"),
			}
			limit := cmd.Int("limit")
			offset := cmd.Int("offset")
			fields := cmd.StringSlice("fields")

			entries, total, err := runCatalogEntries(ctx, c, kind, cmd, filters, limit, offset, fields)
			if err != nil {
				_ = output.WriteError(w, err)
				return err
			}

			data := map[string]any{"entries": entries}
			return output.WriteSuccess(w, data, output.CatalogMeta(string(kind), total, limit, offset))
		},
	}
}

// catalogRunFilters stores client-side filters for report and watchlist runs.
type catalogRunFilters struct {
	MinComposite *int
	MinRS        *int
	ExcludeSPACs bool
}

// validateCatalogRunKind validates the requested catalog kind for dispatch.
func validateCatalogRunKind(value string) (models.CatalogKind, error) {
	if value == "" {
		return "", mserrors.NewValidationError("kind is required", nil)
	}

	kind, err := parseCatalogKind(value)
	if err != nil {
		return "", err
	}
	if kind == nil {
		return "", mserrors.NewValidationError("kind is required", nil)
	}
	if *kind == models.CatalogKindScreen {
		return "", mserrors.NewValidationError("screens cannot be run directly, use catalog list to view them", nil)
	}

	return *kind, nil
}

// optionalIntFlag returns a pointer when the CLI flag was explicitly set.
func optionalIntFlag(cmd *cli.Command, name string) *int {
	if !cmd.IsSet(name) {
		return nil
	}
	value := cmd.Int(name)
	return &value
}

// runCatalogEntries dispatches the requested catalog kind and shapes its output.
func runCatalogEntries(
	ctx context.Context,
	c *client.Client,
	kind models.CatalogKind,
	cmd *cli.Command,
	filters catalogRunFilters,
	limit int,
	offset int,
	fields []string,
) (any, int, error) {
	switch kind {
	case models.CatalogKindReport:
		reportID := cmd.Int("report-id")
		if reportID == 0 {
			return nil, 0, mserrors.NewValidationError("report-id is required when kind=report", nil)
		}

		result, err := c.RunReport(ctx, reportID)
		if err != nil {
			return nil, 0, err
		}

		entries := applyCatalogRunFilters(result.Entries, filters)
		return projectWatchlistEntries(paginateWatchlistEntries(entries, limit, offset), fields), len(entries), nil
	case models.CatalogKindWatchlist:
		watchlistID := cmd.Int64("watchlist-id")
		if watchlistID == 0 {
			return nil, 0, mserrors.NewValidationError("watchlist-id is required when kind=watchlist", nil)
		}

		result, err := c.RunWatchlist(ctx, watchlistID)
		if err != nil {
			return nil, 0, err
		}

		entries := applyCatalogRunFilters(result.Entries, filters)
		return projectWatchlistEntries(paginateWatchlistEntries(entries, limit, offset), fields), len(entries), nil
	case models.CatalogKindCoachScreen:
		coachScreenID := cmd.String("coach-screen-id")
		if coachScreenID == "" {
			return nil, 0, mserrors.NewValidationError("coach-screen-id is required when kind=coach_screen", nil)
		}

		result, err := c.RunCoachScreen(ctx, coachScreenID)
		if err != nil {
			return nil, 0, err
		}

		rows := paginateCoachScreenRows(result.Rows, limit, offset)
		return rows, len(result.Rows), nil
	default:
		return nil, 0, mserrors.NewValidationError("kind must be one of: watchlist, screen, report, coach_screen", nil)
	}
}

// applyCatalogRunFilters applies rating and SPAC filters to watchlist-style entries.
func applyCatalogRunFilters(entries []models.WatchlistEntry, filters catalogRunFilters) []models.WatchlistEntry {
	filtered := make([]models.WatchlistEntry, 0, len(entries))
	for _, entry := range entries {
		if filters.MinComposite != nil {
			if entry.CompositeRating == nil || *entry.CompositeRating < *filters.MinComposite {
				continue
			}
		}
		if filters.MinRS != nil {
			if entry.RSRating == nil || *entry.RSRating < *filters.MinRS {
				continue
			}
		}
		if filters.ExcludeSPACs && isSPACEntry(entry) {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

// isSPACEntry reports whether the entry is a blank-check instrument.
func isSPACEntry(entry models.WatchlistEntry) bool {
	if entry.InstrumentSubType == nil {
		return false
	}
	return strings.EqualFold(*entry.InstrumentSubType, "BLANK_CHECK")
}

// paginateWatchlistEntries slices watchlist entries using limit and offset.
func paginateWatchlistEntries(entries []models.WatchlistEntry, limit int, offset int) []models.WatchlistEntry {
	start := clampCatalogOffset(offset, len(entries))
	if limit < 0 {
		limit = 0
	}
	end := len(entries)
	if limit > 0 && start+limit < end {
		end = start + limit
	}
	if start > end {
		start = end
	}
	return entries[start:end]
}

// paginateCoachScreenRows slices coach screen rows using limit and offset.
func paginateCoachScreenRows(rows []map[string]*string, limit int, offset int) []map[string]*string {
	start := clampCatalogOffset(offset, len(rows))
	if limit < 0 {
		limit = 0
	}
	end := len(rows)
	if limit > 0 && start+limit < end {
		end = start + limit
	}
	if start > end {
		start = end
	}
	return rows[start:end]
}

// clampCatalogOffset bounds the offset to the available entry count.
func clampCatalogOffset(offset int, length int) int {
	if offset < 0 {
		return 0
	}
	if offset > length {
		return length
	}
	return offset
}

// projectWatchlistEntries narrows output entries to the requested fields when present.
func projectWatchlistEntries(entries []models.WatchlistEntry, fields []string) any {
	if len(fields) == 0 {
		return entries
	}

	projected := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		projected = append(projected, projectWatchlistEntry(entry, fields))
	}
	return projected
}

// projectWatchlistEntry converts a watchlist entry into a field-filtered map.
func projectWatchlistEntry(entry models.WatchlistEntry, fields []string) map[string]any {
	encoded, err := json.Marshal(entry)
	if err != nil {
		return map[string]any{}
	}

	decoded := map[string]any{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return map[string]any{}
	}

	projected := map[string]any{}
	for _, field := range fields {
		key, ok := normalizeWatchlistField(field)
		if !ok {
			continue
		}
		value, exists := decoded[key]
		if !exists {
			continue
		}
		projected[key] = value
	}

	return projected
}

// normalizeWatchlistField maps CLI field names to watchlist JSON keys.
func normalizeWatchlistField(field string) (string, bool) {
	trimmed := strings.TrimSpace(field)
	if trimmed == "" {
		return "", false
	}

	if mapped, ok := watchlistFieldAliases[strings.ToLower(trimmed)]; ok {
		return mapped, true
	}

	normalized := strings.ToLower(trimmed)
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, "_", "")
	if mapped, ok := watchlistFieldAliases[normalized]; ok {
		return mapped, true
	}

	return "", false
}
