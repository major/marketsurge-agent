package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

const (
	weeklyTrendLookbackYears = -1
	chartDateTimeLayout      = "2006-01-02T15:04:05.000Z"
)

// OverviewCmd retrieves a compact, LLM-friendly summary for a stock or ETF.
type OverviewCmd struct {
	Symbol string `arg:"" help:"Stock or ETF symbol to summarize, such as AMD."`
}

type overviewItem struct {
	Ticker                string                    `json:"ticker"`
	Name                  *string                   `json:"name,omitempty"`
	BusinessDescription   *string                   `json:"businessDescription,omitempty"`
	Type                  *string                   `json:"type,omitempty"`
	IPODate               *string                   `json:"ipoDate,omitempty"`
	IPOPrice              *overviewFloat            `json:"ipoPrice,omitempty"`
	Ratings               *overviewRatings          `json:"ratings,omitempty"`
	Valuation             *overviewValuation        `json:"valuation,omitempty"`
	Price                 *overviewPrice            `json:"price,omitempty"`
	HistoricalPrices      []overviewHistoricalPrice `json:"historicalPrices,omitempty"`
	VolumeAverages        []overviewVolumeAvg       `json:"volumeAverages,omitempty"`
	RelativeStrengthTrend *overviewRS               `json:"relativeStrengthTrend,omitempty"`
	ANTs                  *overviewANTs             `json:"ants,omitempty"`
	Patterns              []overviewPattern         `json:"patterns,omitempty"`
	TightAreas            []overviewTight           `json:"tightAreas,omitempty"`
	Industry              *overviewIndustry         `json:"industry,omitempty"`
	Ownership             *overviewOwnership        `json:"ownership,omitempty"`
	Fundamentals          *overviewFundamentals     `json:"fundamentals,omitempty"`
	EarningsCalendar      *overviewEarningsCalendar `json:"earningsCalendar,omitempty"`
	CorporateActions      *overviewCorporateActions `json:"corporateActions,omitempty"`
	ProfitMargins         []overviewProfitMargin    `json:"profitMargins,omitempty"`
	GrowthRates           *overviewGrowthRates      `json:"growthRates,omitempty"`
	Risk                  *overviewRisk             `json:"risk,omitempty"`
	WeeklyTrend           *overviewWeeklyTrend      `json:"weeklyTrend,omitempty"`
}

type overviewRatings struct {
	Composite                *int    `json:"composite,omitempty"`
	EPSRating                *int    `json:"epsRating,omitempty"`
	RelativeStrength         *int    `json:"relativeStrength,omitempty"`
	SalesMarginsROE          *string `json:"salesMarginsROE,omitempty"`
	AccumulationDistribution *string `json:"accumulationDistribution,omitempty"`
}

type overviewPrice struct {
	MarketCap           *overviewFloat `json:"marketCap,omitempty"`
	AvgDollarVol50D     *overviewFloat `json:"avgDollarVolume50d,omitempty"`
	ATRPercent          *overviewFloat `json:"atrPercent,omitempty"`
	UpDownVolumeRatio   *overviewFloat `json:"upDownVolumeRatio,omitempty"`
	QuarterChange       *overviewFloat `json:"quarterPercentChange,omitempty"`
	BlueDotDaily        *bool          `json:"blueDotDaily,omitempty"`
	BlueDotWeekly       *bool          `json:"blueDotWeekly,omitempty"`
	BlueDotDailyEvents  []string       `json:"blueDotDailyEvents,omitempty"`
	BlueDotWeeklyEvents []string       `json:"blueDotWeeklyEvents,omitempty"`
}

type overviewRS struct {
	NewHigh *bool                `json:"newHigh,omitempty"`
	History []overviewRSSnapshot `json:"history,omitempty"`
}

type overviewRSSnapshot struct {
	Value  *int    `json:"value,omitempty"`
	Letter *string `json:"letter,omitempty"`
	Period *string `json:"period,omitempty"`
	Offset *string `json:"offset,omitempty"`
}

type overviewANTs struct {
	Count int      `json:"count"`
	Dates []string `json:"dates,omitempty"`
}

type overviewPattern struct {
	Type        *string        `json:"type,omitempty"`
	Status      *string        `json:"status,omitempty"`
	Stage       *string        `json:"stage,omitempty"`
	Length      *int           `json:"length,omitempty"`
	Depth       *overviewFloat `json:"depth,omitempty"`
	Start       *string        `json:"start,omitempty"`
	End         *string        `json:"end,omitempty"`
	Pivot       *overviewFloat `json:"pivot,omitempty"`
	PivotDate   *string        `json:"pivotDate,omitempty"`
	HandleDepth *overviewFloat `json:"handleDepth,omitempty"`
	HandleLen   *int           `json:"handleLen,omitempty"`
}

type overviewTight struct {
	Start  *string `json:"start,omitempty"`
	End    *string `json:"end,omitempty"`
	Length *int    `json:"length,omitempty"`
}

type overviewIndustry struct {
	Name   *string                `json:"name,omitempty"`
	Sector *string                `json:"sector,omitempty"`
	Code   *string                `json:"code,omitempty"`
	Stocks *int                   `json:"stocks,omitempty"`
	Ranks  []overviewPeriodInt    `json:"ranks,omitempty"`
	RS     []overviewPeriodLetter `json:"rs,omitempty"`
}

