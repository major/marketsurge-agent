package queries

import (
	"testing"
)

func TestLoadAllQueries(t *testing.T) {
	t.Parallel()
	queries := []string{
		"other_market_data.graphql",
		"fundamentals.graphql",
		"ownership.graphql",
		"rs_rating_ri_panel.graphql",
		"chart_market_data.graphql",
		"chart_market_data_weekly.graphql",
		"chart_markups.graphql",
		"adhoc_screen.graphql",
		"run_screen.graphql",
		"flagged_symbols.graphql",
		"watchlist_names.graphql",
		"screens.graphql",
		"coach_tree.graphql",
	}

	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
	t.Parallel()
	content, err := Load(query)
			if err != nil {
				t.Fatalf("failed to load %s: %v", query, err)
			}
			if content == "" {
				t.Fatalf("query %s is empty", query)
			}
})
	}
}

func TestLoadUnknownQuery(t *testing.T) {
	t.Parallel()
	_, err := Load("nonexistent.graphql")
	if err == nil {
		t.Fatal("expected error for nonexistent query, got nil")
	}
}
