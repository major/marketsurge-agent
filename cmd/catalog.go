// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"github.com/spf13/cobra"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/internal/output"
)

func init() { rootCmd.AddCommand(newCatalogCmd()) }

func newCatalogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "Catalog commands",
	}
	cmd.AddCommand(newCatalogListCmd())
	// NOTE: Task 11 will add newCatalogRunCmd() here
	return cmd
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
