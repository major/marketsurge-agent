package constants

import (
	"testing"
)

func TestGraphQLHeaders(t *testing.T) {
	t.Parallel()
	headers := GraphQLHeaders()

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "apollographql-client-name",
			key:      "Apollographql-Client-Name",
			expected: "marketsurge",
		},
		{
			name:     "dylan-entitlement-token",
			key:      "Dylan-Entitlement-Token",
			expected: "x4ckyhshg90pdq6bwf6n1voijs7r3fdk",
		},
		{
			name:     "Referer",
			key:      "Referer",
			expected: "https://marketsurge.investors.com/",
		},
		{
			name:     "Origin",
			key:      "Origin",
			expected: "https://marketsurge.investors.com",
		},
		{
			name:     "Content-Type",
			key:      "Content-Type",
			expected: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := headers.Get(tt.key)
			if got != tt.expected {
				t.Errorf("header %q = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

func TestPredefinedReports(t *testing.T) {
	t.Parallel()
	if len(PredefinedReports) != 63 {
		t.Errorf("PredefinedReports length = %d, want 63", len(PredefinedReports))
	}

	missingBrowserDescriptions := map[int]bool{
		128: true,
		129: true,
	}
	wantNames := map[int]string{
		44:  "Daily % Change",
		46:  "Legacy 197 Industry Groups",
		49:  "Market Indices",
		128: "IBD Live Ready",
		129: "IBD Live Watch",
		132: "S&P Sectors",
		133: "IBD Industry Groups",
	}
	seenIDs := make(map[int]bool, len(PredefinedReports))

	for i, report := range PredefinedReports {
		if report.ID == 0 {
			t.Errorf("PredefinedReports[%d] has zero ID", i)
		}
		if seenIDs[report.ID] {
			t.Errorf("PredefinedReports[%d] has duplicate ID %d", i, report.ID)
		}
		seenIDs[report.ID] = true
		if report.Name == "" {
			t.Errorf("PredefinedReports[%d] has empty Name", i)
		}
		if report.Description == "" && !missingBrowserDescriptions[report.ID] {
			t.Errorf("PredefinedReports[%d] (%d) has empty Description", i, report.ID)
		}
		if wantName, ok := wantNames[report.ID]; ok && report.Name != wantName {
			t.Errorf("PredefinedReports[%d] name = %q, want %q", i, report.Name, wantName)
		}
	}

	for id := range wantNames {
		if !seenIDs[id] {
			t.Errorf("PredefinedReports is missing ID %d", id)
		}
	}
}

func TestWatchlistColumns(t *testing.T) {
	t.Parallel()
	if len(WatchlistColumns) != 23 {
		t.Errorf("WatchlistColumns length = %d, want 23", len(WatchlistColumns))
	}

	for i, col := range WatchlistColumns {
		if col == "" {
			t.Errorf("WatchlistColumns[%d] is empty", i)
		}
	}
}
