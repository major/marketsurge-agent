// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

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

func init() { rootCmd.AddCommand(newCatalogCmd()) }

func newCatalogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "Catalog commands",
	}
	cmd.AddCommand(newCatalogListCmd())
	cmd.AddCommand(newCatalogRunCmd())
	return cmd
}

func newCatalogRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a catalog report, watchlist, or coach screen",
		RunE: func(cmd *cobra.Command, args []string) error {
			kind, err := validateCatalogRunKind(cmd)
			if err != nil {
				return err
			}

			entries, total, err := runCatalogEntries(cmd.Context(), ClientFromContext(cmd.Context()), cmd, kind)
			if err != nil {
				return err
			}

			limit, err := cmd.Flags().GetInt("limit")
			if err != nil {
				return err
			}
			offset, err := cmd.Flags().GetInt("offset")
			if err != nil {
				return err
			}

			data := map[string]any{"entries": entries}
			return output.WriteSuccess(cmd.OutOrStdout(), data, output.CatalogMeta(string(kind), total, limit, offset))
		},
	}
	cmd.Flags().String("kind", "", "Catalog kind (required): report, watchlist, or coach_screen")
	cmd.Flags().Int("report-id", 0, "Report ID (required when --kind=report)")
	cmd.Flags().Int64("watchlist-id", 0, "Watchlist ID (required when --kind=watchlist)")
	cmd.Flags().String("coach-screen-id", "", "Coach screen ID (required when --kind=coach_screen)")
	cmd.Flags().Int("limit", defaultCatalogRunLimit, "Maximum number of entries to return")
	cmd.Flags().Int("offset", 0, "Starting offset for pagination")
	cmd.Flags().StringSlice("fields", []string{}, "Optional fields to include in each entry")
	cmd.Flags().Int("min-composite", 0, "Minimum composite rating filter")
	cmd.Flags().Int("min-rs", 0, "Minimum RS rating filter")
	cmd.Flags().Bool("exclude-spacs", false, "Exclude SPAC/blank-check entries")
	return cmd
}

type catalogRunFilters struct {
	MinComposite *int
	MinRS        *int
	ExcludeSPACs bool
}

