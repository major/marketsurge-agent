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

func TestJWTExchangeHeaders(t *testing.T) {
	t.Parallel()
	headers := JWTExchangeHeaders()

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "x-original-host",
			key:      "X-Original-Host",
			expected: "marketsurge.investors.com",
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
	if len(PredefinedReports) != 57 {
		t.Errorf("PredefinedReports length = %d, want 57", len(PredefinedReports))
	}

	for i, report := range PredefinedReports {
		if report.ID == 0 {
			t.Errorf("PredefinedReports[%d] has zero ID", i)
		}
		if report.Name == "" {
			t.Errorf("PredefinedReports[%d] has empty Name", i)
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
