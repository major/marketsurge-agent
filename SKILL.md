---
name: marketsurge-agent
description: |
  marketsurge-agent queries the MarketSurge stock research API.
  All output is structured JSON with semantic exit codes.
  
  Auth
  
    MarketSurge authentication uses Firefox browser cookies. The CLI exchanges
    the local browser session for the API JWT automatically. Cookie databases
    resolve in this order:
  
      1. --cookie-db path to a Firefox cookies.sqlite file
      2. Auto-discovery from local Firefox profiles
  
    Sign into MarketSurge in Firefox before running API commands.
  
  Output
  
    Success envelope:
      {"data": {...}, "metadata": {"symbol": "AAPL"}, "timestamp": "..."}
  
    Error response:
      {"error": "symbol not found", "code": 31, "message": "...", "timestamp": "..."}
  
    Partial response (stock analyze with multiple symbols):
      {"data": {...}, "errors": [...], "metadata": {...}, "timestamp": "..."}
  
  Exit Codes
  
       0 - Success
      30 - Validation error (bad args, missing fields)
      31 - Symbol not found
      32 - Authentication error (missing/expired token, cookie failures)
      33 - API or HTTP error...
metadata:
  author: major
  version: dev
  mcp-server: marketsurge-agent --mcp
---

# marketsurge-agent

## Instructions

### Available Commands

#### `marketsurge-agent`

marketsurge-agent queries the MarketSurge stock research API.
All output is structured JSON with semantic exit codes.

Auth

  MarketSurge authentication uses Firefox browser cookies. The CLI exchanges
  the local browser session for the API JWT automatically. Cookie databases
  resolve in this order:

    1. --cookie-db path to a Firefox cookies.sqlite file
    2. Auto-discovery from local Firefox profiles

  Sign into MarketSurge in Firefox before running API commands.

Output

  Success envelope:
    {"data": {...}, "metadata": {"symbol": "AAPL"}, "timestamp": "..."}

  Error response:
    {"error": "symbol not found", "code": 31, "message": "...", "timestamp": "..."}

  Partial response (stock analyze with multiple symbols):
    {"data": {...}, "errors": [...], "metadata": {...}, "timestamp": "..."}

Exit Codes

     0 - Success
    30 - Validation error (bad args, missing fields)
    31 - Symbol not found
    32 - Authentication error (missing/expired token, cookie failures)
    33 - API or HTTP error (GraphQL errors, rate limiting, server errors)

Gotchas

  - Auth expiry: exit code 32 means Firefox needs an active MarketSurge
    session or the explicit --cookie-db path is not usable.
  - Chart date params: --start-date/--end-date and --lookback are
    mutually exclusive.
  - Catalog kind: catalog run requires --kind and the matching ID flag.
  - Multi-symbol: stock analyze and rs-history get return data keyed by
    ticker when given multiple symbols.
  - Summary mode: stock analyze --summary returns compact screening
    objects for ranking many candidates. Metadata includes mode: "summary".
  - Compact mode: stock analyze --compact removes formatted duplicates,
    profile metadata, internal IDs, empty fields, and stale arrays while
    keeping decision-relevant raw values.
  - Flat mode: stock analyze --flat flattens nested objects into
    single-level keys.
  - Batch tickers: stock analyze --tickers AAPL,NVDA,TSLA accepts
    comma-separated symbols. Positional and --tickers can be combined.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--cookie-db` | string | - | no | Path to Firefox cookies.sqlite file; omit to auto-discover Firefox profiles |
| `--verbose` | bool | false | no | Enable verbose logging for auth and API requests |

**Environment Variables:**

| Variable | Flag | Description |
|----------|------|-------------|
| `MARKETSURGE_AGENT_COOKIEDB` | `--cookie-db` | Path to Firefox cookies.sqlite file; omit to auto-discover Firefox profiles |
| `MARKETSURGE_AGENT_COOKIE_DB` | `--cookie-db` | Path to Firefox cookies.sqlite file; omit to auto-discover Firefox profiles |
| `MARKETSURGE_AGENT_VERBOSE` | `--verbose` | Enable verbose logging for auth and API requests |

#### `marketsurge-agent catalog list`

Lists catalog entries. The --kind flag is optional.

Valid --kind values: watchlist, screen, report, coach_screen.

Omit --kind to aggregate all sources. Partial source failures can
still return entries from working sources.

Output: entries[] with name, kind, description, and the relevant ID field.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--kind` | string | - | no | Filter by catalog kind (watchlist, report, coach_screen, screen); omit to list all sources |

#### `marketsurge-agent catalog run`

Runs a catalog entry and returns its contents.

Required by kind:

  Kind           Required flag          Runnable
  ----           -------------          --------
  watchlist      --watchlist-id         Yes
  report         --report-id            Yes
  coach_screen   --coach-screen-id      Yes
  screen         (none)                 No, list only

