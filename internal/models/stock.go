package models

// Ratings represents stock composite, EPS, RS, SMR, and A/D ratings.
type Ratings struct {
	Composite *int    `json:"composite,omitempty"`
	EPS       *int    `json:"eps,omitempty"`
	RS        *int    `json:"rs,omitempty"`
	SMR       *string `json:"smr,omitempty"`
	AD        *string `json:"ad,omitempty"`
}

// Company represents company profile information.
type Company struct {
	Name                   *string `json:"name,omitempty"`
	Industry               *string `json:"industry,omitempty"`
	Sector                 *string `json:"sector,omitempty"`
	IndustryGroupRank      *int    `json:"industry_group_rank,omitempty"`
	IndustryGroupRS        *int    `json:"industry_group_rs,omitempty"`
	IndustryGroupRSLetter  *string `json:"industry_group_rs_letter,omitempty"`
	Description            *string `json:"description,omitempty"`
	Website                *string `json:"website,omitempty"`
	Address                *string `json:"address,omitempty"`
	Address2               *string `json:"address2,omitempty"`
	Phone                  *string `json:"phone,omitempty"`
	IPODate                *string `json:"ipo_date,omitempty"`
	IPOPrice               *float64 `json:"ipo_price,omitempty"`
	IPOPriceFormatted      *string `json:"ipo_price_formatted,omitempty"`
	City                   *string `json:"city,omitempty"`
	Country                *string `json:"country,omitempty"`
	StateProvince          *string `json:"state_province,omitempty"`
	InstrumentSubType      *string `json:"instrument_sub_type,omitempty"`
}

// PricePercentChanges represents price percent changes vs various reference periods.
type PricePercentChanges struct {
	YTD           *float64 `json:"ytd,omitempty"`
	MTD           *float64 `json:"mtd,omitempty"`
	QTD           *float64 `json:"qtd,omitempty"`
	WTD           *float64 `json:"wtd,omitempty"`
	Vs1D          *float64 `json:"vs_1d,omitempty"`
	Vs1M          *float64 `json:"vs_1m,omitempty"`
	Vs3M          *float64 `json:"vs_3m,omitempty"`
	VsYearHigh    *float64 `json:"vs_year_high,omitempty"`
	VsYearLow     *float64 `json:"vs_year_low,omitempty"`
	Vs5D          *float64 `json:"vs_5d,omitempty"`
	VsSP50026W    *float64 `json:"vs_sp500_26w,omitempty"`
	VsMA10D       *float64 `json:"vs_ma10d,omitempty"`
	VsMA21D       *float64 `json:"vs_ma21d,omitempty"`
	VsMA50D       *float64 `json:"vs_ma50d,omitempty"`
	VsMA150D      *float64 `json:"vs_ma150d,omitempty"`
	VsMA200D      *float64 `json:"vs_ma200d,omitempty"`
}

// HistoricalPriceStatistic represents a single period of historical price statistics.
type HistoricalPriceStatistic struct {
	Period              *string  `json:"period,omitempty"`
	PeriodOffset        *string  `json:"period_offset,omitempty"`
	PeriodEndDate       *string  `json:"period_end_date,omitempty"`
	PriceHighDate       *string  `json:"price_high_date,omitempty"`
	PriceHigh           *float64 `json:"price_high,omitempty"`
	PriceLowDate        *string  `json:"price_low_date,omitempty"`
	PriceLow            *float64 `json:"price_low,omitempty"`
	PriceClose          *float64 `json:"price_close,omitempty"`
	PricePercentChange  *float64 `json:"price_percent_change,omitempty"`
}

// VolumeMovingAverage represents a single volume moving average with period metadata.
type VolumeMovingAverage struct {
	Value        *float64 `json:"value,omitempty"`
	Period       *string  `json:"period,omitempty"`
	PeriodOffset *string  `json:"period_offset,omitempty"`
}