type overviewPeriodInt struct {
	Value  *int    `json:"value,omitempty"`
	Period *string `json:"period,omitempty"`
	Offset *string `json:"offset,omitempty"`
}

type overviewPeriodLetter struct {
	Value  *int    `json:"value,omitempty"`
	Letter *string `json:"letter,omitempty"`
	Period *string `json:"period,omitempty"`
	Offset *string `json:"offset,omitempty"`
}

type overviewOwnership struct {
	FundsFloatPct *overviewFloat       `json:"fundsFloatPct,omitempty"`
	Funds         []overviewFundHolder `json:"funds,omitempty"`
}

type overviewFundHolder struct {
	Date  *string `json:"date,omitempty"`
	Funds *string `json:"funds,omitempty"`
}

type overviewFundamentals struct {
	DebtPct           *overviewFloat             `json:"debtPct,omitempty"`
	RNDPct            *overviewFloat             `json:"rndPct,omitempty"`
	CashFlowPerShare  *overviewFloat             `json:"cashFlowPerShare,omitempty"`
	EarningsStability *int                       `json:"earningsStability,omitempty"`
	CEODate           *string                    `json:"ceoDate,omitempty"`
	ReportedEPS       *overviewFundamentalPeriod `json:"reportedEPS,omitempty"`
	ReportedSales     *overviewFundamentalPeriod `json:"reportedSales,omitempty"`
	EstimatedEPS      *overviewFundamentalPeriod `json:"estimatedEPS,omitempty"`
	EstimatedSales    *overviewFundamentalPeriod `json:"estimatedSales,omitempty"`
}

type overviewFloat struct {
	Value     *float64 `json:"v,omitempty"`
	Formatted *string  `json:"f,omitempty"`
}

type overviewFundamentalPeriod struct {
	Value               *overviewFloat `json:"value,omitempty"`
	YearOverYearPercent *overviewFloat `json:"yearOverYearPercent,omitempty"`
	PeriodOffset        *string        `json:"periodOffset,omitempty"`
	PeriodEndDate       *string        `json:"periodEndDate,omitempty"`
	Period              *string        `json:"period,omitempty"`
	RevisionDirection   *string        `json:"revisionDirection,omitempty"`
}

type overviewRisk struct {
	Alpha         *overviewFloat         `json:"alpha,omitempty"`
	Beta          *overviewFloat         `json:"beta,omitempty"`
	ShortInterest *overviewShortInterest `json:"shortInterest,omitempty"`
}

type overviewShortInterest struct {
	DaysToCover              *overviewFloat `json:"daysToCover,omitempty"`
	DaysToCoverPercentChange *overviewFloat `json:"daysToCoverPercentChange,omitempty"`
	PercentOfFloat           *overviewFloat `json:"percentOfFloat,omitempty"`
	Volume                   *overviewFloat `json:"volume,omitempty"`
}

type overviewValuation struct {
	PriceToEarnings        *overviewFloat `json:"priceToEarnings,omitempty"`
	ForwardPriceToEarnings *overviewFloat `json:"forwardPriceToEarnings,omitempty"`
	PriceToSales           *overviewFloat `json:"priceToSales,omitempty"`
	PriceToCashFlow        *overviewFloat `json:"priceToCashFlow,omitempty"`
	Yield                  *overviewFloat `json:"yield,omitempty"`
	PriceToEarningsVsSP500 *overviewFloat `json:"priceToEarningsVsSP500,omitempty"`
}

type overviewHistoricalPrice struct {
	Period        *string        `json:"period,omitempty"`
	Offset        *string        `json:"offset,omitempty"`
	EndDate       *string        `json:"endDate,omitempty"`
	High          *overviewFloat `json:"high,omitempty"`
	HighDate      *string        `json:"highDate,omitempty"`
	Low           *overviewFloat `json:"low,omitempty"`
	LowDate       *string        `json:"lowDate,omitempty"`
	Close         *overviewFloat `json:"close,omitempty"`
	PercentChange *overviewFloat `json:"percentChange,omitempty"`
}

type overviewVolumeAvg struct {
	Value  *float64 `json:"value,omitempty"`
	Period *string  `json:"period,omitempty"`
	Offset *string  `json:"offset,omitempty"`
}

type overviewEarningsCalendar struct {
	EPSDueDate       *string `json:"epsDueDate,omitempty"`
	EPSDueDateStatus *string `json:"epsDueDateStatus,omitempty"`
	LastReportedDate *string `json:"lastReportedDate,omitempty"`
}

type overviewCorporateActions struct {
	NextExDividendDate *string            `json:"nextExDividendDate,omitempty"`
	Dividends          []overviewDividend `json:"dividends,omitempty"`
	Splits             []string           `json:"splits,omitempty"`
	Spinoffs           []string           `json:"spinoffs,omitempty"`
}

type overviewDividend struct {
	Amount *overviewFloat `json:"amount,omitempty"`
	Change *string        `json:"change,omitempty"`
	ExDate *string        `json:"exDate,omitempty"`
}