Examples:

  catalog run --kind report --report-id 124 --fields symbol,price,composite_rating
  catalog run --kind watchlist --watchlist-id 99 --limit 25 --exclude-spacs
  catalog run --kind coach_screen --coach-screen-id screen-1 --limit 10

Useful flags:

  --limit, --offset   Page large lists (default limit: 50)
  --fields            Project columns: symbol, price, composite_rating,
                      eps_rating, rs_rating, acc_dis_rating, smr_rating,
                      industry_name, market_cap, volume_dollar_avg_50d
  --min-composite     Minimum composite rating for report/watchlist rows
  --min-rs            Minimum RS rating for report/watchlist rows
  --exclude-spacs     Exclude SPAC/blank-check entries

Coach screen rows are paginated, but field projection and filters do
not behave like report or watchlist rows.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--coach-screen-id` | string | - | no | Coach screen ID; required when kind=coach_screen. Example coach screen run: --kind coach_screen --coach-screen-id screen-1 |
| `--exclude-spacs` | bool | false | no | Exclude SPAC/blank-check entries from results |
| `--fields` | stringSlice | [] | no | Project specific result fields; accepts repeated --fields flags or comma-separated values. Examples: --fields symbol --fields price, or --fields symbol,price,composite_rating. Common fields: symbol, price, composite_rating, eps_rating, rs_rating, acc_dis_rating, smr_rating, industry_name, market_cap, volume_dollar_avg_50d |
| `--kind` | string | - | no | Required catalog kind to run: watchlist uses --watchlist-id, report uses --report-id, coach_screen uses --coach-screen-id; screens are list-only. Example report: --kind report --report-id 124 |
| `--limit` | int | 50 | no | Maximum number of results to return |
| `--min-composite` | int | 0 | no | Minimum composite rating for report/watchlist rows (0-99); omitted when unset. Example: --min-composite 90 |
| `--min-rs` | int | 0 | no | Minimum RS rating for report/watchlist rows (0-99); omitted when unset. Example: --min-rs 80 |
| `--offset` | int | 0 | no | Number of results to skip for pagination |
| `--report-id` | int | 0 | no | Report ID; required when kind=report. Example report run: --kind report --report-id 124 |
| `--watchlist-id` | int64 | 0 | no | Watchlist ID; required when kind=watchlist. Example watchlist run: --kind watchlist --watchlist-id 99 |

**Example:**

```bash
marketsurge-agent catalog run --kind report --report-id 124 --fields symbol,price,composite_rating
  marketsurge-agent catalog run --kind watchlist --watchlist-id 99 --limit 25 --exclude-spacs
  marketsurge-agent catalog run --kind coach_screen --coach-screen-id screen-1 --limit 10
```

#### `marketsurge-agent chart history`

Fetches price history for a symbol. Exactly one date mode is required:

  Date mode           Example
  ---------           -------
  Relative lookback   chart history AAPL --lookback 3M
                      Valid: 1W, 1M, 3M, 6M, 1Y, YTD
  Explicit range      chart history AAPL --start-date 2024-01-01 --end-date 2024-04-21
                      Both dates required

Other flags:

  --period daily|weekly    Defaults to daily; weekly maps to P1W
  --benchmark 0S&P5       Includes benchmark_time_series for relative
                           strength calculations

Output: time_series.data_points with OHLCV fields, quote, exchange,
market state, and optional benchmark candles.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--benchmark` | string | - | no | Benchmark symbol for relative strength comparison |
| `--end-date` | string | - | no | End date in YYYY-MM-DD format, for example 2024-06-30; must be paired with --start-date; mutually exclusive with --lookback. Example explicit range: --start-date 2024-01-01 --end-date 2024-06-30 |
| `--lookback` | string | - | no | Relative lookback period (1W, 1M, 3M, 6M, 1Y, YTD); mutually exclusive with --start-date/--end-date. Example relative range: --lookback 3M |
| `--period` | string | daily | no | Data period granularity (daily or weekly) |
| `--start-date` | string | - | no | Start date in YYYY-MM-DD format, for example 2024-01-01; must be paired with --end-date; mutually exclusive with --lookback. Example explicit range: --start-date 2024-01-01 --end-date 2024-06-30 |
| `--symbol` | string | - | no | Stock symbol to fetch, for example AAPL; positional <symbol> remains supported for shell use |

**Example:**

```bash
marketsurge-agent chart history AAPL --lookback 3M
  marketsurge-agent chart history AAPL --start-date 2024-01-01 --end-date 2024-06-30
  marketsurge-agent chart history AAPL --lookback 1Y --period weekly --benchmark 0S&P5
```

#### `marketsurge-agent chart markups`

Fetches user-saved annotations and drawings for a symbol.

Flags:

  --frequency DAILY|WEEKLY   Default: DAILY
  --sort-dir ASC|DESC        Default: ASC

