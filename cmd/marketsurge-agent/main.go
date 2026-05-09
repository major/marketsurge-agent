package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/major/marketsurge-agent/cmd"
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
	if err := ctx.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
