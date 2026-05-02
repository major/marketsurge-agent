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

MarketSurge requires an active browser session. The CLI exchanges Firefox cookies for the API JWT automatically. Cookie databases resolve in this order:

1. `--cookie-db` path to a Firefox `cookies.sqlite` file
2. Auto-discovery from local Firefox profiles

Sign into MarketSurge in Firefox before running API commands. For automation, point the CLI at the Firefox cookie database used by that logged-in profile:

```bash
marketsurge-agent --cookie-db "$HOME/.mozilla/firefox/profile/cookies.sqlite" stock get AAPL
```

Non-secret root options can also come from structcli-managed environment variables:

```bash
export MARKETSURGE_AGENT_COOKIE_DB="$HOME/.mozilla/firefox/profile/cookies.sqlite"
export MARKETSURGE_AGENT_VERBOSE=true
```

## Configuration file

Config files can provide non-secret root options such as `cookie-db` and `verbose`:

```yaml
cookie-db: /home/alice/.mozilla/firefox/profile/cookies.sqlite
verbose: false
```

Use `--config` to load a specific file:

```bash
marketsurge-agent --config ~/.config/marketsurge-agent/config.yaml stock get AAPL
```

Without `--config`, structcli searches its default config locations for `config.yaml`, `config.json`, or `config.toml`, including `/etc/marketsurge-agent`, an executable-adjacent `.marketsurge-agent` directory, and `$HOME/.marketsurge-agent`. Set `MARKETSURGE_AGENT_CONFIG` to point at a config file from the environment. For root options managed by structcli, precedence is CLI flag, environment variable, config file, then default.

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
| `completion <shell>` | Generate shell completion script (bash, zsh, fish, powershell) |
| `--version` | Print version information |

### Shell completion and self-documentation

```bash
# Generate shell completion (bash, zsh, fish, powershell)
marketsurge-agent completion zsh > ~/.zsh/completions/_marketsurge-agent

# Print machine-readable JSON schema for the current command
marketsurge-agent --jsonschema

# Print the full command tree schema for LLM tool definitions
marketsurge-agent --jsonschema=tree

# Run as an MCP server over stdio for agent tool discovery and calls
marketsurge-agent --mcp

# List all supported environment variables
marketsurge-agent help env-vars

# List all supported config file keys
marketsurge-agent help config-keys

# Load config from a specific file
marketsurge-agent --config ~/.config/marketsurge-agent/config.yaml --debug-options

# Print resolved flag values for debugging
marketsurge-agent --debug-options

# Print version
marketsurge-agent --version
```

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

### Schema discovery

`--jsonschema` prints a machine-readable JSON schema for the current command and exits without making any API calls. For LLM tool definitions, prefer `--jsonschema=tree` so the agent sees every subcommand plus flag groups, enum values, defaults, environment bindings, and descriptions in one response:

```bash
marketsurge-agent --jsonschema
marketsurge-agent --jsonschema=tree
```

Required and conditional inputs are documented in flag descriptions. Some rules are intentionally validated by the command instead of structcli tags so errors keep the normal JSON envelope and MarketSurge exit codes, for example `chart history` requires either `--lookback` or `--start-date/--end-date`, and `catalog run` requires the ID flag that matches `--kind`.

Each command also carries a detailed `--help` description covering inputs, outputs, and gotchas. The `env-vars` and `config-keys` help topics list every supported environment variable and config file key:

```bash
marketsurge-agent help env-vars
marketsurge-agent stock analyze --help
```

### MCP server

`--mcp` runs `marketsurge-agent` as a Model Context Protocol server over stdio. Agents can use MCP `initialize`, `tools/list`, and `tools/call` requests to discover and invoke CLI commands without scraping help output:

```bash
marketsurge-agent --mcp
```

MCP discovery does not require MarketSurge authentication. Tool calls that fetch MarketSurge data still use the same Firefox cookie authentication chain as normal CLI commands and return the same JSON envelopes.

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
cmd/
  marketsurge-agent/     Entry point (main.go)
  root.go                Root command, auth, Execute()
  symbol.go              Shared symbol-fetcher pattern
  <group>.go             One file per command group
internal/
  auth/                  Cookie-based JWT exchange
  client/                GraphQL client + API methods
  constants/             API endpoints, column names
  cookies/               Firefox cookie extraction
  errors/                Typed error hierarchy
  models/                Data structures
  output/                JSON envelope formatting
queries/                 Embedded GraphQL queries
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

### Dependencies

- [`github.com/spf13/cobra`](https://github.com/spf13/cobra) - CLI framework
- [`github.com/leodido/structcli`](https://github.com/leodido/structcli) - Flag binding, validation, self-documentation
- [`github.com/browserutils/kooky`](https://github.com/nicholasgasior/kooky) - Firefox cookie extraction
- [`resty.dev/v3`](https://resty.dev) - HTTP client
- [`github.com/stretchr/testify`](https://github.com/stretchr/testify) - Test assertions
