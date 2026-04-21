# Stock Analysis Skill

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
`bash
marketsurge-agent stock get AAPL
`

**Expected Output Shape:**
`json
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
`

### analyze_stock
Analyze a stock with comprehensive data from MarketSurge.

Fetches stock ratings, fundamentals, and ownership data concurrently.
Partial failures in fundamentals or ownership are returned in the errors list rather than failing the entire request.

Stock data now includes valuation ratios, risk metrics, short interest data, and blue dot event flags.

**Parameters:**
- symbol (required): Stock ticker symbol, e.g. AAPL, NVDA, TSLA

**Example:**
`bash
marketsurge-agent stock analyze AAPL NVDA
`

**Expected Output Shape:**
`json
{
  "symbol": "AAPL",
  "stock": { ... },
  "fundamentals": { ... },
  "ownership": { ... },
  "errors": []
}
`

## Workflow Guidance

1. Use **get_stock** for quick lookups of current ratings and pricing
2. Use **analyze_stock** when you need comprehensive data including fundamentals and ownership
3. Combine with chart history for technical analysis
4. Use RS rating to identify relative strength vs market
