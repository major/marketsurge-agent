---
name: marketsurge-agent
description: "MarketSurge stock research CLI for AI agents. Structured JSON output, semantic exit codes. Covers: stock data and ratings, multi-symbol concurrent analysis with compact/flat/summary modes, fundamental earnings and sales, institutional ownership, RS rating history, OHLCV price history with benchmarks, chart markups, catalog listing and running of watchlists/reports/coach screens with pagination and filters. Triggers: 'marketsurge', 'stock', 'stock analysis', 'stock ratings', 'fundamentals', 'earnings', 'ownership', 'institutional', 'RS rating', 'relative strength', 'chart history', 'price history', 'OHLCV', 'watchlist', 'catalog', 'screen', 'marketsurge-agent'."
metadata:
  author: "Major Hayden"
  version: "1.0.0"
---

# marketsurge-agent CLI

Go CLI for MarketSurge stock research API. All output is structured JSON with semantic exit codes. This project is unofficial and is not affiliated with MarketSurge or Investor's Business Daily.

## Feedback

If you hit bugs, confusing usability, missing guidance, or token-inefficient workflows while using `marketsurge-agent`, encourage the user to open a GitHub issue at `github.com/major/marketsurge-agent`. Offer to open the issue for them with GitHub's `gh` CLI if it is installed and the user wants you to file it.

## Companion Files

This is the entry point. Command details are split by domain:

| File | Scope |
|------|-------|
| [stock.md](stock.md) | Stock data, ratings, pricing, multi-symbol analysis (compact/flat/summary modes) |
| [fundamental.md](fundamental.md) | Earnings and sales (historical, estimates, quarterly breakdowns) |
| [ownership.md](ownership.md) | Institutional fund ownership (quarterly counts, float percentage) |
| [rs-history.md](rs-history.md) | Relative strength rating history, RS line new highs |
| [chart.md](chart.md) | OHLCV price history (daily/weekly, benchmarks), chart markups |
| [catalog.md](catalog.md) | List and run watchlists, reports, coach screens with pagination and filters |

## Auth

MarketSurge requires a valid JWT. The CLI resolves credentials in this order (first non-empty wins):

1. `--jwt` flag
2. `MARKETSURGE_TOKEN` env var
3. `--cookie-db` path to a Firefox `cookies.sqlite` file
4. Auto-discovery from local Firefox profiles

For automation, set the env var:

```bash
export MARKETSURGE_TOKEN="your-jwt-here"
```

## Output

### Success envelope

```json
{
  "data": { ... },
  "metadata": { "symbol": "AAPL" },
  "timestamp": "2026-04-21T12:00:00Z"
}
```

Access results via `.data`. Check `.metadata` for context (symbol, mode, symbols list).

### Error response

```json
{
  "error": "symbol not found",
  "code": 2,
  "message": "XYZZY: no matching symbol",
  "timestamp": "2026-04-21T12:00:00Z"
}
```

### Partial responses

`stock analyze` with multiple symbols returns partial results: successful symbols in `.data`, failures in `.errors`. Check both.

### Exit codes

| Code | Meaning |
|------|---------|
| 1 | Validation error (bad args, missing fields) |
| 2 | Symbol not found |
| 3 | Authentication error (missing/expired token, cookie failures) |
| 4 | API or HTTP error (GraphQL errors, rate limiting, server errors) |

## Gotchas

- **JWT expiry**: MarketSurge JWTs expire. If you get exit code 3, the user needs to refresh their token or ensure Firefox has an active MarketSurge session for cookie auto-discovery.
- **Chart date params**: `chart history` date parameters are mutually exclusive: use `--start-date` + `--end-date` OR `--lookback`, never both.
- **Catalog kind**: `catalog run` requires `--kind` and the matching ID flag (`--watchlist-id`, `--report-id`, or `--coach-screen-id`). Screen entries cannot be dispatched.
- **Multi-symbol output**: `stock analyze` and `rs-history get` accept multiple symbols. Output shape differs from single-symbol commands: data is keyed by ticker or returned as an array.
- **Summary mode**: `stock analyze --summary` returns small screening objects optimized for ranking many candidates with minimal tokens. Metadata includes `mode: "summary"`.
- **Compact mode**: `stock analyze --compact` strips duplicate formatted string fields (e.g., `market_cap_formatted`) while keeping raw numeric values.
- **Flat mode**: `stock analyze --flat` flattens nested objects into single-level keys (e.g., `stock.pricing.market_cap` becomes `pricing_market_cap`).
- **Batch tickers**: `stock analyze --tickers AAPL,NVDA,TSLA` accepts comma-separated symbols. Positional symbols and `--tickers` can be combined.
