// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/leodido/structcli"
	"github.com/leodido/structcli/config"
	"github.com/leodido/structcli/debug"
	"github.com/leodido/structcli/helptopics"
	structclimcp "github.com/leodido/structcli/mcp"
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
	CookieDB string `flag:"cookie-db" flaggroup:"Authentication & Logging" flagdescr:"Path to Firefox cookies.sqlite file; omit to auto-discover Firefox profiles" flagenv:"true"`
	Verbose  bool   `flag:"verbose" flagshort:"v" flaggroup:"Authentication & Logging" flagdescr:"Enable verbose logging for auth and API requests" flagenv:"true"`
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

  MarketSurge authentication uses Firefox browser cookies. The CLI exchanges
  the local browser session for the API JWT automatically. Cookie databases
  resolve in this order:

    1. --cookie-db path to a Firefox cookies.sqlite file
    2. Auto-discovery from local Firefox profiles

  Sign into MarketSurge in Firefox before running API commands.

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

  - Auth expiry: exit code 32 means Firefox needs an active MarketSurge
    session or the explicit --cookie-db path is not usable.
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
	Version:            version,
	SilenceUsage:       true,
	SilenceErrors:      true,
	TraverseChildren:   true,
	PersistentPreRunE:  persistentPreRunE,
	PersistentPostRunE: persistentPostRunE,
}

func init() {
	if err := structcli.Bind(rootCmd, rootOpts); err != nil {
		panic(err)
	}
	if err := structcli.Setup(rootCmd,
		structcli.WithAppName("marketsurge-agent"),
		structcli.WithConfig(config.Options{ValidateKeys: true}),
		structcli.WithJSONSchema(),
		structcli.WithFlagErrors(),
		structcli.WithHelpTopics(helptopics.Options{ReferenceSection: true}),
		structcli.WithDebug(debug.Options{Exit: true}),
		structcli.WithMCP(structclimcp.Options{
			Name:    "marketsurge-agent",
			Version: version,
			Exclude: []string{
				"completion-bash",
				"completion-fish",
				"completion-powershell",
				"completion-zsh",
			},
		}),
	); err != nil {
		panic(err)
	}
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

func persistentPreRunE(cmd *cobra.Command, args []string) error {
	if isNonAPICommand(cmd) {
		return nil
	}

	// structcli intercepts --jsonschema before hooks run, but
	// --debug-options prints during RunE and exits afterward.
	// Skip auth so the debug output can complete cleanly.
	if cmd.Root().Flags().Changed("debug-options") {
		return nil
	}

	// With TraverseChildren enabled, unknown subcommands fall through
	// to the root command as positional args. Skip auth so Execute()
	// can detect and report them as "unknown command" errors.
	if !cmd.HasParent() && len(args) > 0 {
		return nil
	}

	if rootOpts.Verbose {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	}

	token, err := auth.ResolveJWT(cmd.Context(), rootOpts.CookieDB)
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
		case "completion", "help", "env-vars", "config-keys":
			return true
		}
	}
	return false
}

// Execute runs the root command and writes errors as JSON envelopes.
func Execute() {
	rootCmd.SetArgs(os.Args[1:])

	originalOut := rootCmd.OutOrStdout()
	filter := &configStatusFilter{w: originalOut}
	rootCmd.SetOut(filter)
	executed, err := structcli.ExecuteC(rootCmd)
	if flushErr := filter.Flush(); err == nil && flushErr != nil {
		err = flushErr
	}
	rootCmd.SetOut(nil)

	if err == nil && executed == rootCmd && len(rootCmd.Flags().Args()) > 0 {
		err = fmt.Errorf("unknown command %q", rootCmd.Flags().Arg(0))
	}
	if err != nil {
		_ = output.WriteError(rootCmd.OutOrStderr(), err)
		var mserr interface{ ExitCode() int }
		if errors.As(err, &mserr) {
			os.Exit(mserr.ExitCode())
		}
		os.Exit(1)
	}
}

type configStatusFilter struct {
	w       io.Writer
	pending []byte
}

func (f *configStatusFilter) Write(p []byte) (int, error) {
	f.pending = append(f.pending, p...)

	for {
		line, rest, found := bytes.Cut(f.pending, []byte("\n"))
		if !found {
			break
		}

		line = append(line, '\n')
		if err := f.writeLine(line); err != nil {
			return 0, err
		}
		f.pending = rest
	}

	return len(p), nil
}

func (f *configStatusFilter) Flush() error {
	if len(f.pending) == 0 {
		return nil
	}

	err := f.writeLine(f.pending)
	f.pending = nil
	return err
}

func (f *configStatusFilter) writeLine(line []byte) error {
	if isConfigStatusLine(line) {
		return nil
	}

	_, err := f.w.Write(line)
	return err
}

func isConfigStatusLine(line []byte) bool {
	trimmed := strings.TrimSpace(string(line))
	return trimmed == "Running without a configuration file" || strings.HasPrefix(trimmed, "Using config file: ")
}