// Pricing represents pricing statistics and market data.
type Pricing struct {
	MarketCap                              *float64                      `json:"market_cap,omitempty"`
	MarketCapFormatted                     *string                       `json:"market_cap_formatted,omitempty"`
	AvgDollarVolume50D                     *float64                      `json:"avg_dollar_volume_50d,omitempty"`
	AvgDollarVolume50DFormatted            *string                       `json:"avg_dollar_volume_50d_formatted,omitempty"`
	UpDownVolumeRatio                      *float64                      `json:"up_down_volume_ratio,omitempty"`
	UpDownVolumeRatioFormatted             *string                       `json:"up_down_volume_ratio_formatted,omitempty"`
	ATRPercent21D                          *float64                      `json:"atr_percent_21d,omitempty"`
	ATRPercent21DFormatted                 *string                       `json:"atr_percent_21d_formatted,omitempty"`
	ShortInterestPercentFloat              *float64                      `json:"short_interest_percent_float,omitempty"`
	ShortInterestPercentFloatFormatted     *string                       `json:"short_interest_percent_float_formatted,omitempty"`
	BlueDotDailyDates                      []string                      `json:"blue_dot_daily_dates,omitempty"`
	BlueDotWeeklyDates                     []string                      `json:"blue_dot_weekly_dates,omitempty"`
	AntDates                               []string                      `json:"ant_dates,omitempty"`
	PricePercentChanges                    *PricePercentChanges          `json:"price_percent_changes,omitempty"`
	VolumePercentChangeVs50D               *float64                      `json:"volume_percent_change_vs_50d,omitempty"`
	HistoricalPriceStatistics              []HistoricalPriceStatistic    `json:"historical_price_statistics,omitempty"`
	VolumeMovingAverages                   []VolumeMovingAverage         `json:"volume_moving_averages,omitempty"`
	VolumePercentChangeVs6M                *float64                      `json:"volume_percent_change_vs_6m,omitempty"`
	VolumePercentChangeVs10W               *float64                      `json:"volume_percent_change_vs_10w,omitempty"`
	DividendYield                          *float64                      `json:"dividend_yield,omitempty"`
	DividendYieldFormatted                 *string                       `json:"dividend_yield_formatted,omitempty"`
	PriceToCashFlowRatio                   *float64                      `json:"price_to_cash_flow_ratio,omitempty"`
	PriceToCashFlowRatioFormatted          *string                       `json:"price_to_cash_flow_ratio_formatted,omitempty"`
	ForwardPriceToEarningsRatio            *float64                      `json:"forward_price_to_earnings_ratio,omitempty"`
	ForwardPriceToEarningsRatioFormatted   *string                       `json:"forward_price_to_earnings_ratio_formatted,omitempty"`
	PriceToSalesRatio                      *float64                      `json:"price_to_sales_ratio,omitempty"`
	PriceToSalesRatioFormatted             *string                       `json:"price_to_sales_ratio_formatted,omitempty"`
	PriceToEarningsRatio                   *float64                      `json:"price_to_earnings_ratio,omitempty"`
	PriceToEarningsRatioFormatted          *string                       `json:"price_to_earnings_ratio_formatted,omitempty"`
	PEVsSP500                              *float64                      `json:"pe_vs_sp500,omitempty"`
	PEVsSP500Formatted                     *string                       `json:"pe_vs_sp500_formatted,omitempty"`
	Alpha                                  *float64                      `json:"alpha,omitempty"`
	AlphaFormatted                         *string                       `json:"alpha_formatted,omitempty"`
	Beta                                   *float64                      `json:"beta,omitempty"`
	BetaFormatted                          *string                       `json:"beta_formatted,omitempty"`
	ShortInterestDaysToCover               *float64                      `json:"short_interest_days_to_cover,omitempty"`
	ShortInterestDaysToCoverFormatted      *string                       `json:"short_interest_days_to_cover_formatted,omitempty"`
	ShortInterestDaysToCoverPctChange      *float64                      `json:"short_interest_days_to_cover_pct_change,omitempty"`
	ShortInterestDaysToCoverPctChangeFormatted *string                   `json:"short_interest_days_to_cover_pct_change_formatted,omitempty"`
	ShortInterestVolume                    *int                          `json:"short_interest_volume,omitempty"`
	ShortInterestVolumeFormatted           *string                       `json:"short_interest_volume_formatted,omitempty"`
	IsDailyBlueDotEvent                    *bool                         `json:"is_daily_blue_dot_event,omitempty"`
	IsWeeklyBlueDotEvent                   *bool                         `json:"is_weekly_blue_dot_event,omitempty"`
	PricingStartDate                       *string                       `json:"pricing_start_date,omitempty"`
	PricingEndDate                         *string                       `json:"pricing_end_date,omitempty"`
}

