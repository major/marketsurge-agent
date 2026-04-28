# Chart Skill

Use chart commands for OHLCV candles, benchmark series, and user-saved chart annotations.

## `chart history <symbol>`

Fetches price history. You must provide exactly one date mode:

| Date mode | Example | Notes |
|---|---|---|
| Relative lookback | `marketsurge-agent chart history AAPL --lookback 3M` | Valid: `1W`, `1M`, `3M`, `6M`, `1Y`, `YTD` |
| Explicit range | `marketsurge-agent chart history AAPL --start-date 2024-01-01 --end-date 2024-04-21` | Must provide both dates |

Other flags:

- `--period daily|weekly`: defaults to `daily`; `weekly` maps to `P1W`.
- `--benchmark 0S&P5`: includes `benchmark_time_series` for relative strength calculations.

Output focus: `time_series.data_points` with OHLCV fields, quote, exchange, market state, and optional benchmark candles.

## `chart markups <symbol>`

Fetches user-saved annotations and drawings.

```bash
marketsurge-agent chart markups AAPL --frequency WEEKLY --sort-dir DESC
```

Flags:

- `--frequency DAILY|WEEKLY`, default `DAILY`.
- `--sort-dir ASC|DESC`, default `ASC`.

Markup `data` is opaque serialized chart data. Do not parse it unless a downstream MarketSurge-specific renderer understands the format.
