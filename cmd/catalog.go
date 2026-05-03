// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
)

const (
	defaultCatalogRunLimit       = 50
	structcliFlagGroupAnnotation = "leodido/structcli/flag-groups"
	validCatalogKindList         = "watchlist, screen, report, coach_screen"
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
		return nil, mserrors.NewValidationError(fmt.Sprintf("invalid --kind %q: use one of %s", value, validCatalogKindList), nil)
	}

	return &kind, nil
}
