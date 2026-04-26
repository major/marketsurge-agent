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
2. `MARKETSURGE_TOKEN` environment variable
3. `--cookie-db` path to a Firefox `cookies.sqlite` file
4. Auto-discovery from local Firefox profiles

The simplest approach for automation:

```bash
export MARKETSURGE_TOKEN="your-jwt-here"
```

## Usage

```bash
# Get stock data for a single symbol
marketsurge-agent stock get AAPL

# Analyze multiple symbols concurrently
marketsurge-agent stock analyze AAPL MSFT NVDA GOOG

# Analyze a comma-separated batch and remove formatted duplicate fields
marketsurge-agent stock analyze --tickers AAPL,MSFT,NVDA --compact

# Flatten each analysis result for lower-token agent parsing
marketsurge-agent stock analyze AAPL --flat

# Fundamental data
marketsurge-agent fundamental get TSLA

# Institutional ownership
marketsurge-agent ownership get AMZN

# Relative strength history
marketsurge-agent rs-history get META

# Chart price history (daily, last 90 days)
marketsurge-agent chart history AAPL --period 90

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
  "metadata": { "symbol": "AAPL" },
  "timestamp": "2026-04-21T12:00:00Z"
}
```

Errors follow the same pattern:

```json
{
  "error": "symbol not found",
  "code": 2,
  "message": "XYZZY: no matching symbol",
  "timestamp": "2026-04-21T12:00:00Z"
}
```

## Commands

| Command | Description |
|---|---|
| `stock get <symbol>` | Stock data (ratings, pricing, financials) |
| `stock analyze [symbols...]` | Concurrent single-symbol or multi-symbol analysis with optional compact, flat, and comma-separated batch modes |
| `fundamental get <symbol>` | Fundamental analysis data |
| `ownership get <symbol>` | Institutional ownership |
| `rs-history get <symbol>` | Relative strength rating history |
| `chart history <symbol>` | Price history (daily or weekly) |
| `chart markups <symbol>` | Chart annotations and markups |
| `catalog list` | List watchlists, screens, reports |
| `catalog run` | Run a watchlist, screen, or report |
| `skills generate` | Generate agent skill documentation |

## Agent integration

### Token-efficient stock analysis

`stock analyze` supports output modes designed for AI agents and batch comparison workflows:

```bash
marketsurge-agent stock analyze --tickers AAPL,MSFT,NVDA --compact --flat
```

- `--tickers AAPL,MSFT,NVDA` analyzes comma-separated symbols in one command. Positional symbols still work and can be combined with `--tickers`.
- `--compact` removes duplicate formatted string fields such as `market_cap_formatted`, while keeping raw numeric values.
- `--flat` flattens each analysis result inside the standard JSON envelope, for example `stock.pricing.market_cap` becomes `pricing_market_cap`.

The `skills generate` command writes Markdown skill files to `skills/` that describe each command's inputs, outputs, and usage. These files are designed for consumption by AI agent frameworks that support tool/skill discovery.

```bash
marketsurge-agent skills generate
```

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
skills/                  Generated agent skill docs
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
