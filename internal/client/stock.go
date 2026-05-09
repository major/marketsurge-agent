package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/major/marketsurge-agent/internal/constants"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/queries"
)

// GetStock returns stock details from the OtherMarketData query.
func (c *Client) GetStock(ctx context.Context, symbol string) (*models.StockData, error) {
	query, err := queries.Load("other_market_data.graphql")
	if err != nil {
		return nil, err
	}

	patternEnd := time.Now().UTC().Format("2006-01-02")
	patternStart := time.Now().UTC().AddDate(-4, 0, 0).Format("2006-01-02")
	query = strings.ReplaceAll(query, "{pattern_start_date}", patternStart)
	query = strings.ReplaceAll(query, "{pattern_end_date}", patternEnd)

	raw, err := c.Execute(ctx, Request{
		OperationName: "OtherMarketData",
		Variables: map[string]any{
			"symbols":                             []string{symbol},
			"symbolDialectType":                   constants.SymbolDialectType,
			"upToHistoricalPeriodForProfitMargin": "P12Q_AGO",
			"upToHistoricalPeriodOffset":          "P24Q_AGO",
			"upToQueryPeriodOffset":               "P4Q_FUTURE",
		},
		Query: query,
	})
	if err != nil {
		return nil, err
	}

	item, err := firstMarketData(raw, symbol)
	if err != nil {
		return nil, err
	}

	ratings := getNestedMap(item, "ratings")
	pricingStatistics := getNestedMap(item, "pricingStatistics")
	pricingEOD := getNestedMap(pricingStatistics, "endOfDayStatistics")
	pricingIntraday := getNestedMap(pricingStatistics, "intradayStatistics")
	financials := getNestedMap(item, "financials")
	consensus := getNestedMap(financials, "consensusFinancials")
	epsConsensus := getNestedMap(consensus, "eps")
	salesConsensus := getNestedMap(consensus, "sales")
	industry := getNestedMap(item, "industry")
	ownership := getNestedMap(item, "ownership")
	fundamentals := getNestedMap(item, "fundamentals")
	corporateActions := getNestedMap(item, "corporateActions")
	patternInfo := getNestedMap(item, "patternInfo")
	symbology := getNestedMap(item, "symbology")
	company := firstMap(getNestedSlice(symbology, "company"))
	instrument := firstMap(getNestedSlice(symbology, "instrument"))
	compRating := firstMap(getNestedSlice(ratings, "compRating"))
	epsRating := firstMap(getNestedSlice(ratings, "epsRating"))
	rsRating := firstMap(getNestedSlice(ratings, "rsRating"))
	smrRating := firstMap(getNestedSlice(ratings, "smrRating"))
	adRating := firstMap(getNestedSlice(ratings, "adRating"))
	groupRank := firstMap(getNestedSlice(industry, "groupRanks"))
	groupRS := firstMap(getNestedSlice(industry, "groupRS"))
	profitMargin := firstMap(getNestedSlice(financials, "profitMarginValues"))
	epsGrowth := firstMap(getNestedSlice(epsConsensus, "growthRate"))
	salesGrowth := firstMap(getNestedSlice(salesConsensus, "growthRate"))
	atr21d := firstMap(getNestedSlice(pricingEOD, "averageTrueRangePercent"))
	patterns := buildSlice(getNestedSlice(patternInfo, "patterns"), func(item map[string]any) models.Pattern {
		return models.Pattern{
			ID:                               stringPtr(item["id"]),
			PatternType:                      stringPtr(item["patternType"]),
			Periodicity:                      stringPtr(item["periodicity"]),
			BaseStage:                        stringPtr(item["baseStage"]),
			BaseNumber:                       intPtr(item["baseNumber"]),
			BaseStatus:                       stringPtr(item["baseStatus"]),
			BaseLength:                       intPtr(item["baseLength"]),
			BaseDepth:                        floatPtr(item["baseDepth"]),
			BaseDepthFormatted:               formattedValue(item["baseDepth"]),
			BaseStartDate:                    stringPtr(item["baseStartDate"]),
			BaseEndDate:                      stringPtr(item["baseEndDate"]),
			BaseBottomDate:                   stringPtr(item["baseBottomDate"]),
			LeftSideHighDate:                 stringPtr(item["leftSideHighDate"]),
			PivotPrice:                       floatPtr(item["pivotPrice"]),
			PivotPriceFormatted:              formattedValue(item["pivotPrice"]),
			PivotDate:                        stringPtr(item["pivotDate"]),
			PivotPriceDate:                   stringPtr(item["pivotPriceDate"]),
			AvgVolumeRatePctOnPivot:          floatPtr(item["avgVolumeRatePctOnPivot"]),
			AvgVolumeRatePctOnPivotFormatted: formattedValue(item["avgVolumeRatePctOnPivot"]),
			PricePctChangeOnPivot:            floatPtr(item["pricePctChangeOnPivot"]),
			PricePctChangeOnPivotFormatted:   formattedValue(item["pricePctChangeOnPivot"]),
		}
	})

	blueDotDailyItems := getNestedSlice(pricingEOD, "blueDotDailyEvents")
	blueDotDailyDates := make([]string, 0, len(blueDotDailyItems))
	for _, item := range blueDotDailyItems {
		if v := stringPtr(item); v != nil {
			blueDotDailyDates = append(blueDotDailyDates, *v)
		}
	}

	blueDotWeeklyItems := getNestedSlice(pricingEOD, "blueDotWeeklyEvents")
	blueDotWeeklyDates := make([]string, 0, len(blueDotWeeklyItems))
	for _, item := range blueDotWeeklyItems {
		if v := stringPtr(item); v != nil {
			blueDotWeeklyDates = append(blueDotWeeklyDates, *v)
		}
	}

	antItems := getNestedSlice(pricingEOD, "antEvents")
	antDates := make([]string, 0, len(antItems))
	for _, item := range antItems {
		if v := stringPtr(item); v != nil {
			antDates = append(antDates, *v)
		}
	}

	pricing := &models.Pricing{
		MarketCap:                            floatPtr(pricingEOD["marketCapitalization"]),
		MarketCapFormatted:                   formattedValue(pricingEOD["marketCapitalization"]),
		AvgDollarVolume50D:                   floatPtr(pricingEOD["avgDollarVolume50Day"]),
		AvgDollarVolume50DFormatted:          formattedValue(pricingEOD["avgDollarVolume50Day"]),
		UpDownVolumeRatio:                    floatPtr(pricingEOD["upDownVolumeRatio"]),
		UpDownVolumeRatioFormatted:           formattedValue(pricingEOD["upDownVolumeRatio"]),
		ATRPercent21D:                        floatPtr(atr21d),
		ATRPercent21DFormatted:               formattedValue(atr21d),
		BlueDotDailyDates:                    blueDotDailyDates,
		BlueDotWeeklyDates:                   blueDotWeeklyDates,
		AntDates:                             antDates,
		DividendYield:                        floatPtr(pricingIntraday["yield"]),
		DividendYieldFormatted:               formattedValue(pricingIntraday["yield"]),
		PriceToCashFlowRatio:                 floatPtr(pricingIntraday["priceToCashFlowRatio"]),
		PriceToCashFlowRatioFormatted:        formattedValue(pricingIntraday["priceToCashFlowRatio"]),
		ForwardPriceToEarningsRatio:          floatPtr(pricingIntraday["forwardPriceToEarningsRatio"]),
		ForwardPriceToEarningsRatioFormatted: formattedValue(pricingIntraday["forwardPriceToEarningsRatio"]),
		PriceToSalesRatio:                    floatPtr(pricingIntraday["priceToSalesRatio"]),
		PriceToSalesRatioFormatted:           formattedValue(pricingIntraday["priceToSalesRatio"]),
		PriceToEarningsRatio:                 floatPtr(pricingIntraday["priceToEarningsRatio"]),
		PriceToEarningsRatioFormatted:        formattedValue(pricingIntraday["priceToEarningsRatio"]),
		PEVsSP500:                            floatPtr(pricingIntraday["priceToEarningsVsSP500"]),
		PEVsSP500Formatted:                   formattedValue(pricingIntraday["priceToEarningsVsSP500"]),
		Alpha:                                floatPtr(pricingEOD["alpha"]),
		AlphaFormatted:                       formattedValue(pricingEOD["alpha"]),
		Beta:                                 floatPtr(pricingEOD["beta"]),
		BetaFormatted:                        formattedValue(pricingEOD["beta"]),
		IsDailyBlueDotEvent:                  boolPtr(pricingIntraday["isDailyBlueDotEvent"]),
		IsWeeklyBlueDotEvent:                 boolPtr(pricingIntraday["isWeeklyBlueDotEvent"]),
		PricingStartDate:                     stringPtr(pricingEOD["pricingStartDate"]),
		PricingEndDate:                       stringPtr(pricingEOD["pricingEndDate"]),
	}

	stock := &models.StockData{
		Symbol: symbol,
		Ratings: &models.Ratings{
			Composite: intPtr(compRating["value"]),
			EPS:       intPtr(epsRating["value"]),
			RS:        intPtr(rsRating["value"]),
			SMR:       stringPtr(smrRating["letterValue"]),
			AD:        stringPtr(adRating["letterValue"]),
		},
		Company: &models.Company{
			Name:                  stringPtr(company["companyName"]),
			Industry:              stringPtr(industry["name"]),
			Sector:                stringPtr(industry["sector"]),
			IndustryGroupRank:     intPtr(groupRank["value"]),
			IndustryGroupRS:       intPtr(groupRS["value"]),
			IndustryGroupRSLetter: stringPtr(groupRS["letterValue"]),
			Description:           stringPtr(company["businessDescription"]),
			Website:               stringPtr(company["url"]),
			Address:               stringPtr(company["address"]),
			Address2:              stringPtr(company["address2"]),
			Phone:                 stringPtr(company["phone"]),
			IPODate:               stringPtr(instrument["ipoDate"]),
			IPOPrice:              floatPtr(instrument["ipoPrice"]),
			IPOPriceFormatted:     formattedValue(instrument["ipoPrice"]),
			City:                  stringPtr(company["city"]),
			Country:               stringPtr(company["country"]),
			StateProvince:         stringPtr(company["stateProvince"]),
			InstrumentSubType:     stringPtr(instrument["subType"]),
		},
		Pricing:     pricing,
		BasePattern: buildBasePattern(patterns),
		Signals:     buildSignals(pricing),
		Financials: &models.Financials{
			EPSDueDate:                stringPtr(financials["epsDueDate"]),
			EPSDueDateStatus:          stringPtr(financials["epsDueDateStatus"]),
			EPSLastReportedDate:       stringPtr(financials["epsLastReportedDate"]),
			EPSGrowthRate:             floatPtr(epsGrowth),
			SalesGrowthRate3Y:         floatPtr(salesGrowth),
			PreTaxMargin:              floatPtr(profitMargin["preTaxMargin"]),
			AfterTaxMargin:            floatPtr(profitMargin["afterTaxMargin"]),
			GrossMargin:               floatPtr(profitMargin["grossMargin"]),
			ReturnOnEquity:            floatPtr(profitMargin["returnOnEquity"]),
			EarningsStability:         intPtr(epsConsensus["earningsStability"]),
			CashFlowPerShare:          floatPtr(pricingIntraday["cashFlowPerShareLastYear"]),
			CashFlowPerShareFormatted: formattedValue(pricingIntraday["cashFlowPerShareLastYear"]),
		},
		CorporateActions: &models.CorporateActions{
			NextExDividendDate: stringPtr(corporateActions["dividendNextReportedExDate"]),
		},
		Industry: &models.Industry{
			Name:           stringPtr(industry["name"]),
			Sector:         stringPtr(industry["sector"]),
			Code:           stringPtr(industry["indCode"]),
			NumberOfStocks: intPtr(industry["numberOfStocksInGroup"]),
		},
		Ownership: &models.BasicOwnership{
			FundsFloatPct:          floatPtr(ownership["fundsFloatPercentHeld"]),
			FundsFloatPctFormatted: formattedValue(ownership["fundsFloatPercentHeld"]),
		},
		Fundamentals: &models.Fundamentals{
			RAndDPercentLastQtr:          floatPtr(fundamentals["researchAndDevelopmentPercentLastQtr"]),
			RAndDPercentLastQtrFormatted: formattedValue(fundamentals["researchAndDevelopmentPercentLastQtr"]),
			DebtPercentFormatted:         formattedValue(fundamentals["debtPercent"]),
			NewCEODate:                   stringPtr(fundamentals["newCEODate"]),
		},
		QuarterlyFinancials: &models.QuarterlyFinancials{
			ReportedEarnings: buildSlice(getNestedSlice(epsConsensus, "reportedEarnings"), func(item map[string]any) models.QuarterlyReportedPeriod {
				return models.QuarterlyReportedPeriod{
					Value:           floatPtr(item["value"]),
					PctChangeYoY:    floatPtr(item["percentChangeYOY"]),
					PeriodOffset:    stringify(item["periodOffset"]),
					PeriodEndDate:   stringPtr(item["periodEndDate"]),
					EffectiveDate:   stringPtr(item["effectiveDate"]),
					PercentSurprise: floatPtr(item["percentSurprise"]),
					SurpriseAmount:  floatPtr(item["surpriseAmount"]),
					QuarterNumber:   intPtr(item["quarterNumber"]),
					FiscalYear:      intPtr(item["fiscalYear"]),
					Period:          stringPtr(item["period"]),
				}
			}),
			ReportedSales: buildSlice(getNestedSlice(salesConsensus, "reportedSales"), func(item map[string]any) models.QuarterlyReportedPeriod {
				return models.QuarterlyReportedPeriod{
					Value:           floatPtr(item["value"]),
					PctChangeYoY:    floatPtr(item["percentChangeYOY"]),
					PeriodOffset:    stringify(item["periodOffset"]),
					PeriodEndDate:   stringPtr(item["periodEndDate"]),
					EffectiveDate:   stringPtr(item["effectiveDate"]),
					PercentSurprise: floatPtr(item["percentSurprise"]),
					SurpriseAmount:  floatPtr(item["surpriseAmount"]),
					QuarterNumber:   intPtr(item["quarterNumber"]),
					FiscalYear:      intPtr(item["fiscalYear"]),
					Period:          stringPtr(item["period"]),
				}
			}),
			EPSEstimates: buildSlice(getNestedSlice(financials, "estimates", "epsEstimates"), func(item map[string]any) models.QuarterlyEstimate {
				return models.QuarterlyEstimate{
					Value:             floatPtr(item["value"]),
					PctChangeYoY:      floatPtr(item["percentChangeYOY"]),
					PeriodEndDate:     stringPtr(item["periodEndDate"]),
					EffectiveDate:     stringPtr(item["effectiveDate"]),
					RevisionDirection: stringPtr(item["revisionDirection"]),
					EstimateType:      stringPtr(item["type"]),
				}
			}),
			SalesEstimates: buildSlice(getNestedSlice(financials, "estimates", "salesEstimates"), func(item map[string]any) models.QuarterlyEstimate {
				return models.QuarterlyEstimate{
					Value:             floatPtr(item["value"]),
					PctChangeYoY:      floatPtr(item["percentChangeYOY"]),
					PeriodEndDate:     stringPtr(item["periodEndDate"]),
					EffectiveDate:     stringPtr(item["effectiveDate"]),
					RevisionDirection: stringPtr(item["revisionDirection"]),
					EstimateType:      stringPtr(item["type"]),
				}
			}),
			ProfitMargins: buildSlice(getNestedSlice(financials, "profitMarginValues"), func(item map[string]any) models.QuarterlyProfitMargin {
				return models.QuarterlyProfitMargin{
					PeriodOffset:   stringify(item["periodOffset"]),
					PeriodEndDate:  stringPtr(item["periodEndDate"]),
					PreTaxMargin:   floatPtr(item["preTaxMargin"]),
					AfterTaxMargin: floatPtr(item["afterTaxMargin"]),
					GrossMargin:    floatPtr(item["grossMargin"]),
					ReturnOnEquity: floatPtr(item["returnOnEquity"]),
				}
			}),
		},
		Patterns: patterns,
		TightAreas: buildSlice(getNestedSlice(patternInfo, "tightAreas"), func(item map[string]any) models.TightArea {
			return models.TightArea{
				PatternID: intPtr(item["patternID"]),
				StartDate: stringPtr(item["startDate"]),
				EndDate:   stringPtr(item["endDate"]),
				Length:    intPtr(item["length"]),
			}
		}),
	}
	if !hasStockData(stock) {
		return nil, mserrors.NewSymbolNotFoundError(fmt.Sprintf("symbol not found: %q", symbol), nil, symbol)
	}
	return stock, nil
}

