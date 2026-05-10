package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

var defaultCompareColumns = []marketsurge.ColumnName{
	marketsurge.ColumnSymbol,
	marketsurge.ColumnCompanyName,
	marketsurge.ColumnPrice,
	marketsurge.ColumnPriceNetChg,
	marketsurge.ColumnPricePctChg,
	marketsurge.ColumnPricePctOff52WHigh,
	marketsurge.ColumnCompositeRating,
	marketsurge.ColumnEPSRating,
	marketsurge.ColumnRSRating,
	marketsurge.ColumnSMRRating,
	marketsurge.ColumnAccDisRating,
	marketsurge.ColumnVolumePctChgVs50DAvgVolume,
	marketsurge.ColumnVolumeAvg50Day,
	marketsurge.ColumnVolumeDollarAvg50D,
	marketsurge.ColumnUpDownVolumeRatio,
	marketsurge.ColumnPricePctChgVs50DaySMA,
	marketsurge.ColumnPricePctChgVs200DaySMA,
	marketsurge.ColumnPricePctChgLast1M,
	marketsurge.ColumnPricePctChgLast3M,
	marketsurge.ColumnPricePctChgLast6M,
	marketsurge.ColumnATRPct21D,
	marketsurge.ColumnEpsPctChgYoyLastReportedQ,
	marketsurge.ColumnEpsEstPctChgQ1,
	marketsurge.ColumnEpsEstPctChgY1,
	marketsurge.ColumnSalesPctChgYoy1QAgo,
	marketsurge.ColumnPriceEarningsRatioForward,
	marketsurge.ColumnIndustryGroupRank,
	marketsurge.ColumnIndustryName,
	marketsurge.ColumnIndustryGroupRSRatingLetter,
	marketsurge.ColumnFundsNumberHoldingOwnership,
	marketsurge.ColumnEPSDueDate,
	marketsurge.ColumnIPODate,
	marketsurge.ColumnMarketCapIntraday,
	marketsurge.ColumnBlueDotMostRecentDate,
	marketsurge.ColumnBlueDotCount45Day,
}

// CompareCmd retrieves compact, LLM-friendly comparison data for multiple stocks or ETFs.
type CompareCmd struct {
	Symbols []string `arg:"" help:"Stock or ETF symbols to compare, such as AMD NVDA MSFT."`
	Columns []string `help:"Response columns to include." env:"MARKETSURGE_AGENT_COMPARE_COLUMNS" sep:","`
}

type compareItem struct {
	Ticker       string               `json:"ticker"`
	Name         *string              `json:"name,omitempty"`
	Ratings      *compareRatings      `json:"ratings,omitempty"`
	Price        *comparePrice        `json:"price,omitempty"`
	Volume       *compareVolume       `json:"volume,omitempty"`
	Momentum     *compareMomentum     `json:"momentum,omitempty"`
	Fundamentals *compareFundamentals `json:"fundamentals,omitempty"`
	Industry     *compareIndustry     `json:"industry,omitempty"`
	Ownership    *compareOwnership    `json:"ownership,omitempty"`
	Events       *compareEvents       `json:"events,omitempty"`
	Columns      map[string]*string   `json:"columns,omitempty"`
}

type compareRatings struct {
	Composite                *string `json:"composite,omitempty"`
	EPSRating                *string `json:"epsRating,omitempty"`
	RelativeStrength         *string `json:"relativeStrength,omitempty"`
	SalesMarginsROE          *string `json:"salesMarginsROE,omitempty"`
	AccumulationDistribution *string `json:"accumulationDistribution,omitempty"`
}

type comparePrice struct {
	Last                 *string `json:"last,omitempty"`
	NetChange            *string `json:"netChange,omitempty"`
	PercentChange        *string `json:"percentChange,omitempty"`
	PercentOff52WeekHigh *string `json:"percentOff52WeekHigh,omitempty"`
	ATRPercent21D        *string `json:"atrPercent21d,omitempty"`
	MarketCap            *string `json:"marketCap,omitempty"`
}

