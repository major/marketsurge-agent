package client

import (
	"context"

	"github.com/major/marketsurge-agent/internal/constants"
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

	return &models.FundamentalData{
		Symbol:           symbol,
		CompanyName:      companyName,
		ReportedEarnings: buildReportedPeriods(getNestedSlice(consensus, "eps", "reportedEarnings")),
		ReportedSales:    buildReportedPeriods(getNestedSlice(consensus, "sales", "reportedSales")),
		EPSEstimates:     buildEstimatePeriods(getNestedSlice(estimates, "epsEstimates")),
		SalesEstimates:   buildEstimatePeriods(getNestedSlice(estimates, "salesEstimates")),
	}, nil
}

func buildReportedPeriods(items []any) []models.ReportedPeriod {
	result := make([]models.ReportedPeriod, 0, len(items))
	for _, entry := range items {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		valueMap, _ := item["value"].(map[string]any)
		pctMap, _ := item["percentChangeYOY"].(map[string]any)
		result = append(result, models.ReportedPeriod{
			Value:              floatPtr(item["value"]),
			FormattedValue:     stringPtr(valueMap["formattedValue"]),
			PctChangeYoY:       floatPtr(item["percentChangeYOY"]),
			FormattedPctChange: stringPtr(pctMap["formattedValue"]),
			PeriodOffset:       stringify(item["periodOffset"]),
			PeriodEndDate:      stringPtr(item["periodEndDate"]),
		})
	}
	return result
}

func buildEstimatePeriods(items []any) []models.EstimatePeriod {
	result := make([]models.EstimatePeriod, 0, len(items))
	for _, entry := range items {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		valueMap, _ := item["value"].(map[string]any)
		pctMap, _ := item["percentChangeYOY"].(map[string]any)
		result = append(result, models.EstimatePeriod{
			Value:              floatPtr(item["value"]),
			FormattedValue:     stringPtr(valueMap["formattedValue"]),
			PctChangeYoY:       floatPtr(item["percentChangeYOY"]),
			FormattedPctChange: stringPtr(pctMap["formattedValue"]),
			PeriodOffset:       stringify(item["periodOffset"]),
			Period:             stringPtr(item["period"]),
			RevisionDirection:  stringPtr(item["revisionDirection"]),
		})
	}
	return result
}
