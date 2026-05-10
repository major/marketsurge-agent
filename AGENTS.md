# marketsurge-agent

Go CLI tool that lets AI agents query the MarketSurge stock research API. Single binary, JSON-first output, and `--help`.

This project is unofficial and is not affiliated with, approved by, or endorsed by MarketSurge or Investor's Business Daily.

Keep `.coderabbit.yaml` and `.github/copilot-instructions.md` plus `.github/instructions/*.instructions.md` aligned with current repo conventions when review-relevant behavior changes.

## Architecture

```text
cmd/
  marketsurge-agent/main.go   Entry point: kong.Parse, lazy auth via BindSingletonProvider, ctx.Run()
  root.go                     CLI struct (CLI, ChartCmd, CoachCmd, ColumnsCmd, CompareCmd, IndustryCmd, OverviewCmd, ReportsCmd, WatchlistCmd), kong flag tags
  chart.go                    ChartCmd.Run(client) - calls ChartMarketData for daily OHLCV + live quotes
  coach.go                    CoachCmd.Run(client) - calls CoachTree for curated watchlist/screen discovery
  columns.go                  ColumnsCmd.Run() - calls Columns()/ColumnsByCategory() for local column catalog (no auth)
  compare.go                  CompareCmd.Run(client) - calls MarketDataAdhocScreen for multi-symbol comparison data
  industry.go                 IndustryCmd.Run(client) - calls IndustryGroupRS for 6-month industry group RS
  overview.go                 OverviewCmd.Run(client) - calls OtherMarketData, RSRatingRIPanel, Ownership, Fundamentals, ChartMarketDataWeekly APIs
  reports_catalog.go          ReportsCatalogCmd.Run() - calls ReportScreens() for built-in report catalog (no auth)
  reports_list.go             ReportsListCmd.Run(client) - calls Screens API
  reports_get.go              ReportsGetCmd.Run(client) - calls RunScreen API
  reports_inspect.go          ReportsInspectCmd.Run(client) - calls Screen API for screen definition/filter criteria
  watchlist_list.go           WatchlistListCmd.Run(client) - calls GetAllWatchlistNames API
  watchlist_get.go            WatchlistGetCmd.Run(client) - calls FlaggedSymbols API
  root_test.go                Binary-level tests (help, version, auth error, missing subcommand)
  chart_test.go               Unit tests for chart
  coach_test.go               Unit tests for coach
  columns_test.go             Unit tests for columns
  compare_test.go             Unit tests for compare
  reports_catalog_test.go     Unit tests for reports catalog
  reports_list_test.go        Unit tests for reports list
  reports_get_test.go         Unit tests for reports get
  reports_inspect_test.go     Unit tests for reports inspect
  overview_test.go            Unit tests for overview
  watchlist_list_test.go      Unit tests for watchlist list
  watchlist_get_test.go       Unit tests for watchlist get
  industry_test.go            Unit tests for industry
internal/
  auth/                       Cookie-based JWT exchange
  cookies/                    Firefox cookie extraction
  errors/                     Custom error hierarchy + WriteJSON(w, err) + ExitCodeFor(err)
```

### Request flow

1. `main.go` calls `kong.Parse(&cli, ...)` which parses flags and selects the matched command
2. `kong.BindSingletonProvider` registers a lazy `newClient` factory; the client is only created when a command's `Run` method accepts `*marketsurge.Client`
3. Commands with `Run() error` (columns, reports catalog) skip auth entirely
4. Commands with `Run(client *marketsurge.Client) error` trigger `newClient`, which calls `auth.ResolveJWT` then `marketsurge.NewClient(marketsurge.WithJWT(jwt))`
5. The command calls the marketsurge-go client, marshals the result as JSON to stdout
6. On command error, `mserrors.WriteJSON(os.Stderr, err)` and `os.Exit(mserrors.ExitCodeFor(err))`

### Auth chain (`internal/auth/chain.go`)

Cookie database precedence:

1. `--cookie-db` explicit path to Firefox cookie DB
2. Firefox profile auto-discovery

Cookies are exchanged at `investors.com` using the JWT exchange URL constant, then the resulting JWT is used for API requests at `dowjones.io`. The CLI does not accept JWT injection through arguments, environment variables, or config files.

Auth is wired lazily via `kong.BindSingletonProvider` in `main.go`. The client is only created when a command's `Run` method requests `*marketsurge.Client`. Commands with `Run() error` signatures (columns, reports catalog) skip auth entirely.

### Error hierarchy (`internal/errors/errors.go`)

All errors embed `MarketSurgeError` base type. Use the constructor functions, not raw structs.

Exit code ranges: 0 = success, 30+ = domain errors.

