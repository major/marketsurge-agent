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

# Equivalent schema-visible flag form for agents and MCP clients
marketsurge-agent stock get --symbol AAPL

# Analyze multiple symbols concurrently
marketsurge-agent stock analyze AAPL MSFT NVDA GOOG

# Analyze a comma-separated batch and remove low-value fields
marketsurge-agent stock analyze --symbols AAPL,MSFT,NVDA --compact

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
marketsurge-agent chart history --symbol AAPL --lookback 3M

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
| --- | --- |
| `stock get <symbol>` / `stock get --symbol SYMBOL` | Stock data (ratings, pricing, financials) |
| `stock analyze [symbols...]` / `stock analyze --symbols SYMBOLS` | Concurrent single-symbol or multi-symbol analysis with optional compact, flat, summary, and comma-separated batch modes |
| `fundamental get <symbol>` / `fundamental get --symbol SYMBOL` | Fundamental analysis data |
| `ownership get <symbol>` / `ownership get --symbol SYMBOL` | Institutional ownership |
| `rs-history get [symbols...]` / `rs-history get --symbols SYMBOLS` | Relative strength rating history for one or more symbols |
| `chart history <symbol>` / `chart history --symbol SYMBOL` | Price history (daily or weekly) |
| `chart markups <symbol>` / `chart markups --symbol SYMBOL` | Chart annotations and markups |
| `catalog list` | List watchlists, screens, reports |
| `catalog run` | Run a watchlist, coach screen, or report |
| `completion <shell>` | Generate shell completion script (bash, zsh, fish, powershell) |
| `--version` | Print version information |

### Shell completion and self-documentation

