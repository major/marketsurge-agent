// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/leodido/structcli"
	structclivalues "github.com/leodido/structcli/values"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/internal/output"
)

const (
	defaultCatalogRunLimit       = 50
	structcliFlagGroupAnnotation = "leodido/structcli/flag-groups"
)

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
		Long: `Catalog commands discover and run MarketSurge watchlists, screens,
reports, and coach screens.

Workflow:

  1. Run "catalog list" to discover entries and their IDs
  2. Run a returned ID with "catalog run" using the matching kind and ID flag
  3. Page and project results before deeper analysis
  4. Feed selected tickers into "stock analyze --summary" or "stock analyze"`,
	}
	cmd.AddCommand(newCatalogListCmd())
	cmd.AddCommand(newCatalogRunCmd())
	return cmd
}

// CatalogRunOptions holds flags for the catalog run command.
type CatalogRunOptions struct {
	Kind          models.CatalogKind `flag:"kind" flaggroup:"Catalog Selection" flagdescr:"Required catalog kind to run (watchlist, report, coach_screen); screens are list-only" flagcustom:"true"`
	ReportID      int                `flag:"report-id" flaggroup:"Kind-Specific IDs" flagdescr:"Report ID; required when kind=report (discover with catalog list --kind report)"`
	WatchlistID   int64              `flag:"watchlist-id" flaggroup:"Kind-Specific IDs" flagdescr:"Watchlist ID; required when kind=watchlist (discover with catalog list --kind watchlist)"`
	CoachScreenID string             `flag:"coach-screen-id" flaggroup:"Kind-Specific IDs" flagdescr:"Coach screen ID; required when kind=coach_screen (discover with catalog list --kind coach_screen)"`
	Limit         int                `flag:"limit" flaggroup:"Pagination" flagdescr:"Maximum number of results to return" default:"50"`
	Offset        int                `flag:"offset" flaggroup:"Pagination" flagdescr:"Number of results to skip for pagination"`
	Fields        []string           `flag:"fields" flaggroup:"Filtering & Projection" flagdescr:"Project specific result fields, such as symbol, price, composite_rating, eps_rating, rs_rating"`
	MinComposite  *int
	MinRS         *int
	ExcludeSPACs  bool `flag:"exclude-spacs" flaggroup:"Filtering & Projection" flagdescr:"Exclude SPAC/blank-check entries from results"`
}

// DefineKind keeps catalog kind parsing in Validate so missing and invalid kind
// values continue to return the CLI's domain-specific ValidationError type.
func (o *CatalogRunOptions) DefineKind(_ string, _ string, descr string, _ reflect.StructField, _ reflect.Value) (pflag.Value, string) { //nolint:gocritic // signature dictated by structcli convention
	value := string(o.Kind)
	return structclivalues.NewString(&value), descr
}

// DecodeKind converts the custom string flag into the typed catalog enum.
func (o *CatalogRunOptions) DecodeKind(input any) (any, error) {
	value, _ := input.(string)
	return models.CatalogKind(value), nil
}

// FromCommand populates pointer fields that require cobra's Changed() detection.
func (o *CatalogRunOptions) FromCommand(cmd *cobra.Command) {
	if cmd.Flags().Changed("min-composite") {
		v, _ := cmd.Flags().GetInt("min-composite")
		o.MinComposite = &v
	}
	if cmd.Flags().Changed("min-rs") {
		v, _ := cmd.Flags().GetInt("min-rs")
		o.MinRS = &v
	}
}

// Validate checks that the catalog run options are consistent and complete.
func (o *CatalogRunOptions) Validate(_ context.Context) []error {
	if o.Kind == "" {
		return []error{mserrors.NewValidationError("kind is required", nil)}
	}

	kind, err := parseCatalogKind(string(o.Kind))
	if err != nil {
		return []error{err}
	}
	if kind == nil {
		return []error{mserrors.NewValidationError("kind is required", nil)}
	}
	if *kind == models.CatalogKindScreen {
		return []error{mserrors.NewValidationError("screens cannot be run directly, use catalog list to view them", nil)}
	}

	switch *kind {
	case models.CatalogKindReport:
		if o.ReportID == 0 {
			return []error{mserrors.NewValidationError("report-id is required when kind=report", nil)}
		}
	case models.CatalogKindWatchlist:
		if o.WatchlistID == 0 {
			return []error{mserrors.NewValidationError("watchlist-id is required when kind=watchlist", nil)}
		}
	case models.CatalogKindCoachScreen:
		if o.CoachScreenID == "" {
			return []error{mserrors.NewValidationError("coach-screen-id is required when kind=coach_screen", nil)}
		}
	}

	return nil
}

// Filters returns the filter configuration derived from the options.
func (o *CatalogRunOptions) Filters() catalogRunFilters {
	return catalogRunFilters{
		MinComposite: o.MinComposite,
		MinRS:        o.MinRS,
		ExcludeSPACs: o.ExcludeSPACs,
	}
}

