# marketsurge-agent

Go CLI tool that lets AI agents query the MarketSurge stock research API. Single binary, JSON-first output, static skill files for agent consumption.

This project is unofficial and is not affiliated with, approved by, or endorsed by MarketSurge or Investor's Business Daily.

## Architecture

```text
cmd/
  marketsurge-agent/main.go      Entry point, calls cmd.Execute()
  root.go                        Root command, PersistentPreRunE (auth), Execute()
  symbol.go                      Shared symbol-fetcher pattern
  <group>.go                     One file per command group (stock, chart, etc.)
internal/
  auth/                          JWT resolution (4-tier chain)
  client/                        GraphQL client + domain methods
  constants/                     API endpoints, columns, report IDs
  cookies/                       Firefox cookie extraction
  errors/                        Custom error hierarchy
  models/                        Data structures
  output/                        JSON envelope formatting
queries/                         Embedded .graphql files (go:embed)
skills/                          Static agent skill docs
```

### Request flow

1. `main.go` calls `cmd.Execute()` which runs the root cobra command
2. `PersistentPreRunE` resolves JWT via the auth chain, injects `client.Client` into context
3. Command `RunE` retrieves client via `ClientFromContext(cmd.Context())`, validates args, calls client method
4. Client loads embedded `.graphql` query, executes HTTP POST to GraphQL endpoint
5. Response parsed into typed model, wrapped in JSON envelope via `output.WriteSuccess`
6. `PersistentPostRunE` closes the client

### Auth chain (`internal/auth/chain.go`)

Four-tier JWT precedence, first non-empty wins:

1. `--jwt` CLI flag
2. `MARKETSURGE_JWT` env var
3. `--cookie-db` explicit path to Firefox cookie DB
4. Firefox profile auto-discovery

The JWT is exchanged at `investors.com` using the DylanToken constant, then used for GraphQL requests at `dowjones.io`.

### Error hierarchy (`internal/errors/errors.go`)

All errors embed `MarketSurgeError` base type. Use the constructor functions, not raw structs.

| Type | Exit Code | When |
|---|---|---|
| `AuthenticationError` | 3 | 401/403, missing token |
| `TokenExpiredError` | 3 | 401 specifically |
| `CookieExtractionError` | 3 | Cookie DB read failures |
| `APIError` | 4 | GraphQL-level errors |
| `SymbolNotFoundError` | 2 | Ticker not recognized |
| `HTTPError` | 4 | 429, 5xx |
| `ValidationError` | 1 | Bad args, missing fields |

Import alias convention: `mserrors` in commands, `mserr` in the output package.

### Output contract (`internal/output/`)

Every command must produce JSON envelopes. Never write raw output to stdout.

```go
// Success
output.WriteSuccess(w, data, output.SymbolMeta(symbol))

// Error
output.WriteError(w, err)

// Partial (some symbols succeeded, some failed)
output.WritePartial(w, results, errors, metadata)
```

Envelope shape: `{ data, metadata }` for success, `{ data, errors, metadata }` for partial success, and `{ error: { code, message, details } }` for errors.

## Conventions

### Code style

- **`init()` for cobra wiring**: Each command file uses `func init() { rootCmd.AddCommand(newXxxCmd()) }` to register commands
- **Command constructors**: `newXxxCmd() *cobra.Command` (unexported, called by init and tests)
- **Error wrapping**: Always use `fmt.Errorf("context: %w", err)` with `%w`
- **Typed errors**: Use `errors.As()` to match, constructor functions to create
- **Concurrency**: `sync.WaitGroup` + `sync.Mutex` for parallel ops (see `stock_analyze.go`)
- **GraphQL queries**: Embedded via `queries/embed.go`, loaded with `queries.Load("name")`

### Critical constraints

- JWT and Cookie HTTP headers must be added per-request in `client.Execute()`, not in base/default headers
- Chart history has mutually exclusive date params: explicit start/end dates XOR lookback period
- `kind` is required for catalog commands; each kind requires its own ID flag (report-id, watchlist-id, coach-screen-id)

### Adding a new command

1. Create `cmd/<group>.go` with constructor function and `init()` wiring
2. Add client method in `internal/client/<group>.go`
3. Add GraphQL query in `queries/<operation>.graphql`
4. Add model structs in `internal/models/` if needed
5. Add tests in `cmd/<group>_test.go`
6. Update skill files in `skills/marketsurge-agent/`

Follow `fundamental.go` (21 lines) as the canonical simple command template.

## Testing

- Framework: Go stdlib `testing` + `testify/assert` + `testify/require`
- CI runs: `go test -v -race -coverprofile=coverage.out ./...`
- Mock pattern: `httptest.NewServer` with request capture (no external mock libraries)
- Shared helpers in `cmd/helpers_test.go`: `testClient()`, `jsonServer()`, fixture builders
- Table-driven subtests with `t.Run()`, typed error checks with `assert.ErrorAs()`
- CLI tests call constructors directly, inject client via context, capture output to `bytes.Buffer`

## Build

Requires Go 1.26+.

```bash
make build     # Build binary
make test      # go test -v -race -coverprofile
make lint      # golangci-lint run
make clean     # Remove binary + coverage
```

Linting uses [golangci-lint](https://golangci-lint.run/) v2 with config in `.golangci.yml`. The standard linter set is enabled plus `bodyclose`, `errorlint`, `gocritic`, `misspell`, `modernize`, `nolintlint`, `revive`, `unconvert`, and `unparam`.

CI pipeline: `golangci-lint` (separate job) + `go test -v -race` -> `go build`

Release: push `v*` tag -> goreleaser v2 -> multi-platform binaries (linux/darwin, amd64/arm64, CGO disabled) -> GitHub Releases

## Maintenance

- **Keep this file updated**: When adding, removing, or changing commands, error types, conventions, or architecture, update this file and subdirectory AGENTS.md files to match.
- **Keep README.md updated**: When changing commands, flags, output format, install instructions, or development workflow, update README.md as well.
- **Keep skill files updated**: When changing command inputs, outputs, flags, or usage patterns, update the corresponding skill file in `skills/marketsurge-agent/` to match.

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/browserutils/kooky` - Firefox cookie extraction
- `github.com/stretchr/testify` - Test assertions
