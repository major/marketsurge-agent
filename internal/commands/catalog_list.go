package commands

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/client"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/internal/output"
)

var validCatalogKinds = map[string]models.CatalogKind{
	string(models.CatalogKindWatchlist):   models.CatalogKindWatchlist,
	string(models.CatalogKindScreen):      models.CatalogKindScreen,
	string(models.CatalogKindReport):      models.CatalogKindReport,
	string(models.CatalogKindCoachScreen): models.CatalogKindCoachScreen,
}

// CatalogListCommand returns the CLI command for listing catalog entries.
func CatalogListCommand(c *client.Client, w io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List catalog entries from MarketSurge sources",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "kind",
				Usage: "Filter by kind: watchlist, screen, report, or coach_screen",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			kind, err := parseCatalogKind(cmd.String("kind"))
			if err != nil {
				return err
			}

			catalog, err := c.ListCatalog(ctx, kind)
			if err != nil {
				return err
			}

			data := map[string]any{"entries": catalog.Entries}
			meta := output.CatalogMeta(cmd.String("kind"), len(catalog.Entries), 0, 0)
			if len(catalog.Errors) > 0 {
				return output.WritePartial(w, data, catalog.Errors, meta)
			}
			return output.WriteSuccess(w, data, meta)
		},
	}
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
