package models

// RSRatingSnapshot represents a single RS rating value at a specific period and time offset.
type RSRatingSnapshot struct {
	LetterValue  *string `json:"letter_value,omitempty"`
	Period       *string `json:"period,omitempty"`
	PeriodOffset *string `json:"period_offset,omitempty"`
	Value        *int    `json:"value,omitempty"`
}

// RSRatingHistory represents RS rating history from RSRatingRIPanel query.
type RSRatingHistory struct {
	Symbol        string               `json:"symbol"`
	Ratings       []RSRatingSnapshot   `json:"ratings"`
	RSLineNewHigh *bool                `json:"rs_line_new_high,omitempty"`
}