type compareVolume struct {
	PercentChangeVs50D *string `json:"percentChangeVs50dAvg,omitempty"`
	Average50D         *string `json:"average50d,omitempty"`
	AverageDollar50D   *string `json:"averageDollar50d,omitempty"`
	UpDownRatio        *string `json:"upDownRatio,omitempty"`
}

type compareMomentum struct {
	Vs50DaySMA  *string `json:"vs50DaySMA,omitempty"`
	Vs200DaySMA *string `json:"vs200DaySMA,omitempty"`
	OneMonth    *string `json:"oneMonth,omitempty"`
	ThreeMonths *string `json:"threeMonths,omitempty"`
	SixMonths   *string `json:"sixMonths,omitempty"`
}

type compareFundamentals struct {
	LatestQuarterEPSChange *string `json:"latestQuarterEPSChange,omitempty"`
	NextQuarterEPSChange   *string `json:"nextQuarterEPSChange,omitempty"`
	CurrentYearEPSChange   *string `json:"currentYearEPSChange,omitempty"`
	LatestQuarterSales     *string `json:"latestQuarterSalesChange,omitempty"`
	ForwardPE              *string `json:"forwardPE,omitempty"`
}

type compareIndustry struct {
	Name          *string `json:"name,omitempty"`
	GroupRank     *string `json:"groupRank,omitempty"`
	GroupRSRating *string `json:"groupRSRating,omitempty"`
}

type compareOwnership struct {
	FundsHolding *string `json:"fundsHolding,omitempty"`
}

type compareEvents struct {
	EPSDueDate            *string `json:"epsDueDate,omitempty"`
	IPODate               *string `json:"ipoDate,omitempty"`
	BlueDotMostRecentDate *string `json:"blueDotMostRecentDate,omitempty"`
	BlueDotCount45Day     *string `json:"blueDotCount45d,omitempty"`
}

// Run executes the comparison query and writes row objects as a JSON array.
func (c *CompareCmd) Run(ctx context.Context, client *marketsurge.Client) error {
	return c.run(ctx, client, os.Stdout)
}

func (c *CompareCmd) run(ctx context.Context, client *marketsurge.Client, w io.Writer) error {
	symbols := normalizeSymbols(c.Symbols)
	if len(symbols) == 0 {
		return mserrors.NewValidationError("at least one symbol is required", errors.New("empty symbols"))
	}

	columns, err := compareResponseColumns(c.Columns)
	if err != nil {
		return err
	}

	req := marketsurge.NewMarketDataAdhocScreenRequest(columns)
	req.IncludeSource.Instruments = &marketsurge.AdhocScreenInstruments{
		Symbols: symbols,
		Dialect: marketsurge.DefaultChartSymbolDialectType,
	}

	resp, err := client.MarketDataAdhocScreen(ctx, req)
	if err != nil {
		return clientError("comparison data request failed", err)
	}

	if err := json.NewEncoder(w).Encode(compareRows(resp)); err != nil {
		return mserrors.NewAPIError("failed to write comparison output", err)
	}

	return nil
}

func compareResponseColumns(names []string) ([]marketsurge.AdhocScreenResponseColumn, error) {
	if len(names) == 0 {
		columns := make([]marketsurge.AdhocScreenResponseColumn, len(defaultCompareColumns))
		for i, name := range defaultCompareColumns {
			columns[i] = marketsurge.AdhocScreenResponseColumn{Name: name}
		}
		return columns, nil
	}

	columns := make([]marketsurge.AdhocScreenResponseColumn, 0, len(names))
	hasSymbol := false
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		if trimmed == string(marketsurge.ColumnSymbol) {
			hasSymbol = true
		}
		columns = append(columns, marketsurge.AdhocScreenResponseColumn{Name: marketsurge.ColumnName(trimmed)})
	}
	if len(columns) == 0 {
		return nil, mserrors.NewValidationError("at least one column is required", errors.New("empty columns"))
	}
	if !hasSymbol {
		columns = append([]marketsurge.AdhocScreenResponseColumn{{Name: marketsurge.ColumnSymbol}}, columns...)
	}

	return columns, nil
}

