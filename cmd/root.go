// Package cmd provides the kong command tree for marketsurge-agent.
package cmd

import "github.com/alecthomas/kong"

// CLI is the root command struct for marketsurge-agent.
type CLI struct {
	CookieDB string           `help:"Path to Firefox cookie database." env:"MARKETSURGE_AGENT_COOKIE_DB" name:"cookie-db"`
	Verbose  bool             `help:"Enable verbose logging to stderr." env:"MARKETSURGE_AGENT_VERBOSE"`
	Version  kong.VersionFlag `help:"Show version and exit." short:"V"`

	Compare   CompareCmd   `cmd:"" help:"Compare key MarketSurge data for multiple stocks or ETFs."`
	Industry  IndustryCmd  `cmd:"" help:"Show industry group relative strength for stocks or ETFs."`
	Overview  OverviewCmd  `cmd:"" help:"Summarize high-level MarketSurge data for a stock or ETF."`
	Reports   ReportsCmd   `cmd:"" help:"List and retrieve MarketSurge predefined reports."`
	Watchlist WatchlistCmd `cmd:"" help:"List and retrieve MarketSurge watchlists."`
}

// ReportsCmd groups report-related subcommands.
type ReportsCmd struct {
	List ReportsListCmd `cmd:"" help:"List all available screens and reports."`
	Get  ReportsGetCmd  `cmd:"" help:"Get report data for a specific screen ID."`
}

// WatchlistCmd groups watchlist-related subcommands.
type WatchlistCmd struct {
	List WatchlistListCmd `cmd:"" help:"List all saved watchlists."`
	Get  WatchlistGetCmd  `cmd:"" help:"Get symbols in a watchlist by ID."`
}