func hasStockData(stock *models.StockData) bool {
	if stock == nil {
		return false
	}
	if stock.Ratings != nil && (stock.Ratings.Composite != nil || stock.Ratings.EPS != nil || stock.Ratings.RS != nil || stock.Ratings.SMR != nil || stock.Ratings.AD != nil) {
		return true
	}
	if stock.Company != nil && (stock.Company.Name != nil || stock.Company.Industry != nil || stock.Company.Sector != nil || stock.Company.IndustryGroupRank != nil || stock.Company.IndustryGroupRS != nil) {
		return true
	}
	if stock.Pricing != nil && (stock.Pricing.MarketCap != nil || stock.Pricing.AvgDollarVolume50D != nil || stock.Pricing.UpDownVolumeRatio != nil || stock.Pricing.ATRPercent21D != nil || stock.Pricing.PricingStartDate != nil || stock.Pricing.PricingEndDate != nil) {
		return true
	}
	if hasBasePatternData(stock.BasePattern) {
		return true
	}
	if stock.Financials != nil && (stock.Financials.EPSDueDate != nil || stock.Financials.EPSLastReportedDate != nil || stock.Financials.EPSGrowthRate != nil || stock.Financials.SalesGrowthRate3Y != nil || stock.Financials.CashFlowPerShare != nil) {
		return true
	}
	if stock.Ownership != nil && stock.Ownership.FundsFloatPct != nil {
		return true
	}
	if stock.Fundamentals != nil && (stock.Fundamentals.RAndDPercentLastQtr != nil || stock.Fundamentals.DebtPercentFormatted != nil || stock.Fundamentals.NewCEODate != nil) {
		return true
	}
	return hasQuarterlyFinancialData(stock.QuarterlyFinancials)
}

