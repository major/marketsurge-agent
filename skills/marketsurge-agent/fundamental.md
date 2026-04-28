# Fundamental Data Skill

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
`bash
marketsurge-agent fundamental get AAPL
`

**Expected Output Shape:**
`json
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
`

## Workflow Guidance

1. Use for earnings analysis and growth trends
2. Compare historical vs estimated earnings
3. Analyze quarterly breakdowns for seasonal patterns
4. Review cash flow per share for financial health