type overviewProfitMargin struct {
	Period         *string        `json:"period,omitempty"`
	Offset         *string        `json:"offset,omitempty"`
	EndDate        *string        `json:"endDate,omitempty"`
	GrossMargin    *overviewFloat `json:"grossMargin,omitempty"`
	PreTaxMargin   *overviewFloat `json:"preTaxMargin,omitempty"`
	AfterTaxMargin *overviewFloat `json:"afterTaxMargin,omitempty"`
	ReturnOnEquity *overviewFloat `json:"returnOnEquity,omitempty"`
}

type overviewGrowthRates struct {
	EPS   []overviewGrowthRate `json:"eps,omitempty"`
	Sales []overviewGrowthRate `json:"sales,omitempty"`
}

type overviewGrowthRate struct {
	Value  *overviewFloat `json:"value,omitempty"`
	Period *string        `json:"period,omitempty"`
}

type overviewWeeklyTrend struct {
	Period *string             `json:"period,omitempty"`
	Latest *overviewChartPoint `json:"latest,omitempty"`
	Quote  *overviewChartQuote `json:"quote,omitempty"`
}

type overviewChartPoint struct {
	Start  *string        `json:"start,omitempty"`
	End    *string        `json:"end,omitempty"`
	Open   *overviewFloat `json:"open,omitempty"`
	High   *overviewFloat `json:"high,omitempty"`
	Low    *overviewFloat `json:"low,omitempty"`
	Last   *overviewFloat `json:"last,omitempty"`
	Volume *overviewFloat `json:"volume,omitempty"`
}

type overviewChartQuote struct {
	TradeDateTime *string        `json:"tradeDateTime,omitempty"`
	Timeliness    *string        `json:"timeliness,omitempty"`
	QuoteType     *string        `json:"quoteType,omitempty"`
	Last          *overviewFloat `json:"last,omitempty"`
	PercentChange *overviewFloat `json:"percentChange,omitempty"`
	NetChange     *overviewFloat `json:"netChange,omitempty"`
	Volume        *overviewFloat `json:"volume,omitempty"`
}

// Run executes the overview query and writes a compact JSON array.
func (c *OverviewCmd) Run(ctx context.Context, client *marketsurge.Client) error {
	return c.run(ctx, client, os.Stdout)
}

func (c *OverviewCmd) run(ctx context.Context, client *marketsurge.Client, w io.Writer) error {
	symbol := strings.ToUpper(strings.TrimSpace(c.Symbol))
	if symbol == "" {
		return mserrors.NewValidationError("symbol is required", errors.New("empty symbol"))
	}

	marketData, err := client.OtherMarketData(ctx, marketsurge.NewOtherMarketDataRequest(symbol))
	if err != nil {
		return clientError("market data request failed", err)
	}
	if marketData == nil || len(marketData.MarketData) == 0 {
		return mserrors.NewSymbolNotFoundError("symbol not found", nil, symbol)
	}

	item := newOverviewItem(symbol, &marketData.MarketData[0])

	rsData, err := client.RSRatingRIPanel(ctx, marketsurge.NewRSRatingRIPanelRequest(symbol))
	if err != nil {
		return clientError("RS rating request failed", err)
	}
	addRSOverview(&item, rsData)

	ownership, err := client.Ownership(ctx, marketsurge.NewOwnershipRequest(symbol))
	if err != nil {
		return clientError("ownership request failed", err)
	}
	addOwnershipOverview(&item, ownership)

	fundamentals, err := client.Fundamentals(ctx, marketsurge.NewFundamentalsRequest(symbol))
	if err != nil {
		return clientError("fundamentals request failed", err)
	}
	addFundamentalsOverview(&item, fundamentals)

	weeklyTrend, err := client.ChartMarketDataWeekly(ctx, newWeeklyTrendRequest(symbol, time.Now().UTC()))
	if err != nil {
		return clientError("weekly trend request failed", err)
	}
	item.WeeklyTrend = overviewWeeklyTrendFrom(weeklyTrend)

	if err := json.NewEncoder(w).Encode([]overviewItem{item}); err != nil {
		return mserrors.NewAPIError("failed to write overview output", err)
	}

	return nil
}

func newOverviewItem(symbol string, data *marketsurge.MarketDataItem) overviewItem {
	item := overviewItem{Ticker: symbol}
	if data.Symbology != nil {
		if data.Symbology.Company != nil {
			item.Name = data.Symbology.Company.CompanyName
			item.BusinessDescription = data.Symbology.Company.BusinessDescription
		}
		if data.Symbology.Instrument != nil {
			item.Type = data.Symbology.Instrument.SubType
			item.IPODate = dateValue(data.Symbology.Instrument.IPODate)
			item.IPOPrice = overviewCurrencyFloat(data.Symbology.Instrument.IPOPrice)
		}
	}
	item.Ratings = overviewRatingsFrom(data.Ratings)
	item.Valuation = overviewValuationFrom(data.PricingStatistics)
	item.Price = overviewPriceFrom(data.PricingStatistics)
	item.HistoricalPrices = overviewHistoricalPricesFrom(data.PricingStatistics)
	item.VolumeAverages = overviewVolumeAvgsFrom(data.PricingStatistics)
	item.ANTs = overviewANTsFrom(data.PricingStatistics)
	item.Patterns = overviewPatternsFrom(data.PatternInfo)
	item.TightAreas = overviewTightAreasFrom(data.PatternInfo)
	item.Industry = overviewIndustryFrom(data.Industry)
	item.Ownership = overviewMarketDataOwnershipFrom(data.Ownership)
	item.Fundamentals = overviewFundFrom(data.Fundamentals, data.Financials)
	item.CorporateActions = overviewCorporateActionsFrom(data.CorporateActions)
	item.EarningsCalendar = overviewEarningsCalendarFrom(data.Financials)
	item.ProfitMargins = overviewProfitMarginsFrom(data.Financials)
	item.GrowthRates = overviewGrowthRatesFrom(data.Financials)
	item.Risk = overviewRiskFrom(data.PricingStatistics)
	return item
}