func hasBasePatternData(pattern *models.BasePattern) bool {
	if pattern == nil {
		return false
	}
	return pattern.PatternType != nil || pattern.BaseStage != nil || pattern.PivotPrice != nil || pattern.PivotPriceDate != nil || pattern.BaseLengthWeeks != nil || pattern.BaseDepthPercent != nil || pattern.VolumeAtPivotPct != nil
}

func hasQuarterlyFinancialData(financials *models.QuarterlyFinancials) bool {
	if financials == nil {
		return false
	}
	return hasQuarterlyReportedPeriodData(financials.ReportedEarnings) || hasQuarterlyReportedPeriodData(financials.ReportedSales) || hasQuarterlyEstimateData(financials.EPSEstimates) || hasQuarterlyEstimateData(financials.SalesEstimates) || hasQuarterlyProfitMarginData(financials.ProfitMargins)
}

func hasQuarterlyReportedPeriodData(periods []models.QuarterlyReportedPeriod) bool {
	for _, period := range periods {
		if period.Value != nil || period.PctChangeYoY != nil || period.PeriodEndDate != nil || period.EffectiveDate != nil || period.PercentSurprise != nil || period.SurpriseAmount != nil || period.QuarterNumber != nil || period.FiscalYear != nil || period.Period != nil {
			return true
		}
	}
	return false
}

