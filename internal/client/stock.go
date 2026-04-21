package client

import (
	"context"
	"strings"
	"time"

	"github.com/tidwall/gjson"

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

	ratings := item.Get("ratings")
	pricingEOD := item.Get("pricingStatistics.endOfDayStatistics")
	pricingIntraday := item.Get("pricingStatistics.intradayStatistics")
	financials := item.Get("financials")
	epsConsensus := financials.Get("consensusFinancials.eps")
	salesConsensus := financials.Get("consensusFinancials.sales")
	industry := item.Get("industry")
	ownership := item.Get("ownership")
	fundmtls := item.Get("fundamentals")
	corporateActions := item.Get("corporateActions")
	company := item.Get("symbology.company.0")
	instrument := item.Get("symbology.instrument.0")
	profitMargin := financials.Get("profitMarginValues.0")
	atr := pricingEOD.Get("averageTrueRangePercent.0")

	return &models.StockData{
		Symbol: symbol,
		Ratings: &models.Ratings{
			Composite: gInt(ratings.Get("compRating.0.value")),
			EPS:       gInt(ratings.Get("epsRating.0.value")),
			RS:        gInt(ratings.Get("rsRating.0.value")),
			SMR:       gStr(ratings.Get("smrRating.0.letterValue")),
			AD:        gStr(ratings.Get("adRating.0.letterValue")),
		},
		Company: &models.Company{
			Name:                  gStr(company.Get("companyName")),
			Industry:              gStr(industry.Get("name")),
			Sector:                gStr(industry.Get("sector")),
			IndustryGroupRank:     gInt(industry.Get("groupRanks.0.value")),
			IndustryGroupRS:       gInt(industry.Get("groupRS.0.value")),
			IndustryGroupRSLetter: gStr(industry.Get("groupRS.0.letterValue")),
			Description:           gStr(company.Get("businessDescription")),
			Website:               gStr(company.Get("url")),
			Address:               gStr(company.Get("address")),
			Address2:              gStr(company.Get("address2")),
			Phone:                 gStr(company.Get("phone")),
			IPODate:               gStr(instrument.Get("ipoDate")),
			IPOPrice:              gFloat(instrument.Get("ipoPrice")),
			IPOPriceFormatted:     gStr(instrument.Get("ipoPrice.formattedValue")),
			City:                  gStr(company.Get("city")),
			Country:               gStr(company.Get("country")),
			StateProvince:         gStr(company.Get("stateProvince")),
			InstrumentSubType:     gStr(instrument.Get("subType")),
		},
		Pricing: &models.Pricing{
			MarketCap:                            gFloat(pricingEOD.Get("marketCapitalization")),
			MarketCapFormatted:                   gStr(pricingEOD.Get("marketCapitalization.formattedValue")),
			AvgDollarVolume50D:                   gFloat(pricingEOD.Get("avgDollarVolume50Day")),
			AvgDollarVolume50DFormatted:          gStr(pricingEOD.Get("avgDollarVolume50Day.formattedValue")),
			UpDownVolumeRatio:                    gFloat(pricingEOD.Get("upDownVolumeRatio")),
			UpDownVolumeRatioFormatted:           gStr(pricingEOD.Get("upDownVolumeRatio.formattedValue")),
			ATRPercent21D:                        gFloat(atr),
			ATRPercent21DFormatted:               gStr(atr.Get("formattedValue")),
			DividendYield:                        gFloat(pricingIntraday.Get("yield")),
			DividendYieldFormatted:               gStr(pricingIntraday.Get("yield.formattedValue")),
			PriceToCashFlowRatio:                 gFloat(pricingIntraday.Get("priceToCashFlowRatio")),
			PriceToCashFlowRatioFormatted:        gStr(pricingIntraday.Get("priceToCashFlowRatio.formattedValue")),
			ForwardPriceToEarningsRatio:          gFloat(pricingIntraday.Get("forwardPriceToEarningsRatio")),
			ForwardPriceToEarningsRatioFormatted: gStr(pricingIntraday.Get("forwardPriceToEarningsRatio.formattedValue")),
			PriceToSalesRatio:                    gFloat(pricingIntraday.Get("priceToSalesRatio")),
			PriceToSalesRatioFormatted:           gStr(pricingIntraday.Get("priceToSalesRatio.formattedValue")),
			PriceToEarningsRatio:                 gFloat(pricingIntraday.Get("priceToEarningsRatio")),
			PriceToEarningsRatioFormatted:        gStr(pricingIntraday.Get("priceToEarningsRatio.formattedValue")),
			PEVsSP500:                            gFloat(pricingIntraday.Get("priceToEarningsVsSP500")),
			PEVsSP500Formatted:                   gStr(pricingIntraday.Get("priceToEarningsVsSP500.formattedValue")),
			Alpha:                                gFloat(pricingEOD.Get("alpha")),
			AlphaFormatted:                       gStr(pricingEOD.Get("alpha.formattedValue")),
			Beta:                                 gFloat(pricingEOD.Get("beta")),
			BetaFormatted:                        gStr(pricingEOD.Get("beta.formattedValue")),
			PricingStartDate:                     gStr(pricingEOD.Get("pricingStartDate")),
			PricingEndDate:                       gStr(pricingEOD.Get("pricingEndDate")),
		},
		Financials: &models.Financials{
			EPSDueDate:                gStr(financials.Get("epsDueDate")),
			EPSDueDateStatus:          gStr(financials.Get("epsDueDateStatus")),
			EPSLastReportedDate:       gStr(financials.Get("epsLastReportedDate")),
			EPSGrowthRate:             gFloat(epsConsensus.Get("growthRate.0")),
			SalesGrowthRate3Y:         gFloat(salesConsensus.Get("growthRate.0")),
			PreTaxMargin:              gFloat(profitMargin.Get("preTaxMargin")),
			AfterTaxMargin:            gFloat(profitMargin.Get("afterTaxMargin")),
			GrossMargin:               gFloat(profitMargin.Get("grossMargin")),
			ReturnOnEquity:            gFloat(profitMargin.Get("returnOnEquity")),
			EarningsStability:         gInt(epsConsensus.Get("earningsStability")),
			CashFlowPerShare:          gFloat(pricingIntraday.Get("cashFlowPerShareLastYear")),
			CashFlowPerShareFormatted: gStr(pricingIntraday.Get("cashFlowPerShareLastYear.formattedValue")),
		},
		CorporateActions: &models.CorporateActions{
			NextExDividendDate: gStr(corporateActions.Get("dividendNextReportedExDate")),
		},
		Industry: &models.Industry{
			Name:           gStr(industry.Get("name")),
			Sector:         gStr(industry.Get("sector")),
			Code:           gStr(industry.Get("indCode")),
			NumberOfStocks: gInt(industry.Get("numberOfStocksInGroup")),
		},
		Ownership: &models.BasicOwnership{
			FundsFloatPct:          gFloat(ownership.Get("fundsFloatPercentHeld")),
			FundsFloatPctFormatted: gStr(ownership.Get("fundsFloatPercentHeld.formattedValue")),
		},
		Fundamentals: &models.Fundamentals{
			RAndDPercentLastQtr:          gFloat(fundmtls.Get("researchAndDevelopmentPercentLastQtr")),
			RAndDPercentLastQtrFormatted: gStr(fundmtls.Get("researchAndDevelopmentPercentLastQtr.formattedValue")),
			DebtPercentFormatted:         gStr(fundmtls.Get("debtPercent.formattedValue")),
			NewCEODate:                   gStr(fundmtls.Get("newCEODate")),
		},
		QuarterlyFinancials: &models.QuarterlyFinancials{
			ReportedEarnings: buildQuarterlyReported(epsConsensus.Get("reportedEarnings").Array()),
			ReportedSales:    buildQuarterlyReported(salesConsensus.Get("reportedSales").Array()),
			EPSEstimates:     buildQuarterlyEstimates(financials.Get("estimates.epsEstimates").Array()),
			SalesEstimates:   buildQuarterlyEstimates(financials.Get("estimates.salesEstimates").Array()),
			ProfitMargins:    buildQuarterlyProfitMargins(financials.Get("profitMarginValues").Array()),
		},
		Patterns:   []models.Pattern{},
		TightAreas: []models.TightArea{},
	}, nil
}

