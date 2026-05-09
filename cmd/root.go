// Package cmd provides the kong command tree for marketsurge-agent.
package cmd

import (
	"errors"

	"github.com/alecthomas/kong"
)

// CLI is the root command struct for marketsurge-agent.
type CLI struct {
	CookieDB string           `help:"Path to Firefox cookie database." env:"MARKETSURGE_AGENT_COOKIE_DB" name:"cookie-db"`
	Verbose  bool             `help:"Enable verbose logging to stderr." env:"MARKETSURGE_AGENT_VERBOSE"`
	Version  kong.VersionFlag `help:"Show version and exit." short:"V"`

	Reports ReportsCmd `cmd:"" help:"List and retrieve MarketSurge predefined reports."`
}

// ReportsCmd groups report-related subcommands.
type ReportsCmd struct {
	List ReportsListCmd `cmd:"" help:"List all available screens and reports."`
	Get  ReportsGetCmd  `cmd:"" help:"Get report data for a specific screen ID."`
}

// ReportsGetCmd retrieves report data for a specific screen ID.
// Full implementation in cmd/reports_get.go (Task 8).
type ReportsGetCmd struct {
	ScreenID string `arg:"" help:"Screen ID from 'reports list' output."`
}

// Run is a placeholder - real implementation in reports_get.go.
func (c *ReportsGetCmd) Run() error {
	return errors.New("not implemented")
}