func hasQuarterlyEstimateData(estimates []models.QuarterlyEstimate) bool {
	for _, estimate := range estimates {
		if estimate.Value != nil || estimate.PctChangeYoY != nil || estimate.PeriodEndDate != nil || estimate.EffectiveDate != nil || estimate.RevisionDirection != nil || estimate.EstimateType != nil {
			return true
		}
	}
	return false
}

func hasQuarterlyProfitMarginData(margins []models.QuarterlyProfitMargin) bool {
	for _, margin := range margins {
		if margin.PeriodEndDate != nil || margin.PreTaxMargin != nil || margin.AfterTaxMargin != nil || margin.GrossMargin != nil || margin.ReturnOnEquity != nil {
			return true
		}
	}
	return false
}

func buildSignals(pricing *models.Pricing) *models.Signals {
	blueDot := hasBlueDot(pricing)
	antSignal := len(pricing.AntDates) > 0

	return &models.Signals{
		BlueDot:     &blueDot,
		BlueDotDate: firstString(pricing.BlueDotDailyDates, pricing.BlueDotWeeklyDates),
		AntSignal:   &antSignal,
	}
}

func hasBlueDot(pricing *models.Pricing) bool {
	if pricing.IsDailyBlueDotEvent != nil && *pricing.IsDailyBlueDotEvent {
		return true
	}
	if pricing.IsWeeklyBlueDotEvent != nil && *pricing.IsWeeklyBlueDotEvent {
		return true
	}
	return len(pricing.BlueDotDailyDates) > 0 || len(pricing.BlueDotWeeklyDates) > 0
}

func firstString(slices ...[]string) *string {
	for _, values := range slices {
		if len(values) == 0 {
			continue
		}
		return &values[0]
	}
	return nil
}

func buildBasePattern(patterns []models.Pattern) *models.BasePattern {
	pattern := latestPattern(patterns)
	if pattern == nil {
		return nil
	}

	return &models.BasePattern{
		PatternType:      pattern.PatternType,
		BaseStage:        pattern.BaseStage,
		PivotPrice:       pattern.PivotPrice,
		PivotPriceDate:   pattern.PivotPriceDate,
		BaseLengthWeeks:  pattern.BaseLength,
		BaseDepthPercent: pattern.BaseDepth,
		VolumeAtPivotPct: pattern.AvgVolumeRatePctOnPivot,
	}
}

func latestPattern(patterns []models.Pattern) *models.Pattern {
	var latest *models.Pattern
	for i := range patterns {
		pattern := &patterns[i]
		if latest == nil || dateValue(pattern.BaseEndDate) > dateValue(latest.BaseEndDate) {
			latest = pattern
		}
	}
	return latest
}

func dateValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
