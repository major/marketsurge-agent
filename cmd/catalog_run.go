package cmd

import (
	"context"
	"encoding/json"
	"fmt"
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

// CatalogRunOptions holds flags for the catalog run command.
type CatalogRunOptions struct {
	Kind          models.CatalogKind `flag:"kind" flaggroup:"Catalog Selection" flagdescr:"Required catalog kind to run: watchlist uses --watchlist-id, report uses --report-id, coach_screen uses --coach-screen-id; screens are list-only. Example report: --kind report --report-id 124" flagcustom:"true"`
	ReportID      int                `flag:"report-id" flaggroup:"Kind-Specific IDs" flagdescr:"Report ID; required when kind=report. Example report run: --kind report --report-id 124"`
	WatchlistID   int64              `flag:"watchlist-id" flaggroup:"Kind-Specific IDs" flagdescr:"Watchlist ID; required when kind=watchlist. Example watchlist run: --kind watchlist --watchlist-id 99"`
	CoachScreenID string             `flag:"coach-screen-id" flaggroup:"Kind-Specific IDs" flagdescr:"Coach screen ID; required when kind=coach_screen. Example coach screen run: --kind coach_screen --coach-screen-id screen-1"`
	Limit         int                `flag:"limit" flaggroup:"Pagination" flagdescr:"Maximum number of results to return" default:"50"`
	Offset        int                `flag:"offset" flaggroup:"Pagination" flagdescr:"Number of results to skip for pagination"`
	Fields        []string           `flag:"fields" flaggroup:"Filtering & Projection" flagdescr:"Project specific result fields; accepts repeated --fields flags or comma-separated values. Examples: --fields symbol --fields price, or --fields symbol,price,composite_rating. Common fields: symbol, price, composite_rating, eps_rating, rs_rating, acc_dis_rating, smr_rating, industry_name, market_cap, volume_dollar_avg_50d"`
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
		return []error{mserrors.NewValidationError("missing --kind: use --kind watchlist --watchlist-id 12345, --kind report --report-id 67890, or --kind coach_screen --coach-screen-id ID", nil)}
	}

	kind, err := parseCatalogKind(string(o.Kind))
	if err != nil {
		return []error{mserrors.NewValidationError(
			fmt.Sprintf("invalid --kind %q for catalog run: use one of watchlist, report, coach_screen; screen is list-only", o.Kind), nil,
		)}
	}
	if kind == nil {
		return []error{mserrors.NewValidationError("missing --kind: use --kind watchlist --watchlist-id 12345, --kind report --report-id 67890, or --kind coach_screen --coach-screen-id ID", nil)}
	}
	if *kind == models.CatalogKindScreen {
		return []error{mserrors.NewValidationError("invalid --kind screen for catalog run: screens are list-only; use catalog list --kind screen", nil)}
	}

	switch *kind {
	case models.CatalogKindReport:
		if o.ReportID == 0 {
			return []error{mserrors.NewValidationError("missing --report-id: --kind report requires --report-id 67890", nil)}
		}
	case models.CatalogKindWatchlist:
		if o.WatchlistID == 0 {
			return []error{mserrors.NewValidationError("missing --watchlist-id: --kind watchlist requires --watchlist-id 12345", nil)}
		}
	case models.CatalogKindCoachScreen:
		if o.CoachScreenID == "" {
			return []error{mserrors.NewValidationError("missing --coach-screen-id: --kind coach_screen requires --coach-screen-id ID", nil)}
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
		Example: `  marketsurge-agent catalog run --kind report --report-id 124 --fields symbol,price,composite_rating
  marketsurge-agent catalog run --kind watchlist --watchlist-id 99 --limit 25 --exclude-spacs
  marketsurge-agent catalog run --kind coach_screen --coach-screen-id screen-1 --limit 10`,
		Long: `Runs a catalog entry and returns its contents.

Required by kind:

  Kind           Required flag          Runnable
  ----           -------------          --------
  watchlist      --watchlist-id         Yes
  report         --report-id            Yes
  coach_screen   --coach-screen-id      Yes
  screen         (none)                 No, list only

Examples:

  catalog run --kind report --report-id 124 --fields symbol,price,composite_rating
  catalog run --kind watchlist --watchlist-id 99 --limit 25 --exclude-spacs
  catalog run --kind coach_screen --coach-screen-id screen-1 --limit 10

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
	cmd.Flags().Int("min-composite", 0, "Minimum composite rating for report/watchlist rows (0-99); omitted when unset. Example: --min-composite 90")
	cmd.Flags().Int("min-rs", 0, "Minimum RS rating for report/watchlist rows (0-99); omitted when unset. Example: --min-rs 80")
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
		return nil, 0, mserrors.NewValidationError("invalid --kind: use one of watchlist, report, coach_screen for catalog run; screen is list-only", nil)
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