// Financials represents financial metrics and earnings data.
type Financials struct {
	EPSDueDate              *string  `json:"eps_due_date,omitempty"`
	EPSDueDateStatus        *string  `json:"eps_due_date_status,omitempty"`
	EPSLastReportedDate     *string  `json:"eps_last_reported_date,omitempty"`
	EPSGrowthRate           *float64 `json:"eps_growth_rate,omitempty"`
	SalesGrowthRate3Y       *float64 `json:"sales_growth_rate_3y,omitempty"`
	PreTaxMargin            *float64 `json:"pre_tax_margin,omitempty"`
	AfterTaxMargin          *float64 `json:"after_tax_margin,omitempty"`
	GrossMargin             *float64 `json:"gross_margin,omitempty"`
	ReturnOnEquity          *float64 `json:"return_on_equity,omitempty"`
	EarningsStability       *int     `json:"earnings_stability,omitempty"`
	CashFlowPerShare        *float64 `json:"cash_flow_per_share,omitempty"`
	CashFlowPerShareFormatted *string `json:"cash_flow_per_share_formatted,omitempty"`
}

// Dividend represents an individual dividend event.
type Dividend struct {
	ExDate            *string `json:"ex_date,omitempty"`
	Amount            *string `json:"amount,omitempty"`
	ChangeIndicator   *string `json:"change_indicator,omitempty"`
}

// CorporateActions represents corporate action history (dividends, splits, spinoffs).
type CorporateActions struct {
	NextExDividendDate *string     `json:"next_ex_dividend_date,omitempty"`
	Dividends          []Dividend  `json:"dividends,omitempty"`
	Splits             []string    `json:"splits,omitempty"`
	Spinoffs           []any       `json:"spinoffs,omitempty"`
}

