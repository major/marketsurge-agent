package models

import "github.com/leodido/structcli"

// Frequency represents a chart data frequency.
type Frequency string

// Frequency constants.
const (
	FrequencyDaily  Frequency = "DAILY"
	FrequencyWeekly Frequency = "WEEKLY"
)

// SortDirection represents a sort order.
type SortDirection string

// Sort direction constants.
const (
	SortDirectionAsc  SortDirection = "ASC"
	SortDirectionDesc SortDirection = "DESC"
)

// Lookback represents a relative time lookback period.
type Lookback string

// Lookback constants.
const (
	Lookback1W  Lookback = "1W"
	Lookback1M  Lookback = "1M"
	Lookback3M  Lookback = "3M"
	Lookback6M  Lookback = "6M"
	Lookback1Y  Lookback = "1Y"
	LookbackYTD Lookback = "YTD"
)

// Period represents a chart history period granularity.
type Period string

// Period constants.
const (
	PeriodDaily  Period = "daily"
	PeriodWeekly Period = "weekly"
)

func init() {
	structcli.RegisterEnum[Frequency](map[Frequency][]string{
		FrequencyDaily:  {"DAILY"},
		FrequencyWeekly: {"WEEKLY"},
	})
	structcli.RegisterEnum[SortDirection](map[SortDirection][]string{
		SortDirectionAsc:  {"ASC"},
		SortDirectionDesc: {"DESC"},
	})
	structcli.RegisterEnum[Lookback](map[Lookback][]string{
		Lookback1W:  {"1W"},
		Lookback1M:  {"1M"},
		Lookback3M:  {"3M"},
		Lookback6M:  {"6M"},
		Lookback1Y:  {"1Y"},
		LookbackYTD: {"YTD"},
	})
	structcli.RegisterEnum[Period](map[Period][]string{
		PeriodDaily:  {"daily"},
		PeriodWeekly: {"weekly"},
	})
	structcli.RegisterEnum[CatalogKind](map[CatalogKind][]string{
		CatalogKindWatchlist:   {"watchlist"},
		CatalogKindScreen:      {"screen"},
		CatalogKindReport:      {"report"},
		CatalogKindCoachScreen: {"coach_screen"},
	})
}
