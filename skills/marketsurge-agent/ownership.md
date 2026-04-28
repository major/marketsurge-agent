# Ownership Data Skill

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
`bash
marketsurge-agent ownership get AAPL
`

**Expected Output Shape:**
`json
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
`

## Workflow Guidance

1. Monitor institutional ownership trends
2. High ownership percentage indicates strong institutional support
3. Increasing fund count suggests growing institutional interest
4. Use with stock analysis for comprehensive view