// TightArea represents a tight price consolidation area on a chart.
type TightArea struct {
	PatternID *int    `json:"pattern_id,omitempty"`
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`
	Length    *int    `json:"length,omitempty"`
}

// Pattern represents a base class for all technical chart patterns.
type Pattern struct {
	ID                              *string  `json:"id,omitempty"`
	PatternType                     *string  `json:"pattern_type,omitempty"`
	Periodicity                     *string  `json:"periodicity,omitempty"`
	BaseStage                       *string  `json:"base_stage,omitempty"`
	BaseNumber                      *int     `json:"base_number,omitempty"`
	BaseStatus                      *string  `json:"base_status,omitempty"`
	BaseLength                      *int     `json:"base_length,omitempty"`
	BaseDepth                       *float64 `json:"base_depth,omitempty"`
	BaseDepthFormatted              *string  `json:"base_depth_formatted,omitempty"`
	BaseStartDate                   *string  `json:"base_start_date,omitempty"`
	BaseEndDate                     *string  `json:"base_end_date,omitempty"`
	BaseBottomDate                  *string  `json:"base_bottom_date,omitempty"`
	LeftSideHighDate                *string  `json:"left_side_high_date,omitempty"`
	PivotPrice                      *float64 `json:"pivot_price,omitempty"`
	PivotPriceFormatted             *string  `json:"pivot_price_formatted,omitempty"`
	PivotDate                       *string  `json:"pivot_date,omitempty"`
	PivotPriceDate                  *string  `json:"pivot_price_date,omitempty"`
	AvgVolumeRatePctOnPivot         *float64 `json:"avg_volume_rate_pct_on_pivot,omitempty"`
	AvgVolumeRatePctOnPivotFormatted *string `json:"avg_volume_rate_pct_on_pivot_formatted,omitempty"`
	PricePctChangeOnPivot           *float64 `json:"price_pct_change_on_pivot,omitempty"`
	PricePctChangeOnPivotFormatted  *string  `json:"price_pct_change_on_pivot_formatted,omitempty"`
}

// CupPattern represents a cup or saucer pattern with or without handle.
type CupPattern struct {
	Pattern
	HandleDepth       *float64 `json:"handle_depth,omitempty"`
	HandleDepthFormatted *string `json:"handle_depth_formatted,omitempty"`
	HandleLength      *int     `json:"handle_length,omitempty"`
	CupLength         *int     `json:"cup_length,omitempty"`
	CupEndDate        *string  `json:"cup_end_date,omitempty"`
	HandleLowDate     *string  `json:"handle_low_date,omitempty"`
	HandleStartDate   *string  `json:"handle_start_date,omitempty"`
}

// DoubleBottomPattern represents a double bottom (W-shaped) chart pattern.
type DoubleBottomPattern struct {
	Pattern
	FirstBottomDate  *string `json:"first_bottom_date,omitempty"`
	SecondBottomDate *string `json:"second_bottom_date,omitempty"`
	MidPeakDate      *string `json:"mid_peak_date,omitempty"`
}

// AscendingBasePattern represents an ascending base pattern.
type AscendingBasePattern struct {
	Pattern
	FirstBottomDate           *string  `json:"first_bottom_date,omitempty"`
	SecondAscendingHighDate   *string  `json:"second_ascending_high_date,omitempty"`
	SecondBottomDate          *string  `json:"second_bottom_date,omitempty"`
	ThirdAscendingHighDate    *string  `json:"third_ascending_high_date,omitempty"`
	ThirdBottomDate           *string  `json:"third_bottom_date,omitempty"`
	PullBack1Depth            *float64 `json:"pull_back_1_depth,omitempty"`
	PullBack1DepthFormatted   *string  `json:"pull_back_1_depth_formatted,omitempty"`
	PullBack2Depth            *float64 `json:"pull_back_2_depth,omitempty"`
	PullBack2DepthFormatted   *string  `json:"pull_back_2_depth_formatted,omitempty"`
	PullBack3Depth            *float64 `json:"pull_back_3_depth,omitempty"`
	PullBack3DepthFormatted   *string  `json:"pull_back_3_depth_formatted,omitempty"`
}

// IpoBasePattern represents the first base pattern after an IPO.
type IpoBasePattern struct {
	Pattern
	UpBars                      *int     `json:"up_bars,omitempty"`
	BlueBars                    *int     `json:"blue_bars,omitempty"`
	StallBars                   *int     `json:"stall_bars,omitempty"`
	DownBars                    *int     `json:"down_bars,omitempty"`
	RedBars                     *int     `json:"red_bars,omitempty"`
	SupportBars                 *int     `json:"support_bars,omitempty"`
	UpVolumeTotal               *float64 `json:"up_volume_total,omitempty"`
	UpVolumeTotalFormatted      *string  `json:"up_volume_total_formatted,omitempty"`
	DownVolumeTotal             *float64 `json:"down_volume_total,omitempty"`
	DownVolumeTotalFormatted    *string  `json:"down_volume_total_formatted,omitempty"`
	VolumePctChangeOnPivot      *float64 `json:"volume_pct_change_on_pivot,omitempty"`
	VolumePctChangeOnPivotFormatted *string `json:"volume_pct_change_on_pivot_formatted,omitempty"`
}

// IndustryGroupSnapshot represents a single time snapshot of industry group rank or RS.
type IndustryGroupSnapshot struct {
	PeriodOffset *string `json:"period_offset,omitempty"`
	Value        *int    `json:"value,omitempty"`
	LetterValue  *string `json:"letter_value,omitempty"`
}

// Industry represents industry group information.
type Industry struct {
	Name              *string                    `json:"name,omitempty"`
	Sector            *string                    `json:"sector,omitempty"`
	Code              *string                    `json:"code,omitempty"`
	NumberOfStocks    *int                       `json:"number_of_stocks,omitempty"`
	GroupRankHistory  []IndustryGroupSnapshot    `json:"group_rank_history,omitempty"`
	GroupRSHistory    []IndustryGroupSnapshot    `json:"group_rs_history,omitempty"`
}

// BasicOwnership represents basic fund ownership metrics.
type BasicOwnership struct {
	FundsFloatPct          *float64 `json:"funds_float_pct,omitempty"`
	FundsFloatPctFormatted *string  `json:"funds_float_pct_formatted,omitempty"`
}

// Fundamentals represents fundamental financial data.
type Fundamentals struct {
	RAndDPercentLastQtr          *float64 `json:"r_and_d_percent_last_qtr,omitempty"`
	RAndDPercentLastQtrFormatted *string  `json:"r_and_d_percent_last_qtr_formatted,omitempty"`
	DebtPercentFormatted         *string  `json:"debt_percent_formatted,omitempty"`
	NewCEODate                   *string  `json:"new_ceo_date,omitempty"`
}

// StockData represents complete stock data response.
type StockData struct {
	Symbol            string              `json:"symbol"`
	Ratings           *Ratings            `json:"ratings,omitempty"`
	Company           *Company            `json:"company,omitempty"`
	Pricing           *Pricing            `json:"pricing,omitempty"`
	Financials        *Financials         `json:"financials,omitempty"`
	CorporateActions  *CorporateActions   `json:"corporate_actions,omitempty"`
	Industry          *Industry           `json:"industry,omitempty"`
	Ownership         *BasicOwnership     `json:"ownership,omitempty"`
	Fundamentals      *Fundamentals       `json:"fundamentals,omitempty"`
	QuarterlyFinancials *QuarterlyFinancials `json:"quarterly_financials,omitempty"`
	Patterns          []Pattern           `json:"patterns,omitempty"`
	TightAreas        []TightArea         `json:"tight_areas,omitempty"`
}