| Type | Exit Code | When |
|---|---|---|
| `ValidationError` | 30 | Bad args, missing fields |
| `SymbolNotFoundError` | 31 | Ticker not recognized |
| `AuthenticationError` | 32 | 401/403, missing token |
| `TokenExpiredError` | 32 | 401 specifically |
| `CookieExtractionError` | 32 | Cookie DB read failures |
| `APIError` | 33 | API-level errors |
| `HTTPError` | 33 | 429, 5xx |

Import alias convention: `mserrors` in commands and main.

Error output uses `mserrors.WriteJSON(w, err)` which writes `{"code":"AUTH_FAILED","message":"..."}` to stderr. Success output goes to stdout as a raw JSON array (no envelope wrapper).

### Kong CLI struct (`cmd/root.go`)

The CLI uses [kong](https://github.com/alecthomas/kong) for flag parsing and command dispatch.

```go
type CLI struct {
    CookieDB string           `help:"..." env:"MARKETSURGE_AGENT_COOKIE_DB" name:"cookie-db"`
    Verbose  bool             `help:"..." env:"MARKETSURGE_AGENT_VERBOSE"`
    Version  kong.VersionFlag `help:"..." short:"V"`
    Chart     ChartCmd     `cmd:"" help:"..."`
    Coach     CoachCmd     `cmd:"" help:"..."`
    Columns   ColumnsCmd   `cmd:"" help:"..."`
    Compare   CompareCmd   `cmd:"" help:"..."`
    Industry  IndustryCmd  `cmd:"" help:"..."`
    Overview  OverviewCmd  `cmd:"" help:"..."`
    Reports   ReportsCmd   `cmd:"" help:"..."`
    Watchlist WatchlistCmd `cmd:"" help:"..."`
}

type ReportsCmd struct {
    Catalog ReportsCatalogCmd `cmd:"" help:"..."`
    Get     ReportsGetCmd     `cmd:"" help:"..."`
    Inspect ReportsInspectCmd `cmd:"" help:"..."`
    List    ReportsListCmd    `cmd:"" help:"..."`
}

type WatchlistCmd struct {
    List WatchlistListCmd `cmd:"" help:"..."`
    Get  WatchlistGetCmd  `cmd:"" help:"..."`
}
```

Key behaviors:
- `--cookie-db` and `--verbose` bind to `MARKETSURGE_AGENT_COOKIE_DB` and `MARKETSURGE_AGENT_VERBOSE` via `env:` tags
- `--version` / `-V` prints the version set via ldflags and exits
- Kong exits with code 80 when a required subcommand is missing
- `[]string` flags use `sep:","` to accept comma-separated values
- `columns` is a top-level porcelain command for local column catalog (no auth); supports `--category` filtering
- `chart <symbol>` is a top-level porcelain command for daily OHLCV chart data and live/extended-hours quotes
- `coach` is a top-level porcelain command for discovering MarketSurge curated watchlists and screens
- `compare <symbols...>` is a top-level porcelain command for comparing short symbol lists before deeper review
- `overview <symbol>` is a top-level porcelain command for one stock or ETF, not part of the `reports` group
- `industry <symbols...>` is a top-level porcelain command accepting comma-separated or space-separated symbols
- `watchlist` is a group command mirroring the `reports` subcommand pattern (list + get)

## Conventions

### Code style

- **Kong struct pattern**: Commands are structs with kong struct tags; no `init()` wiring needed
- **Command dispatch**: Commands that call MarketSurge APIs implement `Run(ctx context.Context, client *marketsurge.Client) error`; Kong calls them via `ctx.Run()` with a signal-aware context and lazy client provider. Commands that need no auth implement `Run() error` instead
- **Error wrapping**: Always use `fmt.Errorf("context: %w", err)` with `%w`
- **Typed errors**: Use `errors.As()` to match, constructor functions to create
- **Output**: Commands write JSON directly to `os.Stdout` (or an `io.Writer` for testability); no envelope wrapper

### Critical constraints

- Auth is wired in `main.go` through Kong's lazy singleton provider before `ctx.Run()` invokes an auth-backed command; there is no per-command auth hook
- `chart` emits a one-element JSON array with LLM-oriented keys (`ticker`, `days`, `exchange`, `dataPoints`, `quote`, `premarketQuote`, `postmarketQuote`, `currentMarketState`); days reflects actual data point count; DataPoints use `close` (mapped from API's `Last`)
- `coach` emits a flat JSON array of `coachNode` objects with a synthetic `category` field (`"watchlist"` or `"screen"`); supports `--type=watchlist|screen|all` filtering; empty result is `[]`
- `compare` emits one JSON object per returned symbol, keeps all requested MarketSurge columns under `columns`, and groups default columns into LLM-friendly keys (`ratings`, `price`, `volume`, `momentum`, `fundamentals`, `industry`, `ownership`, `events`)
- `overview` emits a one-element JSON array with LLM-oriented keys (`ticker`, `ratings`, `price`, `relativeStrengthTrend`, `ants`, `patterns`, `tightAreas`, `industry`, `ownership`, `fundamentals`, `risk`, `weeklyTrend`) and returns `SymbolNotFoundError` when `OtherMarketData` returns no rows
- `overview` also includes enrichment sections from the same OtherMarketData response: `businessDescription`, `ipoDate`, `ipoPrice`, `valuation` (P/E, fwd P/E, P/S, P/CF, yield, P/E vs S&P), `historicalPrices` (multi-period high/low/close/change), `volumeAverages`, `earningsCalendar` (EPSDueDate, status, last reported), `corporateActions` (dividends, splits, spinoffs), `profitMargins` (gross, pre-tax, after-tax, ROE per period), `growthRates` (EPS and sales from consensus)
- `columns` emits a JSON array of `columnItem` objects with keys `name`, `displayName`, `description`, `category`; supports `--category` filtering; empty result is `[]`
- `reports catalog` emits a JSON array of `reportScreenItem` objects with keys `id`, `name`, `description`; uses local catalog, no auth required
- `reports get` accepts `--columns` as a comma-separated list with `sep:","` and a default of 23 column names
- `reports list` emits a JSON array of `marketsurge.ScreenEntry` objects; empty result is `[]`, not `null`
- `reports get` reshapes `[][]RunScreenCell` into `[]map[string]any` keyed by `MDItem.Name`
- `reports inspect` emits a one-element JSON array with LLM-oriented keys (`id`, `name`, `description`, `type`, `filters`, `filterType`, `resultLimit`, `sortBy`, `lastResult`, `createdAt`, `updatedAt`); supports `--coach` flag for MarketSurge coach screens
- `watchlist list` emits a JSON array of `marketsurge.WatchlistSummary` objects; empty result is `[]`, not `null`
- `watchlist get` emits a one-element JSON array with LLM-oriented keys (`id`, `name`, `lastModifiedDateUtc`, `description`, `symbols`), extracting ticker symbols from `WatchlistItem.Key`
- `industry` emits a JSON array of `{ticker, industryGroupRS}` objects from the 6-month industry group RS data; nil RS values are preserved as JSON `null`

### Adding a new command

1. Add a new struct to `cmd/root.go` (or a new file for larger commands) with kong struct tags
2. Embed the struct in the appropriate parent (`ReportsCmd` or a new group on `CLI`)
3. Implement `Run(ctx context.Context, client *marketsurge.Client) error` for auth-backed commands, or `Run() error` for local no-auth commands
4. Write JSON output to `os.Stdout` and use an internal `run(..., w io.Writer)` helper when tests need to capture output without swapping global stdout
5. Add tests in `cmd/<command>_test.go` using the external `cmd_test` package

## Testing

- Framework: Go stdlib `testing` + `testify/assert` + `testify/require`
- CI runs: `go test -v -race -coverprofile=coverage.out ./...`
- Mock pattern: `marketsurge.WithHTTPClient()` + `httptest.NewServer` (no external mock libraries)
- Tests live in the external `cmd_test` package and construct command structs directly
- Table-driven subtests with `t.Run()`, typed error checks with `assert.ErrorAs()`
- Commands that write to `os.Stdout` accept an `io.Writer` parameter in their internal `run` method for output capture in tests

## Build

Requires Go 1.26+.

```bash
make build     # Build binary
make test      # go test -v -race -coverprofile
make smoke     # Local-only live API smoke tests with -tags=smoke
make lint      # golangci-lint run
make clean     # Remove binary + coverage
```

Local smoke tests live in `cmd/smoke_test.go` behind `//go:build smoke` and run with `make smoke`. They require a local Firefox MarketSurge session or `MARKETSURGE_AGENT_COOKIE_DB`.

Linting uses [golangci-lint](https://golangci-lint.run/) v2 with config in `.golangci.yml`. The standard linter set is enabled plus `bodyclose`, `errorlint`, `gocritic`, `misspell`, `modernize`, `nolintlint`, `revive`, `unconvert`, and `unparam`.

CI pipeline: `golangci-lint` (separate job) + `go test -v -race` -> `go build`

Release: push `v*` tag -> goreleaser v2 -> multi-platform binaries (linux/darwin, amd64/arm64, CGO disabled) -> GitHub Releases

## Maintenance

- **Keep this file updated**: When adding, removing, or changing commands, error types, conventions, or architecture, update this file to match.
- **Keep README.md updated**: When changing commands, flags, output format, install instructions, or development workflow, update README.md as well.

## Dependencies

- `github.com/alecthomas/kong` - CLI framework (flag parsing, command dispatch)
- `github.com/major/marketsurge-go/marketsurge` - MarketSurge API client
- `github.com/browserutils/kooky` - Firefox cookie extraction
- `github.com/stretchr/testify` - Test assertions