Markup data is opaque serialized chart data. Do not parse it unless
a downstream MarketSurge-specific renderer understands the format.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--frequency` | string | DAILY | no | Chart candle frequency for markup lookup (DAILY or WEEKLY) |
| `--sort-dir` | string | ASC | no | Sort direction for markup annotations (ASC or DESC) |
| `--symbol` | string | - | no | Stock symbol to fetch, for example AAPL; positional <symbol> remains supported for shell use |

#### `marketsurge-agent fundamental get`

Get fundamental data for a symbol

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--symbol` | string | - | no | Stock symbol to fetch, for example AAPL; positional <symbol> remains supported for shell use |

#### `marketsurge-agent ownership get`

Get ownership data for a symbol

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--symbol` | string | - | no | Stock symbol to fetch, for example AAPL; positional <symbol> remains supported for shell use |

#### `marketsurge-agent rs-history get`

Get RS rating history for one or more symbols

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--symbols` | stringSlice | [] | no | Stock symbols to fetch, for example AAPL,MSFT; accepts comma-separated or repeated values; positional symbols remain supported for shell use |

#### `marketsurge-agent stock analyze`

Fetches stock, fundamentals, and ownership concurrently for one or more
symbols. Accepts positional symbols, --symbols values, and backward-compatible
--tickers values together. Multi-symbol requests are concurrent and can return partial
results when only some symbols or subresources fail.

Flags:

  --summary   Compact screening keys: symbol, composite, eps, rs, ad, smr,
              blue_dot, ant_signal, ant_dates, ant_explanation,
              base_type, base_stage, pivot,
              pivot_price_date, pricing_start_date, pricing_end_date,
              base_depth_percent,
              industry_group_rs, up_down_volume, atr_percent,
              avg_dollar_volume, funds_float_percent
  --setup     Setup-focused trade triage keys: all --summary keys plus
              base_length_weeks, volume_at_pivot_percent,
              ownership_funds_float_percent, quarterly_funds
  --compact   Removes formatted duplicates, profile metadata, internal IDs,
              empty fields, and stale arrays while keeping raw decision fields
  --flat      Flattens nested analysis fields inside the JSON envelope

Start with "stock analyze --summary" for candidate ranking, then rerun
interesting symbols without --summary for detail.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--compact` | bool | false | no | Remove low-value fields such as formatted duplicates, profile metadata, internal IDs, empty fields, and stale arrays while keeping decision-relevant raw values; compatible with --flat. Example: stock analyze AAPL --compact |
| `--flat` | bool | false | no | Flatten nested analysis fields into single-level JSON keys, for example stock.pricing.market_cap becomes pricing_market_cap; compatible with --compact. Example: stock analyze AAPL --flat |
| `--setup` | bool | false | no | Return --setup trade triage fields: summary fields plus base_length_weeks, volume_at_pivot_percent, ownership_funds_float_percent, and quarterly_funds. Example: stock analyze AAPL --setup |
| `--summary` | bool | false | no | Return compact screening objects for ranking many symbols. Example: stock analyze --summary AAPL MSFT NVDA |
| `--symbols` | stringSlice | [] | no | Stock symbols to analyze, for example AAPL,MSFT; accepts comma-separated or repeated values; positional symbols remain supported for shell use |
| `--tickers` | stringSlice | [] | no | Additional stock symbols to analyze; accepts comma-separated or repeated values |

**Example:**

```bash
marketsurge-agent stock analyze AAPL
  marketsurge-agent stock analyze --tickers AAPL,MSFT,NVDA --compact
  marketsurge-agent stock analyze --summary AAPL MSFT NVDA
  marketsurge-agent stock analyze AAPL --setup
  marketsurge-agent stock analyze AAPL --flat --compact
```

#### `marketsurge-agent stock get`

Get stock data for a symbol

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--symbol` | string | - | no | Stock symbol to fetch, for example AAPL; positional <symbol> remains supported for shell use |

### Environment Variable Prefix

All environment variables use the `MARKETSURGE_AGENT_` prefix.

### Examples

#### marketsurge-agent catalog run

```bash
marketsurge-agent catalog run --kind report --report-id 124 --fields symbol,price,composite_rating
  marketsurge-agent catalog run --kind watchlist --watchlist-id 99 --limit 25 --exclude-spacs
  marketsurge-agent catalog run --kind coach_screen --coach-screen-id screen-1 --limit 10
```

#### marketsurge-agent chart history

```bash
marketsurge-agent chart history AAPL --lookback 3M
  marketsurge-agent chart history AAPL --start-date 2024-01-01 --end-date 2024-06-30
  marketsurge-agent chart history AAPL --lookback 1Y --period weekly --benchmark 0S&P5
```

#### marketsurge-agent stock analyze

```bash
marketsurge-agent stock analyze AAPL
  marketsurge-agent stock analyze --tickers AAPL,MSFT,NVDA --compact
  marketsurge-agent stock analyze --summary AAPL MSFT NVDA
  marketsurge-agent stock analyze AAPL --setup
  marketsurge-agent stock analyze AAPL --flat --compact
```
