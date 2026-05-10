package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/major/marketsurge-go/marketsurge"

	"github.com/major/marketsurge-agent/cmd"
	"github.com/major/marketsurge-agent/internal/auth"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// version is set via ldflags at build time.
var version = "dev"

func main() {
	var cli cmd.CLI
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	parser, err := kong.New(&cli,
		kong.Name("marketsurge-agent"),
		kong.Description("Query the MarketSurge stock research API."),
		kong.Vars{"version": version},
		kong.BindFor(rootCtx),
		// Lazy client creation: only commands whose Run method accepts
		// *marketsurge.Client will trigger authentication. Commands like
		// "columns" and "reports catalog" work without cookies.
		kong.BindSingletonProvider(func(ctx context.Context) (*marketsurge.Client, error) {
			return newClient(ctx, cli.CookieDB)
		}),
	)
	if err != nil {
		stop()
		mserrors.WriteJSON(os.Stderr, fmt.Errorf("create parser: %w", err))
		os.Exit(mserrors.ExitCodeFor(err))
	}

	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	if cli.Verbose {
		configureLogging()
	}

	if err := ctx.Run(); err != nil {
		stop()
		mserrors.WriteJSON(os.Stderr, err)
		os.Exit(mserrors.ExitCodeFor(err))
	}
	stop()
}

// newClient creates a marketsurge API client by exchanging Firefox cookies for a JWT.
func newClient(ctx context.Context, cookieDBPath string) (*marketsurge.Client, error) {
	jwt, err := auth.ResolveJWT(ctx, cookieDBPath)
	if err != nil {
		return nil, fmt.Errorf("resolve jwt: %w", err)
	}

	client, err := marketsurge.NewClient(marketsurge.WithJWT(jwt))
	if err != nil {
		return nil, fmt.Errorf("create marketsurge client: %w", err)
	}

	return client, nil
}

func configureLogging() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			switch attr.Key {
			case "path":
				return slog.String(attr.Key, "[REDACTED_BROWSER_PROFILE]")
			case "error":
				return slog.String(attr.Key, "[REDACTED_ERROR]")
			default:
				return attr
			}
		},
	})
	slog.SetDefault(slog.New(handler))
}
