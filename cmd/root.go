// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/major/marketsurge-agent/internal/auth"
	"github.com/major/marketsurge-agent/internal/client"
	"github.com/major/marketsurge-agent/internal/output"
)

// version is set via ldflags at build time.
var version = "dev"

type clientKeyType struct{}

var clientKey = clientKeyType{}

// RootOptions holds persistent flag values for the root command.
type RootOptions struct {
	JWT      string
	CookieDB string
	Verbose  bool
}

// BindFlags registers persistent flags on the root command and binds them to RootOptions fields.
func (o *RootOptions) BindFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&o.JWT, "jwt", "", "JWT token for authentication (overrides env var and cookie)")
	cmd.PersistentFlags().StringVar(&o.CookieDB, "cookie-db", "", "Path to Firefox cookie database file")
	cmd.PersistentFlags().BoolVar(&o.Verbose, "verbose", false, "Enable verbose logging")
}

// rootOpts is the package-level instance of RootOptions, initialized by init().
var rootOpts = &RootOptions{}

// rootCmd is the root cobra command. Initialized as a package-level variable so
// that each command file's init() can safely call rootCmd.AddCommand() regardless
// of source-file initialization order.
var rootCmd = &cobra.Command{
	Use:   "marketsurge-agent",
	Short: "Query the MarketSurge stock research API",
	Long: `marketsurge-agent queries the MarketSurge stock research API.
All output is structured JSON with semantic exit codes.

Auth

  MarketSurge requires a valid JWT. Credentials resolve in this order
  (first non-empty wins):

    1. --jwt flag
    2. MARKETSURGE_JWT env var
    3. --cookie-db path to a Firefox cookies.sqlite file
    4. Auto-discovery from local Firefox profiles

  For automation, set the env var:

    export MARKETSURGE_JWT="your-jwt-here"

Output

  Success envelope:
    {"data": {...}, "metadata": {"symbol": "AAPL"}, "timestamp": "..."}

  Error response:
    {"error": "symbol not found", "code": 31, "message": "...", "timestamp": "..."}

  Partial response (stock analyze with multiple symbols):
    {"data": {...}, "errors": [...], "metadata": {...}, "timestamp": "..."}

Exit Codes

     0 - Success
    30 - Validation error (bad args, missing fields)
    31 - Symbol not found
    32 - Authentication error (missing/expired token, cookie failures)
    33 - API or HTTP error (GraphQL errors, rate limiting, server errors)

Gotchas

  - JWT expiry: exit code 32 means the user needs to refresh their token
    or ensure Firefox has an active MarketSurge session.
  - Chart date params: --start-date/--end-date and --lookback are
    mutually exclusive.
  - Catalog kind: catalog run requires --kind and the matching ID flag.
  - Multi-symbol: stock analyze and rs-history get return data keyed by
    ticker when given multiple symbols.
  - Summary mode: stock analyze --summary returns compact screening
    objects for ranking many candidates. Metadata includes mode: "summary".
  - Compact mode: stock analyze --compact strips duplicate formatted
    string fields while keeping raw numeric values.
  - Flat mode: stock analyze --flat flattens nested objects into
    single-level keys.
  - Batch tickers: stock analyze --tickers AAPL,NVDA,TSLA accepts
    comma-separated symbols. Positional and --tickers can be combined.`,
	Version: version,
	SilenceUsage:       true,
	SilenceErrors:      true,
	PersistentPreRunE:  persistentPreRunE,
	PersistentPostRunE: persistentPostRunE,
}

func init() {
	rootOpts.BindFlags(rootCmd)
}

// ClientFromContext returns the API client stored by PersistentPreRunE.
func ClientFromContext(ctx context.Context) *client.Client {
	c, ok := ctx.Value(clientKey).(*client.Client)
	if !ok || c == nil {
		panic("ClientFromContext: no client in context (PersistentPreRunE did not run)")
	}
	return c
}

// ContextWithClient stores the API client in a command context.
func ContextWithClient(ctx context.Context, c *client.Client) context.Context {
	return context.WithValue(ctx, clientKey, c)
}

func persistentPreRunE(cmd *cobra.Command, _ []string) error {
	if isNonAPICommand(cmd) {
		return nil
	}

	if rootOpts.Verbose {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	}

	jwt := rootOpts.JWT
	if jwt == "" {
		jwt = os.Getenv("MARKETSURGE_JWT")
	}

	token, err := auth.ResolveJWT(cmd.Context(), jwt, rootOpts.CookieDB)
	if err != nil {
		return err
	}

	c := client.NewClient(token)
	cmd.SetContext(ContextWithClient(cmd.Context(), c))
	return nil
}

func persistentPostRunE(cmd *cobra.Command, _ []string) error {
	c, _ := cmd.Context().Value(clientKey).(*client.Client)
	if c != nil {
		_ = c.Close()
	}
	return nil
}

func isNonAPICommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "completion", "help":
			return true
		}
	}
	return false
}

// Execute runs the root command and writes errors as JSON envelopes.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_ = output.WriteError(rootCmd.OutOrStderr(), err)
		var mserr interface{ ExitCode() int }
		if errors.As(err, &mserr) {
			os.Exit(mserr.ExitCode())
		}
		os.Exit(1)
	}
}
