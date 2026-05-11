package cmd

import (
	"context"
	"io"

	"github.com/major/marketsurge-go/marketsurge"
)

// RunForTest exposes ChartCmd.run for the external cmd_test package.
func (c *ChartCmd) RunForTest(client *marketsurge.Client, w io.Writer) error {
	return c.run(context.Background(), client, w)
}

// RunForTest exposes ColumnsCmd.run for the external cmd_test package.
func (c *ColumnsCmd) RunForTest(w io.Writer) error {
	return c.run(w)
}

// RunForTest exposes ReportsCatalogCmd.run for the external cmd_test package.
func (c *ReportsCatalogCmd) RunForTest(w io.Writer) error {
	return c.run(w)
}

// RunForTest exposes ReportsGetCmd.run for the external cmd_test package.
func (c *ReportsGetCmd) RunForTest(client *marketsurge.Client, w io.Writer) error {
	return c.run(context.Background(), client, w)
}

// RunForTest exposes CompareCmd.run for the external cmd_test package.
func (c *CompareCmd) RunForTest(client *marketsurge.Client, w io.Writer) error {
	return c.run(context.Background(), client, w)
}

// RunForTest exposes OverviewCmd.run for the external cmd_test package.
func (c *OverviewCmd) RunForTest(client *marketsurge.Client, w io.Writer) error {
	return c.run(context.Background(), client, w)
}

// RunForTest exposes ReportsInspectCmd.run for the external cmd_test package.
func (c *ReportsInspectCmd) RunForTest(client *marketsurge.Client, w io.Writer) error {
	return c.run(context.Background(), client, w)
}

// RunForTest exposes WatchlistGetCmd.run for the external cmd_test package.
func (c *WatchlistGetCmd) RunForTest(client *marketsurge.Client, w io.Writer) error {
	return c.run(context.Background(), client, w)
}

// RunForTest exposes ReportsListCmd.run for the external cmd_test package.
func (c *ReportsListCmd) RunForTest(client *marketsurge.Client, w io.Writer) error {
	return c.run(context.Background(), client, w)
}

// RunForTest exposes WatchlistListCmd.run for the external cmd_test package.
func (c *WatchlistListCmd) RunForTest(client *marketsurge.Client, w io.Writer) error {
	return c.run(context.Background(), client, w)
}

// RunForTest exposes CoachCmd.run for the external cmd_test package.
func (c *CoachCmd) RunForTest(client *marketsurge.Client, w io.Writer) error {
	return c.run(context.Background(), client, w)
}

// RunForTest exposes IndustryCmd.run for the external cmd_test package.
func (c *IndustryCmd) RunForTest(client *marketsurge.Client, w io.Writer) error {
	return c.run(context.Background(), client, w)
}
