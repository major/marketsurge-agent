package commands

// SkillTemplates contains hardcoded skill content for each command group.
var SkillTemplates = map[string]string{
	"stock": `# Stock Analysis Skill

## Overview
Retrieve and analyze stock data from MarketSurge, including ratings, pricing, financials, patterns, and company information.

## Tools

### get_stock
Fetch stock data including ratings, pricing, financials, patterns, and company info from MarketSurge.

Use this for targeted stock data without fundamentals or ownership.
For comprehensive analysis, use analyze_stock instead.

Stock data now includes valuation ratios, risk metrics, short interest data, and blue dot event flags.

**Parameters:**
- symbol (required): Stock ticker symbol, e.g. AAPL, NVDA, TSLA

**Example:**
` + "```" + `bash
marketsurge-agent stock get AAPL
` + "```" + `

**Expected Output Shape:**
` + "```" + `json
{
  "symbol": "AAPL",
  "ratings": {
    "composite_rating": 85,
    "rs_rating": 78
  },
  "price": 150.25,
  "financials": {
    "pe_ratio": 28.5,
    "eps": 5.28
  }
}
` + "```" + `

### analyze_stock
Analyze a stock with comprehensive data from MarketSurge.

Fetches stock ratings, fundamentals, and ownership data concurrently.
Partial failures in fundamentals or ownership are returned in the errors list rather than failing the entire request.

Stock data now includes valuation ratios, risk metrics, short interest data, and blue dot event flags.

**Parameters:**
- symbols (required): One or more stock ticker symbols separated by spaces, e.g. AAPL NVDA TSLA. Each symbol is fetched concurrently.
- tickers (optional): Comma-separated stock ticker symbols, e.g. AAPL,NVDA,TSLA. Useful for larger agent batch comparisons.
- compact (optional): Remove formatted string fields such as market_cap_formatted. Raw numeric values remain when the API provides them.
- flat (optional): Flatten each analysis result inside the standard JSON envelope for lower-token parsing.

**Example:**
` + "```" + `bash
marketsurge-agent stock analyze AAPL NVDA
marketsurge-agent stock analyze --tickers AAPL,NVDA,TSLA --compact --flat
` + "```" + `

**Expected Output Shape:**
` + "```" + `json
{
  "symbol": "AAPL",
  "stock": { ... },
  "fundamentals": { ... },
  "ownership": { ... },
  "errors": []
}
` + "```" + `

With ` + "`" + `--flat` + "`" + `, nested stock fields are emitted as single-level keys, for example ` + "`" + `stock.pricing.market_cap` + "`" + ` becomes ` + "`" + `pricing_market_cap` + "`" + `.

## Workflow Guidance

1. Use **get_stock** for quick lookups of current ratings and pricing
2. Use **analyze_stock** when you need comprehensive data including fundamentals and ownership
3. Combine with chart history for technical analysis
4. Use RS rating to identify relative strength vs market
`,

	"fundamental": `# Fundamental Data Skill

## Overview
Fetch reported and estimated earnings and sales data from MarketSurge.

## Tools

### get_fundamentals
Fetch reported and estimated earnings and sales data from MarketSurge.

Returns historical EPS/sales with YoY changes and future estimates.
For comprehensive analysis, use analyze_stock instead.

Quarterly breakdowns (earnings, sales, margins) and cash flow per share are now included.

**Parameters:**
- symbol (required): Stock ticker symbol, e.g. AAPL, NVDA, TSLA

**Example:**
` + "`" + `bash
marketsurge-agent fundamental get AAPL
` + "`" + `

**Expected Output Shape:**
` + "`" + `json
{
  "symbol": "AAPL",
  "earnings": {
    "historical": [
      {
        "period": "Q1 2024",
        "eps": 2.18,
        "yoy_change": 5.3
      }
    ],
    "estimates": [
      {
        "period": "Q2 2024",
        "eps_estimate": 2.45
      }
    ]
  },
  "sales": { ... }
}
` + "`" + `

## Workflow Guidance

1. Use for earnings analysis and growth trends
2. Compare historical vs estimated earnings
3. Analyze quarterly breakdowns for seasonal patterns
4. Review cash flow per share for financial health
`,

	"ownership": `# Ownership Data Skill

## Overview
Fetch institutional fund ownership data from MarketSurge.

## Tools

### get_ownership
Fetch institutional fund ownership data from MarketSurge.

Returns quarterly fund ownership counts and funds as percentage of float.
For comprehensive analysis, use analyze_stock instead.

**Parameters:**
- symbol (required): Stock ticker symbol, e.g. AAPL, NVDA, TSLA

**Example:**
` + "`" + `bash
marketsurge-agent ownership get AAPL
` + "`" + `

**Expected Output Shape:**
` + "`" + `json
{
  "symbol": "AAPL",
  "ownership": {
    "quarterly_data": [
      {
        "quarter": "Q4 2023",
        "fund_count": 3245,
        "percentage_of_float": 28.5
      }
    ]
  }
}
` + "`" + `

## Workflow Guidance

1. Monitor institutional ownership trends
2. High ownership percentage indicates strong institutional support
3. Increasing fund count suggests growing institutional interest
4. Use with stock analysis for comprehensive view
`,

	"rs-history": `# RS Rating History Skill

## Overview
Fetch reported relative strength rating history for a stock from MarketSurge.

## Tools

### get_rs_rating_history
Fetch reported relative strength rating history for a stock from MarketSurge.

Returns a time series of RS rating snapshots showing relative price performance
vs the market over various periods. Includes the rs_line_new_high flag indicating
when the RS line hits a new high ahead of price.

**Parameters:**
- symbol (required): Stock ticker symbol, e.g. AAPL, NVDA, TSLA

**Example:**
` + "`" + `bash
marketsurge-agent rs-history get AAPL
` + "`" + `

**Expected Output Shape:**
` + "`" + `json
{
  "symbol": "AAPL",
  "rs_history": [
    {
      "date": "2024-04-21",
      "rs_rating": 78,
      "rs_line_new_high": true
    }
  ]
}
` + "`" + `

## Workflow Guidance

1. Track RS rating trends over time
2. RS line new highs indicate strong relative strength
3. Compare RS rating with price action for divergences
4. Use for identifying leading stocks in uptrends
`,

	"chart": `# Chart Data Skill

## Overview
Fetch OHLCV price history and user-saved chart markups from MarketSurge.

## Tools

### get_price_history
Fetch OHLCV price history for a stock from MarketSurge.

Provide either (start_date + end_date) or lookback, not both.

**Parameters:**
- symbol (required): Stock ticker symbol, e.g. AAPL, NVDA, TSLA
- start_date (optional): Start date in ISO format (YYYY-MM-DD). Use with end_date.
- end_date (optional): End date in ISO format (YYYY-MM-DD). Use with start_date.
- lookback (optional): Relative lookback period: 1W, 1M, 3M, 6M, 1Y, or YTD. Cannot be used with start_date/end_date.
- period (optional): Chart period: daily or weekly. Defaults to daily.
- benchmark (optional): Benchmark symbol for relative strength computation (e.g. '0S&P5' for S&P 500). When provided, the response includes a benchmark_time_series for computing RS line ratios.

**Examples:**
` + "`" + `bash
marketsurge-agent chart history AAPL --lookback 1M
marketsurge-agent chart history AAPL --lookback 3M --period weekly
marketsurge-agent chart history AAPL --start-date 2024-01-01 --end-date 2024-04-21
marketsurge-agent chart history AAPL --lookback 1Y --benchmark 0S&P5
` + "`" + `

**Expected Output Shape:**
` + "`" + `json
{
  "symbol": "AAPL",
  "time_series": {
    "period": "P1D",
    "data_points": [
      {
        "start_date_time": "2024-03-21",
        "end_date_time": "2024-03-21",
        "open": 148.5,
        "high": 150.2,
        "low": 148.0,
        "close": 149.8,
        "volume": 52000000
      }
    ]
  },
  "quote": { "last": 149.8, "change": 1.3, "change_percent": 0.87 },
  "current_market_state": "REGULAR_MARKET",
  "exchange": "NASDAQ"
}
` + "`" + `

### get_chart_markups
Fetch user-saved chart markups (annotations/drawings) for a stock from MarketSurge.

Returns chart markups saved by the user on MarketSurge charts. The data field
in each markup contains opaque serialized markup data and should not be parsed.
Supported frequency values: DAILY, WEEKLY. Supported sort_dir values: ASC, DESC.

**Parameters:**
- symbol (required): Stock ticker symbol, e.g. AAPL, NVDA, TSLA
- frequency (optional): Chart frequency: DAILY or WEEKLY. Defaults to DAILY.
- sort_dir (optional): Sort direction: ASC (oldest first) or DESC (newest first). Defaults to ASC.

**Examples:**
` + "`" + `bash
marketsurge-agent chart markups AAPL
marketsurge-agent chart markups AAPL --frequency WEEKLY --sort-dir DESC
` + "`" + `

**Expected Output Shape:**
` + "`" + `json
{
  "cursor_id": "abc123",
  "markups": [
    {
      "id": "markup_123",
      "name": "Trendline 1",
      "frequency": "DAILY",
      "site": "marketsurge",
      "created_at": "2024-04-20T10:30:00Z",
      "updated_at": "2024-04-21T08:00:00Z",
      "data": "opaque_serialized_data"
    }
  ]
}
` + "`" + `

## Workflow Guidance

1. Use **get_price_history** for technical analysis
2. Combine with benchmarks to compute relative strength lines
3. Use **get_chart_markups** to retrieve saved annotations
4. Analyze price patterns with OHLCV data
5. Support multiple timeframes: 1W, 1M, 3M, 6M, 1Y, YTD
`,

	"catalog": `# Catalog Skill

## Overview
List and run stock lists from MarketSurge, including watchlists, screens, reports, and coach screens.

## Tools

### get_catalog
List all available stock lists from all sources.

Aggregates user screens, predefined reports, coach screens, and
watchlists into a single catalog. Tolerates partial failures: if one
source errors, entries from other sources are still returned with the
error message collected.

Use run_catalog_entry to fetch the stocks in a specific entry.

**Parameters:**
- kind (optional): Filter entries by kind: 'watchlist', 'screen', 'report', or 'coach_screen'. Omit to return all entries.

**Examples:**
` + "`" + `bash
marketsurge-agent catalog list
marketsurge-agent catalog list --kind watchlist
marketsurge-agent catalog list --kind report
` + "`" + `

**Expected Output Shape:**
` + "`" + `json
{
  "entries": [
    {
      "name": "My Watchlist",
      "kind": "watchlist",
      "description": "",
      "watchlist_id": 123
    },
    {
      "name": "Growth Stocks",
      "kind": "report",
      "description": "Top growth stocks by earnings",
      "report_id": 456
    }
  ]
}
` + "`" + `

### run_catalog_entry
Run a catalog entry and return its results.

Dispatches to the appropriate run method based on the entry's kind.
Pass the identifying fields from a get_catalog entry directly.

Screen entries cannot be dispatched (raises an error).

Reports and coach screens can return hundreds of stocks. Use limit
and offset for manageable pages, fields to select only the
columns you need, and the filter parameters to narrow results
server-side before they reach you.

**Parameters:**
- kind (required): Entry kind: 'watchlist', 'report', or 'coach_screen'. Screen entries cannot be dispatched.
- report_id (optional): Report ID (required when kind='report').
- coach_screen_id (optional): Coach screen ID (required when kind='coach_screen').
- watchlist_id (optional): Watchlist ID (required when kind='watchlist').
- limit (optional): Maximum number of results to return. Defaults to 50.
- offset (optional): Starting index for pagination (0-based). Use with limit.
- fields (optional): Comma-separated list of fields to include per entry (e.g., 'symbol,composite_rating,rs_rating,price,industry_name'). Omit for all fields.
- min_composite (optional): Minimum Composite Rating filter (0-99).
- min_rs (optional): Minimum RS Rating filter (0-99).
- exclude_spacs (optional): Exclude SPAC/blank-check entries.

**Examples:**
` + "`" + `bash
marketsurge-agent catalog run --kind watchlist --watchlist-id 123
marketsurge-agent catalog run --kind report --report-id 456 --limit 50 --offset 0
marketsurge-agent catalog run --kind coach_screen --coach-screen-id abc123 --fields symbol,composite_rating,rs_rating --min-composite 70
` + "`" + `

**Expected Output Shape:**
` + "`" + `json
{
  "entries": [
    {
      "symbol": "AAPL",
      "composite_rating": 85,
      "rs_rating": 78,
      "price": 150.25,
      "industry_name": "Technology"
    }
  ],
  "total": 245,
  "limit": 50,
  "offset": 0
}
` + "`" + `

## Workflow Guidance

1. Use **get_catalog** to discover available lists
2. Filter by kind to find specific list types
3. Use **run_catalog_entry** to fetch stocks from a list
4. Use pagination (limit/offset) for large result sets
5. Use fields parameter to select only needed columns
6. Apply filters (min_composite, min_rs, exclude_spacs) server-side
7. Combine with stock analysis for deeper insights
`,
}
