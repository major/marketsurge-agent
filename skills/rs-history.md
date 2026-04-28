# RS Rating History Skill

## Overview
Fetch reported relative strength rating history for one or more stocks from MarketSurge.

## Tools

### get_rs_rating_history
Fetch reported relative strength rating history for one or more stocks from MarketSurge.

Returns a time series of RS rating snapshots showing relative price performance
vs the market over various periods. Includes the rs_line_new_high flag indicating
when the RS line hits a new high ahead of price.

**Parameters:**
- symbols (required): One or more stock ticker symbols separated by spaces, e.g. AAPL NVDA TSLA

**Example:**
```bash
marketsurge-agent rs-history get AAPL NVDA
```

**Expected Output Shape:**
```json
{
  "data": {
    "AAPL": {
      "symbol": "AAPL",
      "ratings": [
        {
          "period": "Current",
          "value": 78
        }
      ],
      "rs_line_new_high": true
    }
  },
  "metadata": {
    "symbols": ["AAPL", "NVDA"]
  }
}
```

## Workflow Guidance

1. Track RS rating trends over time
2. Pass multiple symbols at once when comparing candidates; multi-symbol output is keyed by ticker
3. RS line new highs indicate strong relative strength
4. Compare RS rating with price action for divergences
5. Use for identifying leading stocks in uptrends
