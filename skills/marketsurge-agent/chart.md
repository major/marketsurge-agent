# Chart Data Skill

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
`bash
marketsurge-agent chart history AAPL --lookback 1M
marketsurge-agent chart history AAPL --lookback 3M --period weekly
marketsurge-agent chart history AAPL --start-date 2024-01-01 --end-date 2024-04-21
marketsurge-agent chart history AAPL --lookback 1Y --benchmark 0S&P5
`

**Expected Output Shape:**
`json
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
`

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
`bash
marketsurge-agent chart markups AAPL
marketsurge-agent chart markups AAPL --frequency WEEKLY --sort-dir DESC
`

**Expected Output Shape:**
`json
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
`

## Workflow Guidance

1. Use **get_price_history** for technical analysis
2. Combine with benchmarks to compute relative strength lines
3. Use **get_chart_markups** to retrieve saved annotations
4. Analyze price patterns with OHLCV data
5. Support multiple timeframes: 1W, 1M, 3M, 6M, 1Y, YTD