```bash
# Generate shell completion (bash, zsh, fish, powershell)
marketsurge-agent completion zsh > ~/.zsh/completions/_marketsurge-agent

# Print the full command tree JSON schema for LLM tool definitions
marketsurge-agent --jsonschema

# Explicit full-tree mode, kept for scripts that already request it
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

marketsurge-agent is built for coding agents and LLM tools that need structured stock research data without scraping MarketSurge pages. The repository uses a generated root `SKILL.md` for skill-aware tools plus hand-maintained `AGENTS.md` files, live `--jsonschema`, `--help`, and `--mcp` output from the binary.

### Claude Code and skill-aware tools

Use the checked-in root `SKILL.md` when your tool supports Agent Skills, including Claude Code. It contains trigger phrases, command descriptions, flag tables, examples, and MCP server hints generated from the live Cobra and structcli command tree.

For Claude Code, copy or symlink the generated skill into your local skills directory if you want it available outside this repository:

```bash
mkdir -p ~/.claude/skills/marketsurge-agent
ln -sf "$(pwd)/SKILL.md" ~/.claude/skills/marketsurge-agent/SKILL.md
```

If you already have a custom skill at that path, review or back it up first because `ln -sf` replaces the existing file or symlink. Regenerate `SKILL.md` after command, flag, default, or help text changes so the skill matches the current command surface:

```bash
make docs
```

### OpenCode and Codex

OpenCode and Codex use `AGENTS.md` as project instructions. Start in the repository root so the tool can load the root `AGENTS.md`; package-specific guidance lives in `cmd/AGENTS.md` and `internal/client/AGENTS.md` for command wiring, MCP behavior, JSON Schema behavior, client methods, and test patterns.

Use `--jsonschema`, `--help`, `help env-vars`, and `help config-keys` for the generated command and configuration contract. Update the relevant `AGENTS.md` files by hand when project structure, command patterns, build steps, auth behavior, or safety rules change.

### Token-efficient stock analysis

`stock analyze` supports output modes designed for AI agents and batch comparison workflows:

```bash
marketsurge-agent stock analyze --symbols AAPL,MSFT,NVDA --compact --flat
```

- `--symbols AAPL,MSFT,NVDA` analyzes comma-separated symbols in one command. Positional symbols still work and can be combined with `--symbols`.
- `--tickers AAPL,MSFT,NVDA` remains supported as a backward-compatible alias for `stock analyze`.
- `--compact` removes low-value fields such as formatted duplicates, profile metadata, internal IDs, empty fields, and stale arrays, while keeping decision-relevant raw values.
- `--summary` returns one small screening object per symbol with rankings, signal flags, ANTS dates and explanation, base details, liquidity, volatility, and ownership fields. Response metadata includes `mode: "summary"`.
- `--flat` flattens each analysis result inside the standard JSON envelope, for example `stock.pricing.market_cap` becomes `pricing_market_cap`.

`stock analyze` also includes MarketSurge technical context for chart-driven screening: `stock.base_pattern` summarizes the current base with pattern type, base stage, pivot price, base length, depth, and volume at pivot; `stock.signals` reports blue dot and ANTS signal flags when the API provides them. ANTS marks flag institutional accumulation: repeated upside price action with rising volume over a recent 15-day window. Full output exposes the mark dates at `stock.pricing.ant_dates`; summary output includes `ant_dates` and `ant_explanation` whenever `ant_signal` is true.

`rs-history get` accepts multiple symbols in one request through positional arguments or `--symbols`. Multi-symbol output uses a `data` object keyed by ticker so agents can compare RS trends without shell loops.

### Schema discovery

`--jsonschema` prints a machine-readable JSON schema for the full command tree and exits without making any API calls. Agents see every subcommand plus flag groups, enum values, defaults, environment bindings, and descriptions in one response. `--jsonschema=tree` remains supported and returns the same full-tree contract for scripts that already request it:

```bash
marketsurge-agent --jsonschema
marketsurge-agent --jsonschema=tree
```

The schema is a JSON array. Each entry is one command schema with `title`, `description`, `properties`, `x-structcli-groups`, `x-structcli-env-prefix`, and `x-structcli-config-flag` metadata. Enum flags expose both machine-readable `enum` arrays and the original enum tokens in their descriptions, for example `{DAILY,WEEKLY}` or `{daily,weekly}`.

Required and conditional inputs are documented in flag descriptions with concrete examples. Some rules are intentionally validated by the command instead of structcli tags so errors keep the normal JSON envelope and MarketSurge exit codes, for example `chart history` requires either `--lookback` or `--start-date/--end-date`, and `catalog run` requires the ID flag that matches `--kind`. Complex command descriptions and flag descriptions include copyable valid invocation shapes so LLM agents can infer usable calls from schema output.

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

MCP tool names use underscores between command path segments, such as `stock_analyze`, `chart_history`, and `catalog_run`. Existing dashes inside a command segment remain, so `rs-history get` becomes `rs-history_get`. The tool list is limited to runnable MarketSurge data commands; shell completion and parent help commands are not exposed.

## Development

Requires Go 1.26+.

```bash
make build     # Build binary
make test      # Run tests with race detector
make smoke     # Run local live-data smoke tests against MarketSurge
make lint      # Run golangci-lint (install: https://golangci-lint.run/welcome/install/)
make docs      # Refresh root SKILL.md
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

### Local live smoke tests

Smoke tests exercise every MarketSurge API leaf command against live data from your local machine. They are excluded from normal test runs and CI by the `smoke` build tag.

Sign into MarketSurge in Firefox first, then run:

```bash
make smoke
```

For a specific Firefox profile, set the same cookie database environment variable used by the CLI:

```bash
export MARKETSURGE_AGENT_COOKIE_DB="$HOME/.mozilla/firefox/profile/cookies.sqlite"
make smoke
```

`catalog run` needs an account-specific ID. Set one of these optional variables to include it; otherwise that subtest is skipped while the schema coverage check still ensures the command has a smoke case:

```bash
export MARKETSURGE_SMOKE_WATCHLIST_ID=12345
export MARKETSURGE_SMOKE_REPORT_ID=67890
export MARKETSURGE_SMOKE_COACH_SCREEN_ID=screen-1
```

The smoke suite validates exit status and JSON envelope shape, not exact live prices, counts, timestamps, or rankings.

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