func normalizeSymbols(symbols []string) []string {
	seen := make(map[string]bool, len(symbols))
	normalized := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		trimmed := strings.ToUpper(strings.TrimSpace(symbol))
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		normalized = append(normalized, trimmed)
	}

	return normalized
}

func compareRows(resp *marketsurge.AdhocScreenResponse) []compareItem {
	if resp == nil || resp.MarketDataAdhocScreen == nil {
		return []compareItem{}
	}

	rows := make([]compareItem, 0, len(resp.MarketDataAdhocScreen.ResponseValues))
	for _, cells := range resp.MarketDataAdhocScreen.ResponseValues {
		columns := compareColumnValues(cells)
		rows = append(rows, newCompareItem(columns))
	}

	return rows
}

func compareColumnValues(cells []marketsurge.AdhocScreenCell) map[string]*string {
	columns := make(map[string]*string, len(cells))
	for _, cell := range cells {
		if cell.MDItem == nil || cell.MDItem.Name == nil {
			continue
		}
		columns[*cell.MDItem.Name] = cell.Value
	}

	return columns
}

func newCompareItem(columns map[string]*string) compareItem {
	item := compareItem{
		Ticker:  compareValue(columns, marketsurge.ColumnSymbol),
		Name:    columns[string(marketsurge.ColumnCompanyName)],
		Columns: columns,
	}
	item.Ratings = compareRatingsFrom(columns)
	item.Price = comparePriceFrom(columns)
	item.Volume = compareVolumeFrom(columns)
	item.Momentum = compareMomentumFrom(columns)
	item.Fundamentals = compareFundamentalsFrom(columns)
	item.Industry = compareIndustryFrom(columns)
	item.Ownership = compareOwnershipFrom(columns)
	item.Events = compareEventsFrom(columns)

	return item
}

func compareRatingsFrom(columns map[string]*string) *compareRatings {
	ratings := &compareRatings{
		Composite:                columns[string(marketsurge.ColumnCompositeRating)],
		EPSRating:                columns[string(marketsurge.ColumnEPSRating)],
		RelativeStrength:         columns[string(marketsurge.ColumnRSRating)],
		SalesMarginsROE:          columns[string(marketsurge.ColumnSMRRating)],
		AccumulationDistribution: columns[string(marketsurge.ColumnAccDisRating)],
	}
	if ratings.Composite == nil && ratings.EPSRating == nil && ratings.RelativeStrength == nil && ratings.SalesMarginsROE == nil && ratings.AccumulationDistribution == nil {
		return nil
	}

	return ratings
}

func comparePriceFrom(columns map[string]*string) *comparePrice {
	price := &comparePrice{
		Last:                 columns[string(marketsurge.ColumnPrice)],
		NetChange:            columns[string(marketsurge.ColumnPriceNetChg)],
		PercentChange:        columns[string(marketsurge.ColumnPricePctChg)],
		PercentOff52WeekHigh: columns[string(marketsurge.ColumnPricePctOff52WHigh)],
		ATRPercent21D:        columns[string(marketsurge.ColumnATRPct21D)],
		MarketCap:            columns[string(marketsurge.ColumnMarketCapIntraday)],
	}
	if price.Last == nil && price.NetChange == nil && price.PercentChange == nil && price.PercentOff52WeekHigh == nil && price.ATRPercent21D == nil && price.MarketCap == nil {
		return nil
	}

	return price
}

