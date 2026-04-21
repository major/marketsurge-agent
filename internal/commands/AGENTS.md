# internal/commands

CLI command implementations. Each file is a command factory that returns a `*cli.Command`.

## Pattern

Every command follows the same factory signature:

```go
func FooCommand(c *client.Client, w io.Writer) *cli.Command {
    return &cli.Command{
        Name:  "foo",
        Usage: "...",
        Flags: []cli.Flag{...},
        Action: func(ctx context.Context, cmd *cli.Command) error {
            // 1. Validate args
            // 2. Call client method
            // 3. output.WriteSuccess(w, result, metadata)
            return nil
        },
    }
}
```

Use `fundamental_get.go` (33 lines) as the canonical minimal example.

## Command groups

| Group | Files | Description |
|---|---|---|
| stock | `stock_get.go`, `stock_analyze.go` | Single/multi-symbol stock data |
| fundamental | `fundamental_get.go` | Fundamental analysis data |
| ownership | `ownership_get.go` | Institutional ownership |
| rs_history | `rs_history_get.go` | Relative strength history |
| chart | `chart_history.go`, `chart_markups.go` | Price history + chart annotations |
| catalog | `catalog_list.go`, `catalog_run.go` | List/run watchlists, screens, reports |
| skills | `skills_generate.go` | Auto-generate agent skill docs |

## Error handling

Import the errors package as `mserrors`:

```go
import mserrors "github.com/major/marketsurge-agent/internal/errors"
```

Return typed errors from constructors, never raw structs:

```go
return mserrors.NewValidationError("symbol is required", nil)
return mserrors.NewSymbolNotFoundError(symbol, nil)
```

All errors flow through `output.WriteError(w, err)` which maps them to the correct exit code and JSON envelope.

## Concurrency pattern (`stock_analyze.go`)

Multi-symbol operations use `sync.WaitGroup` + `sync.Mutex`:

```go
var wg sync.WaitGroup
var mu sync.Mutex
var results []Result
var errs []error

for _, symbol := range symbols {
    wg.Add(1)
    go func(s string) {
        defer wg.Done()
        result, err := c.GetStock(s)
        mu.Lock()
        defer mu.Unlock()
        if err != nil {
            errs = append(errs, err)
        } else {
            results = append(results, result)
        }
    }(symbol)
}
wg.Wait()
```

Use `output.WritePartial()` when some symbols succeed and others fail.

## Complex command: `catalog_run.go`

The most complex command (345 lines). Key details:

- `kind` flag is required, determines which GraphQL operation to use
- Each kind has its own ID flag: `--report-id`, `--watchlist-id`, `--coach-screen-id`
- Watchlist fields have aliases (e.g., "symbol" maps to "Symbol") defined in `watchlistFieldAliases`
- Supports filters: `--min-composite`, `--min-rs`, `--exclude-spacs`
- Pagination via `--limit` and `--offset`
- Field projection with `--fields` (comma-separated)

## Testing

Tests live in `_test.go` files alongside their commands. Shared helpers in `helpers_test.go`.

### Test helpers

- `testClient(handler http.HandlerFunc) *client.Client` - Creates a client pointing at an httptest.Server
- `jsonServer(response string) http.HandlerFunc` - Returns a handler that serves fixed JSON
- Fixture builders for stock data, fundamental data, etc.

### Test patterns

```go
func TestFooCommand(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        response string
        wantErr  bool
    }{...}

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := testClient(jsonServer(tt.response))
            var buf bytes.Buffer
            cmd := FooCommand(c, &buf)
            // ... run command, assert output
        })
    }
}
```

- CLI tests suppress `os.Exit` via `ExitErrHandler: func(_ context.Context, _ *cli.Command, _ error) {}`
- Use `assert.ErrorAs()` for typed error checks, not string matching
- Use `t.Setenv()` for environment variable tests, `t.TempDir()` for filesystem tests

## Adding a new command

1. Create `<group>_<action>.go` with the factory function
2. Add corresponding client method in `internal/client/<group>.go`
3. Add GraphQL query in `queries/<operation>.graphql`
4. Register in `main.go` under the appropriate command group's `Commands` slice
5. Create `<group>_<action>_test.go` with table-driven tests
6. Use `testClient()` + `jsonServer()` from `helpers_test.go`
