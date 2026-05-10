package cmd

import (
	"io"

	"github.com/major/marketsurge-go/marketsurge"
)

// RunForTest exposes ChartCmd.run for the external cmd_test package.
func (c *ChartCmd) RunForTest(client *marketsurge.Client, w io.Writer) error {
	return c.run(client, w)
}

// RunForTest exposes ColumnsCmd.run for the external cmd_test package.
func (c *ColumnsCmd) RunForTest(w io.Writer) error {
	return c.run(w)
}

// RunForTest exposes ReportsCatalogCmd.run for the external cmd_test package.
func (c *ReportsCatalogCmd) RunForTest(w io.Writer) error {
	return c.run(w)
}