func buildQuarterlyReported(items []gjson.Result) []models.QuarterlyReportedPeriod {
	return buildSlice(items, func(item gjson.Result) models.QuarterlyReportedPeriod {
		return models.QuarterlyReportedPeriod{
			Value:           gFloat(item.Get("value")),
			PctChangeYoY:    gFloat(item.Get("percentChangeYOY")),
			PeriodOffset:    stringify(item.Get("periodOffset")),
			PeriodEndDate:   gStr(item.Get("periodEndDate")),
			EffectiveDate:   gStr(item.Get("effectiveDate")),
			PercentSurprise: gFloat(item.Get("percentSurprise")),
			SurpriseAmount:  gFloat(item.Get("surpriseAmount")),
			QuarterNumber:   gInt(item.Get("quarterNumber")),
			FiscalYear:      gInt(item.Get("fiscalYear")),
			Period:          gStr(item.Get("period")),
		}
	})
}

func buildQuarterlyEstimates(items []gjson.Result) []models.QuarterlyEstimate {
	return buildSlice(items, func(item gjson.Result) models.QuarterlyEstimate {
		return models.QuarterlyEstimate{
			Value:             gFloat(item.Get("value")),
			PctChangeYoY:      gFloat(item.Get("percentChangeYOY")),
			PeriodEndDate:     gStr(item.Get("periodEndDate")),
			EffectiveDate:     gStr(item.Get("effectiveDate")),
			RevisionDirection: gStr(item.Get("revisionDirection")),
			EstimateType:      gStr(item.Get("type")),
		}
	})
}

func buildQuarterlyProfitMargins(items []gjson.Result) []models.QuarterlyProfitMargin {
	return buildSlice(items, func(item gjson.Result) models.QuarterlyProfitMargin {
		return models.QuarterlyProfitMargin{
			PeriodOffset:   stringify(item.Get("periodOffset")),
			PeriodEndDate:  gStr(item.Get("periodEndDate")),
			PreTaxMargin:   gFloat(item.Get("preTaxMargin")),
			AfterTaxMargin: gFloat(item.Get("afterTaxMargin")),
			GrossMargin:    gFloat(item.Get("grossMargin")),
			ReturnOnEquity: gFloat(item.Get("returnOnEquity")),
		}
	})
}
