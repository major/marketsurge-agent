package cmd

import (
	"reflect"

	"github.com/leodido/structcli"
	structclivalues "github.com/leodido/structcli/values"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/internal/output"
)

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
