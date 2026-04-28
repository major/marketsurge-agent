# marketsurge-agent

CLI tool that gives AI agents structured access to [MarketSurge](https://marketsurge.investors.com) stock research data. Every command returns JSON, making it easy to integrate with agent frameworks, pipelines, or scripts.

> **Disclaimer**: This project is unofficial and is not affiliated with, approved by, or endorsed by MarketSurge or Investor's Business Daily. Use at your own risk.

## Install

Pre-built binaries are available on the [Releases](https://github.com/major/marketsurge-agent/releases) page for Linux and macOS (amd64/arm64).

From source:

```bash
go install github.com/major/marketsurge-agent/cmd/marketsurge-agent@latest
```

## Authentication

MarketSurge requires a valid JWT. The CLI resolves credentials in this order:

1. `--jwt` flag
2. `MARKETSURGE_JWT` environment variable
3. `--cookie-db` path to a Firefox `cookies.sqlite` file
4. Auto-discovery from local Firefox profiles

The simplest approach for automation:

```bash
export MARKETSURGE_JWT="your-jwt-here"
```

## Usage

```bash
# Get stock data for a single symbol
marketsurge-agent stock get AAPL

# Analyze multiple symbols concurrently
marketsurge-agent stock analyze AAPL MSFT NVDA GOOG

# Analyze a comma-separated batch and remove formatted duplicate fields
marketsurge-agent stock analyze --tickers AAPL,MSFT,NVDA --compact

# Return screening fields for ranking many candidates
marketsurge-agent stock analyze --summary AAPL MSFT NVDA

# Flatten each analysis result for lower-token agent parsing
marketsurge-agent stock analyze AAPL --flat

# Fundamental data
marketsurge-agent fundamental get TSLA

# Institutional ownership
marketsurge-agent ownership get AMZN

# Relative strength history for one or more symbols
marketsurge-agent rs-history get META NVDA

# Chart price history (daily, last 3 months)
marketsurge-agent chart history AAPL --lookback 3M

# Chart markups and annotations
marketsurge-agent chart markups AAPL

# List available watchlists, screens, and reports
marketsurge-agent catalog list

# Run a specific watchlist
marketsurge-agent catalog run --kind watchlist --watchlist-id 12345

# Run a report with filters
marketsurge-agent catalog run --kind report --report-id 67890 --min-composite 90 --min-rs 80
```

All commands return a JSON envelope:

```json
{
  "data": { ... },
  "metadata": { "symbol": "AAPL" }
}
```

Errors follow the same pattern:

```json
{
  "error": {
    "code": "SYMBOL_NOT_FOUND",
    "message": "XYZZY: no matching symbol",
    "details": "symbol: XYZZY"
  }
}
```

## Commands

| Command | Description |
|---|---|
| `stock get <symbol>` | Stock data (ratings, pricing, financials) |
| `stock analyze [symbols...]` | Concurrent single-symbol or multi-symbol analysis with optional compact, flat, summary, and comma-separated batch modes |
| `fundamental get <symbol>` | Fundamental analysis data |
| `ownership get <symbol>` | Institutional ownership |
| `rs-history get [symbols...]` | Relative strength rating history for one or more symbols |
| `chart history <symbol>` | Price history (daily or weekly) |
| `chart markups <symbol>` | Chart annotations and markups |
| `catalog list` | List watchlists, screens, reports |
| `catalog run` | Run a watchlist, coach screen, or report |

## Agent integration

### Token-efficient stock analysis

`stock analyze` supports output modes designed for AI agents and batch comparison workflows:

```bash
marketsurge-agent stock analyze --tickers AAPL,MSFT,NVDA --compact --flat
```

- `--tickers AAPL,MSFT,NVDA` analyzes comma-separated symbols in one command. Positional symbols still work and can be combined with `--tickers`.
- `--compact` removes duplicate formatted string fields such as `market_cap_formatted`, while keeping raw numeric values.
- `--summary` returns one small screening object per symbol with rankings, signal flags, base details, liquidity, volatility, and ownership fields. Response metadata includes `mode: "summary"`.
- `--flat` flattens each analysis result inside the standard JSON envelope, for example `stock.pricing.market_cap` becomes `pricing_market_cap`.

`stock analyze` also includes MarketSurge technical context for chart-driven screening: `stock.base_pattern` summarizes the current base with pattern type, base stage, pivot price, base length, depth, and volume at pivot; `stock.signals` reports blue dot and ant signal flags when the API provides them.

`rs-history get` accepts multiple symbols in one request. Multi-symbol output uses a `data` object keyed by ticker so agents can compare RS trends without shell loops.

The static Markdown files in `skills/marketsurge-agent/` describe each command group's inputs, outputs, and gotchas for AI agent frameworks that support tool or skill discovery. Keep them updated with CLI behavior changes.

## Development

Requires Go 1.26+.

```bash
make build     # Build binary
make test      # Run tests with race detector
make lint      # Run golangci-lint (install: https://golangci-lint.run/welcome/install/)
make clean     # Remove binary and build artifacts
```

### Project layout

```text
cmd/marketsurge-agent/   Entry point
internal/
  auth/                  JWT resolution (4-tier chain)
  client/                GraphQL client + API methods
  commands/              CLI command implementations
  constants/             API endpoints, column names
  cookies/               Firefox cookie extraction
  errors/                Typed error hierarchy
  models/                Data structures
  output/                JSON envelope formatting
queries/                 Embedded GraphQL queries
skills/                  Static agent skill docs
```

### Running tests

```bash
go test -v -race ./...
```

Tests use `httptest.NewServer` for HTTP mocking with no external mock libraries.

### Releasing

Push a version tag to trigger a [GoReleaser](https://goreleaser.com) build:

```bash
git tag v1.2.3
git push origin v1.2.3
```

This produces binaries for linux/darwin on amd64/arm64, published to GitHub Releases.