func compareVolumeFrom(columns map[string]*string) *compareVolume {
	volume := &compareVolume{
		PercentChangeVs50D: columns[string(marketsurge.ColumnVolumePctChgVs50DAvgVolume)],
		Average50D:         columns[string(marketsurge.ColumnVolumeAvg50Day)],
		AverageDollar50D:   columns[string(marketsurge.ColumnVolumeDollarAvg50D)],
		UpDownRatio:        columns[string(marketsurge.ColumnUpDownVolumeRatio)],
	}
	if volume.PercentChangeVs50D == nil && volume.Average50D == nil && volume.AverageDollar50D == nil && volume.UpDownRatio == nil {
		return nil
	}

	return volume
}

func compareMomentumFrom(columns map[string]*string) *compareMomentum {
	momentum := &compareMomentum{
		Vs50DaySMA:  columns[string(marketsurge.ColumnPricePctChgVs50DaySMA)],
		Vs200DaySMA: columns[string(marketsurge.ColumnPricePctChgVs200DaySMA)],
		OneMonth:    columns[string(marketsurge.ColumnPricePctChgLast1M)],
		ThreeMonths: columns[string(marketsurge.ColumnPricePctChgLast3M)],
		SixMonths:   columns[string(marketsurge.ColumnPricePctChgLast6M)],
	}
	if momentum.Vs50DaySMA == nil && momentum.Vs200DaySMA == nil && momentum.OneMonth == nil && momentum.ThreeMonths == nil && momentum.SixMonths == nil {
		return nil
	}

	return momentum
}

func compareFundamentalsFrom(columns map[string]*string) *compareFundamentals {
	fundamentals := &compareFundamentals{
		LatestQuarterEPSChange: columns[string(marketsurge.ColumnEpsPctChgYoyLastReportedQ)],
		NextQuarterEPSChange:   columns[string(marketsurge.ColumnEpsEstPctChgQ1)],
		CurrentYearEPSChange:   columns[string(marketsurge.ColumnEpsEstPctChgY1)],
		LatestQuarterSales:     columns[string(marketsurge.ColumnSalesPctChgYoy1QAgo)],
		ForwardPE:              columns[string(marketsurge.ColumnPriceEarningsRatioForward)],
	}
	if fundamentals.LatestQuarterEPSChange == nil && fundamentals.NextQuarterEPSChange == nil && fundamentals.CurrentYearEPSChange == nil && fundamentals.LatestQuarterSales == nil && fundamentals.ForwardPE == nil {
		return nil
	}

	return fundamentals
}

func compareIndustryFrom(columns map[string]*string) *compareIndustry {
	industry := &compareIndustry{
		Name:          columns[string(marketsurge.ColumnIndustryName)],
		GroupRank:     columns[string(marketsurge.ColumnIndustryGroupRank)],
		GroupRSRating: columns[string(marketsurge.ColumnIndustryGroupRSRatingLetter)],
	}
	if industry.Name == nil && industry.GroupRank == nil && industry.GroupRSRating == nil {
		return nil
	}

	return industry
}

func compareOwnershipFrom(columns map[string]*string) *compareOwnership {
	ownership := &compareOwnership{FundsHolding: columns[string(marketsurge.ColumnFundsNumberHoldingOwnership)]}
	if ownership.FundsHolding == nil {
		return nil
	}

	return ownership
}

func compareEventsFrom(columns map[string]*string) *compareEvents {
	events := &compareEvents{
		EPSDueDate:            columns[string(marketsurge.ColumnEPSDueDate)],
		IPODate:               columns[string(marketsurge.ColumnIPODate)],
		BlueDotMostRecentDate: columns[string(marketsurge.ColumnBlueDotMostRecentDate)],
		BlueDotCount45Day:     columns[string(marketsurge.ColumnBlueDotCount45Day)],
	}
	if events.EPSDueDate == nil && events.IPODate == nil && events.BlueDotMostRecentDate == nil && events.BlueDotCount45Day == nil {
		return nil
	}

	return events
}

func compareValue(columns map[string]*string, name marketsurge.ColumnName) string {
	value := columns[string(name)]
	if value == nil {
		return ""
	}

	return *value
}
