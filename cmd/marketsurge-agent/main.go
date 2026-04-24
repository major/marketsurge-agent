// marketsurge-agent is a CLI tool that lets AI agents query the MarketSurge
// stock research API.
//
// This project is unofficial and is not affiliated with, approved by, or
// endorsed by MarketSurge or Investor's Business Daily.
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/major/marketsurge-agent/internal/auth"
	"github.com/major/marketsurge-agent/internal/client"
	"github.com/major/marketsurge-agent/internal/commands"
	"github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/output"
)

// version is set via ldflags at build time.
var version = "dev"

func main() {
	app := buildApp(os.Stdout)
	ctx := context.Background()

	if err := app.Run(ctx, os.Args); err != nil {
		_ = output.WriteError(os.Stdout, err)
		os.Exit(errors.ExitCodeFor(err))
	}
}

// buildApp constructs the root CLI command with all subcommands wired.
// w receives JSON output from command actions.
func buildApp(w io.Writer) *cli.Command {
	// Pre-allocate the client struct. The Before handler populates it
	// after JWT resolution, so command closures see the live client.
	apiClient := &client.Client{}

	return &cli.Command{
		Name:    "marketsurge-agent",
		Usage:   "MarketSurge data agent for AI assistants",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "jwt",
				Usage: "JWT token override",
			},
			&cli.StringFlag{
				Name:  "cookie-db",
				Usage: "Path to Firefox cookies.sqlite",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Enable debug logging to stderr",
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if cmd.Bool("verbose") {
				slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
					Level: slog.LevelDebug,
				})))
			} else {
				slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
			}

			// Skills commands don't require authentication.
			if cmd.Args().Len() > 0 && cmd.Args().First() == "skills" {
				return ctx, nil
			}

			jwt, err := auth.ResolveJWT(ctx, cmd.String("jwt"), cmd.String("cookie-db"))
			if err != nil {
				return ctx, err
			}

			*apiClient = *client.NewClient(jwt)
			return ctx, nil
		},
		CommandNotFound: func(_ context.Context, _ *cli.Command, name string) {
			err := errors.NewValidationError(fmt.Sprintf("unknown command %q", name), nil)
			_ = output.WriteError(w, err)
		},
		Commands: []*cli.Command{
			{
				Name:  "stock",
				Usage: "Stock data commands",
				Commands: []*cli.Command{
					commands.StockGetCommand(apiClient, w),
					commands.StockAnalyzeCommand(apiClient, w),
				},
			},
			{
				Name:  "fundamental",
				Usage: "Fundamental data commands",
				Commands: []*cli.Command{
					commands.FundamentalGetCommand(apiClient, w),
				},
			},
			{
				Name:  "ownership",
				Usage: "Ownership data commands",
				Commands: []*cli.Command{
					commands.OwnershipGetCommand(apiClient, w),
				},
			},
			{
				Name:  "rs-history",
				Usage: "RS rating history commands",
				Commands: []*cli.Command{
					commands.RSHistoryGetCommand(apiClient, w),
				},
			},
			{
				Name:  "chart",
				Usage: "Chart data commands",
				Commands: []*cli.Command{
					commands.ChartHistoryCommand(apiClient, w),
					commands.ChartMarkupsCommand(apiClient, w),
				},
			},
			{
				Name:  "catalog",
				Usage: "Catalog commands",
				Commands: []*cli.Command{
					commands.CatalogListCommand(apiClient, w),
					commands.CatalogRunCommand(apiClient, w),
				},
			},
			{
				Name:  "skills",
				Usage: "Agent skill commands",
				Commands: []*cli.Command{
					commands.SkillsGenerateCommand(w),
				},
			},
		},
	}
}
