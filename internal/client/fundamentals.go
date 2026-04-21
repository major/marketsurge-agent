package client

import (
	"context"

	"github.com/tidwall/gjson"

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

	consensus := item.Get("financials.consensusFinancials")
	companyName := gStr(item.Get("symbology.company.0.companyName"))

	return &models.FundamentalData{
		Symbol:           symbol,
		CompanyName:      companyName,
		ReportedEarnings: buildReportedPeriods(consensus.Get("eps.reportedEarnings").Array()),
		ReportedSales:    buildReportedPeriods(consensus.Get("sales.reportedSales").Array()),
		EPSEstimates:     buildEstimatePeriods(item.Get("financials.estimates.epsEstimates").Array()),
		SalesEstimates:   buildEstimatePeriods(item.Get("financials.estimates.salesEstimates").Array()),
	}, nil
}

func buildReportedPeriods(items []gjson.Result) []models.ReportedPeriod {
	return buildSlice(items, func(item gjson.Result) models.ReportedPeriod {
		return models.ReportedPeriod{
			Value:              gFloat(item.Get("value")),
			FormattedValue:     gStr(item.Get("value.formattedValue")),
			PctChangeYoY:       gFloat(item.Get("percentChangeYOY")),
			FormattedPctChange: gStr(item.Get("percentChangeYOY.formattedValue")),
			PeriodOffset:       stringify(item.Get("periodOffset")),
			PeriodEndDate:      gStr(item.Get("periodEndDate")),
		}
	})
}

func buildEstimatePeriods(items []gjson.Result) []models.EstimatePeriod {
	return buildSlice(items, func(item gjson.Result) models.EstimatePeriod {
		return models.EstimatePeriod{
			Value:              gFloat(item.Get("value")),
			FormattedValue:     gStr(item.Get("value.formattedValue")),
			PctChangeYoY:       gFloat(item.Get("percentChangeYOY")),
			FormattedPctChange: gStr(item.Get("percentChangeYOY.formattedValue")),
			PeriodOffset:       stringify(item.Get("periodOffset")),
			Period:             gStr(item.Get("period")),
			RevisionDirection:  gStr(item.Get("revisionDirection")),
		}
	})
}