func overviewRatingsFrom(ratings *marketsurge.MDRatings) *overviewRatings {
	if ratings == nil {
		return nil
	}
	return &overviewRatings{
		Composite:                firstRatingValue(ratings.CompRating),
		EPSRating:                firstRatingValue(ratings.EPSRating),
		RelativeStrength:         firstRatingValue(ratings.RSRating),
		SalesMarginsROE:          firstRatingLetter(ratings.SMRRating),
		AccumulationDistribution: firstRatingLetter(ratings.ADRating),
	}
}

func overviewPriceFrom(stats *marketsurge.MDPricingStatistics) *overviewPrice {
	if stats == nil {
		return nil
	}
	price := &overviewPrice{}
	if stats.EndOfDayStatistics != nil {
		eod := stats.EndOfDayStatistics
		price.MarketCap = overviewFormattedFloat(eod.MarketCapitalization)
		price.AvgDollarVol50D = overviewFormattedFloat(eod.AvgDollarVolume50Day)
		price.ATRPercent = firstATRPct(eod.AverageTrueRangePercent)
		price.UpDownVolumeRatio = overviewScaledFloat(eod.UpDownVolumeRatio)
		price.QuarterChange = firstQuarterChange(eod.HistoricalPriceStatistics)
		price.BlueDotDailyEvents = formattedStringValues(eod.BlueDotDailyEvents)
		price.BlueDotWeeklyEvents = formattedStringValues(eod.BlueDotWeeklyEvents)
	}
	if stats.IntradayStatistics != nil {
		price.BlueDotDaily = stats.IntradayStatistics.IsDailyBlueDotEvent
		price.BlueDotWeekly = stats.IntradayStatistics.IsWeeklyBlueDotEvent
	}
	return price
}

func overviewANTsFrom(stats *marketsurge.MDPricingStatistics) *overviewANTs {
	if stats == nil || stats.EndOfDayStatistics == nil || len(stats.EndOfDayStatistics.AntEvents) == 0 {
		return nil
	}
	dates := make([]string, 0, len(stats.EndOfDayStatistics.AntEvents))
	for _, event := range stats.EndOfDayStatistics.AntEvents {
		if event.Value != "" {
			dates = append(dates, event.Value)
		}
	}
	return &overviewANTs{Count: len(stats.EndOfDayStatistics.AntEvents), Dates: dates}
}

func overviewPatternsFrom(info *marketsurge.MDPatternInfo) []overviewPattern {
	if info == nil || len(info.Patterns) == 0 {
		return nil
	}
	patterns := make([]overviewPattern, 0, len(info.Patterns))
	for i := range info.Patterns {
		pattern := &info.Patterns[i]
		patterns = append(patterns, overviewPattern{
			Type:        pattern.PatternType,
			Status:      pattern.BaseStatus,
			Stage:       pattern.BaseStage,
			Length:      pattern.BaseLength,
			Depth:       overviewScaledFloat(pattern.BaseDepth),
			Start:       dateValue(pattern.BaseStartDate),
			End:         dateValue(pattern.BaseEndDate),
			Pivot:       overviewCurrencyFloat(pattern.PivotPrice),
			PivotDate:   dateValue(pattern.PivotDate),
			HandleDepth: overviewScaledFloat(pattern.HandleDepth),
			HandleLen:   pattern.HandleLength,
		})
	}
	return patterns
}

func overviewTightAreasFrom(info *marketsurge.MDPatternInfo) []overviewTight {
	if info == nil || len(info.TightAreas) == 0 {
		return nil
	}
	tightAreas := make([]overviewTight, 0, len(info.TightAreas))
	for _, area := range info.TightAreas {
		tightAreas = append(tightAreas, overviewTight{
			Start:  dateValue(area.StartDate),
			End:    dateValue(area.EndDate),
			Length: area.Length,
		})
	}
	return tightAreas
}

func overviewIndustryFrom(industry *marketsurge.MDIndustry) *overviewIndustry {
	if industry == nil {
		return nil
	}
	result := &overviewIndustry{
		Name:   industry.Name,
		Sector: industry.Sector,
		Code:   industry.IndCode,
		Stocks: industry.NumberOfStocksInGroup,
		Ranks:  make([]overviewPeriodInt, 0, len(industry.GroupRanks)),
		RS:     make([]overviewPeriodLetter, 0, len(industry.GroupRS)),
	}
	for _, rank := range industry.GroupRanks {
		result.Ranks = append(result.Ranks, overviewPeriodInt{Value: rank.Value, Period: rank.Period, Offset: rank.PeriodOffset})
	}
	for _, rs := range industry.GroupRS {
		result.RS = append(result.RS, overviewPeriodLetter{Value: rs.Value, Letter: rs.LetterValue, Period: rs.Period, Offset: rs.PeriodOffset})
	}
	return result
}

