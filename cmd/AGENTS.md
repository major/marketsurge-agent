# cmd package

Kong command tree for marketsurge-agent.

## Structure

- `root.go` - `CLI` root struct, top-level `CompareCmd`, `IndustryCmd`, `OverviewCmd`, `ReportsCmd` group, `WatchlistCmd` group; kong flag tags
- `compare.go` - `CompareCmd.Run(client)` - calls `client.MarketDataAdhocScreen()` for multi-symbol comparison data
- `industry.go` - `IndustryCmd.Run(client)` - calls `client.IndustryGroupRS()` for 6-month industry group RS
- `overview.go` - `OverviewCmd.Run(client)` - calls `client.OtherMarketData()`, `client.RSRatingRIPanel()`, `client.Ownership()`, `client.Fundamentals()`, and `client.ChartMarketDataWeekly()`
- `reports_list.go` - `ReportsListCmd.Run(client)` - calls `client.Screens()`
- `reports_get.go` - `ReportsGetCmd.Run(client)` - calls `client.RunScreen()`
- `watchlist_list.go` - `WatchlistListCmd.Run(client)` - calls `client.GetAllWatchlistNames()`
- `watchlist_get.go` - `WatchlistGetCmd.Run(client)` - calls `client.FlaggedSymbols()`
- `root_test.go` - Binary-level tests (help, version, auth error, missing subcommand)
- `reports_list_test.go` - Unit tests for reports list
- `reports_get_test.go` - Unit tests for reports get
- `watchlist_list_test.go` - Unit tests for watchlist list
- `watchlist_get_test.go` - Unit tests for watchlist get
- `industry_test.go` - Unit tests for industry

## Patterns

### Kong struct pattern

Commands are plain structs with kong struct tags. No `init()` wiring, no constructor functions.

```go
type CLI struct {
    CookieDB string           `help:"..." env:"MARKETSURGE_AGENT_COOKIE_DB" name:"cookie-db"`
    Verbose  bool             `help:"..." env:"MARKETSURGE_AGENT_VERBOSE"`
    Version  kong.VersionFlag `help:"..." short:"V"`
    Compare   CompareCmd   `cmd:"" help:"..."`
    Industry  IndustryCmd  `cmd:"" help:"..."`
    Overview  OverviewCmd  `cmd:"" help:"..."`
    Reports   ReportsCmd   `cmd:"" help:"..."`
    Watchlist WatchlistCmd `cmd:"" help:"..."`
}

type ReportsCmd struct {
    List ReportsListCmd `cmd:"" help:"..."`
    Get  ReportsGetCmd  `cmd:"" help:"..."`
}

type WatchlistCmd struct {
    List WatchlistListCmd `cmd:"" help:"..."`
    Get  WatchlistGetCmd  `cmd:"" help:"..."`
}
```

Each leaf command struct implements `Run(client *marketsurge.Client) error`. Kong dispatches to it via `ctx.Run(client)` in `main.go`.

### Client injection

Auth runs in `main.go` before `ctx.Run`. The client is constructed once and passed directly to `ctx.Run(client)`. Commands receive an already-authenticated client; there is no per-command auth hook.

```go
// main.go
client, err := newClient(cli.CookieDB)
// ...
if err := ctx.Run(client); err != nil { ... }
```

### Command flags

Flags are struct fields with kong tags. `[]string` fields use `sep:","` for comma-separated input.

```go
type ReportsGetCmd struct {
    ScreenID string   `arg:"" help:"Screen ID from 'reports list' output."`
    Columns  []string `help:"Response columns to include." default:"Symbol,..." sep:","`
}
```

### Output

Commands write raw JSON arrays to stdout (no envelope wrapper). Empty results are `[]`, never `null`. Errors are returned to `main.go`, which calls `mserrors.WriteJSON(os.Stderr, err)`. `overview <symbol>` is a single-symbol porcelain command, but it still returns a one-element JSON array to preserve this contract.

### Porcelain compare output

`CompareCmd` uses `MarketDataAdhocScreen` with `IncludeSource.Instruments` to compare a short list of stock or ETF symbols. Normalize symbols by trimming whitespace, uppercasing, and de-duplicating while preserving first occurrence. The response keeps every returned MarketSurge value in the `columns` map and also groups curated defaults into LLM-oriented keys (`ratings`, `price`, `volume`, `momentum`, `fundamentals`, `industry`, `ownership`, `events`). Empty API results are a successful `[]`; blank or missing symbols are validation errors.

