package models

// CatalogKind represents the type of catalog entry.
type CatalogKind string

// Catalog entry kind constants.
const (
	CatalogKindWatchlist   CatalogKind = "watchlist"
	CatalogKindScreen      CatalogKind = "screen"
	CatalogKindReport      CatalogKind = "report"
	CatalogKindCoachScreen CatalogKind = "coach_screen"
)

// ScreenSource represents a data source linked to a saved screen.
type ScreenSource struct {
	ID   *string `json:"id,omitempty"`
	Type *string `json:"type,omitempty"`
	Pub  *string `json:"pub,omitempty"`
}

// Screen represents a saved screen definition.
type Screen struct {
	ID             *string       `json:"id,omitempty"`
	Name           *string       `json:"name,omitempty"`
	Type           *string       `json:"type,omitempty"`
	Source         *ScreenSource `json:"source,omitempty"`
	Description    *string       `json:"description,omitempty"`
	FilterCriteria *string       `json:"filter_criteria,omitempty"`
	CreatedAt      *string       `json:"created_at,omitempty"`
	UpdatedAt      *string       `json:"updated_at,omitempty"`
}

// ScreenResult represents the result of running a named screen.
type ScreenResult struct {
	ScreenName      *string              `json:"screen_name,omitempty"`
	ElapsedTime     *string              `json:"elapsed_time,omitempty"`
	NumInstruments  *int                 `json:"num_instruments,omitempty"`
	Rows            []map[string]*string `json:"rows,omitempty"`
}

// WatchlistEntry represents a single row from an AdhocScreen watchlist query.
type WatchlistEntry struct {
	Symbol                *string  `json:"symbol,omitempty"`
	CompanyName           *string  `json:"company_name,omitempty"`
	ListRank              *int     `json:"list_rank,omitempty"`
	Price                 *float64 `json:"price,omitempty"`
	PriceNetChange        *float64 `json:"price_net_change,omitempty"`
	PricePctChange        *float64 `json:"price_pct_change,omitempty"`
	PricePctOff52WHighs   *float64 `json:"price_pct_off_52w_high,omitempty"`
	Volume                *int     `json:"volume,omitempty"`
	VolumeChange          *int     `json:"volume_change,omitempty"`
	VolumePctChange       *float64 `json:"volume_pct_change,omitempty"`
	CompositeRating       *int     `json:"composite_rating,omitempty"`
	EPSRating             *int     `json:"eps_rating,omitempty"`
	RSRating              *int     `json:"rs_rating,omitempty"`
	AccDisRating          *string  `json:"acc_dis_rating,omitempty"`
	SMRRating             *string  `json:"smr_rating,omitempty"`
	IndustryGroupRank     *int     `json:"industry_group_rank,omitempty"`
	IndustryName          *string  `json:"industry_name,omitempty"`
	MarketCap             *float64 `json:"market_cap,omitempty"`
	VolumeDollarAvg50D    *float64 `json:"volume_dollar_avg_50d,omitempty"`
	IPODate               *string  `json:"ipo_date,omitempty"`
	DowJonesKey           *string  `json:"dow_jones_key,omitempty"`
	ChartingSymbol        *string  `json:"charting_symbol,omitempty"`
	InstrumentType        *string  `json:"instrument_type,omitempty"`
	InstrumentSubType     *string  `json:"instrument_sub_type,omitempty"`
}

// AdhocScreenResult represents the result of running an adhoc screen.
type AdhocScreenResult struct {
	Entries     []WatchlistEntry `json:"entries"`
	ErrorValues []string         `json:"error_values,omitempty"`
}

// WatchlistSummary represents a summary of a user watchlist.
type WatchlistSummary struct {
	ID           *int    `json:"id,omitempty"`
	Name         *string `json:"name,omitempty"`
	LastModified *string `json:"last_modified,omitempty"`
	Description  *string `json:"description,omitempty"`
}

// WatchlistSymbol represents a single symbol entry in a watchlist.
type WatchlistSymbol struct {
	Key        *string `json:"key,omitempty"`
	DowJonesKey *string `json:"dow_jones_key,omitempty"`
}

// WatchlistDetail represents a full watchlist with its symbol items.
type WatchlistDetail struct {
	ID           *string            `json:"id,omitempty"`
	Name         *string            `json:"name,omitempty"`
	LastModified *string            `json:"last_modified,omitempty"`
	Description  *string            `json:"description,omitempty"`
	Items        []WatchlistSymbol   `json:"items"`
}

// CatalogEntry represents a unified entry from the stock list catalog.
type CatalogEntry struct {
	Name          string      `json:"name"`
	Kind          CatalogKind `json:"kind"`
	Description   *string     `json:"description,omitempty"`
	ReportID      *int        `json:"report_id,omitempty"`
	CoachScreenID *string     `json:"coach_screen_id,omitempty"`
	WatchlistID   *int64      `json:"watchlist_id,omitempty"`
}

// Catalog represents the result of catalog discovery across all stock list sources.
type Catalog struct {
	Entries []CatalogEntry `json:"entries"`
	Errors  []string       `json:"errors"`
}

// CatalogResult represents the result of running a catalog entry.
type CatalogResult struct {
	Kind             CatalogKind         `json:"kind"`
	Total            *int                `json:"total,omitempty"`
	ScreenResult     *ScreenResult       `json:"screen_result,omitempty"`
	AdhocResult      *AdhocScreenResult  `json:"adhoc_result,omitempty"`
	WatchlistEntries []WatchlistEntry    `json:"watchlist_entries,omitempty"`
}
