package client

import (
	"context"
	"fmt"

	"github.com/major/marketsurge-agent/internal/constants"
	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/queries"
)

// GetFundamentals returns reported and estimated fundamentals for a symbol.
func (c *Client) GetFundamentals(ctx context.Context, symbol string) (*models.FundamentalData, error) {
	query, err := queries.Load("fundamentals.graphql")
	if err != nil {
		return nil, err
	}

	raw, err := c.Execute(ctx, Request{
		OperationName: "FundermentalDataBox",
		Variables: map[string]any{
			"symbols":                            []string{symbol},
			"symbolDialectType":                  constants.SymbolDialectType,
			"upToHistoricalPeriodOffset":         "P7Y_AGO",
			"upToQueryPeriodOffset":              "P2Y_FUTURE",
			"reportedSalesUpToHistoricalPeriod2": "P7Y_AGO",
			"salesEstimatesUpToQueryPeriod2":     "P2Y_FUTURE",
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

	financials := getNestedMap(item, "financials")
	consensus := getNestedMap(financials, "consensusFinancials")
	estimates := getNestedMap(financials, "estimates")
	symbology := getNestedMap(item, "symbology")
	companyName := stringPtr(firstMap(getNestedSlice(symbology, "company"))["companyName"])

	fundamentals := &models.FundamentalData{
		Symbol:           symbol,
		CompanyName:      companyName,
		ReportedEarnings: buildReportedPeriods(getNestedSlice(consensus, "eps", "reportedEarnings")),
		ReportedSales:    buildReportedPeriods(getNestedSlice(consensus, "sales", "reportedSales")),
		EPSEstimates:     buildEstimatePeriods(getNestedSlice(estimates, "epsEstimates")),
		SalesEstimates:   buildEstimatePeriods(getNestedSlice(estimates, "salesEstimates")),
	}
	if !hasFundamentalData(fundamentals) {
		return nil, mserrors.NewSymbolNotFoundError(fmt.Sprintf("symbol not found: %q", symbol), nil, symbol)
	}
	return fundamentals, nil
}

func hasFundamentalData(data *models.FundamentalData) bool {
	if data == nil {
		return false
	}
	if data.CompanyName != nil {
		return true
	}
	return hasReportedPeriodData(data.ReportedEarnings) || hasReportedPeriodData(data.ReportedSales) || hasEstimatePeriodData(data.EPSEstimates) || hasEstimatePeriodData(data.SalesEstimates)
}

func hasReportedPeriodData(periods []models.ReportedPeriod) bool {
	for _, period := range periods {
		if period.Value != nil || period.FormattedValue != nil || period.PctChangeYoY != nil || period.FormattedPctChange != nil || period.PeriodEndDate != nil {
			return true
		}
	}
	return false
}

func hasEstimatePeriodData(periods []models.EstimatePeriod) bool {
	for _, period := range periods {
		if period.Value != nil || period.FormattedValue != nil || period.PctChangeYoY != nil || period.FormattedPctChange != nil || period.Period != nil || period.RevisionDirection != nil {
			return true
		}
	}
	return false
}

func buildReportedPeriods(items []any) []models.ReportedPeriod {
	return buildSlice(items, func(item map[string]any) models.ReportedPeriod {
		valueMap, _ := item["value"].(map[string]any)
		pctMap, _ := item["percentChangeYOY"].(map[string]any)
		return models.ReportedPeriod{
			Value:              floatPtr(item["value"]),
			FormattedValue:     stringPtr(valueMap["formattedValue"]),
			PctChangeYoY:       floatPtr(item["percentChangeYOY"]),
			FormattedPctChange: stringPtr(pctMap["formattedValue"]),
			PeriodOffset:       stringify(item["periodOffset"]),
			PeriodEndDate:      stringPtr(item["periodEndDate"]),
		}
	})
}

func buildEstimatePeriods(items []any) []models.EstimatePeriod {
	return buildSlice(items, func(item map[string]any) models.EstimatePeriod {
		valueMap, _ := item["value"].(map[string]any)
		pctMap, _ := item["percentChangeYOY"].(map[string]any)
		return models.EstimatePeriod{
			Value:              floatPtr(item["value"]),
			FormattedValue:     stringPtr(valueMap["formattedValue"]),
			PctChangeYoY:       floatPtr(item["percentChangeYOY"]),
			FormattedPctChange: stringPtr(pctMap["formattedValue"]),
			PeriodOffset:       stringify(item["periodOffset"]),
			Period:             stringPtr(item["period"]),
			RevisionDirection:  stringPtr(item["revisionDirection"]),
		}
	})
}
