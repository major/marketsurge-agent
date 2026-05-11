package cmd_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentcmd "github.com/major/marketsurge-agent/cmd"
)

func TestReportsCatalog(t *testing.T) {
	output := runReportsCatalog(t)

	var items []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items), "json.Unmarshal(ReportsCatalogCmd output)")
	assert.Greater(t, len(items), 0, "ReportsCatalogCmd should return at least one report screen")

	// Verify first item has required fields.
	assert.Greater(t, items[0].ID, 0, "first report screen ID should be positive")
	assert.NotEmpty(t, items[0].Name, "first report screen name should not be empty")
}

func runReportsCatalog(t *testing.T) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, (&agentcmd.ReportsCatalogCmd{}).RunForTest(&buf), "ReportsCatalogCmd.RunForTest() error")

	return buf.String()
}