func overviewMarketDataOwnershipFrom(ownership *marketsurge.MDOwnership) *overviewOwnership {
	if ownership == nil {
		return nil
	}
	return &overviewOwnership{FundsFloatPct: overviewScaledFloat(ownership.FundsFloatPercentHeld)}
}

// overviewFundFrom extracts fundamentals from MDFundamentals (OtherMarketData)
// and enriches with cash flow and earnings stability from MDFinancials.
func overviewFundFrom(fund *marketsurge.MDFundamentals, fin *marketsurge.MDFinancials) *overviewFundamentals {
	if fund == nil && fin == nil {
		return nil
	}
	result := &overviewFundamentals{}
	if fund != nil {
		result.DebtPct = overviewFormattedFloat(fund.DebtPercent)
		result.RNDPct = overviewScaledFloat(fund.ResearchAndDevelopmentPercentLastQtr)
		result.CEODate = dateValue(fund.NewCEODate)
	}
	if fin != nil {
		result.CashFlowPerShare = overviewFormattedFloat(fin.CashFlowPerShareLastYear)
		if fin.ConsensusFinancials != nil && fin.ConsensusFinancials.EPS != nil {
			result.EarningsStability = fin.ConsensusFinancials.EPS.EarningsStability
		}
	}
	return result
}

func addRSOverview(item *overviewItem, resp *marketsurge.RSRatingRIPanelResponse) {
	if resp == nil || len(resp.MarketData) == 0 {
		return
	}
	data := resp.MarketData[0]
	result := &overviewRS{}
	if data.PricingStatistics != nil && data.PricingStatistics.IntradayStatistics != nil {
		result.NewHigh = data.PricingStatistics.IntradayStatistics.RSLineNewHigh
	}
	if data.Ratings != nil && len(data.Ratings.RSRating) > 0 {
		result.History = make([]overviewRSSnapshot, 0, len(data.Ratings.RSRating))
		for _, rating := range data.Ratings.RSRating {
			result.History = append(result.History, overviewRSSnapshot{
				Value:  rating.Value,
				Letter: rating.LetterValue,
				Period: rating.Period,
				Offset: rating.PeriodOffset,
			})
		}
	}
	item.RelativeStrengthTrend = result
}

func addOwnershipOverview(item *overviewItem, resp *marketsurge.OwnershipResponse) {
	if resp == nil || len(resp.MarketData) == 0 || resp.MarketData[0].Ownership == nil {
		return
	}
	if item.Ownership == nil {
		item.Ownership = &overviewOwnership{}
	}
	ownership := resp.MarketData[0].Ownership
	if item.Ownership.FundsFloatPct == nil && ownership.FundsFloatPercentHeld != nil {
		item.Ownership.FundsFloatPct = &overviewFloat{Formatted: ownership.FundsFloatPercentHeld.FormattedValue}
	}
	item.Ownership.Funds = make([]overviewFundHolder, 0, len(ownership.FundOwnershipSummary))
	for _, summary := range ownership.FundOwnershipSummary {
		item.Ownership.Funds = append(item.Ownership.Funds, overviewFundHolder{
			Date:  ownershipDateValue(summary.Date),
			Funds: ownershipFormattedValue(summary.NumberOfFundsHeld),
		})
	}
}

func addFundamentalsOverview(item *overviewItem, resp *marketsurge.FundamentalsResponse) {
	if resp == nil || len(resp.MarketData) == 0 || resp.MarketData[0].Financials == nil {
		return
	}
	if item.Fundamentals == nil {
		item.Fundamentals = &overviewFundamentals{}
	}
	financials := resp.MarketData[0].Financials

	if financials.ConsensusFinancials != nil {
		consensus := financials.ConsensusFinancials
		if consensus.EPS != nil && len(consensus.EPS.ReportedEarnings) > 0 {
			item.Fundamentals.ReportedEPS = overviewReportedFundamentalFrom(&consensus.EPS.ReportedEarnings[0])
		}
		if consensus.Sales != nil && len(consensus.Sales.ReportedSales) > 0 {
			item.Fundamentals.ReportedSales = overviewReportedFundamentalFrom(&consensus.Sales.ReportedSales[0])
		}
	}
	if financials.Estimates != nil {
		estimates := financials.Estimates
		if len(estimates.EPSEstimates) > 0 {
			item.Fundamentals.EstimatedEPS = overviewEstimateFundamentalFrom(&estimates.EPSEstimates[0])
		}
		if len(estimates.SalesEstimates) > 0 {
			item.Fundamentals.EstimatedSales = overviewEstimateFundamentalFrom(&estimates.SalesEstimates[0])
		}
	}
}

func newWeeklyTrendRequest(symbol string, now time.Time) marketsurge.ChartMarketDataWeeklyRequest {
	end := now.UTC()
	start := end.AddDate(weeklyTrendLookbackYears, 0, 0)
	return marketsurge.NewChartMarketDataWeeklyRequest(
		[]string{symbol},
		start.Format(chartDateTimeLayout),
		end.Format(chartDateTimeLayout),
		"ONE_WEEK",
	)
}

