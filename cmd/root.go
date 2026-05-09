// Package cmd provides the kong command tree for marketsurge-agent.
package cmd

import (
	"errors"

	"github.com/alecthomas/kong"
)

// CLI is the root command struct for marketsurge-agent.
type CLI struct {
	Version  kong.VersionFlag `help:"Show version and exit." short:"V"`
	CookieDB string           `help:"Path to Firefox cookie database." env:"MARKETSURGE_AGENT_COOKIE_DB"`
	Verbose  bool             `help:"Enable verbose logging to stderr." env:"MARKETSURGE_AGENT_VERBOSE"`

	Reports ReportsCmd `cmd:"" help:"List and retrieve MarketSurge predefined reports."`
}

// ReportsCmd groups report-related subcommands.
type ReportsCmd struct {
	List ReportsListCmd `cmd:"" help:"List all available screens and reports."`
	Get  ReportsGetCmd  `cmd:"" help:"Get report data for a specific screen ID."`
}

// ReportsListCmd lists all available screens and reports.
// Implementation is in cmd/reports_list.go (added in a later task).
type ReportsListCmd struct{}

// Run is a placeholder - real implementation in reports_list.go.
func (c *ReportsListCmd) Run() error {
	return errors.New("not implemented")
}

// ReportsGetCmd retrieves report data for a specific screen ID.
// Implementation is in cmd/reports_get.go (added in a later task).
type ReportsGetCmd struct {
	ScreenID string `arg:"" help:"Screen ID from 'reports list' output."`
}

// Run is a placeholder - real implementation in reports_get.go.
func (c *ReportsGetCmd) Run() error {
	return errors.New("not implemented")
}