func validateCatalogRunKind(cmd *cobra.Command) (models.CatalogKind, error) {
	value, err := cmd.Flags().GetString("kind")
	if err != nil {
		return "", err
	}
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

func optionalIntFlag(cmd *cobra.Command, name string) *int {
	if !cmd.Flags().Changed(name) {
		return nil
	}
	v, _ := cmd.Flags().GetInt(name)
	return &v
}

func runCatalogEntries(ctx context.Context, c *client.Client, cmd *cobra.Command, kind models.CatalogKind) (any, int, error) {
	filters, err := catalogRunFiltersFromFlags(cmd)
	if err != nil {
		return nil, 0, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return nil, 0, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return nil, 0, err
	}
	fields, err := cmd.Flags().GetStringSlice("fields")
	if err != nil {
		return nil, 0, err
	}

	switch kind {
	case models.CatalogKindReport:
		reportID, err := cmd.Flags().GetInt("report-id")
		if err != nil {
			return nil, 0, err
		}
		if reportID == 0 {
			return nil, 0, mserrors.NewValidationError("report-id is required when kind=report", nil)
		}

		result, err := c.RunReport(ctx, reportID)
		if err != nil {
			return nil, 0, err
		}

		entries := applyCatalogRunFilters(result.Entries, filters)
		return projectWatchlistEntries(paginateSlice(entries, limit, offset), fields), len(entries), nil
	case models.CatalogKindWatchlist:
		watchlistID, err := cmd.Flags().GetInt64("watchlist-id")
		if err != nil {
			return nil, 0, err
		}
		if watchlistID == 0 {
			return nil, 0, mserrors.NewValidationError("watchlist-id is required when kind=watchlist", nil)
		}

		result, err := c.RunWatchlist(ctx, watchlistID)
		if err != nil {
			return nil, 0, err
		}

		entries := applyCatalogRunFilters(result.Entries, filters)
		return projectWatchlistEntries(paginateSlice(entries, limit, offset), fields), len(entries), nil
	case models.CatalogKindCoachScreen:
		coachScreenID, err := cmd.Flags().GetString("coach-screen-id")
		if err != nil {
			return nil, 0, err
		}
		if coachScreenID == "" {
			return nil, 0, mserrors.NewValidationError("coach-screen-id is required when kind=coach_screen", nil)
		}

		result, err := c.RunCoachScreen(ctx, coachScreenID)
		if err != nil {
			return nil, 0, err
		}

		rows := paginateSlice(result.Rows, limit, offset)
		return rows, len(result.Rows), nil
	default:
		return nil, 0, mserrors.NewValidationError("kind must be one of: watchlist, screen, report, coach_screen", nil)
	}
}

func catalogRunFiltersFromFlags(cmd *cobra.Command) (catalogRunFilters, error) {
	excludeSPACs, err := cmd.Flags().GetBool("exclude-spacs")
	if err != nil {
		return catalogRunFilters{}, err
	}
	return catalogRunFilters{
		MinComposite: optionalIntFlag(cmd, "min-composite"),
		MinRS:        optionalIntFlag(cmd, "min-rs"),
		ExcludeSPACs: excludeSPACs,
	}, nil
}

func applyCatalogRunFilters(entries []models.WatchlistEntry, filters catalogRunFilters) []models.WatchlistEntry {
	filtered := make([]models.WatchlistEntry, 0, len(entries))
	for i := range entries {
		if filters.MinComposite != nil {
			if entries[i].CompositeRating == nil || *entries[i].CompositeRating < *filters.MinComposite {
				continue
			}
		}
		if filters.MinRS != nil {
			if entries[i].RSRating == nil || *entries[i].RSRating < *filters.MinRS {
				continue
			}
		}
		if filters.ExcludeSPACs && isSPACEntry(&entries[i]) {
			continue
		}
		filtered = append(filtered, entries[i])
	}
	return filtered
}

func isSPACEntry(entry *models.WatchlistEntry) bool {
	if entry.InstrumentSubType == nil {
		return false
	}
	return strings.EqualFold(*entry.InstrumentSubType, "BLANK_CHECK")
}

func paginateSlice[T any](slice []T, limit, offset int) []T {
	start := clampCatalogOffset(offset, len(slice))
	if limit < 0 {
		limit = 0
	}
	end := len(slice)
	if limit > 0 && start+limit < end {
		end = start + limit
	}
	if start > end {
		start = end
	}
	return slice[start:end]
}

func clampCatalogOffset(offset, length int) int {
	return max(0, min(offset, length))
}

func projectWatchlistEntries(entries []models.WatchlistEntry, fields []string) any {
	if len(fields) == 0 {
		return entries
	}

	projected := make([]map[string]any, 0, len(entries))
	for i := range entries {
		projected = append(projected, projectWatchlistEntry(&entries[i], fields))
	}
	return projected
}

func projectWatchlistEntry(entry *models.WatchlistEntry, fields []string) map[string]any {
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

func newCatalogListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List catalog entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			kindStr, err := cmd.Flags().GetString("kind")
			if err != nil {
				return err
			}

			kind, err := parseCatalogKind(kindStr)
			if err != nil {
				return err
			}

			c := ClientFromContext(cmd.Context())
			catalog, err := c.ListCatalog(cmd.Context(), kind)
			if err != nil {
				return err
			}

			data := map[string]any{"entries": catalog.Entries}
			meta := output.CatalogMeta(kindStr, len(catalog.Entries), 0, 0)
			if len(catalog.Errors) > 0 {
				return output.WritePartial(cmd.OutOrStdout(), data, catalog.Errors, meta)
			}
			return output.WriteSuccess(cmd.OutOrStdout(), data, meta)
		},
	}
	cmd.Flags().String("kind", "", "Catalog kind (required): report, watchlist, coach_screen, screen")
	return cmd
}

var validCatalogKinds = map[string]models.CatalogKind{
	string(models.CatalogKindWatchlist):   models.CatalogKindWatchlist,
	string(models.CatalogKindScreen):      models.CatalogKindScreen,
	string(models.CatalogKindReport):      models.CatalogKindReport,
	string(models.CatalogKindCoachScreen): models.CatalogKindCoachScreen,
}

// parseCatalogKind validates the CLI flag and returns a typed catalog kind pointer.
func parseCatalogKind(value string) (*models.CatalogKind, error) {
	if value == "" {
		return nil, nil
	}

	kind, ok := validCatalogKinds[value]
	if !ok {
		return nil, mserrors.NewValidationError("kind must be one of: watchlist, screen, report, coach_screen", nil)
	}

	return &kind, nil
}
