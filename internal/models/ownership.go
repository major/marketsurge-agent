package models

// QuarterlyFundOwnership represents fund ownership count for a single quarter.
type QuarterlyFundOwnership struct {
	Date  *string `json:"date,omitempty"`
	Count *string `json:"count,omitempty"`
}

// OwnershipData represents complete ownership data response from the Ownership query.
type OwnershipData struct {
	Symbol           string                     `json:"symbol"`
	FundsFloatPct    *string                    `json:"funds_float_pct,omitempty"`
	QuarterlyFunds   []QuarterlyFundOwnership   `json:"quarterly_funds"`
}
