package client

import (
	"context"
	"strings"
	"time"

	"github.com/major/marketsurge-agent/internal/constants"
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

	return &models.StockData{
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
		Pricing: &models.Pricing{
			MarketCap:                            floatPtr(pricingEOD["marketCapitalization"]),
			MarketCapFormatted:                   formattedValue(pricingEOD["marketCapitalization"]),
			AvgDollarVolume50D:                   floatPtr(pricingEOD["avgDollarVolume50Day"]),
			AvgDollarVolume50DFormatted:          formattedValue(pricingEOD["avgDollarVolume50Day"]),
			UpDownVolumeRatio:                    floatPtr(pricingEOD["upDownVolumeRatio"]),
			UpDownVolumeRatioFormatted:           formattedValue(pricingEOD["upDownVolumeRatio"]),
			ATRPercent21D:                        floatPtr(atr21d),
			ATRPercent21DFormatted:               formattedValue(atr21d),
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
			PricingStartDate:                     stringPtr(pricingEOD["pricingStartDate"]),
			PricingEndDate:                       stringPtr(pricingEOD["pricingEndDate"]),
		},
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
			ReportedEarnings: buildQuarterlyReported(getNestedSlice(epsConsensus, "reportedEarnings")),
			ReportedSales:    buildQuarterlyReported(getNestedSlice(salesConsensus, "reportedSales")),
			EPSEstimates:     buildQuarterlyEstimates(getNestedSlice(financials, "estimates", "epsEstimates")),
			SalesEstimates:   buildQuarterlyEstimates(getNestedSlice(financials, "estimates", "salesEstimates")),
			ProfitMargins:    buildQuarterlyProfitMargins(getNestedSlice(financials, "profitMarginValues")),
		},
		Patterns:   []models.Pattern{},
		TightAreas: []models.TightArea{},
	}, nil
}

func buildQuarterlyReported(items []any) []models.QuarterlyReportedPeriod {
	result := make([]models.QuarterlyReportedPeriod, 0, len(items))
	for _, entry := range items {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, models.QuarterlyReportedPeriod{
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
		})
	}
	return result
}

func buildQuarterlyEstimates(items []any) []models.QuarterlyEstimate {
	result := make([]models.QuarterlyEstimate, 0, len(items))
	for _, entry := range items {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, models.QuarterlyEstimate{
			Value:             floatPtr(item["value"]),
			PctChangeYoY:      floatPtr(item["percentChangeYOY"]),
			PeriodEndDate:     stringPtr(item["periodEndDate"]),
			EffectiveDate:     stringPtr(item["effectiveDate"]),
			RevisionDirection: stringPtr(item["revisionDirection"]),
			EstimateType:      stringPtr(item["type"]),
		})
	}
	return result
}

func buildQuarterlyProfitMargins(items []any) []models.QuarterlyProfitMargin {
	result := make([]models.QuarterlyProfitMargin, 0, len(items))
	for _, entry := range items {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, models.QuarterlyProfitMargin{
			PeriodOffset:   stringify(item["periodOffset"]),
			PeriodEndDate:  stringPtr(item["periodEndDate"]),
			PreTaxMargin:   floatPtr(item["preTaxMargin"]),
			AfterTaxMargin: floatPtr(item["afterTaxMargin"]),
			GrossMargin:    floatPtr(item["grossMargin"]),
			ReturnOnEquity: floatPtr(item["returnOnEquity"]),
		})
	}
	return result
}
