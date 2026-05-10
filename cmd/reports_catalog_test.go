package cmd_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
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

	r, w, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.Close() })

	oldStdout := os.Stdout
	os.Stdout = w
	runErr := (&agentcmd.ReportsCatalogCmd{}).Run()
	_ = w.Close()
	os.Stdout = oldStdout

	require.NoError(t, runErr, "ReportsCatalogCmd.Run() error = %v, want nil", runErr)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	return buf.String()
}