func newCatalogRunCmd() *cobra.Command {
	opts := &CatalogRunOptions{}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a catalog report, watchlist, or coach screen",
		Long: `Runs a catalog entry and returns its contents.

Required by kind:

  Kind           Required flag          Runnable
  ----           -------------          --------
  watchlist      --watchlist-id         Yes
  report         --report-id            Yes
  coach_screen   --coach-screen-id      Yes
  screen         (none)                 No, list only

Useful flags:

  --limit, --offset   Page large lists (default limit: 50)
  --fields            Project columns: symbol, price, composite_rating,
                      eps_rating, rs_rating, acc_dis_rating, smr_rating,
                      industry_name, market_cap, volume_dollar_avg_50d
  --min-composite     Minimum composite rating for report/watchlist rows
  --min-rs            Minimum RS rating for report/watchlist rows
  --exclude-spacs     Exclude SPAC/blank-check entries

Coach screen rows are paginated, but field projection and filters do
not behave like report or watchlist rows.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.FromCommand(cmd)

			if errs := opts.Validate(cmd.Context()); errs != nil {
				return errs[0]
			}

			kind, _ := parseCatalogKind(string(opts.Kind))

			entries, total, err := runCatalogEntries(cmd.Context(), ClientFromContext(cmd.Context()), opts, *kind)
			if err != nil {
				return err
			}

			data := map[string]any{"entries": entries}
			return output.WriteSuccess(cmd.OutOrStdout(), data, output.CatalogMeta(string(*kind), total, opts.Limit, opts.Offset))
		},
	}
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(err)
	}
	cmd.Flags().Int("min-composite", 0, "Minimum composite rating for report/watchlist rows (0-99); omitted when unset")
	cmd.Flags().Int("min-rs", 0, "Minimum RS rating for report/watchlist rows (0-99); omitted when unset")
	for _, name := range []string{"min-composite", "min-rs"} {
		if err := cmd.Flags().SetAnnotation(name, structcliFlagGroupAnnotation, []string{"Filtering & Projection"}); err != nil {
			panic(err)
		}
	}
	return cmd
}

type catalogRunFilters struct {
	MinComposite *int
	MinRS        *int
	ExcludeSPACs bool
}

func runCatalogEntries(ctx context.Context, c *client.Client, opts *CatalogRunOptions, kind models.CatalogKind) (result any, total int, err error) {
	filters := opts.Filters()

	switch kind {
	case models.CatalogKindReport:
		result, err := c.RunReport(ctx, opts.ReportID)
		if err != nil {
			return nil, 0, err
		}

		entries := applyCatalogRunFilters(result.Entries, filters)
		return projectWatchlistEntries(paginateSlice(entries, opts.Limit, opts.Offset), opts.Fields), len(entries), nil
	case models.CatalogKindWatchlist:
		result, err := c.RunWatchlist(ctx, opts.WatchlistID)
		if err != nil {
			return nil, 0, err
		}

		entries := applyCatalogRunFilters(result.Entries, filters)
		return projectWatchlistEntries(paginateSlice(entries, opts.Limit, opts.Offset), opts.Fields), len(entries), nil
	case models.CatalogKindCoachScreen:
		result, err := c.RunCoachScreen(ctx, opts.CoachScreenID)
		if err != nil {
			return nil, 0, err
		}

		rows := paginateSlice(result.Rows, opts.Limit, opts.Offset)
		return rows, len(result.Rows), nil
	default:
		return nil, 0, mserrors.NewValidationError("kind must be one of: watchlist, screen, report, coach_screen", nil)
	}
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

// CatalogListOptions holds flags for the catalog list command.
type CatalogListOptions struct {
	Kind models.CatalogKind `flag:"kind" flaggroup:"Filtering" flagdescr:"Filter by catalog kind (watchlist, report, coach_screen, screen); omit to list all sources" flagcustom:"true"`
}

// DefineKind keeps catalog kind parsing in RunE so invalid list kinds continue
// to return the CLI's domain-specific ValidationError type.
func (o *CatalogListOptions) DefineKind(_ string, _ string, descr string, _ reflect.StructField, _ reflect.Value) (pflag.Value, string) { //nolint:gocritic // signature dictated by structcli convention
	value := string(o.Kind)
	return structclivalues.NewString(&value), descr
}

// DecodeKind converts the custom string flag into the typed catalog enum.
func (o *CatalogListOptions) DecodeKind(input any) (any, error) {
	value, _ := input.(string)
	return models.CatalogKind(value), nil
}

func newCatalogListCmd() *cobra.Command {
	opts := &CatalogListOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List catalog entries",
		Long: `Lists catalog entries. The --kind flag is optional.

Valid --kind values: watchlist, screen, report, coach_screen.

Omit --kind to aggregate all sources. Partial source failures can
still return entries from working sources.

Output: entries[] with name, kind, description, and the relevant ID field.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			kind, err := parseCatalogKind(string(opts.Kind))
			if err != nil {
				return err
			}

			c := ClientFromContext(cmd.Context())
			catalog, err := c.ListCatalog(cmd.Context(), kind)
			if err != nil {
				return err
			}

			data := map[string]any{"entries": catalog.Entries}
			meta := output.CatalogMeta(string(opts.Kind), len(catalog.Entries), 0, 0)
			if len(catalog.Errors) > 0 {
				return output.WritePartial(cmd.OutOrStdout(), data, catalog.Errors, meta)
			}
			return output.WriteSuccess(cmd.OutOrStdout(), data, meta)
		},
	}
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(err)
	}
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
