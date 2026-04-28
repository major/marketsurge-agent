# Catalog Skill

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
`bash
marketsurge-agent catalog list
marketsurge-agent catalog list --kind watchlist
marketsurge-agent catalog list --kind report
`

**Expected Output Shape:**
`json
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
`

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
`bash
marketsurge-agent catalog run --kind watchlist --watchlist-id 123
marketsurge-agent catalog run --kind report --report-id 456 --limit 50 --offset 0
marketsurge-agent catalog run --kind coach_screen --coach-screen-id abc123 --fields symbol,composite_rating,rs_rating --min-composite 70
`

**Expected Output Shape:**
`json
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
`

## Workflow Guidance

1. Use **get_catalog** to discover available lists
2. Filter by kind to find specific list types
3. Use **run_catalog_entry** to fetch stocks from a list
4. Use pagination (limit/offset) for large result sets
5. Use fields parameter to select only needed columns
6. Apply filters (min_composite, min_rs, exclude_spacs) server-side
7. Combine with stock analysis for deeper insights
