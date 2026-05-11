# marketsurge-agent

CLI tool that gives AI agents structured access to [MarketSurge](https://marketsurge.investors.com) stock research data. Every command writes JSON to stdout, making it easy to integrate with agent frameworks, pipelines, or scripts.

> **Disclaimer**: This project is unofficial and is not affiliated with, approved by, or endorsed by MarketSurge or Investor's Business Daily. Use at your own risk.

## Install

Pre-built binaries are available on the [Releases](https://github.com/major/marketsurge-agent/releases) page for Linux and macOS (amd64/arm64).

From source:

```bash
go install github.com/major/marketsurge-agent/cmd/marketsurge-agent@latest
```

## Authentication

MarketSurge requires an active browser session. The CLI exchanges Firefox cookies for the API JWT automatically. Cookie databases resolve in this order:

1. `--cookie-db` path to a Firefox `cookies.sqlite` file
2. Auto-discovery from local Firefox profiles

Sign into MarketSurge in Firefox before running any command. For automation, point the CLI at the Firefox cookie database used by that logged-in profile:

```bash
marketsurge-agent --cookie-db "$HOME/.mozilla/firefox/profile/cookies.sqlite" reports list
```

You can also set environment variables instead of passing flags every time:

```bash
export MARKETSURGE_AGENT_COOKIE_DB="$HOME/.mozilla/firefox/profile/cookies.sqlite"
export MARKETSURGE_AGENT_VERBOSE=true
```

## Usage

```bash
# Compare key data for several stocks or ETFs
marketsurge-agent compare AMD NVDA MSFT

# Summarize high-level data for one stock or ETF
marketsurge-agent overview AMD

# Show daily OHLCV chart data and live quotes
marketsurge-agent chart AMD

# Show industry group relative strength
marketsurge-agent industry AMD NVDA MSFT

# Discover curated watchlists and screens
marketsurge-agent coach

# List available data columns (no auth required)
marketsurge-agent columns
marketsurge-agent columns --category Ratings

# List built-in report screens (no auth required)
marketsurge-agent reports catalog

# List available reports
marketsurge-agent reports list

# Get IBD 50 report data (uses 23 default columns)
marketsurge-agent reports get <screen-id>

# Get report with custom columns
marketsurge-agent reports get <screen-id> --columns Symbol,Price,EPSRating,RSRating

# Inspect a screen's definition and filter criteria
marketsurge-agent reports inspect <screen-id>

# List saved watchlists
marketsurge-agent watchlist list

# Get symbols in a watchlist
marketsurge-agent watchlist get <watchlist-id>

# Use environment variable for cookie path
MARKETSURGE_AGENT_COOKIE_DB=/path/to/cookies.sqlite marketsurge-agent reports list

# Print version
marketsurge-agent --version
```

## Commands

| Command | Description |
| --- | --- |
| `chart <symbol>` | Show daily OHLCV chart data and live quotes |
| `coach` | Discover MarketSurge curated watchlists and screens |
| `columns` | List available MarketSurge data columns (no auth required) |
| `compare <symbols...>` | Compare key MarketSurge data for multiple stocks or ETFs |
| `industry <symbols...>` | Show industry group relative strength for stocks or ETFs |
| `overview <symbol>` | Summarize high-level stock or ETF data for LLM context |
| `reports catalog` | List built-in MarketSurge report screens (no auth required) |
| `reports get <screen-id>` | Get report data for a specific screen ID |
| `reports inspect <screen-id>` | Inspect screen definition and filter criteria |
| `reports list` | List all available screens and reports |
| `watchlist list` | List all saved watchlists |
| `watchlist get <watchlist-id>` | Get symbols in a watchlist by ID |

### Global flags

| Flag | Env var | Description |
| --- | --- | --- |
| `--cookie-db PATH` | `MARKETSURGE_AGENT_COOKIE_DB` | Path to Firefox `cookies.sqlite` |
| `--verbose` | `MARKETSURGE_AGENT_VERBOSE` | Enable verbose logging to stderr |
| `--version` / `-V` | | Show version and exit |

### columns

Lists the available MarketSurge data columns from the local catalog. No authentication required. Use `--category` to filter by category (exact match).

```bash
marketsurge-agent columns
marketsurge-agent columns --category Ratings
```

Example output:

```json
[
  {
    "name": "CompositeRating",
    "displayName": "Composite Rating",
    "description": "Overall rating combining EPS, RS, and other factors",
    "category": "Ratings"
  }
]
```

### compare

Fetches a compact side-by-side snapshot for a short list of stocks or ETFs. The command uses MarketSurge screen columns and returns one JSON object per returned symbol. Curated columns are grouped into LLM-oriented keys, while every requested MarketSurge value remains available under `columns` for exact column-level inspection.

```bash
marketsurge-agent compare AMD NVDA MSFT
marketsurge-agent compare AMD NVDA --columns Symbol,Price,CompositeRating,RSRating
```

The `--columns` flag (env: `MARKETSURGE_AGENT_COMPARE_COLUMNS`) controls which fields MarketSurge returns. If omitted, the command requests a default trader-focused set covering ratings, price, volume, moving-average momentum, fundamentals, industry, ownership, earnings dates, IPO date, market cap, and Blue Dot events.

Example output:

```json
[
  {
    "ticker": "AMD",
    "name": "Advanced Micro Devices",
    "ratings": {"composite": "96", "relativeStrength": "91"},
    "price": {"last": "212.30", "atrPercent21d": "4.1%"},
    "industry": {"groupRSRating": "A"},
    "columns": {"Symbol": "AMD", "Price": "212.30", "CompositeRating": "96", "RSRating": "91"}
  }
]
```

### industry

Fetches the 6-month industry group relative strength ranking for one or more symbols. Accepts comma-separated or space-separated tickers.

```bash
marketsurge-agent industry AMD NVDA
marketsurge-agent industry AMD,NVDA,MSFT
```

Example output:

```json
[
  {"ticker": "AMD", "industryGroupRS": 85},
  {"ticker": "NVDA", "industryGroupRS": 92}
]
```

A `null` value for `industryGroupRS` means the API returned no RS data for that symbol.

### overview

Fetches a token-efficient, symbol-centered summary for one stock or ETF. The command combines MarketSurge market data, relative strength history, chart pattern metadata, ANTs event dates, industry ranks, ownership history, fundamentals, risk context, and a compact weekly trend snapshot into a compact JSON object. It also includes valuation ratios (P/E, forward P/E, P/S, P/CF, yield), historical price statistics, volume averages, earnings calendar, corporate actions (dividends, splits, spinoffs), profit margins, growth rates, business description, and IPO date/price. Output is still a JSON array, so a successful single-symbol response is `[{}]` rather than `{}`.

```bash
marketsurge-agent overview AMD
```

Example output:

```json
[
  {
    "ticker": "AMD",
    "name": "Advanced Micro Devices",
    "type": "COMMON_STOCK",
    "ratings": {"composite": 96, "epsRating": 83, "relativeStrength": 91, "salesMarginsROE": "A", "accumulationDistribution": "B+"},
    "price": {"marketCap": {"v": 300000000000, "f": "300B"}, "avgDollarVolume50d": {"v": 2500000000, "f": "2.5B"}},
    "relativeStrengthTrend": {"newHigh": true, "history": [{"value": 91, "letter": "A", "period": "P12M", "offset": "CURRENT"}]},
    "ants": {"count": 2, "dates": ["2026-05-01", "2026-05-08"]},
    "industry": {"name": "Electronics-Semiconductor Fabless", "sector": "Technology", "ranks": [{"value": 5, "period": "P1M", "offset": "CURRENT"}]},
    "ownership": {"fundsFloatPct": {"v": 49.5, "f": "49.5%"}, "funds": [{"date": "2026-03-31", "funds": "3,210"}]},
    "fundamentals": {"debtPct": {"v": 20.1, "f": "20.1%"}, "reportedEPS": {"value": {"v": 1.65, "f": "$1.65"}}},
    "risk": {"beta": {"v": 1.4, "f": "1.4"}},
    "weeklyTrend": {"period": "ONE_WEEK", "latest": {"last": {"v": 212.3}}, "quote": {"timeliness": "REAL_TIME"}}
  }
]
```

Missing symbols return a `SYMBOL_NOT_FOUND` error with exit code 31. MarketSurge exposes RS Rating and RS line data through the current client; the command does not invent a technical RSI value when that indicator is unavailable.

### reports catalog

Lists the built-in MarketSurge report screens from the local catalog. No authentication required. Use these IDs with `reports get` to fetch report data.

```bash
marketsurge-agent reports catalog
```

Example output:

```json
[
  {
    "id": 1,
    "name": "IBD 50",
    "description": "Top-rated growth stocks based on CAN SLIM criteria"
  }
]
```

### reports list

Lists all screens and reports available in your MarketSurge account. Output is a JSON array of screen entries, each with an ID, name, type, and description.

```bash
marketsurge-agent reports list
```

Example output:

```json
[
  {
    "id": "12345",
    "name": "IBD 50",
    "type": "screen",
    "description": "Top-rated growth stocks"
  }
]
```

### reports get

Fetches report data for a screen ID from `reports list`. Output is a JSON array where each element is an object with column names as keys.

```bash
marketsurge-agent reports get <screen-id>
marketsurge-agent reports get <screen-id> --columns Symbol,Price,EPSRating,RSRating
```

The `--columns` flag (env: `MARKETSURGE_AGENT_COLUMNS`) controls which fields appear in each row. Pass a comma-separated list. The default set of 23 columns is:

```text
Symbol, CompanyName, ListRank, Price, PriceNetChg, PricePctChg, PricePctOff52WHigh,
VolumePctChgVs50DAvgVolume, VolumeAvg50Day, MarketCapIntraday, CompositeRating,
EPSRating, RSRating, AccDisRating, SMRRating, IndustryGroupRank, IndustryName,
VolumeDollarAvg50D, IPODate, DowJonesKey, ChartingSymbol, DowJonesInstrumentType,
DowJonesInstrumentSubType
```

Example output:

```json
[
  {
    "Symbol": "NVDA",
    "CompanyName": "NVIDIA Corp",
    "CompositeRating": 99,
    "EPSRating": 99,
    "RSRating": 97
  }
]
```

### watchlist list

Lists all saved watchlists in your MarketSurge account. Output is a JSON array of watchlist entries, each with an ID, name, description, and last-modified timestamp.

```bash
marketsurge-agent watchlist list
```

Example output:

```json
[
  {
    "id": "12345",
    "name": "Growth Leaders",
    "lastModifiedDateUtc": "2026-05-01T12:00:00Z",
    "description": "Top growth stocks"
  }
]
```

### watchlist get

Fetches watchlist details and the list of ticker symbols for a watchlist ID from `watchlist list`. Output is a one-element JSON array with an LLM-friendly shape.

```bash
marketsurge-agent watchlist get <watchlist-id>
```

Example output:

```json
[
  {
    "id": "12345",
    "name": "Growth Leaders",
    "lastModifiedDateUtc": "2026-05-01T12:00:00Z",
    "description": "Top growth stocks",
    "symbols": ["AAPL", "NVDA", "AMD"]
  }
]
```

## Output format

All commands write to stdout and stderr separately:

- **stdout**: JSON arrays (one per command). Empty results are `[]`, never `null`.
- **stderr**: JSON error objects when a command fails.

Error shape on stderr:

```json
{"code": "AUTH_FAILED", "message": "authentication failed: ..."}
```

## Exit codes

| Code | Meaning |
| --- | --- |
| 0 | Success |
| 30 | Validation error (bad arguments) |
| 31 | Symbol not found |
| 32 | Authentication error (missing or expired session) |
| 33 | API or HTTP error |

## Development

Requires Go 1.26+.

```bash
make build     # Build binary
make test      # Run tests with race detector
make smoke     # Run local live-data smoke tests against MarketSurge
make lint      # Run golangci-lint (install: https://golangci-lint.run/welcome/install/)
make clean     # Remove binary and build artifacts
```

### Project layout

```text
cmd/
  marketsurge-agent/     Entry point (main.go)
  root.go                Root command struct (kong)
  chart.go               chart command
  coach.go               coach command
  columns.go             columns command (no auth)
  compare.go             compare command
  industry.go            industry command
  overview.go            overview command
  reports_catalog.go     reports catalog command (no auth)
  reports_get.go         reports get command
  reports_inspect.go     reports inspect command
  reports_list.go        reports list command
  watchlist_get.go       watchlist get command
  watchlist_list.go      watchlist list command
internal/
  auth/                  Cookie-based JWT exchange
  cookies/               Firefox cookie extraction
  errors/                Typed error hierarchy
```

### Running tests

```bash
go test -v -race ./...
```

Tests use `httptest.NewServer` for HTTP mocking with no external mock libraries.

### Smoke target

The smoke target runs the command package test suite with `go test -v ./cmd/`. This keeps command-level coverage quick and local without the removed `smoke` build tag or live-only `cmd/smoke_test.go` setup.

Run the smoke target with:

```bash
make smoke
```

### Releasing

Push a version tag to trigger a [GoReleaser](https://goreleaser.com) build:

```bash
git tag v1.2.3
git push origin v1.2.3
```

This produces binaries for linux/darwin on amd64/arm64, published to GitHub Releases.

### Dependencies

- [`github.com/alecthomas/kong`](https://github.com/alecthomas/kong) - CLI framework
- [`github.com/browserutils/kooky`](https://github.com/browserutils/kooky) - Firefox cookie extraction
- [`github.com/major/marketsurge-go`](https://github.com/major/marketsurge-go) - MarketSurge API client
- [`github.com/stretchr/testify`](https://github.com/stretchr/testify) - Test assertions
