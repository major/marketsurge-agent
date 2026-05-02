package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrequencyEnum(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value    Frequency
		expected string
	}{
		{FrequencyDaily, "DAILY"},
		{FrequencyWeekly, "WEEKLY"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.value))
		})
	}
}

func TestSortDirectionEnum(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value    SortDirection
		expected string
	}{
		{SortDirectionAsc, "ASC"},
		{SortDirectionDesc, "DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.value))
		})
	}
}

func TestLookbackEnum(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value    Lookback
		expected string
	}{
		{Lookback1W, "1W"},
		{Lookback1M, "1M"},
		{Lookback3M, "3M"},
		{Lookback6M, "6M"},
		{Lookback1Y, "1Y"},
		{LookbackYTD, "YTD"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.value))
		})
	}
}

func TestPeriodEnum(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value    Period
		expected string
	}{
		{PeriodDaily, "daily"},
		{PeriodWeekly, "weekly"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.value))
		})
	}
}

func TestCatalogKindEnum(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value    CatalogKind
		expected string
	}{
		{CatalogKindWatchlist, "watchlist"},
		{CatalogKindScreen, "screen"},
		{CatalogKindReport, "report"},
		{CatalogKindCoachScreen, "coach_screen"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.value))
		})
	}
}
