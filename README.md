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
# List available reports
marketsurge-agent reports list

# Get IBD 50 report data (uses 23 default columns)
marketsurge-agent reports get <screen-id>

# Get report with custom columns
marketsurge-agent reports get <screen-id> --columns Symbol,Price,EPSRating,RSRating

# Use environment variable for cookie path
MARKETSURGE_AGENT_COOKIE_DB=/path/to/cookies.sqlite marketsurge-agent reports list

# Print version
marketsurge-agent --version
```

## Commands

| Command | Description |
| --- | --- |
| `reports list` | List all available screens and reports |
| `reports get <screen-id>` | Get report data for a specific screen ID |

### Global flags

| Flag | Env var | Description |
| --- | --- | --- |
| `--cookie-db PATH` | `MARKETSURGE_AGENT_COOKIE_DB` | Path to Firefox `cookies.sqlite` |
| `--verbose` | `MARKETSURGE_AGENT_VERBOSE` | Enable verbose logging to stderr |
| `--version` / `-V` | | Show version and exit |

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
  reports_list.go        reports list command
  reports_get.go         reports get command
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

### Local live smoke tests

Smoke tests run every API command against live MarketSurge data. They're excluded from normal test runs and CI by the `smoke` build tag.

Sign into MarketSurge in Firefox first, then run:

```bash
make smoke
```

For a specific Firefox profile:

```bash
export MARKETSURGE_AGENT_COOKIE_DB="$HOME/.mozilla/firefox/profile/cookies.sqlite"
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