func overviewWeeklyTrendFrom(resp *marketsurge.ChartMarketDataResponse) *overviewWeeklyTrend {
	if resp == nil || len(resp.MarketData) == 0 || resp.MarketData[0].Pricing == nil {
		return nil
	}
	pricing := resp.MarketData[0].Pricing
	result := &overviewWeeklyTrend{Quote: overviewChartQuoteFrom(pricing.Quote)}
	if pricing.TimeSeries == nil {
		return result
	}
	if pricing.TimeSeries.Period != "" {
		result.Period = &pricing.TimeSeries.Period
	}
	points := pricing.TimeSeries.DataPoints
	if len(points) > 0 {
		result.Latest = overviewChartPointFrom(&points[len(points)-1])
	}
	return result
}

func overviewRiskFrom(stats *marketsurge.MDPricingStatistics) *overviewRisk {
	if stats == nil || stats.EndOfDayStatistics == nil {
		return nil
	}
	eod := stats.EndOfDayStatistics
	result := &overviewRisk{
		Alpha:         overviewScaledFloat(eod.Alpha),
		Beta:          overviewScaledFloat(eod.Beta),
		ShortInterest: overviewShortInterestFrom(eod.ShortInterest),
	}
	if result.Alpha == nil && result.Beta == nil && result.ShortInterest == nil {
		return nil
	}
	return result
}

func overviewShortInterestFrom(shortInterest *marketsurge.MDShortInterest) *overviewShortInterest {
	if shortInterest == nil {
		return nil
	}
	return &overviewShortInterest{
		DaysToCover:              overviewFormattedFloat(shortInterest.DaysToCover),
		DaysToCoverPercentChange: overviewFormattedFloat(shortInterest.DaysToCoverPercentChange),
		PercentOfFloat:           overviewScaledFloat(shortInterest.PercentOfFloat),
		Volume:                   overviewScaledFloat(shortInterest.Volume),
	}
}

func overviewReportedFundamentalFrom(period *marketsurge.FundamentalsReportedPeriod) *overviewFundamentalPeriod {
	if period == nil {
		return nil
	}
	return &overviewFundamentalPeriod{
		Value:               overviewFundamentalFloat(period.Value),
		YearOverYearPercent: overviewFundamentalFloat(period.PercentChangeYOY),
		PeriodOffset:        period.PeriodOffset,
		PeriodEndDate:       fundamentalDateValue(period.PeriodEndDate),
	}
}

func overviewEstimateFundamentalFrom(estimate *marketsurge.FundamentalsEstimate) *overviewFundamentalPeriod {
	if estimate == nil {
		return nil
	}
	return &overviewFundamentalPeriod{
		Value:               overviewFundamentalFloat(estimate.Value),
		YearOverYearPercent: overviewFundamentalFloat(estimate.PercentChangeYOY),
		PeriodOffset:        estimate.PeriodOffset,
		Period:              estimate.Period,
		RevisionDirection:   estimate.RevisionDirection,
	}
}

func overviewChartPointFrom(point *marketsurge.ChartDataPoint) *overviewChartPoint {
	if point == nil {
		return nil
	}
	return &overviewChartPoint{
		Start:  stringValue(point.StartDateTime),
		End:    stringValue(point.EndDateTime),
		Open:   overviewChartValue(point.Open),
		High:   overviewChartValue(point.High),
		Low:    overviewChartValue(point.Low),
		Last:   overviewChartValue(point.Last),
		Volume: overviewChartValue(point.Volume),
	}
}

func overviewChartQuoteFrom(quote *marketsurge.ChartQuote) *overviewChartQuote {
	if quote == nil {
		return nil
	}
	return &overviewChartQuote{
		TradeDateTime: quote.TradeDateTime,
		Timeliness:    quote.Timeliness,
		QuoteType:     quote.QuoteType,
		Last:          overviewChartFormattedFloat(quote.Last),
		PercentChange: overviewChartFormattedFloat(quote.PercentChange),
		NetChange:     overviewChartFormattedFloat(quote.NetChange),
		Volume:        overviewChartFormattedFloat(quote.Volume),
	}
}

func formattedStringValues(values []marketsurge.MDFormattedString) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value.Value != nil && *value.Value != "" {
			result = append(result, *value.Value)
			continue
		}
		if value.FormattedValue != nil && *value.FormattedValue != "" {
			result = append(result, *value.FormattedValue)
		}
	}
	return result
}

func overviewFundamentalFloat(value *marketsurge.FundamentalsFormattedValue) *overviewFloat {
	if value == nil {
		return nil
	}
	return &overviewFloat{Value: value.Value, Formatted: value.FormattedValue}
}

func fundamentalDateValue(value *marketsurge.FundamentalsDateValue) *string {
	if value == nil || value.Value == "" {
		return nil
	}
	return &value.Value
}

func overviewChartValue(value *marketsurge.ChartValue) *overviewFloat {
	if value == nil {
		return nil
	}
	return &overviewFloat{Value: value.Value}
}

func overviewChartFormattedFloat(value *marketsurge.ChartFormattedValue) *overviewFloat {
	if value == nil {
		return nil
	}
	return &overviewFloat{Value: value.Value, Formatted: value.FormattedValue}
}

