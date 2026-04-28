# MarketSurge Agent Skill Index

Use `marketsurge-agent` when an AI agent needs MarketSurge stock research data as JSON. Every command requires auth and writes a JSON envelope to stdout.

## Pick the command

| Need | Command | Notes |
|---|---|---|
| Fast quote, ratings, base, signals | `stock get <symbol>` | One symbol, no fundamentals or ownership fetch |
| Best first call for research or ranking | `stock analyze [symbols...]` | Fetches stock, fundamentals, ownership concurrently |
| Lowest-token batch ranking | `stock analyze --summary ...` | Small screening object per symbol |
| Earnings, sales, estimates only | `fundamental get <symbol>` | Use standalone when you do not need price or ownership |
| Fund ownership trend only | `ownership get <symbol>` | Use standalone for institutional support checks |
| RS trend comparison | `rs-history get [symbols...]` | Multi-symbol output is keyed by ticker |
| OHLCV candles | `chart history <symbol>` | Requires `--lookback` or `--start-date` plus `--end-date` |
| Saved user drawings | `chart markups <symbol>` | Markup `data` is opaque, do not parse it |
| Discover lists and IDs | `catalog list` | Get IDs before `catalog run` |
| Fetch a watchlist, report, or coach screen | `catalog run` | `screen` entries cannot be run directly |

## Auth and envelopes

Auth precedence: `--jwt`, `MARKETSURGE_JWT`, `--cookie-db`, Firefox profile discovery.

Success shape: `{ "data": ..., "metadata": ... }`. Errors are JSON too: `{ "error": { "code": "...", "message": "...", "details": "..." } }`. Partial batch commands may return successes plus an `errors` collection when only some symbols or sources fail.

## Token-saving defaults

- Prefer `stock analyze --summary` for screening many symbols.
- Add `--compact` to remove duplicate formatted fields from full analysis.
- Add `--flat` only when a downstream parser needs one-level keys.
- For catalog runs, always use `--limit` and `--fields` when you only need a few columns.
- Prefer multi-symbol commands (`stock analyze`, `rs-history get`) over shell loops.
