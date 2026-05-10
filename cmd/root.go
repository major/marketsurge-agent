// Package cmd provides the kong command tree for marketsurge-agent.
package cmd

import "github.com/alecthomas/kong"

// CLI is the root command struct for marketsurge-agent.
type CLI struct {
	CookieDB string           `help:"Path to Firefox cookie database." env:"MARKETSURGE_AGENT_COOKIE_DB" name:"cookie-db"`
	Verbose  bool             `help:"Enable verbose logging to stderr." env:"MARKETSURGE_AGENT_VERBOSE"`
	Version  kong.VersionFlag `help:"Show version and exit." short:"V"`

	Chart     ChartCmd     `cmd:"" help:"Show daily OHLCV chart data and live quotes for a stock or ETF."`
	Coach     CoachCmd     `cmd:"" help:"Discover MarketSurge curated watchlists and screens."`
	Columns   ColumnsCmd   `cmd:"" help:"List available MarketSurge data columns (local catalog, no auth required)."`
	Compare   CompareCmd   `cmd:"" help:"Compare key MarketSurge data for multiple stocks or ETFs."`
	Industry  IndustryCmd  `cmd:"" help:"Show industry group relative strength for stocks or ETFs."`
	Overview  OverviewCmd  `cmd:"" help:"Summarize high-level MarketSurge data for a stock or ETF."`
	Reports   ReportsCmd   `cmd:"" help:"List and retrieve MarketSurge predefined reports."`
	Watchlist WatchlistCmd `cmd:"" help:"List and retrieve MarketSurge watchlists."`
}

// ReportsCmd groups report-related subcommands.
type ReportsCmd struct {
	Catalog ReportsCatalogCmd `cmd:"" help:"List built-in MarketSurge report screens (local catalog, no auth required)."`
	Get     ReportsGetCmd     `cmd:"" help:"Get report data for a specific screen ID."`
	Inspect ReportsInspectCmd `cmd:"" help:"Inspect the definition and filter criteria of a screen."`
	List    ReportsListCmd    `cmd:"" help:"List all available screens and reports."`
}

// WatchlistCmd groups watchlist-related subcommands.
type WatchlistCmd struct {
	List WatchlistListCmd `cmd:"" help:"List all saved watchlists."`
	Get  WatchlistGetCmd  `cmd:"" help:"Get symbols in a watchlist by ID."`
}
