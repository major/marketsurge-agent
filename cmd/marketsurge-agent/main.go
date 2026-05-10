package main

import (
	"context"
	"fmt"
	"os"

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
	ctx := kong.Parse(&cli,
		kong.Name("marketsurge-agent"),
		kong.Description("Query the MarketSurge stock research API."),
		kong.Vars{"version": version},
	)

	client, err := newClient(cli.CookieDB)
	if err != nil {
		mserrors.WriteJSON(os.Stderr, err)
		os.Exit(mserrors.ExitCodeFor(err))
	}

	if err := ctx.Run(client); err != nil {
		mserrors.WriteJSON(os.Stderr, err)
		os.Exit(mserrors.ExitCodeFor(err))
	}
}

// newClient creates a marketsurge API client by exchanging Firefox cookies for a JWT.
func newClient(cookieDBPath string) (*marketsurge.Client, error) {
	jwt, err := auth.ResolveJWT(context.Background(), cookieDBPath)
	if err != nil {
		return nil, fmt.Errorf("resolve jwt: %w", err)
	}

	client, err := marketsurge.NewClient(marketsurge.WithJWT(jwt))
	if err != nil {
		return nil, fmt.Errorf("create marketsurge client: %w", err)
	}

	return client, nil
}
