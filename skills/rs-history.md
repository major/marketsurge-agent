# RS Rating History Skill

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
`bash
marketsurge-agent rs-history get AAPL
`

**Expected Output Shape:**
`json
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
`

## Workflow Guidance

1. Track RS rating trends over time
2. RS line new highs indicate strong relative strength
3. Compare RS rating with price action for divergences
4. Use for identifying leading stocks in uptrends