### Porcelain industry output

`IndustryCmd` uses `IndustryGroupRS` to fetch 6-month industry group relative strength for one or more symbols. Symbols are normalized via `normalizeSymbols()` (trim, uppercase, deduplicate). The output is a flat JSON array of `{ticker, industryGroupRS}` objects. Nil RS values (when the API returns no data for a symbol) are preserved as JSON `null`. Empty symbols input is a validation error.

### Porcelain watchlist output

`WatchlistListCmd` uses `GetAllWatchlistNames` and emits a raw JSON array of `WatchlistSummary` objects directly. Empty result is `[]`.

`WatchlistGetCmd` uses `FlaggedSymbols` and reshapes the response into a one-element JSON array with LLM-oriented keys (`id`, `name`, `lastModifiedDateUtc`, `description`, `symbols`). Ticker symbols are extracted from `WatchlistItem.Key`, skipping empty keys. The `symbols` field is always a JSON array (never `null`).

### Porcelain overview output

`OverviewCmd` gathers high-level context from `OtherMarketData`, then enriches it with `RSRatingRIPanel`, `Ownership`, `Fundamentals`, and `ChartMarketDataWeekly`. Keep its JSON keys LLM-oriented and explicit enough to avoid MarketSurge shorthand confusion (`ticker`, `ratings`, `price`, `relativeStrengthTrend`, `ants`, `patterns`, `tightAreas`, `industry`, `ownership`, `fundamentals`, `risk`, `weeklyTrend`). Do not expose MarketSurge internal IDs in the overview payload. Treat empty `OtherMarketData.marketData` as `mserrors.NewSymbolNotFoundError(...)` rather than an empty success array.

### Error mapping

Commands map client errors to typed `mserrors` before returning:

```go
if marketsurge.IsAuthError(err) {
    return mserrors.NewAuthenticationError("authentication failed", err)
}
return mserrors.NewAPIError("API request failed", err)
```

### Testability

`ReportsListCmd` writes to `os.Stdout` directly. Tests redirect `os.Stdout` via `os.Pipe()` to capture output.

`ReportsGetCmd`, `CompareCmd`, `OverviewCmd`, `WatchlistListCmd`, `WatchlistGetCmd`, and `IndustryCmd` expose an internal `run(client, w io.Writer)` method so tests can pass a `bytes.Buffer` without redirecting `os.Stdout`.

### Test pattern

Tests live in the external `cmd_test` package. They construct command structs directly and build a test client with `httptest.NewServer`:

```go
func TestReportsListSuccess(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, `{"data":{"user":{"screens":[...]}}}`)
    }))
    t.Cleanup(server.Close)

    client, err := marketsurge.NewClient(
        marketsurge.WithJWT("test-token"),
        marketsurge.WithHTTPClient(server.Client()),
        marketsurge.WithGraphQLURL(server.URL),
    )
    require.NoError(t, err)

    // redirect stdout, run command, restore stdout
    err = (&agentcmd.ReportsListCmd{}).Run(client)
    require.NoError(t, err)
}
```

Binary-level tests in `root_test.go` build the binary in `TestMain` and run it as a subprocess to verify exit codes, help output, and auth error JSON shape.

### Local live smoke tests

`smoke_test.go` is guarded by `//go:build smoke` and is local-only. Run it with `make smoke`. It requires a live Firefox MarketSurge session or `MARKETSURGE_AGENT_COOKIE_DB`. Keep smoke cases serial; do not use `t.Parallel()`.

## Adding a new command

1. Add a new struct to `cmd/root.go` (or a new file for larger commands) with kong struct tags
2. Embed the struct in the appropriate parent (`ReportsCmd`, `WatchlistCmd`, or a new group on `CLI`)
3. Implement `Run(client *marketsurge.Client) error` on the struct
4. Write JSON output to `os.Stdout` (or accept `io.Writer` via an internal `run` method for testability)
5. Map client errors to `mserrors` types before returning
6. Add tests in `cmd/<command>_test.go` using the external `cmd_test` package
