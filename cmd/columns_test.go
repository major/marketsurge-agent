package cmd_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentcmd "github.com/major/marketsurge-agent/cmd"
)

func TestColumnsAll(t *testing.T) {
	output := runColumns(t, agentcmd.ColumnsCmd{})

	var items []struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		Description string `json:"description"`
		Category    string `json:"category"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &items), "json.Unmarshal(ColumnsCmd output)")
	assert.Greater(t, len(items), 0, "ColumnsCmd should return at least one column")

	// Verify the first item has required fields.
	assert.NotEmpty(t, items[0].Name, "first column name should not be empty")
	assert.NotEmpty(t, items[0].DisplayName, "first column displayName should not be empty")
}

func TestColumnsByCategory(t *testing.T) {
	// First get all columns to find a valid category.
	allOutput := runColumns(t, agentcmd.ColumnsCmd{})
	var allItems []struct {
		Category string `json:"category"`
	}
	require.NoError(t, json.Unmarshal([]byte(allOutput), &allItems))
	require.Greater(t, len(allItems), 0, "need at least one column to test category filter")

	// Find a non-empty category.
	var category string
	for _, item := range allItems {
		if item.Category != "" {
			category = item.Category
			break
		}
	}
	require.NotEmpty(t, category, "need at least one column with a non-empty category")

	// Filter by that category.
	filteredOutput := runColumns(t, agentcmd.ColumnsCmd{Category: category})
	var filteredItems []struct {
		Category string `json:"category"`
	}
	require.NoError(t, json.Unmarshal([]byte(filteredOutput), &filteredItems))
	assert.Greater(t, len(filteredItems), 0, "filtered columns should not be empty")

	for i, item := range filteredItems {
		assert.Equal(t, category, item.Category, "filteredItems[%d].category = %q, want %q", i, item.Category, category)
	}
}

func TestColumnsByCategoryEmpty(t *testing.T) {
	output := runColumns(t, agentcmd.ColumnsCmd{Category: "nonexistent-category-xyz"})
	assert.Equal(t, "[]\n", output, "ColumnsCmd with unknown category should return empty array")
}

func runColumns(t *testing.T, cmd agentcmd.ColumnsCmd) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, cmd.RunForTest(&buf), "ColumnsCmd.RunForTest() error")

	return buf.String()
}