func stringValue(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func firstRatingValue(ratings []marketsurge.MDRating) *int {
	if len(ratings) == 0 {
		return nil
	}
	return ratings[0].Value
}

func firstRatingLetter(ratings []marketsurge.MDRating) *string {
	if len(ratings) == 0 {
		return nil
	}
	return ratings[0].LetterValue
}

func firstATRPct(values []marketsurge.MDAverageTrueRangePercent) *overviewFloat {
	if len(values) == 0 {
		return nil
	}
	return &overviewFloat{Value: values[0].Value, Formatted: values[0].FormattedValue}
}

func firstQuarterChange(values []marketsurge.MDHistoricalPriceStatistic) *overviewFloat {
	if len(values) == 0 || values[0].PricePercentChange == nil {
		return nil
	}
	return overviewFormattedFloat(values[0].PricePercentChange)
}

func overviewFormattedFloat(value *marketsurge.MDFormattedFloat) *overviewFloat {
	if value == nil {
		return nil
	}
	return &overviewFloat{Value: value.Value, Formatted: value.FormattedValue}
}

func overviewScaledFloat(value *marketsurge.MDScaledFloat) *overviewFloat {
	if value == nil {
		return nil
	}
	return &overviewFloat{Value: value.Value, Formatted: value.FormattedValue}
}

func overviewCurrencyFloat(value *marketsurge.MDCurrencyValue) *overviewFloat {
	if value == nil {
		return nil
	}
	return &overviewFloat{Value: value.Value, Formatted: value.FormattedValue}
}

func dateValue(value *marketsurge.MDDateValue) *string {
	if value == nil || value.Value == "" {
		return nil
	}
	return &value.Value
}

func ownershipDateValue(value *marketsurge.OwnershipDateValue) *string {
	if value == nil || value.Value == "" {
		return nil
	}
	return &value.Value
}

func ownershipFormattedValue(value *marketsurge.OwnershipFormattedValue) *string {
	if value == nil {
		return nil
	}
	return value.FormattedValue
}

func overviewValuationFrom(stats *marketsurge.MDPricingStatistics) *overviewValuation {
	if stats == nil || stats.IntradayStatistics == nil {
		return nil
	}
	intra := stats.IntradayStatistics
	result := &overviewValuation{
		PriceToEarnings:        overviewScaledFloat(intra.PriceToEarningsRatio),
		ForwardPriceToEarnings: overviewScaledFloat(intra.ForwardPriceToEarningsRatio),
		PriceToSales:           overviewScaledFloat(intra.PriceToSalesRatio),
		PriceToCashFlow:        overviewScaledFloat(intra.PriceToCashFlowRatio),
		Yield:                  overviewScaledFloat(intra.Yield),
		PriceToEarningsVsSP500: overviewScaledFloat(intra.PriceToEarningsVsSP500),
	}
	if result.PriceToEarnings == nil && result.ForwardPriceToEarnings == nil &&
		result.PriceToSales == nil && result.PriceToCashFlow == nil &&
		result.Yield == nil && result.PriceToEarningsVsSP500 == nil {
		return nil
	}
	return result
}

func overviewHistoricalPricesFrom(stats *marketsurge.MDPricingStatistics) []overviewHistoricalPrice {
	if stats == nil || stats.EndOfDayStatistics == nil || len(stats.EndOfDayStatistics.HistoricalPriceStatistics) == 0 {
		return nil
	}
	hps := stats.EndOfDayStatistics.HistoricalPriceStatistics
	result := make([]overviewHistoricalPrice, 0, len(hps))
	for i := range hps {
		h := &hps[i]
		result = append(result, overviewHistoricalPrice{
			Period:        h.Period,
			Offset:        h.PeriodOffset,
			EndDate:       formattedStringValue(h.PeriodEndDate),
			High:          overviewFormattedFloat(h.PriceHigh),
			HighDate:      formattedStringValue(h.PriceHighDate),
			Low:           overviewFormattedFloat(h.PriceLow),
			LowDate:       formattedStringValue(h.PriceLowDate),
			Close:         overviewFormattedFloat(h.PriceClose),
			PercentChange: overviewFormattedFloat(h.PricePercentChange),
		})
	}
	return result
}

func overviewVolumeAvgsFrom(stats *marketsurge.MDPricingStatistics) []overviewVolumeAvg {
	if stats == nil || stats.EndOfDayStatistics == nil || len(stats.EndOfDayStatistics.VolumeMovingAverages) == 0 {
		return nil
	}
	vmas := stats.EndOfDayStatistics.VolumeMovingAverages
	result := make([]overviewVolumeAvg, 0, len(vmas))
	for i := range vmas {
		v := &vmas[i]
		result = append(result, overviewVolumeAvg{
			Value:  v.Value,
			Period: v.Period,
			Offset: v.PeriodOffset,
		})
	}
	return result
}

func overviewCorporateActionsFrom(ca *marketsurge.MDCorporateActions) *overviewCorporateActions {
	if ca == nil {
		return nil
	}
	result := &overviewCorporateActions{}

	if ca.DividendNextReportedExDate != nil {
		if ca.DividendNextReportedExDate.Value != nil && *ca.DividendNextReportedExDate.Value != "" {
			result.NextExDividendDate = ca.DividendNextReportedExDate.Value
		} else if ca.DividendNextReportedExDate.FormattedValue != nil && *ca.DividendNextReportedExDate.FormattedValue != "" {
			result.NextExDividendDate = ca.DividendNextReportedExDate.FormattedValue
		}
	}

	if len(ca.Dividends) > 0 {
		result.Dividends = make([]overviewDividend, 0, len(ca.Dividends))
		for i := range ca.Dividends {
			d := &ca.Dividends[i]
			result.Dividends = append(result.Dividends, overviewDividend{
				Amount: overviewFormattedFloat(d.Amount),
				Change: d.ChangeIndicator,
				ExDate: dateValue(d.ExDate),
			})
		}
	}

	if len(ca.Splits) > 0 {
		result.Splits = make([]string, 0, len(ca.Splits))
		for _, s := range ca.Splits {
			if d := dateValue(s.SplitDate); d != nil {
				result.Splits = append(result.Splits, *d)
			}
		}
	}

	if len(ca.Spinoffs) > 0 {
		result.Spinoffs = make([]string, 0, len(ca.Spinoffs))
		for _, s := range ca.Spinoffs {
			if d := dateValue(s.ExDate); d != nil {
				result.Spinoffs = append(result.Spinoffs, *d)
			}
		}
	}

	// Return nil if nothing was populated.
	if result.NextExDividendDate == nil && len(result.Dividends) == 0 &&
		len(result.Splits) == 0 && len(result.Spinoffs) == 0 {
		return nil
	}
	return result
}

func overviewEarningsCalendarFrom(fin *marketsurge.MDFinancials) *overviewEarningsCalendar {
	if fin == nil {
		return nil
	}
	result := &overviewEarningsCalendar{
		EPSDueDateStatus: fin.EPSDueDateStatus,
		LastReportedDate: dateValue(fin.EPSLastReportedDate),
	}
	if fin.EPSDueDate != nil {
		if fin.EPSDueDate.Value != nil && *fin.EPSDueDate.Value != "" {
			result.EPSDueDate = fin.EPSDueDate.Value
		} else if fin.EPSDueDate.FormattedValue != nil && *fin.EPSDueDate.FormattedValue != "" {
			result.EPSDueDate = fin.EPSDueDate.FormattedValue
		}
	}
	if result.EPSDueDate == nil && result.EPSDueDateStatus == nil && result.LastReportedDate == nil {
		return nil
	}
	return result
}

func overviewProfitMarginsFrom(fin *marketsurge.MDFinancials) []overviewProfitMargin {
	if fin == nil || len(fin.ProfitMarginValues) == 0 {
		return nil
	}
	result := make([]overviewProfitMargin, 0, len(fin.ProfitMarginValues))
	for i := range fin.ProfitMarginValues {
		pm := &fin.ProfitMarginValues[i]
		result = append(result, overviewProfitMargin{
			Period:         pm.Period,
			Offset:         pm.PeriodOffset,
			EndDate:        formattedStringValue(pm.PeriodEndDate),
			GrossMargin:    overviewValueWrapperFloat(pm.GrossMargin),
			PreTaxMargin:   overviewScaledFloat(pm.PreTaxMargin),
			AfterTaxMargin: overviewValueWrapperFloat(pm.AfterTaxMargin),
			ReturnOnEquity: overviewFormattedFloat(pm.ReturnOnEquity),
		})
	}
	return result
}

func overviewGrowthRatesFrom(fin *marketsurge.MDFinancials) *overviewGrowthRates {
	if fin == nil || fin.ConsensusFinancials == nil {
		return nil
	}
	result := &overviewGrowthRates{}
	if fin.ConsensusFinancials.EPS != nil && len(fin.ConsensusFinancials.EPS.GrowthRate) > 0 {
		result.EPS = overviewGrowthRateSlice(fin.ConsensusFinancials.EPS.GrowthRate)
	}
	if fin.ConsensusFinancials.Sales != nil && len(fin.ConsensusFinancials.Sales.GrowthRate) > 0 {
		result.Sales = overviewGrowthRateSlice(fin.ConsensusFinancials.Sales.GrowthRate)
	}
	if len(result.EPS) == 0 && len(result.Sales) == 0 {
		return nil
	}
	return result
}

func overviewGrowthRateSlice(rates []marketsurge.MDGrowthRate) []overviewGrowthRate {
	result := make([]overviewGrowthRate, 0, len(rates))
	for i := range rates {
		r := &rates[i]
		var val *overviewFloat
		if r.Value != nil {
			val = &overviewFloat{Value: r.Value, Formatted: r.FormattedValue}
		}
		result = append(result, overviewGrowthRate{
			Value:  val,
			Period: r.Period,
		})
	}
	return result
}

func overviewValueWrapperFloat(value *marketsurge.MDValueWrapper) *overviewFloat {
	if value == nil || value.Value == nil {
		return nil
	}
	return &overviewFloat{Value: value.Value}
}

func formattedStringValue(value *marketsurge.MDFormattedString) *string {
	if value == nil {
		return nil
	}
	if value.Value != nil && *value.Value != "" {
		return value.Value
	}
	if value.FormattedValue != nil && *value.FormattedValue != "" {
		return value.FormattedValue
	}
	return nil
}

func clientError(message string, err error) error {
	if marketsurge.IsAuthError(err) {
		return mserrors.NewAuthenticationError("authentication failed", err)
	}
	if marketsurge.IsRateLimited(err) {
		return mserrors.NewHTTPError("rate limited", err, 429, "")
	}
	return mserrors.NewAPIError(message, err)
}
