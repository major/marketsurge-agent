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
	Ticker                string                `json:"ticker"`
	Name                  *string               `json:"name,omitempty"`
	Type                  *string               `json:"type,omitempty"`
	Ratings               *overviewRatings      `json:"ratings,omitempty"`
	Price                 *overviewPrice        `json:"price,omitempty"`
	RelativeStrengthTrend *overviewRS           `json:"relativeStrengthTrend,omitempty"`
	ANTs                  *overviewANTs         `json:"ants,omitempty"`
	Patterns              []overviewPattern     `json:"patterns,omitempty"`
	TightAreas            []overviewTight       `json:"tightAreas,omitempty"`
	Industry              *overviewIndustry     `json:"industry,omitempty"`
	Ownership             *overviewOwnership    `json:"ownership,omitempty"`
	Fundamentals          *overviewFundamentals `json:"fundamentals,omitempty"`
	Risk                  *overviewRisk         `json:"risk,omitempty"`
	WeeklyTrend           *overviewWeeklyTrend  `json:"weeklyTrend,omitempty"`
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
	DebtPct        *overviewFloat             `json:"debtPct,omitempty"`
	RNDPct         *overviewFloat             `json:"rndPct,omitempty"`
	CEODate        *string                    `json:"ceoDate,omitempty"`
	ReportedEPS    *overviewFundamentalPeriod `json:"reportedEPS,omitempty"`
	ReportedSales  *overviewFundamentalPeriod `json:"reportedSales,omitempty"`
	EstimatedEPS   *overviewFundamentalPeriod `json:"estimatedEPS,omitempty"`
	EstimatedSales *overviewFundamentalPeriod `json:"estimatedSales,omitempty"`
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
func (c *OverviewCmd) Run(client *marketsurge.Client) error {
	return c.run(client, os.Stdout)
}

func (c *OverviewCmd) run(client *marketsurge.Client, w io.Writer) error {
	symbol := strings.ToUpper(strings.TrimSpace(c.Symbol))
	if symbol == "" {
		return mserrors.NewValidationError("symbol is required", errors.New("empty symbol"))
	}

	ctx := context.Background()
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
		}
		if data.Symbology.Instrument != nil {
			item.Type = data.Symbology.Instrument.SubType
		}
	}
	item.Ratings = overviewRatingsFrom(data.Ratings)
	item.Price = overviewPriceFrom(data.PricingStatistics)
	item.ANTs = overviewANTsFrom(data.PricingStatistics)
	item.Patterns = overviewPatternsFrom(data.PatternInfo)
	item.TightAreas = overviewTightAreasFrom(data.PatternInfo)
	item.Industry = overviewIndustryFrom(data.Industry)
	item.Ownership = overviewMarketDataOwnershipFrom(data.Ownership)
	item.Fundamentals = overviewFundFrom(data.Fundamentals)
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

func overviewFundFrom(fund *marketsurge.MDFundamentals) *overviewFundamentals {
	if fund == nil {
		return nil
	}
	return &overviewFundamentals{
		DebtPct: overviewFormattedFloat(fund.DebtPercent),
		RNDPct:  overviewScaledFloat(fund.ResearchAndDevelopmentPercentLastQtr),
		CEODate: dateValue(fund.NewCEODate),
	}
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

func clientError(message string, err error) error {
	if marketsurge.IsAuthError(err) {
		return mserrors.NewAuthenticationError("authentication failed", err)
	}
	if marketsurge.IsRateLimited(err) {
		return mserrors.NewHTTPError("rate limited", err, 429, "")
	}
	return mserrors.NewAPIError(message, err)
}
