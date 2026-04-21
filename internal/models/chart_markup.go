package models

// ChartMarkup represents a user-saved chart markup/annotation.
type ChartMarkup struct {
	ID        string `json:"id"`
	Name      *string `json:"name,omitempty"`
	Data      string `json:"data"`
	Frequency *string `json:"frequency,omitempty"`
	Site      *string `json:"site,omitempty"`
	CreatedAt *string `json:"created_at,omitempty"`
	UpdatedAt *string `json:"updated_at,omitempty"`
}

// ChartMarkupList represents a paginated collection of chart markups.
type ChartMarkupList struct {
	CursorID string        `json:"cursor_id"`
	Markups  []ChartMarkup `json:"markups"`
}
