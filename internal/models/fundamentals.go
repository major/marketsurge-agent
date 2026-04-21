package models

// QuarterlyReportedPeriod represents a single quarter of reported earnings or sales.
type QuarterlyReportedPeriod struct {
	Value              *float64 `json:"value,omitempty"`
	PctChangeYoY       *float64 `json:"pct_change_yoy,omitempty"`
	PeriodOffset       string   `json:"period_offset"`
	PeriodEndDate      *string  `json:"period_end_date,omitempty"`
	EffectiveDate      *string  `json:"effective_date,omitempty"`
	PercentSurprise    *float64 `json:"percent_surprise,omitempty"`
	SurpriseAmount     *float64 `json:"surprise_amount,omitempty"`
	QuarterNumber      *int     `json:"quarter_number,omitempty"`
	FiscalYear         *int     `json:"fiscal_year,omitempty"`
	Period             *string  `json:"period,omitempty"`
}

// QuarterlyEstimate represents a single quarter of estimated EPS or sales.
type QuarterlyEstimate struct {
	Value             *float64 `json:"value,omitempty"`
	PctChangeYoY      *float64 `json:"pct_change_yoy,omitempty"`
	PeriodEndDate     *string  `json:"period_end_date,omitempty"`
	EffectiveDate     *string  `json:"effective_date,omitempty"`
	RevisionDirection *string  `json:"revision_direction,omitempty"`
	EstimateType      *string  `json:"estimate_type,omitempty"`
}

// QuarterlyProfitMargin represents profit margin data for a single quarter.
type QuarterlyProfitMargin struct {
	PeriodOffset      string   `json:"period_offset"`
	PeriodEndDate     *string  `json:"period_end_date,omitempty"`
	PreTaxMargin      *float64 `json:"pre_tax_margin,omitempty"`
	AfterTaxMargin    *float64 `json:"after_tax_margin,omitempty"`
	GrossMargin       *float64 `json:"gross_margin,omitempty"`
	ReturnOnEquity    *float64 `json:"return_on_equity,omitempty"`
}

// QuarterlyFinancials represents quarterly earnings, sales, estimates, and margins.
type QuarterlyFinancials struct {
	ReportedEarnings []QuarterlyReportedPeriod `json:"reported_earnings"`
	ReportedSales    []QuarterlyReportedPeriod `json:"reported_sales"`
	EPSEstimates     []QuarterlyEstimate       `json:"eps_estimates"`
	SalesEstimates   []QuarterlyEstimate       `json:"sales_estimates"`
	ProfitMargins    []QuarterlyProfitMargin   `json:"profit_margins"`
}

// ReportedPeriod represents historical reported earnings or sales for a single annual period.
type ReportedPeriod struct {
	Value              *float64 `json:"value,omitempty"`
	FormattedValue     *string  `json:"formatted_value,omitempty"`
	PctChangeYoY       *float64 `json:"pct_change_yoy,omitempty"`
	FormattedPctChange *string  `json:"formatted_pct_change,omitempty"`
	PeriodOffset       string   `json:"period_offset"`
	PeriodEndDate      *string  `json:"period_end_date,omitempty"`
}

// EstimatePeriod represents future EPS or sales estimate for a single annual period.
type EstimatePeriod struct {
	Value             *float64 `json:"value,omitempty"`
	FormattedValue    *string  `json:"formatted_value,omitempty"`
	PctChangeYoY      *float64 `json:"pct_change_yoy,omitempty"`
	FormattedPctChange *string `json:"formatted_pct_change,omitempty"`
	PeriodOffset      string   `json:"period_offset"`
	Period            *string  `json:"period,omitempty"`
	RevisionDirection *string  `json:"revision_direction,omitempty"`
}

// FundamentalData represents fundamental financial data from the FundamentalDataBox query.
type FundamentalData struct {
	Symbol            string             `json:"symbol"`
	CompanyName       *string            `json:"company_name,omitempty"`
	ReportedEarnings  []ReportedPeriod   `json:"reported_earnings"`
	ReportedSales     []ReportedPeriod   `json:"reported_sales"`
	EPSEstimates      []EstimatePeriod   `json:"eps_estimates"`
	SalesEstimates    []EstimatePeriod   `json:"sales_estimates"`
}
